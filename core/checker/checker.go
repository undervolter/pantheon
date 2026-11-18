// Package checker polls OLX for every active filter, decides which listings
// are news, and delivers the alerts.
//
// The two halves are decoupled by the database. A poll cycle evaluates a page
// of results and, in one transaction, records which listings it has seen and
// queues an alert for each match. A separate sender drains that queue. So a
// listing is never marked as seen without its alert being durably queued, and
// a Telegram outage delays alerts instead of losing them.
package checker

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"math/rand/v2"
	"sync"
	"time"

	"pantheon/core/config"
	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/textutil"
)

const (
	// A filter OLX keeps rejecting with HTTP 400 is stopped after this many
	// polls in a row: the request itself is wrong and will not get better.
	maxBadRequests = 3

	// Pause after OLX refuses requests by address (403 / 429). It doubles on
	// every refusal in a row and is jittered so that several installations
	// behind one address do not come back in step.
	blockPauseMin = time.Minute
	blockPauseMax = 15 * time.Minute

	// Fetches of one cycle are started this far apart at most, instead of all
	// at the same instant.
	maxFetchSpacing = 750 * time.Millisecond

	// The shortest time between the starts of two cycles, however often a
	// poll is asked for. It equals the smallest interval the config accepts.
	minCycleGap = 3 * time.Second

	pruneEvery     = time.Hour
	alertRetention = 30 * 24 * time.Hour
)

// Fetcher is the OLX search the checker needs.
type Fetcher interface {
	FetchPage(ctx context.Context, search domain.Search, limit, offset int) ([]domain.Offer, error)
}

// Messenger is the part of the Telegram client the checker needs.
type Messenger interface {
	SendOffer(ctx context.Context, chatID string, o domain.Offer, opt telegram.CardOptions) (int, error)
	SendHTML(ctx context.Context, chatID, text string, km *telegram.InlineKeyboardMarkup, silent bool) (int, error)
}

type Checker struct {
	cfg     *config.Config
	db      *store.Store
	fetch   Fetcher
	tg      Messenger
	limiter *telegram.Limiter

	startedAt time.Time
	now       func() time.Time
	trigger   chan struct{}
	wake      chan struct{}

	mu           sync.Mutex
	frontiers    map[string]time.Time
	started      bool
	blockedUntil time.Time
	blockStreak  int
	badRequests  map[string]int
	lastPrune    time.Time

	// Sender state.
	sendMu   sync.Mutex
	workers  map[string]bool
	workerWG sync.WaitGroup
	dead     map[string]time.Time
	// noPhotos: alerts whose last attempt timed out; they are retried without
	// photos. unrecorded: alerts that were delivered while the write recording
	// it failed. Both map an alert id to when the note was made.
	noPhotos   map[int64]time.Time
	unrecorded map[int64]time.Time
}

func New(cfg *config.Config, db *store.Store, fetch Fetcher, tg Messenger) *Checker {
	return &Checker{
		cfg:         cfg,
		db:          db,
		fetch:       fetch,
		tg:          tg,
		limiter:     telegram.NewLimiter(),
		startedAt:   time.Now(),
		now:         time.Now,
		trigger:     make(chan struct{}, 1),
		wake:        make(chan struct{}, 1),
		frontiers:   make(map[string]time.Time),
		badRequests: make(map[string]int),
		workers:     make(map[string]bool),
		dead:        make(map[string]time.Time),
		noPhotos:    make(map[int64]time.Time),
		unrecorded:  make(map[int64]time.Time),
	}
}

// SetLimiter replaces the per-chat alert limiter.
func (c *Checker) SetLimiter(l *telegram.Limiter) { c.limiter = l }

// StartedAt is when this process began polling.
func (c *Checker) StartedAt() time.Time { return c.startedAt }

// Started reports whether the first poll cycle has finished.
func (c *Checker) Started() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started
}

// BlockedUntil is the end of the current pause after OLX refused requests;
// zero when polling is not paused.
func (c *Checker) BlockedUntil() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.blockedUntil
}

// TriggerCheck asks for a poll cycle now instead of at the next tick.
func (c *Checker) TriggerCheck() {
	select {
	case c.trigger <- struct{}{}:
	default:
	}
}

// ActiveSearches returns the filters that are polled right now: enabled, not
// snoozed, in a section the bot serves, and owned by a chat that still exists.
func (c *Checker) ActiveSearches(ctx context.Context) ([]domain.Search, error) {
	all, err := c.db.Searches(ctx)
	if err != nil {
		return nil, err
	}
	now := c.now()
	out := all[:0:0]
	for _, s := range all {
		if !s.Active(now) {
			continue
		}
		// A filter whose owner was removed would poll OLX for nobody.
		if !s.IsShared() && !c.cfg.IsTrusted(s.OwnerChatID) {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// Run polls until ctx is cancelled, then waits for the alert queue to drain.
func (c *Checker) Run(ctx context.Context) {
	sendCtx, stopSender := context.WithCancel(context.WithoutCancel(ctx))
	defer stopSender()
	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		c.senderLoop(sendCtx)
	}()

	// The first pass after a start is silent unless asked otherwise: it only
	// learns what is already on the results pages.
	silent := !c.cfg.GetNotifyOnStartup()
	for ctx.Err() == nil {
		began := c.now()
		c.cycle(ctx, silent)
		silent = false
		c.mu.Lock()
		c.started = true
		c.mu.Unlock()
		c.maybePrune(ctx)

		wait := time.Duration(c.cfg.GetInterval())*time.Second - c.now().Sub(began)
		if wait < time.Second {
			wait = time.Second
		}
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		case <-c.trigger:
			// A nudge shortens the wait, but cycles never run back to back:
			// each one is a burst of requests to OLX, and a nudge comes from
			// a button anyone can keep pressing.
			if rest := minCycleGap - c.now().Sub(began); rest > 0 {
				select {
				case <-ctx.Done():
				case <-time.After(rest):
				}
			}
		}
	}

	c.drain(shutdownDrain)
	stopSender()
	<-senderDone
}

// RunOnce does a single, non-silent poll cycle and delivers what it found.
func (c *Checker) RunOnce(ctx context.Context) {
	sendCtx, stopSender := context.WithCancel(context.WithoutCancel(ctx))
	defer stopSender()
	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		c.senderLoop(sendCtx)
	}()

	c.cycle(ctx, false)
	c.drain(onceDrain)
	stopSender()
	<-senderDone
}

func (c *Checker) maybePrune(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	now := c.now()
	c.mu.Lock()
	due := now.Sub(c.lastPrune) >= pruneEvery
	if due {
		c.lastPrune = now
	}
	c.mu.Unlock()
	if !due {
		return
	}
	var seenBefore time.Time
	if keep := c.cfg.SeenRetention(); keep > 0 {
		seenBefore = now.Add(-keep)
	}
	seen, alerts, err := c.db.Prune(ctx, seenBefore, now.Add(-alertRetention))
	if err != nil {
		if ctx.Err() == nil {
			logx.Warnf("prune: %v", err)
		}
		return
	}
	if seen+alerts > 0 {
		logx.Infof("🧹 pruned %d seen marks and %d old alerts", seen, alerts)
	}
}

type fetchResult struct {
	offers []domain.Offer
	err    error
}

// cycle polls every active filter once.
func (c *Checker) cycle(ctx context.Context, silent bool) {
	if until := c.BlockedUntil(); c.now().Before(until) {
		return
	}
	searches, err := c.ActiveSearches(ctx)
	if err != nil {
		if ctx.Err() == nil {
			logx.Errorf("read filters: %v", err)
		}
		return
	}
	if len(searches) == 0 {
		return
	}

	concurrency, limit := c.cfg.PerfSettings()
	if concurrency < 1 {
		concurrency = 1
	}
	rules := c.rules()

	// Fetch in parallel, a little apart from each other; evaluate in order.
	spacing := time.Duration(c.cfg.GetInterval()) * time.Second / time.Duration(len(searches)+1)
	if spacing > maxFetchSpacing {
		spacing = maxFetchSpacing
	}
	results := make([]fetchResult, len(searches))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, s := range searches {
		if i > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(spacing):
			}
		}
		if ctx.Err() != nil || c.now().Before(c.BlockedUntil()) {
			results[i].err = context.Canceled
			continue
		}
		wg.Add(1)
		go func(i int, s domain.Search) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			// While this fetch waited for its turn, OLX may have refused
			// another one. Sending more requests would only prolong the ban.
			if ctx.Err() != nil || c.now().Before(c.BlockedUntil()) {
				results[i].err = context.Canceled
				return
			}
			offers, err := c.fetchWithBackfill(ctx, s, limit, rules.Floor(s))
			results[i] = fetchResult{offers: offers, err: err}
			if olx.IsBlocked(err) {
				c.block(ctx, err)
			}
		}(i, s)
	}
	wg.Wait()

	blocked := false
	for i, s := range searches {
		res := results[i]
		switch {
		case res.err == nil:
			c.noteRequestOK(s.ID)
			c.process(ctx, s, res.offers, rules, silent)
		case errors.Is(res.err, context.Canceled) || ctx.Err() != nil:
		case olx.IsBlocked(res.err):
			blocked = true
		case olx.IsBadRequest(res.err):
			c.noteBadRequest(ctx, s, res.err)
		default:
			logx.Warnf("'%s': %v", s.Name, res.err)
		}
	}
	if !blocked {
		c.mu.Lock()
		c.blockStreak = 0
		c.mu.Unlock()
	}
}

func (c *Checker) rules() Rules {
	return Rules{
		Now:           c.now(),
		StartedAt:     c.startedAt,
		MaxAge:        time.Duration(c.cfg.GetMaxAlertAgeMin()) * time.Minute,
		SkipBumpedOld: c.cfg.GetSkipBumpedOld(),
	}
}

// process evaluates one page of results for one filter and records the
// outcome: seen marks and queued alerts, in a single transaction.
func (c *Checker) process(ctx context.Context, s domain.Search, offers []domain.Offer, rules Rules, silent bool) {
	// Listings of a section the bot does not serve are dropped before anything
	// else looks at them: not logged, not stored.
	visible := offers[:0:0]
	ids := make([]int64, 0, len(offers))
	inPage := make(map[int64]bool, len(offers))
	for i := range offers {
		o := &offers[i]
		if o.ID == 0 || inPage[o.ID] {
			continue
		}
		if v, _ := Evaluate(o, s, rules); v == SkipRealEstate {
			continue
		}
		inPage[o.ID] = true
		visible = append(visible, *o)
		ids = append(ids, o.ID)
	}
	if len(visible) == 0 {
		return
	}

	seen, err := c.db.SeenSet(ctx, store.SeenScopes(s.ID, s.OwnerChatID), ids)
	if err != nil {
		if ctx.Err() == nil {
			logx.Errorf("'%s': read seen marks: %v", s.Name, err)
		}
		return
	}

	// Stamped now, not with the start of the cycle: an alert's time limit
	// counts from when it was queued, and the fetches before this point may
	// have taken a while.
	commit := store.Commit{SearchID: s.ID, Now: c.now()}
	recipients := c.recipients(s)
	ansiLinks := c.cfg.GetANSIHyperlinks()
	fresh, queryMisses := 0, 0

	for i := range visible {
		o := &visible[i]
		verdict, detail := Evaluate(o, s, rules)
		if verdict == SkipPromoted {
			// Not marked as seen: once the promotion ends the listing shows up
			// in the ordinary results and is judged like any other. Not logged
			// either — a third of every page is promoted, and without a seen
			// mark the same lines would repeat on every poll.
			continue
		}
		commit.SeenIDs = append(commit.SeenIDs, o.ID)
		if seen[o.ID] {
			continue
		}
		fresh++
		if verdict != Match {
			if verdict == SkipQuery {
				queryMisses++
			}
			if !silent {
				logx.Infof("%s", formatAdSkipped(o, verdict, detail))
			}
			continue
		}
		if silent {
			continue
		}
		logx.Infof("%s", formatAdFound(o, s.Name, ansiLinks))

		payload, err := json.Marshal(o)
		if err != nil {
			logx.Errorf("encode listing %d: %v", o.ID, err)
			continue
		}
		for _, chatID := range recipients {
			commit.Alerts = append(commit.Alerts, store.Reservation{
				OfferID:    o.ID,
				ChatID:     chatID,
				SearchName: s.Name,
				Title:      textutil.TruncateRunes(textutil.SanitizeListingText(o.Title), 120),
				URL:        o.URL,
				Price:      textutil.SanitizeListingText(o.GetPrice()),
				Payload:    string(payload),
			})
		}
	}

	queued, ok, err := c.db.CommitPage(ctx, commit)
	if err != nil {
		if ctx.Err() == nil {
			logx.Errorf("'%s': store the poll result: %v", s.Name, err)
		}
		return
	}
	if !ok {
		// Deleted or paused while its page was on the way: nothing was written.
		return
	}
	if silent {
		logx.Infof("primed '%s': %d listings on the page, %d new to the cache", s.Name, len(visible), fresh)
		return
	}
	if queryMisses > 0 {
		logx.Infof("🎯 '%s': %d listings dropped — the query did not match", s.Name, queryMisses)
	}
	if len(queued) > 0 {
		c.wakeSender()
	}
}

// recipients returns the chats an alert of this filter goes to: its owner, or
// every trusted chat for a shared filter. A paused chat gets nothing — the
// listing is still marked as seen, which is what pausing means — and a chat
// that turned out to be unreachable is left alone for a while.
func (c *Checker) recipients(s domain.Search) []string {
	var chats []string
	if s.IsShared() {
		chats = c.cfg.GetTelegramChatIDs()
	} else if c.cfg.IsTrusted(s.OwnerChatID) {
		chats = []string{s.OwnerChatID}
	}
	out := chats[:0:0]
	for _, id := range chats {
		if c.cfg.IsPaused(id) || c.isDead(id) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// ---------- reacting to OLX refusing requests ----------

// block pauses all polling after a 403 or 429. Such a refusal is tied to the
// address, not to one filter, so the other filters would only make it worse.
func (c *Checker) block(ctx context.Context, cause error) {
	c.mu.Lock()
	now := c.now()
	if now.Before(c.blockedUntil) {
		c.mu.Unlock()
		return
	}
	c.blockStreak++
	pause := blockPauseMin << min(c.blockStreak-1, 8)
	if pause > blockPauseMax || pause <= 0 {
		pause = blockPauseMax
	}
	// ±20 % of jitter.
	pause = time.Duration(float64(pause) * (0.8 + 0.4*rand.Float64()))
	c.blockedUntil = now.Add(pause)
	first := c.blockStreak == 1
	c.mu.Unlock()

	status := 0
	var apiErr *olx.APIError
	if errors.As(cause, &apiErr) {
		status = apiErr.Status
	}
	logx.Errorf("OLX is refusing requests (HTTP %d); polling is paused for %s", status, pause.Round(time.Second))
	if !first {
		return
	}
	// One notice per episode, to the people who can do something about it.
	for _, admin := range c.cfg.Admins() {
		lang := i18n.Of(c.cfg.LangFor(admin))
		text := i18n.T(lang, "ops.blocked", status, textutil.FormatModerationDelay(pause))
		if _, err := c.tg.SendHTML(ctx, admin, text, nil, true); err != nil && ctx.Err() == nil {
			logx.Warnf("notify admin %s: %v", admin, err)
		}
	}
}

func (c *Checker) noteRequestOK(id string) {
	c.mu.Lock()
	delete(c.badRequests, id)
	c.mu.Unlock()
}

// noteBadRequest counts HTTP 400 answers for a filter and stops the filter
// once it is clear the request itself is the problem.
func (c *Checker) noteBadRequest(ctx context.Context, s domain.Search, cause error) {
	c.mu.Lock()
	c.badRequests[s.ID]++
	n := c.badRequests[s.ID]
	c.mu.Unlock()
	logx.Warnf("'%s': OLX rejected the request (%d of %d): %v", s.Name, n, maxBadRequests, cause)
	if n < maxBadRequests {
		return
	}

	if _, err := c.db.UpdateSearch(ctx, s.ID, func(sc *domain.Search) error { sc.Disabled = true; return nil }); err != nil {
		logx.Errorf("'%s': stop the filter: %v", s.Name, err)
		return
	}
	c.noteRequestOK(s.ID)
	logx.Errorf("'%s' was stopped: OLX keeps rejecting its request", s.Name)

	notify := c.cfg.Admins()
	if !s.IsShared() {
		notify = []string{s.OwnerChatID}
	}
	for _, chatID := range notify {
		lang := i18n.Of(c.cfg.LangFor(chatID))
		text := i18n.T(lang, "ops.filter_stopped", html.EscapeString(textutil.TruncateRunes(s.Name, 80)))
		if _, err := c.tg.SendHTML(ctx, chatID, text, nil, false); err != nil && ctx.Err() == nil {
			logx.Warnf("notify %s: %v", chatID, err)
		}
	}
}

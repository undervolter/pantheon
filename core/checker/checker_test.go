package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"pantheon/core/config"
	"pantheon/core/domain"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/telegram/tgtest"
)

var ctx = context.Background()

// ---------- harness ----------

type fakeFetcher struct {
	mu    sync.Mutex
	pages map[string][][]domain.Offer // filter id -> pages by index
	errs  map[string][]error          // filter id -> errors for successive calls
	calls []string                    // "id@offset"
	seen  []domain.Search
}

func newFetcher() *fakeFetcher {
	return &fakeFetcher{pages: map[string][][]domain.Offer{}, errs: map[string][]error{}}
}

func (f *fakeFetcher) set(id string, pages ...[]domain.Offer) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pages[id] = pages
}

func (f *fakeFetcher) fail(id string, errs ...error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errs[id] = append(f.errs[id], errs...)
}

func (f *fakeFetcher) FetchPage(_ context.Context, s domain.Search, limit, offset int) ([]domain.Offer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fmt.Sprintf("%s@%d", s.ID, offset))
	f.seen = append(f.seen, s)
	if queue := f.errs[s.ID]; len(queue) > 0 {
		err := queue[0]
		f.errs[s.ID] = queue[1:]
		if err != nil {
			return nil, err
		}
	}
	page := 0
	if limit > 0 {
		page = offset / limit
	}
	if pages := f.pages[s.ID]; page < len(pages) {
		return append([]domain.Offer(nil), pages[page]...), nil
	}
	return nil, nil
}

func (f *fakeFetcher) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

type harness struct {
	t     *testing.T
	cfg   *config.Config
	db    *store.Store
	fetch *fakeFetcher
	tg    *tgtest.Server
	chk   *Checker
	now   time.Time
	log   *bytes.Buffer

	cancelSender context.CancelFunc
	senderDone   chan struct{}
}

// stopSender stops the background sender and waits until it is gone, for
// tests that drive delivery themselves.
func (h *harness) stopSender() {
	h.cancelSender()
	<-h.senderDone
}

// newHarness builds a checker around a temp database, a fake OLX and a fake
// Bot API. Chats 111 (admin) and 222 are trusted.
func newHarness(t *testing.T, configJSON string) *harness {
	t.Helper()
	dir := t.TempDir()
	if configJSON == "" {
		configJSON = `{"telegram_bot_token":"` + tgtest.Token + `","telegram_chat_ids":["111","222"],"admin_chat_ids":["111"],"max_alert_age_min":30,"fetch_limit":20}`
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(configJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(dir, "pantheon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	h := &harness{t: t, cfg: cfg, db: db, fetch: newFetcher(), tg: tgtest.New(t), log: &bytes.Buffer{}}
	t.Cleanup(logx.SetOutput(h.log))
	h.now = time.Now()
	h.chk = New(cfg, db, h.fetch, telegram.New(tgtest.Token).WithBaseURL(h.tg.URL))
	h.chk.SetLimiter(telegram.NewLimiterWith(0, 0))
	// The process "started" an hour ago, so the start itself is not what
	// decides whether a listing is news.
	h.chk.startedAt = h.now.Add(-time.Hour)

	sendCtx, stop := context.WithCancel(context.Background())
	h.cancelSender, h.senderDone = stop, make(chan struct{})
	go func() { defer close(h.senderDone); h.chk.senderLoop(sendCtx) }()
	t.Cleanup(h.stopSender)
	return h
}

func (h *harness) filter(s domain.Search) domain.Search {
	h.t.Helper()
	// Created well before "now": the filter's age floor is not in the way either.
	if s.CreatedAt == 0 {
		s.CreatedAt = h.now.Add(-2 * time.Hour).Unix()
	}
	out, err := h.db.AddSearch(ctx, s)
	if err != nil {
		h.t.Fatalf("AddSearch: %v", err)
	}
	return out
}

// offer builds a listing created the given number of minutes ago.
func (h *harness) offer(id int64, title string, minutesAgo int) domain.Offer {
	at := h.now.Add(-time.Duration(minutesAgo) * time.Minute)
	return domain.Offer{
		ID:              id,
		Title:           title,
		URL:             fmt.Sprintf("https://www.olx.ua/d/uk/obyavlenie/x-ID%d.html", id),
		CreatedTime:     at.Format(time.RFC3339),
		LastRefreshTime: at.Format(time.RFC3339),
		Category:        domain.OfferCategory{ID: 44, Type: "electronics"},
		Params:          []domain.Param{priceParam(`{"value":1000,"type":"price","currency":"UAH"}`)},
	}
}

func priceParam(value string) domain.Param {
	return domain.Param{Key: "price", Name: "Ціна", Type: "price", Value: json.RawMessage(value)}
}

// poll runs one cycle and waits for the queue to empty.
func (h *harness) poll(silent bool) {
	h.t.Helper()
	h.chk.cycle(ctx, silent)
	h.chk.drain(10 * time.Second)
}

// delivered returns the chats that received a card, in order.
func (h *harness) delivered() []string {
	var out []string
	for _, c := range h.tg.Calls("sendRichMessage") {
		out = append(out, c.ChatID())
	}
	return out
}

func (h *harness) sentCount(chatID string) int {
	n, _, err := h.db.AlertCounts(ctx, chatID, h.now.Add(-time.Hour))
	if err != nil {
		h.t.Fatal(err)
	}
	return n
}

// ---------- pipeline ----------

func TestEvaluateVerdicts(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	rules := Rules{Now: now, StartedAt: now.Add(-time.Hour), MaxAge: 30 * time.Minute, SkipBumpedOld: true}
	created := func(minutesAgo int) string {
		return now.Add(-time.Duration(minutesAgo) * time.Minute).Format(time.RFC3339)
	}
	base := func() domain.Offer {
		return domain.Offer{ID: 1, Title: "iPhone 13 128GB", CreatedTime: created(12), LastRefreshTime: created(10),
			Category: domain.OfferCategory{ID: 44, Type: "electronics"},
			Params:   []domain.Param{priceParam(`{"value":9000,"type":"price","currency":"UAH"}`)}}
	}
	filter := domain.Search{ID: "f", Query: "iphone 13", QueryVariants: []string{"iphone 13"}, CreatedAt: now.Add(-48 * time.Hour).Unix()}

	cases := []struct {
		name   string
		mutate func(o *domain.Offer, s *domain.Search)
		want   Verdict
	}{
		{"match", func(*domain.Offer, *domain.Search) {}, Match},
		{"real estate by category type", func(o *domain.Offer, _ *domain.Search) {
			o.Category = domain.OfferCategory{ID: 999999, Type: "real_estate"}
		}, SkipRealEstate},
		{"real estate by tree position", func(o *domain.Offer, _ *domain.Search) { o.Category = domain.OfferCategory{ID: 1758} }, SkipRealEstate},
		{"daily rentals are not real estate", func(o *domain.Offer, _ *domain.Search) {
			o.Category = domain.OfferCategory{ID: 3711, Type: "accommodation"}
		}, Match},
		{"promoted", func(o *domain.Offer, _ *domain.Search) { o.Promotion.TopAd = true }, SkipPromoted},
		{"older than the limit", func(o *domain.Offer, _ *domain.Search) { o.CreatedTime = created(45) }, SkipTooOld},
		{"before the floor", func(o *domain.Offer, _ *domain.Search) { o.CreatedTime = created(600) }, SkipOutdated},
		{"an index lag of 18 minutes is still news", func(o *domain.Offer, _ *domain.Search) { o.CreatedTime = created(18) }, Match},
		{"spam title", func(o *domain.Offer, s *domain.Search) { o.Title = "Тест"; s.Query, s.QueryVariants = "", nil }, SkipSpamTitle},
		{"query mismatch", func(o *domain.Offer, _ *domain.Search) { o.Title = "Samsung S21" }, SkipQuery},
		{"query in the description", func(o *domain.Offer, _ *domain.Search) {
			o.Title, o.Description = "Продам телефон", "iPhone 13, стан ідеальний"
		}, Match},
		{"excluded by word start", func(o *domain.Offer, s *domain.Search) {
			o.Title = "iPhone 13 на запчастини"
			s.ExcludeWords = []string{"запчаст"}
		}, SkipExcluded},
		{"exclusion word absent", func(_ *domain.Offer, s *domain.Search) { s.ExcludeWords = []string{"чохол"} }, Match},
		{"free-only filter, paid listing", func(_ *domain.Offer, s *domain.Search) { s.PriceMode = domain.PriceFree }, SkipPrice},
		{"free-only filter, free listing", func(o *domain.Offer, s *domain.Search) {
			s.PriceMode = domain.PriceFree
			o.Params = []domain.Param{priceParam(`{"value":0,"type":"free","budget":false,"label":"Безкоштовно"}`)}
		}, Match},
		{"free-only filter, exchange listing", func(o *domain.Offer, s *domain.Search) {
			s.PriceMode = domain.PriceFree
			o.Params = []domain.Param{priceParam(`{"value":0,"type":"exchange","label":"Обмін"}`)}
		}, SkipPrice},
		{"free-only filter, listing without a price", func(o *domain.Offer, s *domain.Search) {
			s.PriceMode = domain.PriceFree
			o.Params = nil
		}, SkipPrice},
	}
	for _, c := range cases {
		o, s := base(), filter.Clone()
		c.mutate(&o, &s)
		if got, detail := Evaluate(&o, s, rules); got != c.want {
			t.Errorf("%s: verdict = %s (%s), want %s", c.name, got.Reason(), detail, c.want.Reason())
		}
	}
}

// A filter resumed after a long pause must not pour out everything that was
// posted meanwhile, even with the age limit switched off.
func TestFloorMovesWithResumeAndSnooze(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	rules := Rules{Now: now, StartedAt: now.Add(-72 * time.Hour), MaxAge: 0, SkipBumpedOld: true}
	old := domain.Offer{ID: 1, Title: "x", CreatedTime: now.Add(-3 * time.Hour).Format(time.RFC3339)}
	recent := domain.Offer{ID: 2, Title: "x", CreatedTime: now.Add(-10 * time.Minute).Format(time.RFC3339)}

	longRunning := domain.Search{ID: "f", Category: 44, CreatedAt: now.Add(-48 * time.Hour).Unix()}
	if v, _ := Evaluate(&old, longRunning, rules); v != Match {
		t.Fatalf("with no age limit a 3h-old listing matches a long-running filter, got %s", v.Reason())
	}

	resumed := longRunning
	resumed.EnabledAt = now.Add(-time.Minute).Unix()
	if v, _ := Evaluate(&old, resumed, rules); v != SkipOutdated {
		t.Errorf("a just-resumed filter must skip what was posted during the pause, got %s", v.Reason())
	}
	if v, _ := Evaluate(&recent, resumed, rules); v != Match {
		t.Errorf("a just-resumed filter still alerts fresh listings, got %s", v.Reason())
	}

	woken := longRunning
	woken.SnoozedUntil = now.Add(-2 * time.Minute).Unix()
	if v, _ := Evaluate(&old, woken, rules); v != SkipOutdated {
		t.Errorf("a filter that just woke from snooze must skip the backlog, got %s", v.Reason())
	}

	// A snooze end in the future must not push the floor past now.
	future := longRunning
	future.SnoozedUntil = now.Add(time.Hour).Unix()
	if floor := rules.Floor(future); floor.After(now) {
		t.Errorf("floor = %s is in the future", floor)
	}
	if got, want := rules.Floor(longRunning), now.Add(-48*time.Hour).Add(-minGrace); !got.Equal(want) {
		t.Errorf("floor = %s, want %s", got, want)
	}
}

func TestGraceIsNeverBelowTheIndexLag(t *testing.T) {
	for limit, want := range map[time.Duration]time.Duration{0: 20 * time.Minute, 5 * time.Minute: 20 * time.Minute, 20 * time.Minute: 20 * time.Minute, time.Hour: time.Hour} {
		if got := (Rules{MaxAge: limit}).Grace(); got != want {
			t.Errorf("Grace with limit %s = %s, want %s", limit, got, want)
		}
	}
}

func TestExcludedBy(t *testing.T) {
	text := "Продам iPhone 13, запчастини та чохол у подарунок"
	cases := []struct {
		words []string
		want  string
	}{
		{nil, ""},
		{[]string{"запчаст"}, "запчаст"},
		{[]string{"ЗАПЧАСТ"}, "ЗАПЧАСТ"},
		{[]string{"подарунок чохол"}, "подарунок чохол"},
		{[]string{"подарунок коробка"}, ""},
		{[]string{"частини"}, ""}, // a word start, not a substring
		{[]string{"", "   "}, ""},
	}
	for _, c := range cases {
		if got := ExcludedBy(text, c.words); got != c.want {
			t.Errorf("ExcludedBy(%v) = %q, want %q", c.words, got, c.want)
		}
	}
}

// ---------- frontier ----------

func mkOffer(id int64, refresh string, top bool) domain.Offer {
	return domain.Offer{ID: id, LastRefreshTime: refresh, Promotion: domain.Promotion{TopAd: top}}
}

func TestPageFrontierHelpers(t *testing.T) {
	page := []domain.Offer{
		mkOffer(1, "2026-11-07T10:00:00+02:00", true),
		mkOffer(2, "2026-11-07T12:10:00+02:00", false),
		mkOffer(3, "2026-11-07T12:05:00+02:00", false),
		mkOffer(4, "2026-11-07T12:01:00+02:00", false),
	}
	wantOldest, _ := time.Parse(time.RFC3339, "2026-11-07T12:01:00+02:00")
	wantNewest, _ := time.Parse(time.RFC3339, "2026-11-07T12:10:00+02:00")
	if got := pageOldestRefresh(page); !got.Equal(wantOldest) {
		t.Errorf("oldest = %s, want %s", got, wantOldest)
	}
	if got := pageNewestRefresh(page); !got.Equal(wantNewest) {
		t.Errorf("newest = %s, want %s", got, wantNewest)
	}
	before, _ := time.Parse(time.RFC3339, "2026-11-07T12:00:00+02:00")
	after, _ := time.Parse(time.RFC3339, "2026-11-07T12:30:00+02:00")
	if pageReached(page, before) {
		t.Error("page must not reach 12:00: all organic items are newer")
	}
	if !pageReached(page, after) {
		t.Error("page must reach 12:30")
	}
	if !pageReached([]domain.Offer{mkOffer(1, "2026-11-07T10:00:00+02:00", true)}, time.Now()) {
		t.Error("a page of promoted items only counts as reached")
	}
}

func TestBackfillFetchesUntilTheFrontier(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "x", OwnerChatID: "111"})
	h.chk.setFrontier(s.ID, h.now.Add(-30*time.Minute))

	h.fetch.set(s.ID,
		[]domain.Offer{h.offer(1, "x", 1), h.offer(2, "x", 5)},   // all newer than the frontier
		[]domain.Offer{h.offer(3, "x", 10), h.offer(4, "x", 40)}, // reaches it
		[]domain.Offer{h.offer(5, "x", 50)},
	)
	offers, err := h.chk.fetchWithBackfill(ctx, s, 20, time.Time{})
	if err != nil || len(offers) != 4 {
		t.Fatalf("offers = %d, %v; want pages 1 and 2", len(offers), err)
	}
	if h.fetch.callCount() != 2 {
		t.Errorf("fetches = %d, want 2", h.fetch.callCount())
	}
}

// When a backfill request fails, the frontier must stay where it was, so the
// next poll tries to close the same gap. It used to move on and the gap was
// only mentioned in the log.
func TestBackfillErrorKeepsTheFrontier(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "x", OwnerChatID: "111"})
	prev := h.now.Add(-30 * time.Minute)
	h.chk.setFrontier(s.ID, prev)

	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "x", 1), h.offer(2, "x", 5)})
	h.fetch.fail(s.ID, nil, errors.New("connection reset"))

	offers, err := h.chk.fetchWithBackfill(ctx, s, 20, time.Time{})
	if err != nil || len(offers) != 2 {
		t.Fatalf("the first page must still be returned: %d, %v", len(offers), err)
	}
	if got := h.chk.frontier(s.ID); !got.Equal(prev) {
		t.Errorf("frontier moved to %s after a failed backfill, want it kept at %s", got, prev)
	}
}

func TestFrontierOlderThanTheFloorIsIgnored(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "x", OwnerChatID: "111"})
	h.chk.setFrontier(s.ID, h.now.Add(-24*time.Hour))
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "x", 1)}, []domain.Offer{h.offer(2, "x", 90)})

	// The filter was paused for a day; its floor is a few minutes back.
	if _, err := h.chk.fetchWithBackfill(ctx, s, 20, h.now.Add(-20*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if h.fetch.callCount() != 1 {
		t.Errorf("fetches = %d; a day-old frontier must not trigger a backfill", h.fetch.callCount())
	}
}

// ---------- cycle ----------

func TestSilentFirstPassOnlyPrimes(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "Телефоны", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 5)})

	h.poll(true)
	if got := h.delivered(); len(got) != 0 {
		t.Fatalf("the silent pass sent alerts to %v", got)
	}

	// The same listing is not news on the next pass; a new one is.
	h.fetch.set(s.ID, []domain.Offer{h.offer(2, "iPhone 14", 3), h.offer(1, "iPhone 13", 5)})
	h.poll(false)
	if got := h.delivered(); len(got) != 1 || got[0] != "111" {
		t.Fatalf("delivered = %v, want one alert to the owner", got)
	}
	card, _ := h.tg.Last("sendRichMessage")
	if !strings.Contains(card.Text(), "iPhone 14") || !strings.Contains(card.Text(), "Телефоны") {
		t.Errorf("unexpected card:\n%s", card.Text())
	}
	if h.sentCount("111") != 1 {
		t.Errorf("the delivery must be recorded: %d", h.sentCount("111"))
	}

	// And nothing repeats on a third pass.
	h.poll(false)
	if got := h.delivered(); len(got) != 1 {
		t.Fatalf("a listing was alerted twice: %v", got)
	}
}

// The alert-loss bug of 0.1.x: seen was keyed by owner, so a listing filter A
// rejected by query was never shown to the same owner's filter B.
func TestListingRejectedByOneFilterStillReachesAnother(t *testing.T) {
	h := newHarness(t, "")
	a := h.filter(domain.Search{Name: "iphones", Query: "iphone", OwnerChatID: "111"})
	b := h.filter(domain.Search{Name: "all phones", Category: 44, OwnerChatID: "111"})
	samsung := h.offer(7, "Samsung S21", 4)
	h.fetch.set(a.ID, []domain.Offer{samsung})
	h.fetch.set(b.ID, []domain.Offer{samsung})

	h.poll(false)
	got := h.delivered()
	if len(got) != 1 || got[0] != "111" {
		t.Fatalf("delivered = %v, want exactly one alert through the second filter", got)
	}
	card, _ := h.tg.Last("sendRichMessage")
	if !strings.Contains(card.Text(), "all phones") {
		t.Errorf("the alert must come from the filter that matched:\n%s", card.Text())
	}
}

func TestOneAlertPerChatWhenTwoFiltersMatch(t *testing.T) {
	h := newHarness(t, "")
	a := h.filter(domain.Search{Name: "a", Query: "iphone", OwnerChatID: "111"})
	b := h.filter(domain.Search{Name: "b", Category: 44, OwnerChatID: "111"})
	o := h.offer(9, "iPhone 13", 4)
	h.fetch.set(a.ID, []domain.Offer{o})
	h.fetch.set(b.ID, []domain.Offer{o})

	h.poll(false)
	if got := h.delivered(); len(got) != 1 {
		t.Fatalf("delivered = %v, want one alert", got)
	}
}

func TestSharedFilterAlertsEveryChatExceptPaused(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "shared", Query: "iphone"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})
	h.poll(false)
	got := h.delivered()
	if len(got) != 2 {
		t.Fatalf("delivered = %v, want both trusted chats", got)
	}

	if err := h.cfg.SetPaused("222", true); err != nil {
		t.Fatal(err)
	}
	h.tg.Reset()
	h.fetch.set(s.ID, []domain.Offer{h.offer(2, "iPhone 14", 3)})
	h.poll(false)
	if got := h.delivered(); len(got) != 1 || got[0] != "111" {
		t.Fatalf("delivered = %v, want only the chat that is not paused", got)
	}

	// Pausing means missing what was posted meanwhile, not getting it later.
	_ = h.cfg.SetPaused("222", false)
	h.tg.Reset()
	h.poll(false)
	if got := h.delivered(); len(got) != 0 {
		t.Fatalf("a resumed chat got the backlog: %v", got)
	}
}

func TestFiltersOfARemovedChatAreNotPolled(t *testing.T) {
	h := newHarness(t, "")
	gone := h.filter(domain.Search{Name: "orphan", Query: "x", OwnerChatID: "999"})
	paused := h.filter(domain.Search{Name: "paused", Query: "x", OwnerChatID: "111", Disabled: true})
	snoozed := h.filter(domain.Search{Name: "snoozed", Query: "x", OwnerChatID: "111", SnoozedUntil: h.now.Add(time.Hour).Unix()})
	live := h.filter(domain.Search{Name: "live", Query: "x", OwnerChatID: "111"})

	h.poll(false)
	polled := strings.Join(h.fetch.calls, " ")
	for _, s := range []domain.Search{gone, paused, snoozed} {
		if strings.Contains(polled, s.ID) {
			t.Errorf("filter %q was polled: %s", s.Name, polled)
		}
	}
	if !strings.Contains(polled, live.ID) {
		t.Errorf("the live filter was not polled: %s", polled)
	}
}

// A promoted listing is skipped but not remembered: when the promotion ends
// and it appears in the ordinary results, it is judged like any other.
func TestPromotedListingIsNotMarkedSeen(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	promoted := h.offer(1, "iPhone 13", 4)
	promoted.Promotion.TopAd = true
	h.fetch.set(s.ID, []domain.Offer{promoted})

	h.poll(false)
	if got := h.delivered(); len(got) != 0 {
		t.Fatalf("a promoted listing was alerted: %v", got)
	}
	if seen, _ := h.db.SeenSet(ctx, []string{s.ID}, []int64{1}); seen[1] {
		t.Fatal("a promoted listing must not be marked as seen")
	}
	// Without a seen mark it is evaluated on every poll, so it must not be
	// logged: a third of each page is promoted.
	h.poll(false)
	if strings.Contains(h.log.String(), "Top ad") {
		t.Fatalf("promoted listings must not be logged on every poll: %s", h.log.String())
	}

	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})
	h.poll(false)
	if got := h.delivered(); len(got) != 1 {
		t.Fatalf("the listing must be alerted once it shows up without promotion: %v", got)
	}
}

// Real estate is invisible to the bot: not alerted, not logged, not stored —
// including through a filter without a category.
func TestRealEstateIsInvisible(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "everything", Query: "квартира", OwnerChatID: "111"})
	flat := h.offer(1, "Продам квартира центр", 4)
	flat.Category = domain.OfferCategory{ID: 1758, Type: "real_estate"}
	flatByType := h.offer(2, "Здам квартира", 4)
	flatByType.Category = domain.OfferCategory{ID: 987654, Type: "real_estate"}
	lamp := h.offer(3, "Лампа у квартира", 4)
	h.fetch.set(s.ID, []domain.Offer{flat, flatByType, lamp})

	h.poll(false)
	if got := h.delivered(); len(got) != 1 {
		t.Fatalf("delivered = %v, want only the listing that is not real estate", got)
	}
	card, _ := h.tg.Last("sendRichMessage")
	if !strings.Contains(card.Text(), "Лампа") {
		t.Errorf("wrong listing alerted:\n%s", card.Text())
	}
	seen, _ := h.db.SeenSet(ctx, []string{s.ID}, []int64{1, 2, 3})
	if seen[1] || seen[2] || !seen[3] {
		t.Errorf("seen = %v; real-estate listings must leave no trace", seen)
	}
	if out := h.log.String(); strings.Contains(out, "ad 1,") || strings.Contains(out, "ad 2,") {
		t.Errorf("real-estate listings must not be logged:\n%s", out)
	}
}

func TestFreeOnlyFilterDropsPaidListings(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "даром", Category: 44, PriceMode: domain.PriceFree, OwnerChatID: "111"})
	free := h.offer(1, "Віддам телефон", 4)
	free.Params = []domain.Param{priceParam(`{"value":0,"type":"free","budget":false,"label":"Безкоштовно"}`)}
	paid := h.offer(2, "Продам телефон", 4)
	h.fetch.set(s.ID, []domain.Offer{free, paid})

	h.poll(false)
	if got := h.delivered(); len(got) != 1 {
		t.Fatalf("delivered = %v, want only the free listing", got)
	}
	card, _ := h.tg.Last("sendRichMessage")
	if !strings.Contains(card.Text(), "Отдают бесплатно") {
		t.Errorf("the free card wording is missing:\n%s", card.Text())
	}
	// The request itself asks OLX for free listings.
	asked := h.fetch.seen[0]
	if q := olx.BuildOffersQuery(asked, 20, 0); q.Get("filter_enum_price[0]") != "free" {
		t.Errorf("the request must carry the free checkbox: %s", q.Encode())
	}
}

func TestFilterDeletedDuringThePollGetsNothing(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	offers := []domain.Offer{h.offer(1, "iPhone 13", 4)}

	if _, err := h.db.DeleteSearch(ctx, s.ID); err != nil {
		t.Fatal(err)
	}
	// The page arrives after the filter is gone.
	h.chk.process(ctx, s, offers, h.chk.rules(), false)
	h.chk.drain(5 * time.Second)
	if got := h.delivered(); len(got) != 0 {
		t.Fatalf("a deleted filter produced alerts: %v", got)
	}
	if n, _ := h.db.PendingCount(ctx); n != 0 {
		t.Fatalf("pending = %d", n)
	}
}

// ---------- OLX refusing requests ----------

func TestBlockedByOLXPausesEveryFilterAndTellsTheAdminOnce(t *testing.T) {
	h := newHarness(t, "")
	a := h.filter(domain.Search{Name: "a", Query: "x", OwnerChatID: "111"})
	b := h.filter(domain.Search{Name: "b", Query: "y", OwnerChatID: "222"})
	h.fetch.fail(a.ID, &olx.APIError{Status: 429, Body: "Too Many Requests"})
	h.fetch.fail(b.ID, &olx.APIError{Status: 429, Body: "Too Many Requests"})

	h.poll(false)
	until := h.chk.BlockedUntil()
	if !until.After(time.Now()) {
		t.Fatal("polling must be paused after a 429")
	}
	if pause := time.Until(until); pause < 40*time.Second || pause > 80*time.Second {
		t.Errorf("first pause = %s, want about a minute", pause)
	}
	notices := h.tg.Calls("sendMessage")
	if len(notices) != 1 || notices[0].ChatID() != "111" || !strings.Contains(notices[0].Text(), "429") {
		t.Fatalf("want one notice to the admin, got %d: %+v", len(notices), notices)
	}

	// While paused nothing is requested at all.
	before := h.fetch.callCount()
	h.poll(false)
	if h.fetch.callCount() != before {
		t.Error("filters were polled during the pause")
	}
}

func TestFilterWithABadRequestIsStoppedAndItsOwnerTold(t *testing.T) {
	h := newHarness(t, "")
	bad := h.filter(domain.Search{Name: "<b>сломан</b>", Query: "x", OwnerChatID: "222"})
	good := h.filter(domain.Search{Name: "ok", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(good.ID, []domain.Offer{h.offer(1, "iPhone", 4)})
	for i := 0; i < maxBadRequests; i++ {
		h.fetch.fail(bad.ID, &olx.APIError{Status: 400, Body: "Bad Request"})
	}

	for i := 0; i < maxBadRequests; i++ {
		h.poll(false)
	}
	stored, _ := h.db.Search(ctx, bad.ID)
	if !stored.Disabled {
		t.Fatal("a filter OLX keeps rejecting must be stopped")
	}
	notice, ok := h.tg.Last("sendMessage")
	if !ok || notice.ChatID() != "222" {
		t.Fatalf("the owner must be told, got %+v", notice)
	}
	if !strings.Contains(notice.Text(), "&lt;b&gt;сломан&lt;/b&gt;") {
		t.Errorf("the filter name must be escaped in the notice: %s", notice.Text())
	}
	// A 400 is about one filter: the other keeps working and nothing is paused.
	if h.chk.BlockedUntil().After(time.Now()) {
		t.Error("a 400 must not pause the whole poll")
	}
	if got := h.delivered(); len(got) != 1 || got[0] != "111" {
		t.Errorf("the healthy filter must keep alerting: %v", got)
	}
}

// ---------- sender ----------

func TestSenderHonoursRetryAfter(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})
	h.tg.Queue("sendRichMessage", tgtest.FloodWait(1))

	start := time.Now()
	h.poll(false)
	if got := len(h.tg.Calls("sendRichMessage")); got != 2 {
		t.Fatalf("send attempts = %d, want the 429 and one retry", got)
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Error("the retry did not wait for retry_after")
	}
	if h.sentCount("111") != 1 {
		t.Fatal("the alert must end up delivered")
	}
	if n := len(h.tg.Calls("sendMessage")); n != 0 {
		t.Errorf("a flood limit must not produce a plain-text duplicate, sendMessage calls = %d", n)
	}
}

func TestSenderRetriesServerErrors(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})
	h.tg.Queue("sendRichMessage", tgtest.Fail(502, "Bad Gateway"))

	h.poll(false)
	if h.sentCount("111") != 1 {
		t.Fatalf("a 5xx must be retried until delivered; attempts = %d", len(h.tg.Calls("sendRichMessage")))
	}
}

func TestSenderGivesUpOnADeadChat(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4), h.offer(2, "iPhone 14", 3)})
	h.tg.On("sendRichMessage", func(tgtest.Call) *tgtest.Reply {
		r := tgtest.Fail(403, "Forbidden: bot was blocked by the user")
		return &r
	})

	h.poll(false)
	if got := len(h.tg.Calls("sendRichMessage")); got != 1 {
		t.Fatalf("a blocked chat was tried %d times, want once", got)
	}
	if n, _ := h.db.PendingCount(ctx); n != 0 {
		t.Fatalf("pending = %d; the rest of the queue for a dead chat must be dropped", n)
	}
	if !h.chk.isDead("111") {
		t.Fatal("the chat must be remembered as unreachable")
	}

	// New listings are not even queued for it for a while.
	h.fetch.set(s.ID, []domain.Offer{h.offer(3, "iPhone 15", 2)})
	h.poll(false)
	if got := len(h.tg.Calls("sendRichMessage")); got != 1 {
		t.Fatalf("a dead chat was written to again: %d calls", got)
	}
}

func TestSenderFailsARejectedCardOnce(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["111"],"admin_chat_ids":["111"],"rich_alerts":false}`)
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})
	h.tg.Queue("sendMessage", tgtest.Fail(400, "Bad Request: message is too long"))

	h.poll(false)
	if got := len(h.tg.Calls("sendMessage")); got != 1 {
		t.Fatalf("a rejected card was sent %d times, want once", got)
	}
	if h.sentCount("111") != 0 {
		t.Fatal("a rejected card is not a delivery")
	}
	if n, _ := h.db.PendingCount(ctx); n != 0 {
		t.Fatal("a rejected card must leave the queue")
	}
}

// cancellingMessenger delivers and then cancels the context, the way a
// shutdown signal can arrive right after Telegram answered.
type cancellingMessenger struct {
	cancel context.CancelFunc
	sent   int
}

func (m *cancellingMessenger) SendOffer(context.Context, string, domain.Offer, telegram.CardOptions) (int, error) {
	m.sent++
	m.cancel()
	return 100, nil
}

func (m *cancellingMessenger) SendHTML(context.Context, string, string, *telegram.InlineKeyboardMarkup, bool) (int, error) {
	return 1, nil
}

// The delivery must be recorded even when the context is cancelled right
// after the send: losing that write would send the alert again on restart.
func TestDeliveryIsRecordedEvenIfShutdownArrivesRightAfterTheSend(t *testing.T) {
	h := newHarness(t, "")
	h.stopSender()
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	queued, ok, err := h.db.CommitPage(ctx, store.Commit{SearchID: s.ID, Alerts: []store.Reservation{{
		OfferID: 1, ChatID: "111", Title: "iPhone", Payload: `{"id":1,"title":"iPhone"}`,
	}}})
	if err != nil || !ok || len(queued) != 1 {
		t.Fatal(ok, err)
	}

	sendCtx, cancel := context.WithCancel(context.Background())
	m := &cancellingMessenger{cancel: cancel}
	h.chk.tg = m
	h.chk.deliverChat(sendCtx, "111")

	if m.sent != 1 {
		t.Fatalf("sent = %d", m.sent)
	}
	if h.sentCount("111") != 1 {
		t.Fatal("the delivery was lost to the cancelled context")
	}
}

func TestUnreadablePayloadIsDroppedNotRetried(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	if _, _, err := h.db.CommitPage(ctx, store.Commit{SearchID: s.ID, Alerts: []store.Reservation{{OfferID: 1, ChatID: "111", Payload: "not json"}}}); err != nil {
		t.Fatal(err)
	}
	h.chk.wakeSender()
	h.chk.drain(5 * time.Second)
	if n, _ := h.db.PendingCount(ctx); n != 0 {
		t.Fatal("an unreadable alert must leave the queue")
	}
	if got := len(h.tg.Calls("")); got != 0 {
		t.Fatalf("nothing can be sent for it, got %d calls", got)
	}
	if h.sentCount("111") != 0 {
		t.Fatal("it must not be counted as delivered")
	}
}

// What a crash leaves in the queue is delivered by the next start, unless it
// has waited too long to still be news.
func TestQueuedAlertsSurviveARestartUntilTheyExpire(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	fresh, _ := json.Marshal(h.offer(1, "iPhone свежий", 4))
	stale, _ := json.Marshal(h.offer(2, "iPhone старый", 60))
	if _, _, err := h.db.CommitPage(ctx, store.Commit{SearchID: s.ID, Now: time.Now().Add(-3 * time.Minute),
		Alerts: []store.Reservation{{OfferID: 1, ChatID: "111", Payload: string(fresh)}}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.db.CommitPage(ctx, store.Commit{SearchID: s.ID, Now: time.Now().Add(-alertTTL - time.Minute),
		Alerts: []store.Reservation{{OfferID: 2, ChatID: "111", Payload: string(stale)}}}); err != nil {
		t.Fatal(err)
	}

	h.chk.wakeSender()
	h.chk.drain(5 * time.Second)
	cards := h.tg.Calls("sendRichMessage")
	if len(cards) != 1 || !strings.Contains(cards[0].Text(), "свежий") {
		t.Fatalf("want only the fresh alert delivered, got %d", len(cards))
	}
	if n, _ := h.db.PendingCount(ctx); n != 0 {
		t.Fatalf("pending = %d; the stale alert must be expired", n)
	}
}

func TestQuietHoursSendSilently(t *testing.T) {
	h := newHarness(t, "")
	hour := time.Now().In(h.cfg.Location()).Hour()
	if err := h.cfg.SetQuietHours("111", hour, (hour+1)%24, true); err != nil {
		t.Fatal(err)
	}
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})

	h.poll(false)
	card, ok := h.tg.Last("sendRichMessage")
	if !ok {
		t.Fatal("quiet hours must not hold an alert back")
	}
	if card.Payload["disable_notification"] != true {
		t.Error("an alert in quiet hours must be silent")
	}
}

func TestRunOnceDeliversBeforeReturning(t *testing.T) {
	h := newHarness(t, "")
	h.stopSender() // RunOnce brings its own sender
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})

	h.chk.RunOnce(ctx)
	if got := h.delivered(); len(got) != 1 {
		t.Fatalf("-once returned before the alert went out: %v", got)
	}
}

func TestRunStopsAndDrainsOnCancel(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["111"],"admin_chat_ids":["111"],"notify_on_startup":true,"interval_seconds":3}`)
	h.stopSender()
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "iPhone 13", 4)})

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); h.chk.Run(runCtx) }()

	deadline := time.After(10 * time.Second)
	for len(h.delivered()) == 0 {
		select {
		case <-deadline:
			t.Fatal("no alert was delivered")
		case <-time.After(20 * time.Millisecond):
		}
	}
	if !h.chk.Started() {
		t.Error("the checker must report that it has started")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

// ---------- log ----------

func TestFormatAdFound(t *testing.T) {
	o := domain.Offer{
		ID:              940639374,
		Title:           "продаю анти кваріант",
		URL:             "https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html",
		CreatedTime:     "2026-11-12T05:40:28+02:00",
		LastRefreshTime: "2026-11-12T05:43:51+02:00",
	}
	t.Setenv("PANTHEON_ANSI_LINKS", "1")
	line := formatAdFound(&o, "Електроніка", true)
	if strings.Contains(line, "продаю анти кваріант") {
		t.Errorf("the title does not belong in the log: %s", line)
	}
	if !strings.Contains(line, "[Електроніка]") {
		t.Errorf("the filter name is missing: %s", line)
	}
	if !strings.Contains(line, "\x1b]8;;"+o.URL+"\x1b\\link\x1b]8;;\x1b\\") {
		t.Errorf("expected an OSC 8 hyperlink: %q", line)
	}

	plain := formatAdFound(&o, "Електроніка", false)
	if strings.Contains(plain, "\x1b]8;;") || !strings.Contains(plain, o.URL) {
		t.Errorf("with links off the raw URL is printed: %q", plain)
	}

	// A listing URL cannot smuggle its own escape sequence into the terminal.
	evil := o
	evil.URL = "https://olx.ua/x\x1b]0;pwned\x07"
	if got := formatAdFound(&evil, "f", true); strings.Contains(got, "\x07") {
		t.Errorf("control characters from a URL reached the terminal: %q", got)
	}
}

func TestStartMessageAndBanner(t *testing.T) {
	msg := StartMessage(time.Now(), 3, 15)
	if !strings.HasPrefix(msg, "Started pantheon v") || !strings.Contains(msg, ", active filters: 3, interval - 15s") {
		t.Errorf("start message: %s", msg)
	}
	if !strings.Contains(Banner, "github.com/undervolter/pantheon") || !strings.Contains(Banner, "t.me/pantheonUpdates") {
		t.Error("the banner must keep the project links")
	}
}

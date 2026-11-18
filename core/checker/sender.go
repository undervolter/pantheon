package checker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/store"
	"pantheon/core/telegram"
)

// The sender delivers what the poll cycles queued. The database is the
// queue: nothing about a pending alert lives only in memory, so a crash at
// any point leaves it to be picked up by the next start.

const (
	// An alert that could not be delivered for this long is no longer news.
	alertTTL = 10 * time.Minute

	senderTick   = 2 * time.Second
	senderBatch  = 20
	retryWaitMin = 2 * time.Second
	retryWaitMax = 30 * time.Second

	// A chat that turned out to be unreachable is not written to for a while;
	// its owner may unblock the bot later.
	deadChatFor = 30 * time.Minute

	// How long shutdown waits for the queue to empty.
	shutdownDrain = 20 * time.Second
	onceDrain     = 90 * time.Second

	// The write that records a delivery must not be lost to a cancelled
	// context: losing it would send the alert again after a restart.
	markTimeout = 5 * time.Second
)

// errBadPayload means a queued alert lost the listing it was about. It can
// only be given up on.
var errBadPayload = errors.New("the queued alert has no readable listing")

func (c *Checker) wakeSender() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

func (c *Checker) isDead(chatID string) bool {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	until, ok := c.dead[chatID]
	if !ok {
		return false
	}
	if c.now().After(until) {
		delete(c.dead, chatID)
		return false
	}
	return true
}

func (c *Checker) markDead(chatID string, reason error) {
	c.sendMu.Lock()
	_, known := c.dead[chatID]
	c.dead[chatID] = c.now().Add(deadChatFor)
	c.sendMu.Unlock()
	if !known {
		logx.Errorf("chat %s cannot be written to (%v); its alerts are dropped for %s", chatID, reason, deadChatFor)
	}
}

// senderLoop starts a worker for every chat with pending alerts and keeps
// doing so until ctx is cancelled.
func (c *Checker) senderLoop(ctx context.Context) {
	ticker := time.NewTicker(senderTick)
	defer ticker.Stop()
	for {
		c.dispatchPending(ctx)
		select {
		case <-ctx.Done():
			c.workerWG.Wait()
			return
		case <-ticker.C:
		case <-c.wake:
		}
	}
}

// forgetStale drops what the sender remembers about alerts that have left the
// queue some other way (expired, their chat removed): without it the notes
// would pile up for as long as the process lives.
func (c *Checker) forgetStale() {
	cutoff := c.now().Add(-2 * alertTTL)
	c.sendMu.Lock()
	for id, at := range c.noPhotos {
		if at.Before(cutoff) {
			delete(c.noPhotos, id)
		}
	}
	for id, at := range c.unrecorded {
		if at.Before(cutoff) {
			delete(c.unrecorded, id)
		}
	}
	c.sendMu.Unlock()
}

func (c *Checker) dispatchPending(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	c.forgetStale()
	if n, err := c.db.ExpireAlerts(ctx, c.now().Add(-alertTTL)); err != nil {
		if ctx.Err() == nil {
			logx.Warnf("expire alerts: %v", err)
		}
		return
	} else if n > 0 {
		logx.Warnf("%d alerts waited longer than %s and were dropped", n, alertTTL)
	}
	chats, err := c.db.PendingChats(ctx)
	if err != nil {
		if ctx.Err() == nil {
			logx.Warnf("read the alert queue: %v", err)
		}
		return
	}
	for _, chatID := range chats {
		c.sendMu.Lock()
		busy := c.workers[chatID]
		if !busy {
			c.workers[chatID] = true
			c.workerWG.Add(1)
		}
		c.sendMu.Unlock()
		if busy {
			continue
		}
		go func(chatID string) {
			defer func() {
				c.sendMu.Lock()
				delete(c.workers, chatID)
				c.sendMu.Unlock()
				c.workerWG.Done()
			}()
			c.deliverChat(ctx, chatID)
		}(chatID)
	}
}

// drain waits until the queue is empty or the deadline passes.
func (c *Checker) drain(deadline time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	c.wakeSender()
	for {
		n, err := c.db.PendingCount(ctx)
		if err != nil || n == 0 {
			return
		}
		select {
		case <-ctx.Done():
			logx.Warnf("%d alerts were still queued at shutdown; they will be sent on the next start if they are not too old by then", n)
			return
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// deliverChat sends the pending alerts of one chat in queue order. Alerts of
// a chat never overtake each other; chats do not wait for one another.
func (c *Checker) deliverChat(ctx context.Context, chatID string) {
	retryWait := retryWaitMin
	for ctx.Err() == nil {
		alerts, err := c.db.PendingAlerts(ctx, chatID, senderBatch)
		if err != nil {
			if ctx.Err() == nil {
				logx.Warnf("read alerts of %s: %v", chatID, err)
			}
			return
		}
		if len(alerts) == 0 {
			return
		}
		retry := false
		for _, a := range alerts {
			if retry {
				break
			}
			if c.isDead(chatID) || !c.cfg.IsTrusted(chatID) {
				c.failRest(chatID, "chat is unreachable or was removed")
				return
			}
			if c.now().Sub(a.CreatedAt) > alertTTL {
				// Left for ExpireAlerts; the rest of the batch is just as old.
				return
			}
			if c.isUnrecorded(a.ID) {
				// The card went out, but the write that records it failed, so
				// the alert is still in the queue. It is never sent twice:
				// only the write is tried again, and nothing else goes to
				// this chat until it succeeds.
				if c.markSent(a.ID) {
					continue
				}
				if !sleepCtx(ctx, retryWait) {
					return
				}
				if retryWait < retryWaitMax {
					retryWait *= 2
				}
				retry = true
				continue
			}
			if !sleepCtx(ctx, c.limiter.Reserve(chatID)) {
				return
			}
			// The batch was read a moment ago. Meanwhile the alert may have been
			// taken out of the queue: its filter deleted, its chat removed.
			if pending, err := c.db.AlertPending(ctx, a.ID); err == nil && !pending {
				c.forget(a.ID)
				continue
			} else if ctx.Err() != nil {
				return
			}

			err := c.sendAlert(ctx, a)
			if err != nil && ctx.Err() != nil {
				// Interrupted by shutdown, not refused. The alert stays queued
				// and goes out on the next start.
				return
			}
			switch {
			case err == nil:
				if !c.markSent(a.ID) {
					retry = true
					continue
				}
				retryWait = retryWaitMin

			case errors.Is(err, errBadPayload):
				c.fail(a.ID, err.Error())

			case telegram.IsDeadChat(err):
				c.markDead(chatID, err)
				c.failRest(chatID, err.Error())
				return

			case telegram.IsRetryable(err):
				c.note(a.ID, err)
				wait := retryWait
				if limited, ok := telegram.IsRateLimited(err); ok {
					c.limiter.Penalize(chatID, limited)
					wait = limited
				}
				if telegram.IsTimeout(err) {
					// Whether the card arrived is unknown. It is sent again —
					// a duplicate is better than a lost alert — but without
					// photos, the usual reason Telegram takes too long.
					c.sendMu.Lock()
					c.noPhotos[a.ID] = c.now()
					c.sendMu.Unlock()
				}
				logx.Warnf("alert %d to %s failed, retry in %s: %v", a.OfferID, chatID, wait.Round(time.Second), err)
				if retryWait < retryWaitMax {
					retryWait *= 2
				}
				if !sleepCtx(ctx, wait) {
					return
				}
				// Start over from the queue: this alert is still its head.
				retry = true

			default:
				// Telegram understood the request and refused it; trying
				// again would get the same answer.
				logx.Errorf("alert %d to %s was rejected: %v", a.OfferID, chatID, err)
				c.fail(a.ID, err.Error())
			}
		}
	}
}

func (c *Checker) sendAlert(ctx context.Context, a store.Alert) error {
	var offer domain.Offer
	if err := json.Unmarshal([]byte(a.Payload), &offer); err != nil || offer.ID == 0 {
		return errBadPayload
	}

	c.sendMu.Lock()
	_, skipPhotos := c.noPhotos[a.ID]
	c.sendMu.Unlock()

	opt := telegram.CardOptions{
		Lang:       i18n.Of(c.cfg.LangFor(a.ChatID)),
		FilterName: a.SearchName,
		Rich:       c.cfg.GetRichAlerts(),
		ShowParams: c.cfg.GetShowParams(),
		MaxParams:  c.cfg.GetMaxParams(),
		// Quiet hours deliver without a sound; they never hold an alert back,
		// which would run into the alert's own time limit.
		Silent: c.cfg.InQuietHours(a.ChatID, c.now()),
	}
	if c.cfg.GetSendPhotos() && !skipPhotos {
		opt.MaxPhotos = c.cfg.GetMaxPhotos()
	}
	_, err := c.tg.SendOffer(ctx, a.ChatID, offer, opt)
	return err
}

func (c *Checker) forget(id int64) {
	c.sendMu.Lock()
	delete(c.noPhotos, id)
	delete(c.unrecorded, id)
	c.sendMu.Unlock()
}

func (c *Checker) isUnrecorded(id int64) bool {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	_, ok := c.unrecorded[id]
	return ok
}

// markSent records a delivery. If the write fails, the alert is remembered as
// delivered, so that it is not sent again while it is still in the queue.
func (c *Checker) markSent(id int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), markTimeout)
	defer cancel()
	if err := c.db.MarkAlertSent(ctx, id, c.now()); err != nil {
		c.sendMu.Lock()
		_, known := c.unrecorded[id]
		if !known {
			c.unrecorded[id] = c.now()
		}
		c.sendMu.Unlock()
		if !known {
			logx.Errorf("alert %d was delivered, but recording it failed (%v); it will not be sent again, the write is retried", id, err)
		}
		return false
	}
	c.forget(id)
	return true
}

func (c *Checker) fail(id int64, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), markTimeout)
	defer cancel()
	if err := c.db.MarkAlertFailed(ctx, id, reason); err != nil {
		logx.Errorf("record the failure of alert %d: %v", id, err)
	}
	c.forget(id)
}

func (c *Checker) note(id int64, cause error) {
	ctx, cancel := context.WithTimeout(context.Background(), markTimeout)
	defer cancel()
	if err := c.db.NoteAlertAttempt(ctx, id, cause.Error()); err != nil {
		logx.Warnf("record the attempt of alert %d: %v", id, err)
	}
}

func (c *Checker) failRest(chatID, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), markTimeout)
	defer cancel()
	if _, err := c.db.FailPendingOf(ctx, chatID, reason); err != nil {
		logx.Warnf("drop queued alerts of %s: %v", chatID, err)
	}
}

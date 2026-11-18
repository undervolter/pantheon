package checker

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"pantheon/core/domain"
	"pantheon/core/olx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/telegram/tgtest"
)

// queueAlerts puts alerts for chat 111 straight into the queue.
func (h *harness) queueAlerts(s domain.Search, offerIDs ...int64) []store.Alert {
	h.t.Helper()
	commit := store.Commit{SearchID: s.ID}
	for _, id := range offerIDs {
		commit.Alerts = append(commit.Alerts, store.Reservation{
			OfferID: id, ChatID: "111", Title: "iPhone", Payload: fmt.Sprintf(`{"id":%d,"title":"iPhone"}`, id),
		})
	}
	queued, ok, err := h.db.CommitPage(ctx, commit)
	if err != nil || !ok || len(queued) != len(offerIDs) {
		h.t.Fatalf("queue alerts: %d queued, ok=%v, %v", len(queued), ok, err)
	}
	return queued
}

func (h *harness) pending() int {
	h.t.Helper()
	n, err := h.db.PendingCount(ctx)
	if err != nil {
		h.t.Fatal(err)
	}
	return n
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// hangingMessenger never answers until its context is cancelled: a send that
// is still on the wire when the process is told to stop.
type hangingMessenger struct{ started chan struct{} }

func (m *hangingMessenger) SendOffer(ctx context.Context, _ string, _ domain.Offer, _ telegram.CardOptions) (int, error) {
	close(m.started)
	<-ctx.Done()
	return 0, fmt.Errorf("sendRichMessage: %w", ctx.Err())
}

func (m *hangingMessenger) SendHTML(context.Context, string, string, *telegram.InlineKeyboardMarkup, bool) (int, error) {
	return 0, nil
}

// A send cut short by shutdown was taken for a refusal: the alert was failed
// for good while the log promised to send it on the next start.
func TestSendInterruptedByShutdownStaysQueued(t *testing.T) {
	h := newHarness(t, "")
	h.stopSender()
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.queueAlerts(s, 1)

	sendCtx, cancel := context.WithCancel(context.Background())
	m := &hangingMessenger{started: make(chan struct{})}
	h.chk.tg = m
	done := make(chan struct{})
	go func() { defer close(done); h.chk.deliverChat(sendCtx, "111") }()
	<-m.started
	cancel()
	<-done

	if n := h.pending(); n != 1 {
		t.Fatalf("pending = %d: an interrupted send must leave the alert in the queue", n)
	}
}

// When the write that records a delivery fails, the alert is still "pending"
// in the database. It used to be sent again on every pass.
func TestDeliveredAlertIsNotSentAgainWhenRecordingFails(t *testing.T) {
	h := newHarness(t, "")
	h.stopSender()
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	h.queueAlerts(s, 1, 2)

	// A second connection makes that one write fail, the way a full disk would.
	raw, err := sql.Open("sqlite", h.db.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err := raw.Exec(`CREATE TRIGGER no_sent BEFORE UPDATE OF status ON alerts WHEN NEW.status = 'sent'
		BEGIN SELECT RAISE(ABORT, 'disk full'); END`); err != nil {
		t.Fatal(err)
	}

	sendCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); h.chk.deliverChat(sendCtx, "111") }()

	waitFor(t, "the first card", func() bool { return len(h.delivered()) >= 1 })
	// Long enough for a couple of retries of the write.
	time.Sleep(retryWaitMin + time.Second)
	if n := len(h.delivered()); n != 1 {
		t.Fatalf("cards sent = %d: the delivered alert was sent again, or the next one overtook the failed write", n)
	}

	if _, err := raw.Exec("DROP TRIGGER no_sent"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the queue to empty", func() bool { return h.pending() == 0 })
	<-done
	if n := len(h.delivered()); n != 2 {
		t.Fatalf("cards sent = %d, want each alert exactly once", n)
	}
	if h.sentCount("111") != 2 {
		t.Fatalf("recorded deliveries = %d, want 2", h.sentCount("111"))
	}
}

// The sender works from a batch it read earlier. An alert taken out of the
// queue meanwhile — its filter was deleted — must not go out anyway.
func TestAlertOfAFilterDeletedMeanwhileIsNotSent(t *testing.T) {
	h := newHarness(t, "")
	h.stopSender()
	s := h.filter(domain.Search{Name: "noisy", Query: "iphone", OwnerChatID: "111"})
	h.queueAlerts(s, 1, 2, 3)

	// While the first card is on its way, the user deletes the filter.
	deleted := false
	h.tg.On("sendRichMessage", func(tgtest.Call) *tgtest.Reply {
		if !deleted {
			deleted = true
			if _, err := h.db.DeleteSearch(ctx, s.ID); err != nil {
				t.Error(err)
			}
		}
		return nil
	})
	h.chk.deliverChat(ctx, "111")

	if n := len(h.delivered()); n != 1 {
		t.Fatalf("cards sent = %d: a deleted filter kept alerting from the batch already read", n)
	}
}

// A 403 on a backfill page was logged and swallowed, so polling carried on
// against an address OLX had just refused.
func TestRefusalDuringBackfillPausesPolling(t *testing.T) {
	h := newHarness(t, "")
	s := h.filter(domain.Search{Name: "f", Query: "x", OwnerChatID: "111"})
	h.chk.setFrontier(s.ID, h.now.Add(-30*time.Minute))
	h.fetch.set(s.ID, []domain.Offer{h.offer(1, "x", 1), h.offer(2, "x", 5)}) // does not reach the frontier
	h.fetch.fail(s.ID, nil, &olx.APIError{Status: 403, Body: "Forbidden"})

	h.poll(false)

	if !h.chk.BlockedUntil().After(time.Now()) {
		t.Fatal("a refusal on a backfill page must pause polling")
	}
	// The first page is read again after the pause, so nothing of it may be
	// marked seen or alerted now.
	if seen, _ := h.db.SeenSet(ctx, []string{s.ID}, []int64{1, 2}); len(seen) != 0 {
		t.Errorf("seen = %v, want nothing recorded from an abandoned poll", seen)
	}
	if n := len(h.delivered()); n != 0 {
		t.Errorf("cards sent = %d", n)
	}
	if got := h.chk.frontier(s.ID); !got.Equal(h.now.Add(-30 * time.Minute)) {
		t.Errorf("the frontier moved to %s", got)
	}
}

// With low concurrency, fetches queue up behind the one in progress. Once
// that one is refused, the queued ones must not send their requests.
func TestQueuedFetchIsDroppedOnceOLXRefuses(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["111"],"admin_chat_ids":["111"],"concurrency":1,"interval_seconds":3}`)
	a := h.filter(domain.Search{Name: "a", Query: "x", OwnerChatID: "111"})
	b := h.filter(domain.Search{Name: "b", Query: "y", OwnerChatID: "111"})
	h.fetch.fail(a.ID, &olx.APIError{Status: 429, Body: "Too Many Requests"})
	release := make(chan struct{})
	h.fetch.holdUntil(a.ID, release)

	done := make(chan struct{})
	go func() { defer close(done); h.chk.cycle(ctx, false) }()
	// Let the second fetch get in line behind the first, then let OLX answer.
	waitFor(t, "the first fetch", func() bool { return h.fetch.callCount() == 1 })
	time.Sleep(maxFetchSpacing + 300*time.Millisecond)
	close(release)
	<-done

	h.fetch.mu.Lock()
	calls := append([]string(nil), h.fetch.calls...)
	h.fetch.mu.Unlock()
	if len(calls) != 1 {
		t.Fatalf("requests = %v: filter %s was still requested after the refusal", calls, b.ID)
	}
	if !h.chk.BlockedUntil().After(time.Now()) {
		t.Fatal("polling must be paused")
	}
}

// The time limit of an alert counts from when it was queued, not from the
// start of a cycle whose fetches may have taken minutes.
func TestAlertIsStampedWhenItIsQueued(t *testing.T) {
	h := newHarness(t, "")
	h.stopSender()
	s := h.filter(domain.Search{Name: "f", Query: "iphone", OwnerChatID: "111"})
	rules := h.chk.rules()
	rules.Now = rules.Now.Add(-9 * time.Minute) // the cycle began long ago
	h.chk.process(ctx, s, []domain.Offer{h.offer(1, "iPhone 13", 4)}, rules, false)

	alerts, err := h.db.PendingAlerts(ctx, "111", 10)
	if err != nil || len(alerts) != 1 {
		t.Fatalf("pending = %d, %v", len(alerts), err)
	}
	if age := time.Since(alerts[0].CreatedAt); age > time.Minute {
		t.Fatalf("the alert was queued with an age of %s: most of its time limit was already spent", age.Round(time.Second))
	}
}

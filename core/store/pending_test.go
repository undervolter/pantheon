package store

import (
	"testing"
	"time"

	"pantheon/core/domain"
)

// The sender asks this right before each send, because the batch it works
// from may be out of date.
func TestAlertPending(t *testing.T) {
	s := open(t)
	sc := add(t, s, domain.Search{Name: "f", Query: "x", OwnerChatID: "1"})
	queued, ok, err := s.CommitPage(ctx, Commit{SearchID: sc.ID, Alerts: []Reservation{
		{OfferID: 1, ChatID: "1", Payload: "{}"},
		{OfferID: 2, ChatID: "1", Payload: "{}"},
	}})
	if err != nil || !ok || len(queued) != 2 {
		t.Fatal(ok, err)
	}

	for _, a := range queued {
		if pending, err := s.AlertPending(ctx, a.ID); err != nil || !pending {
			t.Fatalf("alert %d: pending=%v, %v", a.ID, pending, err)
		}
	}
	if err := s.MarkAlertSent(ctx, queued[0].ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if pending, _ := s.AlertPending(ctx, queued[0].ID); pending {
		t.Error("a sent alert is not pending")
	}

	// Deleting the filter takes what it queued out of the queue.
	if _, err := s.DeleteSearch(ctx, sc.ID); err != nil {
		t.Fatal(err)
	}
	if pending, _ := s.AlertPending(ctx, queued[1].ID); pending {
		t.Error("an alert of a deleted filter is not pending")
	}
	if pending, err := s.AlertPending(ctx, 999999); err != nil || pending {
		t.Errorf("unknown alert: pending=%v, %v", pending, err)
	}
}

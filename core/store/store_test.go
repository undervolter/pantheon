package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pantheon/core/domain"
)

var ctx = context.Background()

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "pantheon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func add(t *testing.T, s *Store, sc domain.Search) domain.Search {
	t.Helper()
	out, err := s.AddSearch(ctx, sc)
	if err != nil {
		t.Fatalf("AddSearch(%+v): %v", sc, err)
	}
	return out
}

// ---------- database ----------

func TestOpenIsIdempotentAndUsesWAL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub dir", "bots#1", "pantheon.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	var mode string
	if err := s.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil || mode != "wal" {
		t.Errorf("journal_mode = %q, %v; want wal", mode, err)
	}
	created := add(t, s, domain.Search{Name: "a", Query: "iphone"})
	_ = s.Close()

	// The '#' in the directory name must not cut the path short.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database file is not where it was asked for: %v", err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s.Close()
	got, err := s.Search(ctx, created.ID)
	if err != nil || got.Query != "iphone" {
		t.Fatalf("filter after reopen = %+v, %v", got, err)
	}
}

func TestOpenRefusesANewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pantheon.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("PRAGMA user_version = 99"); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()

	if _, err := Open(path); !errors.Is(err, ErrNewerSchema) {
		t.Fatalf("err = %v, want ErrNewerSchema", err)
	}
}

func TestWithTxRollsBack(t *testing.T) {
	s := open(t)
	boom := errors.New("boom")
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO meta(key, value) VALUES('k', 'v')"); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	if set, _ := s.Flag(ctx, "k"); set {
		t.Fatal("a failed transaction must leave nothing behind")
	}
}

// ---------- filters ----------

func TestSearchRoundTrip(t *testing.T) {
	s := open(t)
	in := domain.Search{
		Name: "Камеры", Query: "canon", QueryVariants: []string{"canon", "кэнон"}, ExcludeWords: []string{"чохол"},
		MinPrice: 100, MaxPrice: 9000, Category: 73, CatName: "Цифрові фотоапарати",
		CityID: 268, CityName: "Київ", RegionID: 25, DistrictID: 11, DistrictName: "Печерський",
		BrandParam: "filter_enum_digital_camera_manufacturers", BrandValues: []string{"2580"}, BrandNames: []string{"Canon"},
		BrandValue: "2580", BrandName: "Canon", OwnerType: "private",
		Filters:     map[string][]string{"filter_enum_state": {"used"}},
		Ranges:      map[string]domain.FilterRange{"filter_float_motor_year": {From: 2010, To: 2020}},
		OwnerChatID: "111", SnoozedUntil: 1_900_000_000,
	}
	created := add(t, s, in)
	if len(created.ID) != 8 || created.CreatedAt == 0 || created.EnabledAt != created.CreatedAt {
		t.Fatalf("id/created/enabled = %q/%d/%d", created.ID, created.CreatedAt, created.EnabledAt)
	}

	got, err := s.Search(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RuleKey() != created.RuleKey() || got.Name != "Камеры" || got.OwnerChatID != "111" ||
		got.CityName != "Київ" || got.DistrictName != "Печерський" || got.SnoozedUntil != 1_900_000_000 ||
		got.BrandNames[0] != "Canon" || got.Ranges["filter_float_motor_year"].To != 2020 {
		t.Fatalf("round trip lost data:\n in=%+v\nout=%+v", created, got)
	}
}

func TestAddSearchValidates(t *testing.T) {
	s := open(t)
	if _, err := s.AddSearch(ctx, domain.Search{Name: "Новый поиск"}); !errors.Is(err, domain.ErrNoCondition) {
		t.Errorf("empty filter: err = %v, want ErrNoCondition", err)
	}
	if _, err := s.AddSearch(ctx, domain.Search{Name: "x", Category: 1758}); !errors.Is(err, domain.ErrRealEstate) {
		t.Errorf("real estate: err = %v, want ErrRealEstate", err)
	}
	if list, _ := s.Searches(ctx); len(list) != 0 {
		t.Fatalf("rejected filters must not be stored: %+v", list)
	}
}

func TestSearchesKeepCreationOrderAndUniqueIDs(t *testing.T) {
	s := open(t)
	ids := map[string]bool{}
	for _, name := range []string{"a", "b", "c", "d"} {
		ids[add(t, s, domain.Search{Name: name, Query: name}).ID] = true
	}
	if len(ids) != 4 {
		t.Fatalf("ids are not unique: %v", ids)
	}
	list, err := s.Searches(ctx)
	if err != nil || len(list) != 4 {
		t.Fatalf("Searches = %d, %v", len(list), err)
	}
	for i, name := range []string{"a", "b", "c", "d"} {
		if list[i].Name != name {
			t.Errorf("position %d = %q, want %q", i, list[i].Name, name)
		}
	}
}

// A result of Searches must not share memory with another call: 0.1.x handed
// out shallow copies whose maps were shared with the checker.
func TestSearchesReturnIndependentCopies(t *testing.T) {
	s := open(t)
	add(t, s, domain.Search{Name: "a", Query: "a", Filters: map[string][]string{"filter_enum_state": {"used"}}})
	first, _ := s.Searches(ctx)
	first[0].Filters["filter_enum_state"][0] = "changed"
	first[0].Filters["filter_enum_x"] = []string{"1"}
	second, _ := s.Searches(ctx)
	if second[0].Filters["filter_enum_state"][0] != "used" || len(second[0].Filters) != 1 {
		t.Fatalf("callers share filter maps: %+v", second[0].Filters)
	}
}

func TestUpdateSearch(t *testing.T) {
	s := open(t)
	created := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "111"})

	updated, err := s.UpdateSearch(ctx, created.ID, func(sc *domain.Search) error {
		sc.Disabled = true
		sc.OwnerChatID = ""
		sc.ID = "hacked"
		sc.CreatedAt = 1
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Disabled || !updated.IsShared() || updated.ID != created.ID || updated.CreatedAt != created.CreatedAt {
		t.Fatalf("update = %+v", updated)
	}
	got, _ := s.Search(ctx, created.ID)
	if !got.Disabled || !got.IsShared() {
		t.Fatalf("update was not stored: %+v", got)
	}

	boom := errors.New("refused")
	if _, err := s.UpdateSearch(ctx, created.ID, func(sc *domain.Search) error { sc.Name = "z"; return boom }); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	if got, _ := s.Search(ctx, created.ID); got.Name != "a" {
		t.Fatal("a refused update must change nothing")
	}
	if _, err := s.UpdateSearch(ctx, "missing", func(*domain.Search) error { return nil }); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing id: err = %v, want ErrNotFound", err)
	}
}

func TestDeleteSearchDropsItsSeenMarksOnly(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "1"})
	b := add(t, s, domain.Search{Name: "b", Query: "b", OwnerChatID: "1"})
	for _, sc := range []domain.Search{a, b} {
		if _, ok, err := s.CommitPage(ctx, Commit{SearchID: sc.ID, SeenIDs: []int64{10, 11}}); err != nil || !ok {
			t.Fatal(ok, err)
		}
	}

	deleted, err := s.DeleteSearch(ctx, a.ID)
	if err != nil || !deleted {
		t.Fatalf("DeleteSearch = %v, %v", deleted, err)
	}
	if _, err := s.Search(ctx, a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted filter still readable: %v", err)
	}
	if seen, _ := s.SeenSet(ctx, []string{a.ID}, []int64{10, 11}); len(seen) != 0 {
		t.Errorf("seen marks of the deleted filter remain: %v", seen)
	}
	if seen, _ := s.SeenSet(ctx, []string{b.ID}, []int64{10, 11}); len(seen) != 2 {
		t.Errorf("another filter lost its seen marks: %v", seen)
	}
	if again, _ := s.DeleteSearch(ctx, a.ID); again {
		t.Error("deleting twice must report false")
	}
}

func TestDisableSearchesOf(t *testing.T) {
	s := open(t)
	mine := add(t, s, domain.Search{Name: "mine", Query: "a", OwnerChatID: "1"})
	other := add(t, s, domain.Search{Name: "other", Query: "b", OwnerChatID: "2"})
	shared := add(t, s, domain.Search{Name: "shared", Query: "c"})

	if n, err := s.DisableSearchesOf(ctx, "1"); err != nil || n != 1 {
		t.Fatalf("DisableSearchesOf = %d, %v", n, err)
	}
	for id, want := range map[string]bool{mine.ID: true, other.ID: false, shared.ID: false} {
		if got, _ := s.Search(ctx, id); got.Disabled != want {
			t.Errorf("%s disabled = %v, want %v", got.Name, got.Disabled, want)
		}
	}
	if n, _ := s.CountSearches(ctx, "1"); n != 1 {
		t.Errorf("CountSearches(1) = %d", n)
	}
}

// ---------- seen ----------

// The bug this replaces: seen was keyed by owner, so a listing rejected by
// filter A (query mismatch) was hidden from the same owner's filter B.
func TestSeenIsPerFilter(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "iphone", OwnerChatID: "1"})
	b := add(t, s, domain.Search{Name: "b", Category: 44, OwnerChatID: "1"})

	if _, ok, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{500}}); err != nil || !ok {
		t.Fatal(ok, err)
	}
	seenA, _ := s.SeenSet(ctx, SeenScopes(a.ID, "1"), []int64{500, 501})
	seenB, _ := s.SeenSet(ctx, SeenScopes(b.ID, "1"), []int64{500, 501})
	if !seenA[500] || seenA[501] {
		t.Errorf("filter a: %v", seenA)
	}
	if seenB[500] {
		t.Error("a listing evaluated by filter a must still be new for filter b of the same owner")
	}
}

func TestSeenSetHandlesLargePages(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a"})
	var ids []int64
	for i := int64(1); i <= 1500; i++ {
		ids = append(ids, i)
	}
	if _, ok, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: ids[:900]}); err != nil || !ok {
		t.Fatal(ok, err)
	}
	seen, err := s.SeenSet(ctx, SeenScopes(a.ID, ""), ids)
	if err != nil || len(seen) != 900 {
		t.Fatalf("SeenSet = %d, %v; want 900", len(seen), err)
	}
}

func TestPruneKeepsRefreshedMarks(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a"})
	old := time.Now().Add(-10 * 24 * time.Hour)

	if _, _, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{1, 2}, Now: old}); err != nil {
		t.Fatal(err)
	}
	// Listing 2 is still on the results page: the next poll refreshes its mark.
	if _, _, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{2}}); err != nil {
		t.Fatal(err)
	}

	removed, _, err := s.Prune(ctx, time.Now().Add(-7*24*time.Hour), time.Time{})
	if err != nil || removed != 1 {
		t.Fatalf("Prune removed %d marks, %v; want 1", removed, err)
	}
	seen, _ := s.SeenSet(ctx, []string{a.ID}, []int64{1, 2})
	if seen[1] || !seen[2] {
		t.Fatalf("after prune: %v; the refreshed mark must survive", seen)
	}
}

// ---------- alerts ----------

func reservation(offerID int64, chatID string) Reservation {
	return Reservation{OfferID: offerID, ChatID: chatID, Title: "t", URL: "https://olx.ua/x", Price: "1 грн.", Payload: `{"id":1}`}
}

func TestCommitPageQueuesAlertsAtomically(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "Телефоны", Query: "a", OwnerChatID: "1"})

	queued, ok, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{1, 2, 3},
		Alerts: []Reservation{reservation(1, "1"), reservation(2, "1")}})
	if err != nil || !ok || len(queued) != 2 {
		t.Fatalf("CommitPage = %d alerts, ok=%v, %v", len(queued), ok, err)
	}
	if queued[0].SearchName != "Телефоны" || queued[0].SearchID != a.ID || queued[0].Status != StatusPending {
		t.Errorf("queued alert = %+v", queued[0])
	}
	pending, _ := s.PendingAlerts(ctx, "1", 10)
	if len(pending) != 2 || pending[0].OfferID != 1 || pending[1].OfferID != 2 || pending[0].Payload == "" {
		t.Fatalf("pending = %+v", pending)
	}
}

// One listing matched by two filters of the same chat is one alert.
func TestAlertsAreDedupedPerChatAcrossFilters(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "1"})
	b := add(t, s, domain.Search{Name: "b", Query: "b", OwnerChatID: "1"})
	c := add(t, s, domain.Search{Name: "c", Query: "c", OwnerChatID: "2"})

	first, _, _ := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{7}, Alerts: []Reservation{reservation(7, "1")}})
	second, _, _ := s.CommitPage(ctx, Commit{SearchID: b.ID, SeenIDs: []int64{7}, Alerts: []Reservation{reservation(7, "1")}})
	third, _, _ := s.CommitPage(ctx, Commit{SearchID: c.ID, SeenIDs: []int64{7}, Alerts: []Reservation{reservation(7, "2")}})

	if len(first) != 1 || len(second) != 0 || len(third) != 1 {
		t.Fatalf("queued %d / %d / %d, want 1 / 0 / 1", len(first), len(second), len(third))
	}
	// The second filter still records that it has seen the listing.
	if seen, _ := s.SeenSet(ctx, []string{b.ID}, []int64{7}); !seen[7] {
		t.Error("the deduplicated filter must still mark the listing as seen")
	}
}

// A filter deleted or paused while its page was being fetched must not get
// seen marks or alerts written afterwards.
func TestCommitPageRefusesADeletedOrPausedFilter(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "1"})
	if _, err := s.UpdateSearch(ctx, a.ID, func(sc *domain.Search) error { sc.Disabled = true; return nil }); err != nil {
		t.Fatal(err)
	}
	queued, ok, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{1}, Alerts: []Reservation{reservation(1, "1")}})
	if err != nil || ok || len(queued) != 0 {
		t.Fatalf("paused filter: queued=%d ok=%v err=%v", len(queued), ok, err)
	}
	if seen, _ := s.SeenSet(ctx, []string{a.ID}, []int64{1}); seen[1] {
		t.Error("nothing may be written for a paused filter")
	}

	_, _ = s.DeleteSearch(ctx, a.ID)
	if _, ok, err := s.CommitPage(ctx, Commit{SearchID: a.ID, SeenIDs: []int64{2}, Alerts: []Reservation{reservation(2, "1")}}); err != nil || ok {
		t.Fatalf("deleted filter: ok=%v err=%v", ok, err)
	}
	if n, _ := s.PendingCount(ctx); n != 0 {
		t.Fatalf("pending = %d, want 0", n)
	}
}

func TestAlertLifecycle(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "1"})
	queued, _, _ := s.CommitPage(ctx, Commit{SearchID: a.ID,
		Alerts: []Reservation{reservation(1, "1"), reservation(2, "1"), reservation(3, "1")}})

	if err := s.NoteAlertAttempt(ctx, queued[0].ID, "timeout"); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkAlertSent(ctx, queued[0].ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkAlertFailed(ctx, queued[1].ID, "chat not found"); err != nil {
		t.Fatal(err)
	}

	pending, _ := s.PendingAlerts(ctx, "1", 10)
	if len(pending) != 1 || pending[0].OfferID != 3 {
		t.Fatalf("pending = %+v", pending)
	}
	recent, _ := s.RecentAlerts(ctx, "1", 10)
	if len(recent) != 1 || recent[0].OfferID != 1 || recent[0].Attempts != 2 || recent[0].Payload != "" || recent[0].SentAt.IsZero() {
		t.Fatalf("recent = %+v", recent)
	}
	chats, _ := s.PendingChats(ctx)
	if len(chats) != 1 || chats[0] != "1" {
		t.Fatalf("pending chats = %v", chats)
	}

	// A failed alert still blocks the same listing for the chat: no retry storm
	// through another filter.
	again, _, _ := s.CommitPage(ctx, Commit{SearchID: a.ID, Alerts: []Reservation{reservation(2, "1")}})
	if len(again) != 0 {
		t.Error("a failed alert must not be queued again")
	}
}

func TestExpireAndFailPending(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a"})
	old := time.Now().Add(-30 * time.Minute)
	_, _, _ = s.CommitPage(ctx, Commit{SearchID: a.ID, Alerts: []Reservation{reservation(1, "1")}, Now: old})
	_, _, _ = s.CommitPage(ctx, Commit{SearchID: a.ID, Alerts: []Reservation{reservation(2, "1"), reservation(3, "2")}})

	if n, err := s.ExpireAlerts(ctx, time.Now().Add(-10*time.Minute)); err != nil || n != 1 {
		t.Fatalf("ExpireAlerts = %d, %v; want 1", n, err)
	}
	if n, _ := s.PendingCount(ctx); n != 2 {
		t.Fatalf("pending after expiry = %d, want 2", n)
	}
	if n, err := s.FailPendingOf(ctx, "2", "chat removed"); err != nil || n != 1 {
		t.Fatalf("FailPendingOf = %d, %v", n, err)
	}
	if n, _ := s.PendingCount(ctx); n != 1 {
		t.Fatalf("pending = %d, want 1", n)
	}
}

func TestDeletingAFilterCancelsItsQueuedAlerts(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "1"})
	_, _, _ = s.CommitPage(ctx, Commit{SearchID: a.ID, Alerts: []Reservation{reservation(1, "1")}})
	_, _ = s.DeleteSearch(ctx, a.ID)
	if n, _ := s.PendingCount(ctx); n != 0 {
		t.Fatalf("pending = %d after the filter was deleted", n)
	}
}

func TestAlertCountsAndHistory(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a", OwnerChatID: "1"})
	b := add(t, s, domain.Search{Name: "b", Query: "b", OwnerChatID: "1"})
	now := time.Now()

	send := func(sc domain.Search, offerID int64, at time.Time) {
		q, _, err := s.CommitPage(ctx, Commit{SearchID: sc.ID, Alerts: []Reservation{reservation(offerID, "1")}, Now: at})
		if err != nil || len(q) != 1 {
			t.Fatalf("queue %d: %v", offerID, err)
		}
		if err := s.MarkAlertSent(ctx, q[0].ID, at); err != nil {
			t.Fatal(err)
		}
	}
	send(a, 1, now.Add(-2*time.Hour))
	send(a, 2, now.Add(-time.Hour))
	send(b, 3, now.Add(-30*time.Minute))
	send(b, 4, now.Add(-3*24*time.Hour))

	total, by, err := s.AlertCounts(ctx, "1", now.Add(-24*time.Hour))
	if err != nil || total != 3 || by[a.ID] != 2 || by[b.ID] != 1 {
		t.Fatalf("day counts = %d %v, %v", total, by, err)
	}
	total, by, _ = s.AlertCounts(ctx, "1", now.Add(-7*24*time.Hour))
	if total != 4 || by[b.ID] != 2 {
		t.Fatalf("week counts = %d %v", total, by)
	}
	if total, _, _ := s.AlertCounts(ctx, "2", now.Add(-24*time.Hour)); total != 0 {
		t.Error("another chat has no alerts")
	}

	recent, _ := s.RecentAlerts(ctx, "1", 2)
	if len(recent) != 2 || recent[0].OfferID != 3 || recent[1].OfferID != 2 {
		t.Fatalf("recent = %+v, want newest first", recent)
	}
	last, _ := s.LastSentAt(ctx, "1")
	if last.Unix() != now.Add(-30*time.Minute).Unix() {
		t.Errorf("LastSentAt = %s", last)
	}
	if last, _ := s.LastSentAt(ctx, "9"); !last.IsZero() {
		t.Error("a chat without alerts has no last alert")
	}

	// The journal survives its filter.
	_, _ = s.DeleteSearch(ctx, a.ID)
	if recent, _ := s.RecentAlerts(ctx, "1", 10); len(recent) != 4 || recent[1].SearchName != "a" {
		t.Fatalf("history after deleting a filter = %+v", recent)
	}
}

func TestPruneKeepsPendingAlerts(t *testing.T) {
	s := open(t)
	a := add(t, s, domain.Search{Name: "a", Query: "a"})
	old := time.Now().Add(-60 * 24 * time.Hour)
	q, _, _ := s.CommitPage(ctx, Commit{SearchID: a.ID, Alerts: []Reservation{reservation(1, "1"), reservation(2, "1")}, Now: old})
	_ = s.MarkAlertSent(ctx, q[0].ID, old)

	_, removed, err := s.Prune(ctx, time.Time{}, time.Now().Add(-30*24*time.Hour))
	if err != nil || removed != 1 {
		t.Fatalf("Prune removed %d alerts, %v; want only the delivered one", removed, err)
	}
	if n, _ := s.PendingCount(ctx); n != 1 {
		t.Fatal("a pending alert is never pruned")
	}
}

// ---------- migration from 0.1.x ----------

func TestImportLegacySearches(t *testing.T) {
	s := open(t)
	legacy := []domain.Search{
		{Name: "phones", Query: "iphone", OwnerChatID: "1", CityName: "📍 Київ", CityID: 268, CreatedAt: 1_700_000_000},
		{Name: "", Category: 73, CatName: "Цифрові фотоапарати"},
		{Name: "flats", Category: 1758},
		{Name: "dup id", ID: "same", Query: "a"},
		{Name: "dup id 2", ID: "same", Query: "b"},
	}
	n, ran, err := s.ImportLegacySearches(ctx, legacy, time.Now())
	if err != nil || !ran || n != 5 {
		t.Fatalf("import = %d, ran=%v, %v", n, ran, err)
	}
	list, _ := s.Searches(ctx)
	if len(list) != 5 {
		t.Fatalf("stored %d filters", len(list))
	}
	ids := map[string]bool{}
	for _, sc := range list {
		ids[sc.ID] = true
	}
	if len(ids) != 5 || ids["same"] {
		t.Errorf("legacy ids must be replaced by fresh unique ones: %v", ids)
	}
	if list[0].CityName != "Київ" || list[0].CreatedAt != 1_700_000_000 || list[0].OwnerChatID != "1" {
		t.Errorf("first filter = %+v", list[0])
	}
	if list[1].Name != "Цифрові фотоапарати" {
		t.Errorf("a nameless filter gets a name: %q", list[1].Name)
	}
	if !list[2].Disabled {
		t.Error("a real-estate filter is imported paused")
	}

	// Run again, as on every later start: nothing is duplicated, and filters
	// the user deleted meanwhile do not come back.
	_, _ = s.DeleteSearch(ctx, list[0].ID)
	n, ran, err = s.ImportLegacySearches(ctx, legacy, time.Now())
	if err != nil || ran || n != 0 {
		t.Fatalf("second import = %d, ran=%v, %v; want a no-op", n, ran, err)
	}
	if list, _ := s.Searches(ctx); len(list) != 4 {
		t.Fatalf("after the second import there are %d filters, want 4", len(list))
	}
}

func TestImportLegacySeen(t *testing.T) {
	s := open(t)
	path := filepath.Join(t.TempDir(), "seen_ids.txt")
	body := strings.Join([]string{"111|500", "shared|600", "700", "", "garbage", "222|abc", "111|500"}, "\n")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	n, ran, err := s.ImportLegacySeen(ctx, path, time.Now())
	if err != nil || !ran || n != 4 {
		t.Fatalf("import = %d, ran=%v, %v; want 4 lines read", n, ran, err)
	}

	mine := SeenScopes("filter1", "111")
	other := SeenScopes("filter2", "222")
	shared := SeenScopes("filter3", "")
	ids := []int64{500, 600, 700}

	seen, _ := s.SeenSet(ctx, mine, ids)
	if !seen[500] || seen[600] || !seen[700] {
		t.Errorf("owner 111: %v, want 500 and 700", seen)
	}
	seen, _ = s.SeenSet(ctx, other, ids)
	if seen[500] || seen[600] || !seen[700] {
		t.Errorf("owner 222: %v, want only the bare id 700", seen)
	}
	seen, _ = s.SeenSet(ctx, shared, ids)
	if seen[500] || !seen[600] || !seen[700] {
		t.Errorf("shared filter: %v, want 600 and 700", seen)
	}

	// The file stays for a downgrade, and a second run imports nothing.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the legacy file must be left in place: %v", err)
	}
	if n, ran, _ := s.ImportLegacySeen(ctx, path, time.Now()); ran || n != 0 {
		t.Errorf("second import = %d, ran=%v", n, ran)
	}
	if _, ran, err := s.ImportLegacySeen(ctx, filepath.Join(t.TempDir(), "missing.txt"), time.Now()); err != nil || ran {
		t.Errorf("a missing file is not an error: ran=%v err=%v", ran, err)
	}
}

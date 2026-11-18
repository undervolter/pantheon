package checker

import (
	"strings"
	"testing"
	"time"

	"pantheon/core/config"
	"pantheon/core/domain"
)

func mkOffer(id int64, refresh string, top bool) domain.Offer {
	return domain.Offer{
		ID:              id,
		LastRefreshTime: refresh,
		Promotion:       domain.Promotion{TopAd: top},
	}
}

func TestPageFrontierHelpers(t *testing.T) {
	page := []domain.Offer{
		mkOffer(1, "2026-11-07T10:00:00+02:00", true),  // promoted, старый
		mkOffer(2, "2026-11-07T12:10:00+02:00", false), // organic
		mkOffer(3, "2026-11-07T12:05:00+02:00", false), // organic
		mkOffer(4, "2026-11-07T12:01:00+02:00", false), // organic
	}

	oldest := pageOldestRefresh(page)
	wantOldest, _ := time.Parse(time.RFC3339, "2026-11-07T12:01:00+02:00")
	if !oldest.Equal(wantOldest) {
		t.Errorf("oldest = %s, want %s", oldest, wantOldest)
	}

	newest := pageNewestRefresh(page)
	wantNewest, _ := time.Parse(time.RFC3339, "2026-11-07T12:10:00+02:00")
	if !newest.Equal(wantNewest) {
		t.Errorf("newest = %s, want %s", newest, wantNewest)
	}

	prevBefore, _ := time.Parse(time.RFC3339, "2026-11-07T12:00:00+02:00")
	if pageReached(page, prevBefore) {
		t.Error("page must NOT reach frontier 12:00 (all items are newer)")
	}

	prevAfter, _ := time.Parse(time.RFC3339, "2026-11-07T12:30:00+02:00")
	if !pageReached(page, prevAfter) {
		t.Error("page must reach frontier 12:30 (oldest 12:01 is older)")
	}
}

func TestPageReachedOnlyPromoted(t *testing.T) {
	page := []domain.Offer{
		mkOffer(1, "2026-11-07T10:00:00+02:00", true),
	}
	if !pageReached(page, time.Now()) {
		t.Error("page with only promoted items must count as reached")
	}
}

func TestSearchKeyStable(t *testing.T) {
	a := domain.Search{Name: "Любая", Category: 73, Query: "canon", CityID: 268}
	b := domain.Search{Name: "Любая", Category: 73, Query: "canon", CityID: 268}
	if searchKey(a) != searchKey(b) {
		t.Errorf("searchKey not stable: %q != %q", searchKey(a), searchKey(b))
	}
	c := a
	c.Query = "nikon"
	if searchKey(a) == searchKey(c) {
		t.Error("searchKey must differ for different queries")
	}
}

func TestAlertStartupGrace(t *testing.T) {
	cases := []struct {
		maxAge int
		want   time.Duration
	}{
		{0, 20 * time.Minute},
		{5, 20 * time.Minute},
		{20, 20 * time.Minute},
		{60, 60 * time.Minute},
	}
	for _, tc := range cases {
		if got := alertStartupGrace(tc.maxAge); got != tc.want {
			t.Errorf("alertStartupGrace(%d) = %s, want %s", tc.maxAge, got, tc.want)
		}
	}
}

func TestCheckerBackfillDisabledSetting(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := &config.Config{
		DisableBackfill: true,
	}
	chk := NewChecker(cfg, nil, nil)
	defer chk.Close()

	if !chk.isBackfillDisabled() {
		t.Error("expected isBackfillDisabled to be true")
	}

	cfgDefault := &config.Config{}
	chkDefault := NewChecker(cfgDefault, nil, nil)
	defer chkDefault.Close()

	if chkDefault.isBackfillDisabled() {
		t.Error("expected isBackfillDisabled to be false by default")
	}
}

func TestCheckerFormatAdFound(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := &config.Config{
		ANSIHyperlinks: true,
	}
	chk := NewChecker(cfg, nil, nil)
	defer chk.Close()

	now := time.Date(2026, 11, 12, 5, 58, 27, 0, time.UTC)
	offer := domain.Offer{
		ID:              940639374,
		Title:           "продаю анти кваріант",
		URL:             "https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html",
		CreatedTime:     "2026-11-12T05:40:28+02:00",
		LastRefreshTime: "2026-11-12T05:43:51+02:00",
	}

	line := chk.FormatAdFound(now, offer, "Електроніка")

	// 1. Title must NOT be present in log output
	if strings.Contains(line, "продаю анти кваріант") {
		t.Errorf("ad title should not be in log line: %s", line)
	}

	// 2. Phone label must NOT be present in log output
	if strings.Contains(line, "phone:") {
		t.Errorf("phone label should not be in log line: %s", line)
	}

	// 3. Filter name must be present in brackets
	if !strings.Contains(line, "[Електроніка]") {
		t.Errorf("expected [Електроніка] in log line: %s", line)
	}

	// 4. Hyperlink must contain OSC 8 sequence with label 'link'
	if !strings.Contains(line, "\x1b]8;;https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html\x1b\\link\x1b]8;;\x1b\\") {
		t.Errorf("expected OSC 8 hyperlink in log line: %s", line)
	}

	// Now test with ANSIHyperlinks disabled
	cfgOff := &config.Config{
		ANSIHyperlinks: false,
	}
	chkOff := NewChecker(cfgOff, nil, nil)
	defer chkOff.Close()

	lineOff := chkOff.FormatAdFound(now, offer, "Електроніка")
	if strings.Contains(lineOff, "\x1b]8;;") {
		t.Errorf("did not expect OSC 8 hyperlink when ansi_hyperlinks=false: %s", lineOff)
	}
	if !strings.Contains(lineOff, "https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html") {
		t.Errorf("expected raw URL when ansi_hyperlinks=false: %s", lineOff)
	}
	if strings.Contains(lineOff, "продаю анти кваріант") {
		t.Errorf("ad title should not be in log line: %s", lineOff)
	}
	if strings.Contains(lineOff, "phone:") {
		t.Errorf("phone label should not be in log line when ansi_hyperlinks=false: %s", lineOff)
	}
}

func TestCheckerFormatStartMessage(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := &config.Config{
		IntervalSeconds: 15,
		Searches: []domain.Search{
			{Name: "Електроніка", Query: ""},
			{Name: "Телефони", Query: "iphone", Disabled: true},
		},
	}
	chk := NewChecker(cfg, nil, nil)
	defer chk.Close()

	msg := chk.FormatStartMessage()
	// Started pantheon v0.x.x at HH:MM:SS DD:MM:YYYY, active filters: 1, interval - 15s
	if !strings.HasPrefix(msg, "Started pantheon v") {
		t.Errorf("expected 'Started pantheon v...' prefix, got: %s", msg)
	}
	if !strings.Contains(msg, ", active filters: 1, interval - 15s") {
		t.Errorf("expected active filters and interval in start message, got: %s", msg)
	}
}

func TestCheckerBanner(t *testing.T) {
	if !strings.Contains(Banner, "made by undervolter | github.com/undervolter/pantheon") {
		t.Error("banner must contain author and github repo link")
	}
	if !strings.Contains(Banner, "Want some updates? t.me/pantheonUpdates") {
		t.Error("banner must contain telegram channel link")
	}
}

func TestSeenCacheIsolatedByOwner(t *testing.T) {
	c := &Checker{SeenIDs: map[string]bool{}}

	c.saveSeenID(10, "111")
	if !c.isSeen(10, "111") {
		t.Fatal("offer must be seen by its owner")
	}
	if c.isSeen(10, "222") {
		t.Fatal("offer must NOT be seen by another owner")
	}

	// Общий фильтр не должен блокировать персональные алерты.
	c.saveSeenID(11, "")
	if !c.isSeen(11, "") {
		t.Fatal("offer must be seen by shared filter")
	}
	if c.isSeen(11, "111") {
		t.Fatal("shared filter must not block personal owner")
	}

	// Старые записи (просто ID) считаются просмотренными всеми.
	c.SeenIDs["12"] = true
	if !c.isSeen(12, "111") || !c.isSeen(12, "222") {
		t.Fatal("legacy seen entry must block all owners")
	}
}

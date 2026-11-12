package test

import (
	"testing"
	"time"

	"pantheon/core/checker"
	"pantheon/core/config"
	"pantheon/core/olx"
)

func mkOffer(id int64, refresh string, top bool) olx.Offer {
	return olx.Offer{
		ID:              id,
		LastRefreshTime: refresh,
		Promotion:       olx.Promotion{TopAd: top},
	}
}

func TestPageFrontierHelpers(t *testing.T) {
	page := []olx.Offer{
		mkOffer(1, "2026-11-07T10:00:00+02:00", true),  // promoted, старый
		mkOffer(2, "2026-11-07T12:10:00+02:00", false), // organic
		mkOffer(3, "2026-11-07T12:05:00+02:00", false), // organic
		mkOffer(4, "2026-11-07T12:01:00+02:00", false), // organic
	}

	oldest := checker.PageOldestRefresh(page)
	wantOldest, _ := time.Parse(time.RFC3339, "2026-11-07T12:01:00+02:00")
	if !oldest.Equal(wantOldest) {
		t.Errorf("oldest = %s, want %s", oldest, wantOldest)
	}

	newest := checker.PageNewestRefresh(page)
	wantNewest, _ := time.Parse(time.RFC3339, "2026-11-07T12:10:00+02:00")
	if !newest.Equal(wantNewest) {
		t.Errorf("newest = %s, want %s", newest, wantNewest)
	}

	prevBefore, _ := time.Parse(time.RFC3339, "2026-11-07T12:00:00+02:00")
	if checker.PageReached(page, prevBefore) {
		t.Error("page must NOT reach frontier 12:00 (all items are newer)")
	}

	prevAfter, _ := time.Parse(time.RFC3339, "2026-11-07T12:30:00+02:00")
	if !checker.PageReached(page, prevAfter) {
		t.Error("page must reach frontier 12:30 (oldest 12:01 is older)")
	}
}

func TestPageReachedOnlyPromoted(t *testing.T) {
	page := []olx.Offer{
		mkOffer(1, "2026-11-07T10:00:00+02:00", true),
	}
	if !checker.PageReached(page, time.Now()) {
		t.Error("page with only promoted items must count as reached")
	}
}

func TestSearchKeyStable(t *testing.T) {
	a := config.SearchConfig{Name: "Любая", Category: 73, Query: "canon", CityID: 268}
	b := config.SearchConfig{Name: "Любая", Category: 73, Query: "canon", CityID: 268}
	if checker.SearchKey(a) != checker.SearchKey(b) {
		t.Errorf("searchKey not stable: %q != %q", checker.SearchKey(a), checker.SearchKey(b))
	}
	c := a
	c.Query = "nikon"
	if checker.SearchKey(a) == checker.SearchKey(c) {
		t.Error("searchKey must differ for different queries")
	}
}

func TestAlertStartupGrace(t *testing.T) {
	cases := []struct {
		maxAge int
		want   time.Duration
	}{
		{0, 15 * time.Minute},
		{5, 15 * time.Minute},
		{20, 20 * time.Minute},
		{60, 60 * time.Minute},
	}
	for _, tc := range cases {
		if got := checker.AlertStartupGrace(tc.maxAge); got != tc.want {
			t.Errorf("alertStartupGrace(%d) = %s, want %s", tc.maxAge, got, tc.want)
		}
	}
}

func TestCheckerBackfillDisabledSetting(t *testing.T) {
	cfg := &config.Config{
		DisableBackfill: true,
	}
	chk := checker.NewChecker(cfg, nil, nil)
	defer chk.Close()

	if !chk.IsBackfillDisabled() {
		t.Error("expected IsBackfillDisabled to be true")
	}

	cfgDefault := &config.Config{}
	chkDefault := checker.NewChecker(cfgDefault, nil, nil)
	defer chkDefault.Close()

	if chkDefault.IsBackfillDisabled() {
		t.Error("expected IsBackfillDisabled to be false by default")
	}
}

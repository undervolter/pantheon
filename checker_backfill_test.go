package main

import (
	"testing"
	"time"
)

func mkOffer(id int64, refresh string, top bool) Offer {
	return Offer{
		ID:              id,
		LastRefreshTime: refresh,
		Promotion:       Promotion{TopAd: top},
	}
}

func TestPageFrontierHelpers(t *testing.T) {
	page := []Offer{
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
	page := []Offer{
		mkOffer(1, "2026-11-07T10:00:00+02:00", true),
	}
	if !pageReached(page, time.Now()) {
		t.Error("page with only promoted items must count as reached")
	}
}

func TestSearchKeyStable(t *testing.T) {
	a := SearchConfig{Name: "Любая", Category: 73, Query: "canon", CityID: 268}
	b := SearchConfig{Name: "Любая", Category: 73, Query: "canon", CityID: 268}
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
		{0, 15 * time.Minute},
		{5, 15 * time.Minute},
		{20, 20 * time.Minute},
		{60, 60 * time.Minute},
	}
	for _, tc := range cases {
		if got := alertStartupGrace(tc.maxAge); got != tc.want {
			t.Errorf("alertStartupGrace(%d) = %s, want %s", tc.maxAge, got, tc.want)
		}
	}
}

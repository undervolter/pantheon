package checker

import (
	"context"
	"time"

	"pantheon/core/domain"
	"pantheon/core/logx"
	"pantheon/core/olx"
)

// The frontier of a filter is the last_refresh_time up to which its results
// have been read. When a poll finds that even the oldest listing of the first
// page is newer than the frontier, more than a page arrived since the last
// poll, and further pages are fetched until the frontier is reached.

const maxBackfillPages = 3

func pageOldestRefresh(offers []domain.Offer) time.Time {
	var oldest time.Time
	for i := range offers {
		if offers[i].IsTopAd() {
			continue
		}
		t, err := offers[i].ParsedRefreshTime()
		if err != nil || t.IsZero() {
			continue
		}
		if oldest.IsZero() || t.Before(oldest) {
			oldest = t
		}
	}
	return oldest
}

func pageNewestRefresh(offers []domain.Offer) time.Time {
	var newest time.Time
	for i := range offers {
		if offers[i].IsTopAd() {
			continue
		}
		t, err := offers[i].ParsedRefreshTime()
		if err != nil || t.IsZero() {
			continue
		}
		if newest.IsZero() || t.After(newest) {
			newest = t
		}
	}
	return newest
}

// pageReached reports that a page goes back at least as far as prev.
func pageReached(offers []domain.Offer, prev time.Time) bool {
	oldest := pageOldestRefresh(offers)
	if oldest.IsZero() {
		return true
	}
	return !oldest.After(prev)
}

func (c *Checker) frontier(id string) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.frontiers[id]
}

func (c *Checker) setFrontier(id string, t time.Time) {
	if t.IsZero() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if cur := c.frontiers[id]; t.After(cur) {
		c.frontiers[id] = t
	}
}

// fetchWithBackfill returns the first page of a filter's results plus as many
// further pages as it takes to reach the frontier.
func (c *Checker) fetchWithBackfill(ctx context.Context, s domain.Search, limit int, floor time.Time) ([]domain.Offer, error) {
	first, err := c.fetch.FetchPage(ctx, s, limit, 0)
	if err != nil {
		return nil, err
	}
	if len(first) == 0 {
		return first, nil
	}

	prev := c.frontier(s.ID)
	// A frontier older than the filter's age floor is of no use: the filter
	// was paused meanwhile, and nothing that far back would be alerted anyway.
	if !prev.IsZero() && prev.Before(floor) {
		prev = time.Time{}
	}
	newest := pageNewestRefresh(first)
	if !c.cfg.GetBackfillEnabled() || prev.IsZero() || pageReached(first, prev) {
		c.setFrontier(s.ID, newest)
		return first, nil
	}

	all, extra, reached := first, 0, false
	for page := 1; page <= maxBackfillPages; page++ {
		// Another filter may have been refused by OLX meanwhile; the frontier
		// stays where it was and the gap is closed after the pause.
		if ctx.Err() != nil || c.now().Before(c.BlockedUntil()) {
			return all, nil
		}
		next, err := c.fetch.FetchPage(ctx, s, limit, page*limit)
		if err != nil {
			if olx.IsBlocked(err) {
				// A refusal by address must reach the caller, which pauses all
				// polling. The first page is not processed either: nothing is
				// marked seen, so nothing is lost by reading it again later.
				return nil, err
			}
			// The frontier stays where it was: the next poll tries to close
			// the same gap again instead of skipping it silently.
			logx.Warnf("'%s': backfill page %d failed (%v), will retry from the same point", s.Name, page+1, err)
			return all, nil
		}
		if len(next) == 0 {
			reached = true
			break
		}
		all = append(all, next...)
		extra += len(next)
		if t := pageNewestRefresh(next); t.After(newest) {
			newest = t
		}
		if pageReached(next, prev) {
			reached = true
			break
		}
	}
	c.setFrontier(s.ID, newest)

	logx.Infof("↕️ '%s': %d more listings loaded (the first page overflowed)", s.Name, extra)
	if !reached {
		logx.Warnf("'%s': the previous point was not reached in %d extra pages — some listings may be missed", s.Name, maxBackfillPages)
	}
	return all, nil
}

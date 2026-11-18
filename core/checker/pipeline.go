package checker

import (
	"fmt"
	"strings"
	"time"

	"pantheon/core/catalog"
	"pantheon/core/domain"
	"pantheon/core/textutil"
)

// Verdict is what the pipeline decided about one listing for one filter.
type Verdict int

const (
	Match Verdict = iota
	SkipRealEstate
	SkipPromoted
	SkipOutdated
	SkipTooOld
	SkipSpamTitle
	SkipQuery
	SkipExcluded
	SkipPrice
)

// Reason is the verdict in the words of the log.
func (v Verdict) Reason() string {
	switch v {
	case SkipRealEstate:
		return "Real estate"
	case SkipPromoted:
		return "Top ad"
	case SkipOutdated:
		return "Outdated"
	case SkipTooOld:
		return "Too old"
	case SkipSpamTitle:
		return "Spam title"
	case SkipQuery:
		return "Query mismatch"
	case SkipExcluded:
		return "Excluded word"
	case SkipPrice:
		return "Price mode"
	}
	return "Match"
}

// minGrace is how far before its age floor a filter still accepts a listing.
// OLX puts a listing into search results 10–18 minutes after it is created,
// so a listing first seen right after the bot started, or right after a
// filter was created, was already that old.
const minGrace = 20 * time.Minute

// Rules are the settings one evaluation runs under.
type Rules struct {
	Now       time.Time
	StartedAt time.Time
	// MaxAge is the oldest a listing may be when first seen; 0 — no limit.
	MaxAge time.Duration
	// SkipBumpedOld judges freshness by created_time, so an old listing that
	// was merely pushed up is not new.
	SkipBumpedOld bool
}

// Grace is the allowance below the age floor.
func (r Rules) Grace() time.Duration {
	if r.MaxAge > minGrace {
		return r.MaxAge
	}
	return minGrace
}

// Floor is the moment before which a listing is not news for this filter:
// the latest of the bot start and the filter being created, resumed or woken
// from snooze, minus the grace. Tying it to the resume keeps a filter that
// was paused for a day from pouring out everything posted meanwhile.
func (r Rules) Floor(s domain.Search) time.Time {
	floor := r.StartedAt
	for _, unix := range []int64{s.CreatedAt, s.EnabledAt, s.SnoozedUntil} {
		if unix <= 0 {
			continue
		}
		if t := time.Unix(unix, 0); t.After(floor) && !t.After(r.Now) {
			floor = t
		}
	}
	return floor.Add(-r.Grace())
}

// variantsOf returns the query variants a filter matches against.
func variantsOf(s domain.Search) []string {
	if len(s.QueryVariants) > 0 {
		return s.QueryVariants
	}
	if norm := textutil.NormalizeQuery(s.Query); norm != "" {
		return []string{norm}
	}
	return nil
}

// ExcludedBy returns the exclusion word that rules a listing out, or "".
// Matching is by word start, so "запчаст" also catches "запчастини". A
// phrase excludes only when every word of it is found.
func ExcludedBy(text string, words []string) string {
	if len(words) == 0 {
		return ""
	}
	tokens := strings.Fields(textutil.NormalizeQuery(text))
	for _, w := range words {
		parts := strings.Fields(textutil.NormalizeQuery(w))
		if len(parts) == 0 {
			continue
		}
		all := true
		for _, p := range parts {
			if !hasTokenWithPrefix(tokens, p) {
				all = false
				break
			}
		}
		if all {
			return w
		}
	}
	return ""
}

func hasTokenWithPrefix(tokens []string, prefix string) bool {
	for _, t := range tokens {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}

// Evaluate runs one listing through the filter. It is a pure function: the
// same listing, filter and rules always give the same verdict.
func Evaluate(o *domain.Offer, s domain.Search, r Rules) (Verdict, string) {
	// 1. A section the bot does not serve. This also covers filters without a
	// category, whose results mix every section of the site.
	if catalog.IsRealEstateOffer(o.Category.ID, o.Category.Type) {
		return SkipRealEstate, ""
	}

	// 2. Paid promotion: the same listings resurface for days.
	if o.IsTopAd() {
		return SkipPromoted, topAdDetail(o)
	}

	// 3. Age.
	if at := o.FreshnessTime(r.SkipBumpedOld); !at.IsZero() {
		if floor := r.Floor(s); at.Before(floor) {
			detail := fmt.Sprintf("created: %s | updated: %s", timeLabel(o.ParsedCreatedTime()), timeLabel(o.ParsedRefreshTime()))
			if r.SkipBumpedOld {
				if refreshed, err := o.ParsedRefreshTime(); err == nil && refreshed.After(floor) {
					detail = "bumped old | " + detail
				}
			}
			return SkipOutdated, detail
		}
		if r.MaxAge > 0 {
			if age := r.Now.Sub(at); age > r.MaxAge {
				return SkipTooOld, fmt.Sprintf("age: %s | limit: %dm | created: %s",
					textutil.FormatModerationDelay(age), int(r.MaxAge.Minutes()), timeLabel(o.ParsedCreatedTime()))
			}
		}
	}

	// 4. Placeholder titles.
	title := textutil.SanitizeListingText(o.Title)
	if textutil.IsLikelySpamTitle(title) {
		return SkipSpamTitle, "title: " + textutil.TruncateRunes(title, 60)
	}

	// 5. The query, checked locally: OLX search is fuzzier than the user asked for.
	text := o.SearchText()
	if !textutil.MatchesQuery(text, variantsOf(s)) {
		return SkipQuery, "query: " + s.Query
	}

	// 6. Exclusion words.
	if w := ExcludedBy(text, s.ExcludeWords); w != "" {
		return SkipExcluded, "word: " + w
	}

	// 7. Free-only mode. The request already asks OLX for free listings; this
	// keeps a paid one from slipping through should the API ever ignore it.
	if s.FreeOnly() && !o.IsFree() {
		return SkipPrice, "not free: " + o.GetPrice()
	}
	return Match, ""
}

func topAdDetail(o *domain.Offer) string {
	switch {
	case o.Promotion.TopAd:
		return "top_ad"
	case o.Promotion.Highlighted:
		return "highlighted"
	case o.Promotion.PremiumAdPage:
		return "premium_ad_page"
	}
	return "promoted"
}

func timeLabel(t time.Time, err error) string {
	if err != nil || t.IsZero() {
		return "unknown"
	}
	return t.Format("02.01.2006 15:04")
}

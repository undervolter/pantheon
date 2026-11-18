package bot

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"pantheon/core/catalog"
	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/telegram"
	"pantheon/core/textutil"
)

type (
	button = telegram.InlineKeyboardButton
	markup = telegram.InlineKeyboardMarkup
)

// step is the kind of text the bot is waiting for from a user.
type step int

const (
	stepNone step = iota
	stepQuery
	stepVariant
	stepExclude
	stepPrice
	stepFilterValue
	stepCity
)

// state is what the bot remembers about one user in one chat between
// updates. It lives in memory: a restart drops unfinished filters.
type state struct {
	// Active is set while a filter is being built.
	Active bool
	// MsgID is the message the wizard currently lives in. A wizard button
	// pressed on any other message is stale and must not touch the draft.
	MsgID int
	Step  step
	Draft domain.Search
	// Brands is the multi-select of the brand step: id -> display name.
	Brands         map[string]string
	FilterIdx      int
	FilterListPage int
	// ExclBack is where "Done" on the exclusions screen leads.
	ExclBack string
	// Notes are the warnings of an imported OLX link, shown on the review.
	Notes []string
	// districts are the districts of the city picked last: id -> name.
	districts map[int]string

	// Import holds the filters of an uploaded file until it is confirmed, and
	// ImportMsgID the message that asks for the confirmation: its buttons are
	// the only ones that may answer for this file.
	Import        []domain.Search
	ImportSkipped int
	ImportMsgID   int
}

// claims reports that the message belongs to this state: it is the screen of
// an open draft or the confirmation of a pending import.
func (st *state) claims(msgID int) bool {
	if msgID == 0 {
		return false
	}
	return (st.Active && st.MsgID == msgID) || (len(st.Import) > 0 && st.ImportMsgID == msgID)
}

type states struct {
	mu sync.Mutex
	m  map[string]*state
}

func (s *states) get(key string) *state {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		s.m = map[string]*state{}
	}
	st, ok := s.m[key]
	if !ok {
		st = &state{}
		s.m[key] = st
	}
	return st
}

func (s *states) reset(key string) *state {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		s.m = map[string]*state{}
	}
	st := &state{}
	s.m[key] = st
	return st
}

// claimedByOther reports that a message of a group is the draft or the import
// confirmation of a user other than the one asking. In a group everybody sees
// everybody's messages, but a draft belongs to the person building it.
//
// The fields of a state are only ever touched by the worker of its chat, and
// all users of a group share that one worker, so reading them here is safe.
func (s *states) claimedByOther(chatID, userID string, msgID int) bool {
	own := stateKey(chatID, userID)
	prefix := chatID + ":"
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, st := range s.m {
		if key != own && strings.HasPrefix(key, prefix) && st.claims(msgID) {
			return true
		}
	}
	return false
}

// stateKey separates users inside a group: two people building filters in
// the same chat must not share a draft.
func stateKey(chatID, userID string) string {
	if chatID == userID || userID == "" {
		return chatID
	}
	return chatID + ":" + userID
}

// ---------- keyboards ----------

func grid(buttons []button, perRow int) [][]button {
	var rows [][]button
	for len(buttons) > 0 {
		n := min(perRow, len(buttons))
		rows = append(rows, buttons[:n])
		buttons = buttons[n:]
	}
	return rows
}

// pageBounds clamps a page number and returns the slice bounds of that page.
func pageBounds(total, size, page int) (start, end, current, pages int) {
	pages = (total + size - 1) / size
	if pages < 1 {
		pages = 1
	}
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start = page * size
	end = min(start+size, total)
	return start, end, page, pages
}

// pagerRow is "‹ 2 / 5 ›"; nil when there is a single page.
func pagerRow(page, pages int, data func(page int) string) []button {
	if pages <= 1 {
		return nil
	}
	var row []button
	if page > 0 {
		row = append(row, telegram.Btn("‹", data(page-1)))
	}
	row = append(row, telegram.Label(fmt.Sprintf("%d / %d", page+1, pages)))
	if page < pages-1 {
		row = append(row, telegram.Btn("›", data(page+1)))
	}
	return row
}

// toggleBtn shows a switch: green with a check mark when on.
func toggleBtn(on bool, text, data string) button {
	if on {
		return telegram.Success(text+" ✓", data)
	}
	return telegram.Btn(text, data)
}

func btnBack(lang i18n.Lang, data string) button {
	return telegram.Btn(i18n.T(lang, "btn.back"), data)
}

func btnCancel(lang i18n.Lang) button {
	return telegram.Btn(i18n.T(lang, "btn.cancel"), "menu_main")
}

// itemTag is a short stable handle for free text (a query variant, an
// exclusion word) that has to travel in callback data, which is limited to
// 64 bytes.
func itemTag(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%08x", h.Sum32())
}

func removeTagged(list []string, tag string) []string {
	out := list[:0:0]
	for _, v := range list {
		if itemTag(v) != tag {
			out = append(out, v)
		}
	}
	return out
}

// ---------- formatting ----------

// money prints an amount with thin thousands separators: 25 000.
func money(v float64) string {
	n := strconv.FormatFloat(v, 'f', -1, 64)
	intPart, frac, _ := strings.Cut(n, ".")
	var b strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	if frac != "" {
		b.WriteString("." + frac)
	}
	return b.String()
}

func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// rangeText prints a numeric range in words: "от 2000 до 2010".
func rangeText(lang i18n.Lang, from, to string) string {
	switch {
	case from != "" && to != "":
		return i18n.T(lang, "rng.both", from, to)
	case from != "":
		return i18n.T(lang, "rng.from", from)
	case to != "":
		return i18n.T(lang, "rng.to", to)
	}
	return ""
}

func filterRangeText(lang i18n.Lang, r domain.FilterRange, unit string) string {
	from, to := "", ""
	if r.From > 0 {
		from = trimFloat(r.From)
	}
	if r.To > 0 {
		to = trimFloat(r.To)
	}
	text := rangeText(lang, from, to)
	if text != "" && unit != "" {
		text += " " + unit
	}
	return text
}

// priceText is the price of a filter in words; short drops the unit-less
// "any" in favour of a dash for tables.
func priceText(lang i18n.Lang, s domain.Search) string {
	if s.FreeOnly() {
		return i18n.T(lang, "sum.price_free")
	}
	from, to := "", ""
	if s.MinPrice > 0 {
		from = money(s.MinPrice)
	}
	if s.MaxPrice > 0 {
		to = money(s.MaxPrice)
	}
	if text := rangeText(lang, from, to); text != "" {
		return text + " " + i18n.T(lang, "cur.uah")
	}
	return i18n.T(lang, "sum.price_any")
}

func priceCell(lang i18n.Lang, s domain.Search) string {
	if s.FreeOnly() {
		return i18n.T(lang, "sum.price_free_tag")
	}
	if s.MinPrice <= 0 && s.MaxPrice <= 0 {
		return "—"
	}
	return priceText(lang, s)
}

// whereText is the location of a filter. Names are stored bare; the icon is
// presentation and is added here.
func whereText(lang i18n.Lang, s domain.Search) string {
	switch {
	case s.CityID > 0:
		name := s.CityName
		if name == "" {
			name = i18n.T(lang, "sum.city_n", s.CityID)
		}
		if s.DistrictName != "" {
			name = i18n.T(lang, "sum.district", name, s.DistrictName)
		}
		return "📍 " + esc(name)
	case s.RegionID > 0:
		name := s.CityName
		if name == "" {
			name = i18n.T(lang, "sum.region_n", s.RegionID)
		}
		return "🌍 " + esc(name)
	}
	return i18n.T(lang, "common.all_ukraine")
}

func categoryText(lang i18n.Lang, s domain.Search) string {
	if s.Category <= 0 {
		return i18n.T(lang, "sum.any_category")
	}
	if crumb := catalog.CategoryBreadcrumb(s.Category); crumb != "" {
		return esc(crumb)
	}
	return esc(s.CatName)
}

// displayName is the name a filter is shown under.
func displayName(lang i18n.Lang, s domain.Search) string {
	switch {
	case strings.TrimSpace(s.Name) != "":
		return s.Name
	case s.Query != "":
		return s.Query
	case s.Brands() != "":
		return s.Brands()
	case s.CatName != "":
		return s.CatName
	}
	return i18n.T(lang, "sum.default_name")
}

// extraFilters returns the attribute filters of a category that are set in
// the wizard's "extra filters" screen: everything except the parameters with
// their own step (price, seller type) and the brand dictionary.
func extraFilters(catID int, brandParam string) []catalog.FilterDef {
	all := catalog.FiltersForCategory(catID)
	out := make([]catalog.FilterDef, 0, len(all))
	for _, d := range all {
		if catalog.IsReservedFilterKey(d.Key) || !domain.ValidFilterKey(d.Key) {
			continue
		}
		if brandParam != "" && d.Key == brandParam {
			continue
		}
		out = append(out, d)
	}
	return out
}

func extraSetCount(s domain.Search, def catalog.FilterDef) int {
	if n := len(s.Filters[def.Key]); n > 0 {
		return n
	}
	if r, ok := s.Ranges[def.Key]; ok && (r.From > 0 || r.To > 0) {
		return 1
	}
	return 0
}

func extrasSet(s domain.Search) int {
	n := 0
	for _, d := range extraFilters(s.Category, s.BrandParam) {
		if extraSetCount(s, d) > 0 {
			n++
		}
	}
	return n
}

// extrasText is the chosen attribute filters in one line, escaped.
func extrasText(lang i18n.Lang, s domain.Search) string {
	var parts []string
	known := map[string]bool{}
	for _, d := range catalog.FiltersForCategory(s.Category) {
		known[d.Key] = true
		if vals := s.Filters[d.Key]; len(vals) > 0 {
			labels := make([]string, 0, len(vals))
			if strings.HasPrefix(d.Key, "filter_bool_") {
				labels = append(labels, i18n.T(lang, "sum.bool_on"))
			} else {
				for _, v := range vals {
					label := v
					for _, fv := range d.Values {
						if fv.Value == v {
							label = fv.Label
							break
						}
					}
					labels = append(labels, label)
				}
			}
			parts = append(parts, d.Label+": "+strings.Join(labels, ", "))
		}
		if r, ok := s.Ranges[d.Key]; ok && (r.From > 0 || r.To > 0) {
			parts = append(parts, d.Label+": "+filterRangeText(lang, r, d.Unit))
		}
	}
	// A filter that came from a link or a file can carry keys the category
	// metadata does not describe; they still apply, so they are still shown.
	for key, vals := range s.Filters {
		if !known[key] && len(vals) > 0 {
			parts = append(parts, key+": "+strings.Join(vals, ", "))
		}
	}
	for key, r := range s.Ranges {
		if !known[key] && (r.From > 0 || r.To > 0) {
			parts = append(parts, key+": "+filterRangeText(lang, r, ""))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	sortStrings(parts)
	return esc(textutil.TruncateRunes(strings.Join(parts, "; "), 400))
}

func sortStrings(list []string) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j] < list[j-1]; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

// summary adds the parameters of a filter to a page as key/value rows.
func summary(p *page, lang i18n.Lang, s domain.Search) *page {
	query := "<i>" + i18n.T(lang, "sum.any_listing") + "</i>"
	if s.Query != "" {
		query = esc(s.Query)
	}
	p.kv(i18n.T(lang, "sum.query"), query)

	if len(s.QueryVariants) > 0 {
		own := textutil.QueryVariants(s.Query)
		if strings.Join(own, "|") != strings.Join(s.QueryVariants, "|") {
			p.kv(i18n.T(lang, "sum.variants"), esc(textutil.TruncateRunes(strings.Join(s.QueryVariants, ", "), 180)))
		}
	}
	p.kv(i18n.T(lang, "sum.category"), categoryText(lang, s))
	p.kv(i18n.T(lang, "sum.brand"), esc(textutil.TruncateRunes(s.Brands(), 160)))
	p.kv(i18n.T(lang, "sum.price"), priceText(lang, s))
	p.kv(i18n.T(lang, "sum.where"), whereText(lang, s))
	switch s.OwnerType {
	case "private":
		p.kv(i18n.T(lang, "sum.seller"), i18n.T(lang, "sum.seller_private"))
	case "business":
		p.kv(i18n.T(lang, "sum.seller"), i18n.T(lang, "sum.seller_business"))
	}
	p.kv(i18n.T(lang, "sum.extras"), extrasText(lang, s))
	if len(s.ExcludeWords) > 0 {
		p.kv(i18n.T(lang, "sum.exclude"), esc(textutil.TruncateRunes(strings.Join(s.ExcludeWords, ", "), 200)))
	}
	return p
}

// ago prints a duration as its largest unit: "5 мин", "3 ч".
func ago(lang i18n.Lang, d time.Duration) string {
	switch {
	case d < time.Minute:
		return i18n.T(lang, "dur.s", int(d.Seconds()))
	case d < time.Hour:
		return i18n.T(lang, "dur.m", int(d.Minutes()))
	case d < 48*time.Hour:
		return i18n.T(lang, "dur.h", int(d.Hours()))
	}
	return i18n.T(lang, "dur.d", int(d.Hours()/24))
}

// parsePriceRange reads what a user typed for a price or a numeric filter:
// "5000-20000", "5 000 – 20 000", "от 5000 до 20000", "до 20000", "от 5000",
// "5000" (an upper bound) and "0".
func parsePriceRange(raw string) (minV, maxV float64, ok bool) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.NewReplacer("–", "-", "—", "-", "−", "-", " ", " ").Replace(s)
	for _, w := range []string{"гривень", "гривен", "грн", "uah", "₴"} {
		s = strings.ReplaceAll(s, w, " ")
	}

	parseNum := func(part string) (float64, bool) {
		var b strings.Builder
		for _, r := range part {
			switch {
			case unicode.IsDigit(r) || r == '.':
				b.WriteRune(r)
			case r == ',':
				b.WriteRune('.')
			}
		}
		if b.Len() == 0 {
			return 0, false
		}
		v, err := strconv.ParseFloat(b.String(), 64)
		if err != nil || !domain.ValidNumber(v) {
			return 0, false
		}
		return v, true
	}

	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		lo, lok := parseNum(parts[0])
		hi, hik := parseNum(parts[1])
		if !lok && !hik {
			return 0, 0, false
		}
		return lo, hi, true
	}

	hasFrom := strings.Contains(s, "от") || strings.Contains(s, "від") || strings.Contains(s, "from")
	hasTo := strings.Contains(s, "до") || strings.Contains(s, "to")

	s = strings.ReplaceAll(s, " ", "")
	var nums []float64
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		if v, err := strconv.ParseFloat(cur.String(), 64); err == nil && domain.ValidNumber(v) {
			nums = append(nums, v)
		}
		cur.Reset()
	}
	for _, r := range s {
		switch {
		case unicode.IsDigit(r) || r == '.':
			cur.WriteRune(r)
		case r == ',':
			cur.WriteRune('.')
		default:
			flush()
		}
	}
	flush()

	switch len(nums) {
	case 0:
		return 0, 0, false
	case 1:
		if hasFrom && !hasTo {
			return nums[0], 0, true
		}
		return 0, nums[0], true
	}
	return nums[0], nums[1], true
}

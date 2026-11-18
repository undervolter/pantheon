package olx

import (
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"pantheon/core/catalog"
	"pantheon/core/domain"
	"pantheon/core/textutil"
)

// Errors of ParseSearchURL.
var (
	// ErrNotOLX — the link does not lead to olx.ua.
	ErrNotOLX = errors.New("not an olx.ua link")
	// ErrListingURL — the link opens one listing, not a search.
	ErrListingURL = errors.New("a link to a single listing, not to a search")
)

// Parsed is what a search URL of the site turned into.
type Parsed struct {
	Search domain.Search

	// Slugs are path segments that are neither a category nor the query:
	// candidates for a city or a region ("kiev", "ko"). The parser stays
	// pure; whoever has a geo index resolves them.
	Slugs []string

	// ExchangeDropped — the link asked for «Обмін» listings, which the bot
	// has no mode for.
	ExchangeDropped bool
	// CurrencyDropped holds the currency of price bounds that were dropped:
	// the bot filters in UAH and has no rate to convert "up to $100" with.
	CurrencyDropped string
}

// IsOLXURL reports whether text is a link to the OLX.ua site. The host is
// parsed, not searched for: "https://evil.example/?x=olx.ua" is not one.
func IsOLXURL(text string) bool {
	u, err := url.Parse(strings.TrimSpace(text))
	if err != nil {
		return false
	}
	return isOLXHost(u)
}

func isOLXHost(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "olx.ua" || strings.HasSuffix(host, ".olx.ua")
}

// Path segments of the site that carry no filter.
var pathNoise = map[string]bool{"": true, "uk": true, "ru": true, "d": true, "list": true}

// ParseSearchURL turns a search page URL of olx.ua into a filter.
func ParseSearchURL(rawURL string) (*Parsed, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !isOLXHost(u) {
		return nil, ErrNotOLX
	}
	if offerHashRe.MatchString(u.Path) || strings.Contains(u.Path, "/obyavlenie/") {
		return nil, ErrListingURL
	}

	out := &Parsed{}
	sc := &out.Search
	q := u.Query()

	if query := firstNonEmpty(q.Get("search[query]"), q.Get("q")); query != "" {
		sc.Query = query
	}
	if id := atoi(firstNonEmpty(q.Get("search[city_id]"), q.Get("city_id"))); id > 0 {
		sc.CityID = id
		if city := catalog.FindCity(id); city != nil {
			sc.CityName = city.Name
			sc.RegionID = city.RegionID
		}
	}
	if id := atoi(firstNonEmpty(q.Get("search[region_id]"), q.Get("region_id"))); id > 0 {
		sc.RegionID = id
	}
	if id := atoi(firstNonEmpty(q.Get("search[district_id]"), q.Get("district_id"))); id > 0 {
		sc.DistrictID = id
	}
	if id := atoi(firstNonEmpty(q.Get("search[category_id]"), q.Get("category_id"))); id > 0 {
		sc.Category = id
	}
	switch firstNonEmpty(q.Get("owner_type"), q.Get("search[owner_type]"), q.Get("private_business"), q.Get("search[private_business]")) {
	case "private":
		sc.OwnerType = "private"
	case "business":
		sc.OwnerType = "business"
	}

	for _, part := range strings.Split(strings.Trim(u.Path, "/"), "/") {
		switch {
		case pathNoise[part]:
		case strings.HasPrefix(part, "q-"):
			sc.Query = strings.ReplaceAll(strings.TrimPrefix(part, "q-"), "-", " ")
		default:
			if cat := catalog.FindCategoryBySlug(part); cat != nil {
				sc.Category = cat.ID
			} else {
				out.Slugs = append(out.Slugs, part)
			}
		}
	}
	if cat := catalog.FindCategory(sc.Category); cat != nil {
		sc.CatName = cat.Name
	}

	// Every search[...] parameter in one pass. Keys are sorted so the result
	// does not depend on map iteration order.
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	enums := map[string][]string{}
	wantFree := false
	for _, k := range keys {
		base, sub := parseFilterParam(k)
		vals := q[k]
		switch {
		case base == "filter_float_price":
			for _, v := range vals {
				f, ok := parseNumber(v)
				if !ok {
					continue
				}
				if sub == "to" {
					sc.MaxPrice = f
				} else {
					sc.MinPrice = f
				}
			}
		case base == "filter_enum_price":
			for _, v := range vals {
				switch strings.TrimSpace(v) {
				case "free":
					wantFree = true
				case "exchange":
					out.ExchangeDropped = true
				}
			}
		case !domain.ValidFilterKey(base):
			// search[order], page, view and anything else that is not a filter.
		case strings.HasPrefix(base, "filter_float_") || strings.HasPrefix(base, "filter_int_"):
			if sc.Ranges == nil {
				sc.Ranges = map[string]domain.FilterRange{}
			}
			r := sc.Ranges[base]
			for _, v := range vals {
				f, ok := parseNumber(v)
				if !ok {
					continue
				}
				if sub == "to" {
					r.To = f
				} else {
					r.From = f
				}
			}
			sc.Ranges[base] = r
		case strings.HasPrefix(base, "filter_bool_"):
			if len(vals) > 0 && strings.TrimSpace(vals[0]) != "" && vals[0] != "0" {
				if sc.Filters == nil {
					sc.Filters = map[string][]string{}
				}
				sc.Filters[base] = []string{"1"}
			}
		case strings.HasPrefix(base, "filter_enum_"):
			for _, v := range vals {
				if v = strings.TrimSpace(v); v != "" && !textutil.StringInSlice(enums[base], v) {
					enums[base] = append(enums[base], v)
				}
			}
		}
	}

	if currency := strings.ToUpper(strings.TrimSpace(q.Get("currency"))); currency != "" && currency != "UAH" &&
		(sc.MinPrice > 0 || sc.MaxPrice > 0) {
		out.CurrencyDropped = currency
		sc.MinPrice, sc.MaxPrice = 0, 0
	}
	if wantFree {
		sc.PriceMode = domain.PriceFree
	}

	// Only a parameter the catalog knows as a brand dictionary becomes the
	// brand; every other enum stays an ordinary attribute filter.
	brandParam := ""
	if info := catalog.GetCategoryBrands(sc.Category); info != nil {
		if _, ok := enums[info.ParamName]; ok {
			brandParam = info.ParamName
		}
	}
	if brandParam == "" {
		var known []string
		for p := range enums {
			if catalog.IsBrandParam(p) {
				known = append(known, p)
			}
		}
		sort.Strings(known)
		if len(known) > 0 {
			brandParam = known[0]
		}
	}
	if brandParam != "" {
		applyBrand(sc, brandParam, enums[brandParam])
		delete(enums, brandParam)
	}
	for key, vals := range enums {
		if sc.Filters == nil {
			sc.Filters = map[string][]string{}
		}
		sc.Filters[key] = vals
	}

	if sc.Query != "" {
		sc.QueryVariants = textutil.QueryVariants(sc.Query)
	}
	switch {
	case sc.Query != "":
		sc.Name = sc.Query
	case sc.Brands() != "":
		sc.Name = sc.Brands()
	default:
		sc.Name = sc.CatName
	}
	sc.Normalize()
	return out, nil
}

// applyBrand fills the brand fields in catalog order, with display names.
func applyBrand(sc *domain.Search, param string, values []string) {
	order := map[string]int{}
	names := map[string]string{}
	for _, info := range catalog.CategoryBrands {
		if info.ParamName != param {
			continue
		}
		for _, b := range info.Brands {
			if _, ok := order[b.ID]; !ok {
				order[b.ID] = len(order)
				names[b.ID] = b.Name
			}
		}
	}
	// A category's own dictionary wins over the merged order, so the list
	// reads the way the wizard shows it.
	if info := catalog.GetCategoryBrands(sc.Category); info != nil && info.ParamName == param {
		order = map[string]int{}
		for i, b := range info.Brands {
			order[b.ID] = i
			names[b.ID] = b.Name
		}
	}

	values = append([]string(nil), values...)
	sort.SliceStable(values, func(i, j int) bool {
		oi, iok := order[values[i]]
		oj, jok := order[values[j]]
		if iok && jok {
			return oi < oj
		}
		if iok != jok {
			return iok
		}
		return values[i] < values[j]
	})

	sc.BrandParam = param
	sc.BrandValues = values
	sc.BrandNames = nil
	for _, id := range values {
		name := names[id]
		if name == "" {
			name = id
		}
		sc.BrandNames = append(sc.BrandNames, name)
	}
	sc.BrandValue, sc.BrandName = "", ""
	if len(values) == 1 {
		sc.BrandValue, sc.BrandName = values[0], sc.BrandNames[0]
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func parseNumber(s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(s), ",", "."), 64)
	return f, err == nil
}

// parseFilterParam splits a query key into the filter name and its
// sub-key (a range bound or a list index). The site writes range bounds as
// search[filter_float_year:from]; the API and older links use
// filter_float_year:from, search[filter_float_year][from] and
// filter_enum_color[0].
func parseFilterParam(raw string) (base, sub string) {
	s := raw
	if strings.HasPrefix(s, "search[") {
		s = s[len("search["):]
		i := strings.Index(s, "]")
		if i < 0 {
			return s, ""
		}
		base = s[:i]
		rest := s[i+1:]
		if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
			sub = rest[1 : len(rest)-1]
		}
		// The bound can sit inside the first pair of brackets.
		if name, bound, found := strings.Cut(base, ":"); found {
			base, sub = name, bound
		}
		return base, sub
	}
	if i := strings.Index(s, "["); i > 0 && strings.HasSuffix(s, "]") {
		return s[:i], s[i+1 : len(s)-1]
	}
	if name, bound, found := strings.Cut(s, ":"); found && name != "" {
		return name, bound
	}
	return s, ""
}

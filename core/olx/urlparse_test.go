package olx

import (
	"errors"
	"testing"

	"pantheon/core/domain"
)

func mustParse(t *testing.T, raw string) *Parsed {
	t.Helper()
	p, err := ParseSearchURL(raw)
	if err != nil {
		t.Fatalf("ParseSearchURL(%s): %v", raw, err)
	}
	return p
}

func TestParseSearchURL(t *testing.T) {
	sc := mustParse(t, "https://www.olx.ua/d/uk/elektronika/telefony-i-aksesuary/q-iphone-15/?search[filter_float_price:to]=30000&search[city_id]=268").Search
	if sc.Query != "iphone 15" || sc.Name != "iphone 15" {
		t.Errorf("query/name = %q / %q", sc.Query, sc.Name)
	}
	if sc.Category != 44 || sc.CatName == "" {
		t.Errorf("category = %d / %q", sc.Category, sc.CatName)
	}
	if sc.CityID != 268 || sc.RegionID != 25 || sc.CityName != "Київ" {
		t.Errorf("city = %d / region %d / %q", sc.CityID, sc.RegionID, sc.CityName)
	}
	if sc.MaxPrice != 30000 || sc.MinPrice != 0 {
		t.Errorf("price = %v..%v", sc.MinPrice, sc.MaxPrice)
	}
	// The price bound must not leak into the generic ranges under a garbage key.
	if len(sc.Ranges) != 0 {
		t.Errorf("ranges = %v, want none", sc.Ranges)
	}
	if err := sc.Validate(); err != nil {
		t.Errorf("a parsed link must be a valid filter: %v", err)
	}
}

// The format the site really produces: the bound sits inside the brackets,
// percent-encoded. It used to be stored under the key
// "filter_float_motor_year:from" and silently never applied.
func TestParseRangesInSiteFormat(t *testing.T) {
	raw := "https://www.olx.ua/uk/transport/legkovye-avtomobili/?currency=UAH" +
		"&search%5Bfilter_float_motor_year%3Afrom%5D=2015" +
		"&search%5Bfilter_float_motor_year%3Ato%5D=2020" +
		"&search%5Bfilter_float_price%3Afrom%5D=1000" +
		"&search%5Bfilter_float_price%3Ato%5D=10000"
	sc := mustParse(t, raw).Search

	r, ok := sc.Ranges["filter_float_motor_year"]
	if !ok || r.From != 2015 || r.To != 2020 {
		t.Fatalf("year range = %+v (ok=%v), ranges = %v", r, ok, sc.Ranges)
	}
	if len(sc.Ranges) != 1 {
		t.Errorf("ranges = %v, want only the year", sc.Ranges)
	}
	if sc.MinPrice != 1000 || sc.MaxPrice != 10000 {
		t.Errorf("price = %v..%v", sc.MinPrice, sc.MaxPrice)
	}
	if err := sc.Validate(); err != nil {
		t.Fatalf("parsed filter must validate: %v", err)
	}

	q := BuildOffersQuery(sc, 20, 0)
	if q.Get("filter_float_motor_year:from") != "2015" || q.Get("filter_float_motor_year:to") != "2020" {
		t.Errorf("request lacks the year bounds: %s", q.Encode())
	}
}

func TestParseFilterParam(t *testing.T) {
	cases := []struct{ in, base, sub string }{
		{"search[filter_enum_color][0]", "filter_enum_color", "0"},
		{"filter_enum_color[3]", "filter_enum_color", "3"},
		{"search[filter_float_motor_year][from]", "filter_float_motor_year", "from"},
		{"search[filter_float_motor_year:from]", "filter_float_motor_year", "from"},
		{"search[filter_float_price:to]", "filter_float_price", "to"},
		{"filter_float_motor_year:to", "filter_float_motor_year", "to"},
		{"filter_enum_color", "filter_enum_color", ""},
		{"search[order]", "order", ""},
	}
	for _, c := range cases {
		base, sub := parseFilterParam(c.in)
		if base != c.base || sub != c.sub {
			t.Errorf("parseFilterParam(%q) = (%q, %q), want (%q, %q)", c.in, base, sub, c.base, c.sub)
		}
	}
}

func TestParseEnumAndBracketRanges(t *testing.T) {
	u := "https://www.olx.ua/uk/transport/legkovye-avtomobili/vaz/?search%5Bfilter_enum_color%5D%5B0%5D=4&search%5Bfilter_float_motor_year%5D%5Bfrom%5D=2000&search%5Bfilter_float_motor_year%5D%5Bto%5D=2010&search%5Border%5D=created_at:desc&page=2"
	sc := mustParse(t, u).Search
	if colors := sc.Filters["filter_enum_color"]; len(colors) != 1 || colors[0] != "4" {
		t.Errorf("color filter = %v, want [4]", colors)
	}
	if r := sc.Ranges["filter_float_motor_year"]; r.From != 2000 || r.To != 2010 {
		t.Errorf("year range = %+v, want 2000..2010", r)
	}
	if len(sc.Filters) != 1 || len(sc.Ranges) != 1 {
		t.Errorf("only real filters must be kept: filters=%v ranges=%v", sc.Filters, sc.Ranges)
	}
}

func TestParseMultiBrand(t *testing.T) {
	raw := "https://www.olx.ua/uk/elektronika/foto-video/tsifrovye-fotoapparaty/" +
		"?search[filter_enum_digital_camera_manufacturers][0]=2591" +
		"&search[filter_enum_digital_camera_manufacturers][1]=2580" +
		"&search[filter_enum_state][0]=used"
	sc := mustParse(t, raw).Search
	if sc.Category != 73 {
		t.Fatalf("category = %d, want 73", sc.Category)
	}
	if sc.BrandParam != "filter_enum_digital_camera_manufacturers" {
		t.Errorf("brand param = %q", sc.BrandParam)
	}
	if len(sc.BrandValues) != 2 || sc.BrandValues[0] != "2580" || sc.BrandValues[1] != "2591" {
		t.Fatalf("brand values = %#v, want catalog order", sc.BrandValues)
	}
	if len(sc.BrandNames) != 2 || sc.BrandNames[0] != "Canon" || sc.BrandNames[1] != "Nikon" {
		t.Fatalf("brand names = %#v", sc.BrandNames)
	}
	// The brand must not be duplicated among the ordinary filters.
	if _, dup := sc.Filters[sc.BrandParam]; dup {
		t.Error("brand param is duplicated in Filters")
	}
	if got := sc.Filters["filter_enum_state"]; len(got) != 1 || got[0] != "used" {
		t.Errorf("state filter = %v", got)
	}
}

// With no brand dictionary in play, an ordinary enum (colour, state) used to
// be picked as "the brand" by random map iteration — a different one on each run.
func TestOrdinaryEnumNeverBecomesTheBrand(t *testing.T) {
	raw := "https://www.olx.ua/uk/list/?search%5Bfilter_enum_state%5D%5B0%5D=used&search%5Bfilter_enum_color%5D%5B0%5D=black"
	for i := 0; i < 50; i++ {
		sc := mustParse(t, raw).Search
		if sc.BrandParam != "" || len(sc.BrandValues) != 0 {
			t.Fatalf("run %d: brand = %q %v, want none", i, sc.BrandParam, sc.BrandValues)
		}
		if len(sc.Filters) != 2 {
			t.Fatalf("run %d: filters = %v", i, sc.Filters)
		}
	}
}

func TestParseFreeAndExchange(t *testing.T) {
	p := mustParse(t, "https://www.olx.ua/uk/elektronika/?search%5Bfilter_enum_price%5D%5B0%5D=free")
	if !p.Search.FreeOnly() || p.ExchangeDropped {
		t.Errorf("free link: mode=%q exchange=%v", p.Search.PriceMode, p.ExchangeDropped)
	}
	if _, leaked := p.Search.Filters["filter_enum_price"]; leaked {
		t.Error("the price checkbox must not become an attribute filter")
	}

	p = mustParse(t, "https://www.olx.ua/uk/elektronika/?search%5Bfilter_enum_price%5D%5B0%5D=exchange")
	if p.Search.FreeOnly() || !p.ExchangeDropped {
		t.Errorf("exchange link: mode=%q exchange=%v", p.Search.PriceMode, p.ExchangeDropped)
	}

	p = mustParse(t, "https://www.olx.ua/uk/elektronika/?search%5Bfilter_enum_price%5D%5B0%5D=free&search%5Bfilter_enum_price%5D%5B1%5D=exchange&search%5Bfilter_float_price%3Ato%5D=500")
	if !p.Search.FreeOnly() || !p.ExchangeDropped || p.Search.MaxPrice != 0 {
		t.Errorf("free+exchange link: mode=%q exchange=%v max=%v", p.Search.PriceMode, p.ExchangeDropped, p.Search.MaxPrice)
	}
}

func TestParseForeignCurrencyDropsPriceBounds(t *testing.T) {
	p := mustParse(t, "https://www.olx.ua/uk/transport/legkovye-avtomobili/?currency=USD&search%5Bfilter_float_price%3Ato%5D=5000")
	if p.CurrencyDropped != "USD" || p.Search.MaxPrice != 0 {
		t.Errorf("USD bounds must be dropped, got currency=%q max=%v", p.CurrencyDropped, p.Search.MaxPrice)
	}
	p = mustParse(t, "https://www.olx.ua/uk/transport/legkovye-avtomobili/?currency=USD")
	if p.CurrencyDropped != "" {
		t.Error("a currency without price bounds is nothing to warn about")
	}
}

func TestParseKeepsLocationSlugsForTheCaller(t *testing.T) {
	p := mustParse(t, "https://www.olx.ua/uk/elektronika/telefony-i-aksesuary/kiev/q-iphone/")
	if len(p.Slugs) != 1 || p.Slugs[0] != "kiev" {
		t.Errorf("slugs = %v, want [kiev]", p.Slugs)
	}
	if p.Search.Category != 44 || p.Search.Query != "iphone" {
		t.Errorf("category/query = %d / %q", p.Search.Category, p.Search.Query)
	}
}

func TestParseRejectsListingAndForeignLinks(t *testing.T) {
	// A single listing used to become a filter named «Поиск по ссылке» with no
	// condition at all.
	for _, raw := range []string{
		"https://www.olx.ua/d/uk/obyavlenie/prodam-iphone-ID11HsSR.html",
		"https://www.olx.ua/d/obyavlenie/prodam-iphone-ID11HsSR.html?reason=x",
	} {
		if _, err := ParseSearchURL(raw); !errors.Is(err, ErrListingURL) {
			t.Errorf("%s: err = %v, want ErrListingURL", raw, err)
		}
	}
	for _, raw := range []string{
		"https://evil.example/?x=olx.ua",
		"https://olx.ua.evil.example/uk/list/",
		"https://notolx.ua/uk/list/",
		"ftp://www.olx.ua/uk/list/",
		"javascript:alert(1)//olx.ua",
		"olx.ua",
	} {
		if _, err := ParseSearchURL(raw); !errors.Is(err, ErrNotOLX) {
			t.Errorf("%s: err = %v, want ErrNotOLX", raw, err)
		}
		if IsOLXURL(raw) {
			t.Errorf("IsOLXURL(%s) = true", raw)
		}
	}
	for _, raw := range []string{"https://www.olx.ua/uk/list/", "https://m.olx.ua/uk/elektronika/", "http://olx.ua/"} {
		if !IsOLXURL(raw) {
			t.Errorf("IsOLXURL(%s) = false", raw)
		}
	}
}

func TestParsedRealEstateLinkFailsValidation(t *testing.T) {
	p := mustParse(t, "https://www.olx.ua/uk/nedvizhimost/kvartiry/prodazha-kvartir/kiev/")
	if err := p.Search.Validate(); !errors.Is(err, domain.ErrRealEstate) {
		t.Errorf("err = %v, want ErrRealEstate (category %d)", err, p.Search.Category)
	}
}

func TestParsedBareListFailsValidation(t *testing.T) {
	p := mustParse(t, "https://www.olx.ua/uk/list/")
	if err := p.Search.Validate(); !errors.Is(err, domain.ErrNoCondition) {
		t.Errorf("err = %v, want ErrNoCondition", err)
	}
}

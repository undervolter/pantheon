package domain

import (
	"errors"
	"testing"
	"time"
)

func TestValidateRequiresACondition(t *testing.T) {
	// The exact filter a stale "Launch" button used to create: no query, no
	// category — every new listing on OLX.
	if err := (Search{Name: "Новый поиск"}).Validate(); !errors.Is(err, ErrNoCondition) {
		t.Fatalf("empty filter: err = %v, want ErrNoCondition", err)
	}
	for _, ok := range []Search{
		{Query: "iphone"},
		{Category: 44},
		{PriceMode: PriceFree},
	} {
		if err := ok.Validate(); err != nil {
			t.Errorf("%+v must be valid, got %v", ok, err)
		}
	}
}

func TestValidateRejectsRealEstate(t *testing.T) {
	for _, cat := range []int{1, 1758, 1760, 3690} {
		if err := (Search{Category: cat, Query: "x"}).Validate(); !errors.Is(err, ErrRealEstate) {
			t.Errorf("category %d: err = %v, want ErrRealEstate", cat, err)
		}
	}
	// Daily rentals are a separate OLX section, not real estate.
	if err := (Search{Category: 3711}).Validate(); err != nil {
		t.Errorf("category 3711 (daily rentals) must stay allowed: %v", err)
	}
}

func TestValidateRejectsForeignQueryParams(t *testing.T) {
	bad := []Search{
		{Query: "x", Filters: map[string][]string{"limit": {"1"}}},
		{Query: "x", Filters: map[string][]string{"sort_by": {"price:asc"}}},
		{Query: "x", Filters: map[string][]string{"category_id": {"1"}}},
		{Query: "x", Filters: map[string][]string{"filter_enum_price": {"exchange"}}},
		{Query: "x", Ranges: map[string]FilterRange{"filter_float_price": {To: 5}}},
		{Query: "x", Ranges: map[string]FilterRange{"filter_float_motor_year:from": {From: 2010}}},
		{Query: "x", BrandParam: "category_id", BrandValues: []string{"1"}},
		{Query: "x", BrandParam: "filter_enum_color", BrandValues: []string{"black"}},
	}
	for _, s := range bad {
		var keyErr *BadKeyError
		if err := s.Validate(); !errors.As(err, &keyErr) {
			t.Errorf("%+v: err = %v, want BadKeyError", s, err)
		}
	}

	good := Search{
		Category:   73,
		Filters:    map[string][]string{"filter_enum_state": {"used"}, "filter_bool_delivery": {"1"}},
		Ranges:     map[string]FilterRange{"filter_float_motor_year": {From: 2010}},
		BrandParam: "filter_enum_digital_camera_manufacturers", BrandValues: []string{"2580"},
	}
	if err := good.Validate(); err != nil {
		t.Errorf("ordinary OLX filters must pass: %v", err)
	}
}

func TestValidateOwnerAndPriceMode(t *testing.T) {
	if err := (Search{Query: "x", OwnerType: "robot"}).Validate(); !errors.Is(err, ErrBadOwner) {
		t.Errorf("owner: %v", err)
	}
	if err := (Search{Query: "x", PriceMode: "cheap"}).Validate(); !errors.Is(err, ErrBadPrice) {
		t.Errorf("price mode: %v", err)
	}
}

func TestNormalize(t *testing.T) {
	s := Search{
		Name:          "  Мой   фильтр  ",
		Query:         " iphone   15 ",
		QueryVariants: []string{"iphone 15", "", "iphone 15", "iphone15"},
		ExcludeWords:  []string{" чохол ", "чохол", "", "коробка"},
		MinPrice:      9000, MaxPrice: 1000,
		Filters:    map[string][]string{"filter_enum_state": {"", "used", "used"}, "filter_enum_color": {""}},
		Ranges:     map[string]FilterRange{"filter_float_x": {}},
		BrandParam: "filter_enum_x",
	}
	s.Normalize()

	if s.Name != "Мой фильтр" || s.Query != "iphone 15" {
		t.Errorf("name/query = %q / %q", s.Name, s.Query)
	}
	if len(s.QueryVariants) != 2 || len(s.ExcludeWords) != 2 {
		t.Errorf("variants = %v, exclude = %v", s.QueryVariants, s.ExcludeWords)
	}
	if s.MinPrice != 1000 || s.MaxPrice != 9000 {
		t.Errorf("price bounds must be ordered: %v..%v", s.MinPrice, s.MaxPrice)
	}
	if len(s.Filters) != 1 || len(s.Filters["filter_enum_state"]) != 1 || s.Ranges != nil {
		t.Errorf("filters = %v, ranges = %v", s.Filters, s.Ranges)
	}
	if s.BrandParam != "" {
		t.Error("a brand param without values must be cleared")
	}

	free := Search{PriceMode: PriceFree, MinPrice: 10, MaxPrice: 20}
	free.Normalize()
	if free.MinPrice != 0 || free.MaxPrice != 0 {
		t.Error("free-only mode carries no price bounds")
	}
}

func TestNormalizeCapsExcludeWords(t *testing.T) {
	var s Search
	for i := 0; i < MaxExcludeWords+7; i++ {
		s.ExcludeWords = append(s.ExcludeWords, string(rune('а'+i))+"слово")
	}
	s.Normalize()
	if len(s.ExcludeWords) != MaxExcludeWords {
		t.Errorf("exclude words = %d, want cap %d", len(s.ExcludeWords), MaxExcludeWords)
	}
}

func TestCloneIsDeep(t *testing.T) {
	a := Search{
		QueryVariants: []string{"a"},
		Filters:       map[string][]string{"filter_enum_state": {"used"}},
		Ranges:        map[string]FilterRange{"filter_float_x": {From: 1}},
	}
	b := a.Clone()
	b.QueryVariants[0] = "changed"
	b.Filters["filter_enum_state"][0] = "new"
	b.Ranges["filter_float_x"] = FilterRange{From: 9}

	if a.QueryVariants[0] != "a" || a.Filters["filter_enum_state"][0] != "used" || a.Ranges["filter_float_x"].From != 1 {
		t.Fatalf("clone shares memory with the original: %+v", a)
	}
}

func TestActive(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	cases := []struct {
		s    Search
		want bool
	}{
		{Search{Query: "x"}, true},
		{Search{Query: "x", Disabled: true}, false},
		{Search{Query: "x", SnoozedUntil: now.Unix() + 60}, false},
		{Search{Query: "x", SnoozedUntil: now.Unix() - 60}, true},
		{Search{Category: 1758}, false},
	}
	for _, c := range cases {
		if got := c.s.Active(now); got != c.want {
			t.Errorf("Active(%+v) = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestRuleKeyIgnoresNameAndOwner(t *testing.T) {
	a := Search{Name: "A", OwnerChatID: "1", ID: "x", Query: "iPhone", Category: 44,
		Filters: map[string][]string{"filter_enum_state": {"used", "new"}}}
	b := Search{Name: "B", OwnerChatID: "2", ID: "y", Query: "iphone", Category: 44, Disabled: true,
		Filters: map[string][]string{"filter_enum_state": {"new", "used"}}}
	if a.RuleKey() != b.RuleKey() {
		t.Errorf("same rule must give the same key:\n%s\n%s", a.RuleKey(), b.RuleKey())
	}
	c := b
	c.MaxPrice = 100
	if a.RuleKey() == c.RuleKey() {
		t.Error("a different price must change the key")
	}
	d := b
	d.PriceMode = PriceFree
	if a.RuleKey() == d.RuleKey() {
		t.Error("free-only mode must change the key")
	}
}

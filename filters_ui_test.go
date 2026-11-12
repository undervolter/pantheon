package main

import (
	"strings"
	"testing"
)

func TestParseFilterParam(t *testing.T) {
	cases := []struct {
		in, base, sub string
	}{
		{"search[filter_enum_color][0]", "filter_enum_color", "0"},
		{"filter_enum_color[3]", "filter_enum_color", "3"},
		{"search[filter_float_motor_year][from]", "filter_float_motor_year", "from"},
		{"filter_float_motor_year:to", "filter_float_motor_year", "to"},
		{"filter_enum_color", "filter_enum_color", ""},
	}
	for _, c := range cases {
		base, sub := parseFilterParam(c.in)
		if base != c.base || sub != c.sub {
			t.Errorf("parseFilterParam(%q) = (%q, %q), want (%q, %q)", c.in, base, sub, c.base, c.sub)
		}
	}
}

func TestParseOLXURLFilters(t *testing.T) {
	u := "https://www.olx.ua/uk/transport/legkovye-avtomobili/vaz/?search%5Bfilter_enum_color%5D%5B0%5D=4&search%5Bfilter_float_motor_year%5D%5Bfrom%5D=2000&search%5Bfilter_float_motor_year%5D%5Bto%5D=2010"
	sc, err := ParseOLXURL(u)
	if err != nil {
		t.Fatal(err)
	}
	colors := sc.Filters["filter_enum_color"]
	if len(colors) != 1 || colors[0] != "4" {
		t.Errorf("color filter = %v, want [4]", colors)
	}
	r, ok := sc.Ranges["filter_float_motor_year"]
	if !ok || r.From != 2000 || r.To != 2010 {
		t.Errorf("year range = %+v, want 2000..2010", r)
	}
}

func TestCategoryFiltersSkipsReserved(t *testing.T) {
	defs := (&BotUI{}).categoryFilters(1115, "")
	if len(defs) == 0 {
		t.Fatal("no filters for category 1115")
	}
	for _, d := range defs {
		if isReservedFilterKey(d.Key) {
			t.Errorf("reserved filter %q leaked into UI", d.Key)
		}
	}
	hasColor := false
	for _, d := range defs {
		if d.Key == "filter_enum_color" {
			hasColor = true
		}
	}
	if !hasColor {
		t.Error("color filter missing for category 1115")
	}
}

func TestFiltersSummary(t *testing.T) {
	draft := SearchConfig{
		Category: 1115,
		Filters: map[string][]string{
			"filter_enum_color": {"4"},
		},
		Ranges: map[string]FilterRange{
			"filter_float_motor_year": {From: 2000, To: 2010},
		},
	}
	got := filtersSummary(draft)
	if !strings.Contains(got, "олір") && !strings.Contains(got, "олiр") {
		t.Errorf("summary has no color label: %q", got)
	}
	if !strings.Contains(got, "2000") || !strings.Contains(got, "2010") {
		t.Errorf("summary has no year range: %q", got)
	}
}

func TestRentCategoryFilters(t *testing.T) {
	cat := FindCategory(1760) // Довгострокова оренда квартир
	if cat == nil {
		t.Fatal("rent category 1760 not found in categories.go")
	}
	if !HasFilterCategory(1760) {
		t.Fatal("rent category has no filters metadata")
	}

	defs := (&BotUI{}).categoryFilters(1760, "")
	has := map[string]bool{}
	for _, d := range defs {
		has[d.Key] = true
	}
	for _, key := range []string{"filter_enum_number_of_rooms_string", "filter_float_total_area", "filter_bool_commission"} {
		if !has[key] {
			t.Errorf("rent filters missing %q", key)
		}
	}

	// Проверяем, что до категории можно дойти из TopCategories.
	root := 1760
	for i := 0; i < 10; i++ {
		c := FindCategory(root)
		if c == nil || c.ParentID == 0 {
			break
		}
		root = c.ParentID
	}
	found := false
	for _, tc := range TopCategories {
		if tc.ID == root {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("root category %d is not in TopCategories", root)
	}
}

func TestRangeLabel(t *testing.T) {
	cases := []struct {
		r    FilterRange
		unit string
		want string
	}{
		{FilterRange{From: 2000, To: 2010}, "", "от 2000 до 2010"},
		{FilterRange{From: 100}, "тис.км", "от 100 тис.км"},
		{FilterRange{To: 1.6}, "л", "до 1.6 л"},
	}
	for _, c := range cases {
		if got := rangeLabel(c.r, c.unit); got != c.want {
			t.Errorf("rangeLabel(%+v, %q) = %q, want %q", c.r, c.unit, got, c.want)
		}
	}
}

package olx

import (
	"testing"

	"pantheon/core/catalog"
)

// Category slugs repeat all over the tree: "shorty" names six categories and
// "drugoe" more than a hundred. Taking the first match in the whole tree put
// a link to men's shorts into «Шорти для вагітних».
func TestCategoryPathIsResolvedUnderItsParent(t *testing.T) {
	parent := catalog.FindCategoryBySlug("muzhskaya-odezhda")
	if parent == nil {
		t.Skip("the catalog has no muzhskaya-odezhda")
	}
	sc := mustParse(t, "https://www.olx.ua/uk/moda-i-stil/muzhskaya-odezhda/shorty/").Search
	if sc.Category == parent.ID || !catalog.CategoryInSubtree(sc.Category, parent.ID) {
		got := catalog.FindCategory(sc.Category)
		t.Fatalf("category = %d (%+v), want the shorts under %q (%d)", sc.Category, got, parent.Name, parent.ID)
	}
	if got := catalog.FindCategory(sc.Category); got == nil || got.Slug != "shorty" || sc.CatName != got.Name {
		t.Fatalf("resolved %+v, name %q", got, sc.CatName)
	}

	// The same last segment under another parent is another category.
	other := mustParse(t, "https://www.olx.ua/uk/moda-i-stil/zhenskaya-odezhda/shorty/").Search
	if other.Category == sc.Category {
		t.Errorf("men's and women's shorts resolved to the same category %d", sc.Category)
	}
}

// What is not a child of the category so far is a location, even if some
// unrelated category elsewhere happens to carry the same slug.
func TestSegmentOutsideTheCategoryIsLeftForTheGeoIndex(t *testing.T) {
	p := mustParse(t, "https://www.olx.ua/uk/elektronika/kiev/q-iphone/")
	top := catalog.FindCategoryBySlug("elektronika")
	if top == nil || p.Search.Category != top.ID {
		t.Fatalf("category = %d, want elektronika", p.Search.Category)
	}
	if len(p.Slugs) != 1 || p.Slugs[0] != "kiev" {
		t.Fatalf("slugs = %v, want the city left for the caller", p.Slugs)
	}
}

func TestParseNumberRejectsWhatAFilterCannotHold(t *testing.T) {
	for _, bad := range []string{"inf", "+Inf", "-inf", "Infinity", "nan", "NaN", "1e400", "1e13", "-5", "", "abc"} {
		if v, ok := parseNumber(bad); ok {
			t.Errorf("parseNumber(%q) = %v, want a refusal", bad, v)
		}
	}
	for in, want := range map[string]float64{"30000": 30000, "1,5": 1.5, " 2015 ": 2015, "0": 0} {
		if v, ok := parseNumber(in); !ok || v != want {
			t.Errorf("parseNumber(%q) = %v, %v; want %v", in, v, ok, want)
		}
	}
}

func TestInfiniteBoundsInALinkAreIgnored(t *testing.T) {
	sc := mustParse(t, "https://www.olx.ua/uk/list/q-x/?search%5Bfilter_float_price%3Ato%5D=Inf"+
		"&search%5Bfilter_float_price%3Afrom%5D=nan&search%5Bfilter_float_year%3Ato%5D=inf").Search
	if sc.MaxPrice != 0 || sc.MinPrice != 0 || len(sc.Ranges) != 0 {
		t.Fatalf("a link stored non-numbers: min=%v max=%v ranges=%v", sc.MinPrice, sc.MaxPrice, sc.Ranges)
	}
	sc.Normalize()
	if err := sc.Validate(); err != nil {
		t.Fatalf("the rest of the link must still make a filter: %v", err)
	}
}

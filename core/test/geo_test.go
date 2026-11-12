package test

import (
	"testing"

	"pantheon/core/olx"
)

func TestNormalizeGeoName(t *testing.T) {
	cases := map[string]string{
		"  Кременьчуг ": "кременьчуг",
		"Київ":          "киив",
		"Кременчук-2":   "кременчук 2",
		"Їжакевич":      "ижакевич",
	}
	for in, want := range cases {
		if got := olx.NormalizeGeoName(in); got != want {
			t.Errorf("normalizeGeoName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLevenshteinTypos(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"креминчуг", "кременчук", 2},
		{"криминчуг", "кременчук", 3},
		{"кременьчуг", "кременчук", 2},
		{"кременчук", "кременчук", 0},
	}
	for _, c := range cases {
		if got := olx.Levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestRankLocationsPrefersCloseName(t *testing.T) {
	results := []olx.GeoLocationSuggestion{
		{City: olx.GeoCity{ID: 1, Name: "Кременчуки"}, Region: olx.GeoRegion{ID: 20, Name: "Хмельницька область"}},
		{City: olx.GeoCity{ID: 2, Name: "Кременчук"}, Region: olx.GeoRegion{ID: 15, Name: "Полтавська область"}},
		{City: olx.GeoCity{ID: 3, Name: "Київ"}, Region: olx.GeoRegion{ID: 25, Name: "Київська область"}},
	}
	got := olx.RankLocations("креминчуг", results)
	if len(got) == 0 {
		t.Fatal("no matches")
	}
	if got[0].City.ID != 2 {
		t.Fatalf("best match = %q (id %d), want Кременчук", got[0].City.Name, got[0].City.ID)
	}
}

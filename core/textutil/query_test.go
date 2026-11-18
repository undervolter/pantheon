package textutil

import (
	"strings"
	"testing"
)

func TestNormalizeQuery(t *testing.T) {
	cases := map[string]string{
		"iPhone 4s":       "iphone 4s",
		"Айфон 4с":        "iphone 4s",
		"4С":              "4s",
		"  iPhone   4S! ": "iphone 4s",
		"Плейстейшн 5":    "playstation 5",
		"Самсунг Галакси": "samsung galaxy",
		"айфон":           "iphone",
	}
	for in, want := range cases {
		if got := NormalizeQuery(in); got != want {
			t.Errorf("NormalizeQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestQueryVariants(t *testing.T) {
	got := QueryVariants("Айфон 4с")
	want := []string{"iphone 4s", "iphone4s"}
	if len(got) != len(want) {
		t.Fatalf("QueryVariants = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("variant[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	if got := QueryVariants("4с"); len(got) != 1 || got[0] != "4s" {
		t.Errorf("QueryVariants(4с) = %#v, want [4s]", got)
	}
}

func TestMatchesQuery(t *testing.T) {
	cases := []struct {
		name  string
		query string
		text  string
		want  bool
	}{
		{"latin title vs cyrillic query", "Айфон 4с", "Apple iPhone 4S 16GB", true},
		{"cyrillic title vs latin query", "iphone 4s", "Айфон 4с 16 гб", true},
		{"model only", "4с", "iPhone 4s, 16gb", true},
		{"no false positive on 14", "4s", "iPhone 14 Pro Max", false},
		{"description match", "iphone 4s", "Продам телефон В идеале iPhone 4S 32gb", true},
		{"all tokens required", "iphone 4s", "iPhone 6s 64gb", false},
		{"empty variants means no filter", "", "что угодно", true},
	}

	for _, c := range cases {
		if got := MatchesQuery(c.text, QueryVariants(c.query)); got != c.want {
			t.Errorf("%s: MatchesQuery(%q) = %v, want %v", c.name, c.text, got, c.want)
		}
	}
}

func TestNormalizeQueryKeepsDigits(t *testing.T) {
	got := NormalizeQuery("iPhone 15 Pro Max 256GB")
	if !strings.Contains(got, "15") || !strings.Contains(got, "256gb") {
		t.Errorf("NormalizeQuery lost digits: %q", got)
	}
}

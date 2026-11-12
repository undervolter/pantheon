package main

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

func TestOfferMatchesQuery(t *testing.T) {
	cases := []struct {
		name  string
		query string
		o     Offer
		want  bool
	}{
		{
			name:  "latin title vs cyrillic query",
			query: "Айфон 4с",
			o:     Offer{Title: "Apple iPhone 4S 16GB"},
			want:  true,
		},
		{
			name:  "cyrillic title vs latin query",
			query: "iphone 4s",
			o:     Offer{Title: "Айфон 4с 16 гб"},
			want:  true,
		},
		{
			name:  "model only",
			query: "4с",
			o:     Offer{Title: "iPhone 4s, 16gb"},
			want:  true,
		},
		{
			name:  "no false positive on 14",
			query: "4s",
			o:     Offer{Title: "iPhone 14 Pro Max"},
			want:  false,
		},
		{
			name:  "description match",
			query: "iphone 4s",
			o:     Offer{Title: "Продам телефон", Description: "В идеале iPhone 4S 32gb"},
			want:  true,
		},
		{
			name:  "all tokens required",
			query: "iphone 4s",
			o:     Offer{Title: "iPhone 6s 64gb"},
			want:  false,
		},
		{
			name:  "empty variants means no filter",
			query: "",
			o:     Offer{Title: "что угодно"},
			want:  true,
		},
	}

	for _, c := range cases {
		variants := QueryVariants(c.query)
		if got := OfferMatchesQuery(&c.o, variants); got != c.want {
			t.Errorf("%s: OfferMatchesQuery(%q) = %v, want %v", c.name, c.o.Title, got, c.want)
		}
	}
}

func TestNormalizeQueryKeepsDigits(t *testing.T) {
	got := NormalizeQuery("iPhone 15 Pro Max 256GB")
	if !strings.Contains(got, "15") || !strings.Contains(got, "256gb") {
		t.Errorf("NormalizeQuery lost digits: %q", got)
	}
}

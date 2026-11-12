package test

import (
	"strings"
	"testing"

	"pantheon/core/olx"
)

func TestSanitizeListingText(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"О Б Ь Я В Л Е Н И Е", "ОБЬЯВЛЕНИЕ"},
		{"м и к р о ф о н", "микрофон"},
		{"M I C R O P H O N E", "MICROPHONE"},
		{"Микрофон   срочно\nпродам", "Микрофон срочно продам"},
		{"микро\u200bфон", "микрофон"},
		{"&nbsp;&nbsp;Микрофон", "Микрофон"},
		{"я и ты", "я и ты"},
		{"О\u00a0Б\u00a0Ь\u00a0Я", "ОБЬЯ"},
		{"О\u200bБ\u200bЬ\u200bЯ", "ОБЬЯ"},
		{"  ", ""},
	}

	for _, c := range cases {
		if got := olx.SanitizeListingText(c.in); got != c.want {
			t.Errorf("sanitizeListingText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeListingDescription(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			"     \n\n   Микрофон   Shure   SM7B   \n\n\n\n   продам   срочно   \n     ",
			"Микрофон Shure SM7B\n\nпродам срочно",
		},
		{
			strings.Repeat(" ", 600) + "реальный текст",
			"реальный текст",
		},
		{
			"м и к р о ф о н\n\n\n\nпродам",
			"микрофон\n\nпродам",
		},
		{
			"\n\n\n",
			"",
		},
	}

	for _, c := range cases {
		if got := olx.SanitizeListingDescription(c.in); got != c.want {
			t.Errorf("sanitizeListingDescription(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsLikelySpamTitle(t *testing.T) {
	spam := []string{
		"О Б Ь Я В Л Е Н И Е",
		"ОБЪЯВЛЕНИЕ",
		"обьявление",
		"Т Е С Т",
		"TEST",
		"",
		"...",
	}
	for _, s := range spam {
		normalized := olx.SanitizeListingText(s)
		if !olx.IsLikelySpamTitle(normalized) {
			t.Errorf("%q must be treated as spam (normalized %q)", s, normalized)
		}
	}

	if olx.IsLikelySpamTitle("Объявление о продаже микрофона Shure") {
		t.Error("normal ad title must not be treated as spam")
	}
}

package i18n

import (
	"regexp"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := map[string]Lang{"ru": RU, "RU-ru": RU, "uk": UK, "ua": UK, "uk-UA": UK, "en_GB": EN, "English": EN, "українська": UK}
	for in, want := range cases {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if got, ok := Parse("de"); ok || got != Default {
		t.Errorf("Parse(de) = %q, %v; want the default and false", got, ok)
	}
	if Of("") != Default || Of("uk") != UK {
		t.Error("Of must fall back to the default language")
	}
}

func TestEveryKeyHasAllLanguages(t *testing.T) {
	for _, key := range Keys() {
		for _, lang := range All {
			if strings.TrimSpace(catalog[key][lang]) == "" {
				t.Errorf("%s: missing %s", key, lang)
			}
		}
	}
}

var verbRe = regexp.MustCompile(`%[-+# 0]*[0-9]*(?:\.[0-9]+)?[a-zA-Z%]`)

// A translation with a different set of verbs would print "%!d(string=…)" or
// drop an argument in one language only.
func TestVerbsMatchAcrossLanguages(t *testing.T) {
	for _, key := range Keys() {
		want := strings.Join(verbRe.FindAllString(catalog[key][Default], -1), " ")
		for _, lang := range All {
			got := strings.Join(verbRe.FindAllString(catalog[key][lang], -1), " ")
			if got != want {
				t.Errorf("%s: verbs in %s are %q, in %s %q", key, lang, got, Default, want)
			}
		}
	}
}

func TestTFallsBack(t *testing.T) {
	if got := T(EN, "no.such.key"); got != "no.such.key" {
		t.Errorf("unknown key = %q, want the key itself", got)
	}
	if got := T(UK, "card.price", "100 грн."); got != "Продається за <b>100 грн.</b>" {
		t.Errorf("formatted = %q", got)
	}
}

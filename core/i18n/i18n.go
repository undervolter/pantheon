// Package i18n holds every user-facing string of the bot in Russian,
// Ukrainian and English.
package i18n

import (
	"fmt"
	"sort"
	"strings"
)

type Lang string

const (
	RU Lang = "ru"
	UK Lang = "uk"
	EN Lang = "en"

	Default = RU
)

var All = []Lang{RU, UK, EN}

// Parse recognises a language code or name. Region suffixes ("uk-UA",
// "en_GB") are ignored, which is the form Telegram reports a user's
// language in.
func Parse(s string) (Lang, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if i := strings.IndexAny(s, "-_"); i > 0 {
		s = s[:i]
	}
	switch s {
	case "ru", "rus", "russian", "русский":
		return RU, true
	case "uk", "ua", "ukr", "ukrainian", "українська":
		return UK, true
	case "en", "eng", "english":
		return EN, true
	}
	return Default, false
}

// Of returns the language for a stored code, falling back to the default.
func Of(code string) Lang {
	l, _ := Parse(code)
	return l
}

func (l Lang) Name() string {
	switch l {
	case UK:
		return "Українська"
	case EN:
		return "English"
	}
	return "Русский"
}

func (l Lang) Flag() string {
	switch l {
	case UK:
		return "🇺🇦"
	case EN:
		return "🇬🇧"
	}
	return "🏳️"
}

var catalog = map[string]map[Lang]string{}

func register(entries map[string]map[Lang]string) {
	for k, v := range entries {
		if _, dup := catalog[k]; dup {
			panic("i18n: duplicate key " + k)
		}
		catalog[k] = v
	}
}

// T returns the string for key in lang, formatted with args. A language
// missing for a key falls back to the default one; an unknown key comes back
// as the key itself so the gap is visible instead of silent.
func T(lang Lang, key string, args ...any) string {
	entry, ok := catalog[key]
	if !ok {
		return key
	}
	text := entry[lang]
	if text == "" {
		text = entry[Default]
	}
	if len(args) == 0 {
		return text
	}
	return fmt.Sprintf(text, args...)
}

// Has reports whether key exists for lang.
func Has(lang Lang, key string) bool {
	return catalog[key][lang] != ""
}

// Keys returns every key of the catalog, sorted.
func Keys() []string {
	out := make([]string, 0, len(catalog))
	for k := range catalog {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

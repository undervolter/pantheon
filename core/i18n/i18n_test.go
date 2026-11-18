package i18n

import (
	"io/fs"
	"os"
	"path/filepath"
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

// Every key of the catalog must be used by the code. A string nobody shows is
// dead weight at best; at worst it is the text of a feature this edition does
// not have, carried over by accident.
func TestEveryKeyIsUsed(t *testing.T) {
	var src strings.Builder
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == "deploy" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "catalog_") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src.Write(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	code := src.String()

	// Keys built at run time: "bc." + command name.
	dynamic := []string{"bc."}
	for _, prefix := range dynamic {
		if !strings.Contains(code, `"`+prefix+`"`) {
			t.Errorf("prefix %q is declared dynamic but the code does not build keys from it", prefix)
		}
	}

	for _, key := range Keys() {
		if strings.Contains(code, `"`+key+`"`) {
			continue
		}
		used := false
		for _, prefix := range dynamic {
			used = used || strings.HasPrefix(key, prefix)
		}
		if !used {
			t.Errorf("key %q is not used anywhere", key)
		}
	}
}

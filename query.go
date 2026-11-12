package main

import (
	"regexp"
	"strings"
)

// Синонимы, которые встречаются в объявлениях и в том, как люди пишут запросы.
var querySynonyms = strings.NewReplacer(
	"айфон", "iphone",
	"ифон", "iphone",
	"афон", "iphone",
	"эпл", "apple",
	"еппл", "apple",
	"самсунг", "samsung",
	"сяоми", "xiaomi",
	"ксиоми", "xiaomi",
	"редми", "redmi",
	"макбук", "macbook",
	"плейстейшн", "playstation",
	"плэйстейшн", "playstation",
	"галакси", "galaxy",
	"пиксель", "pixel",
	"онеплюс", "oneplus",
	"поко", "poco",
	"нива", "niva",
	"лада", "lada",
	"тойота", "toyota",
	"бмв", "bmw",
	"мерседес", "mercedes",
	"фольксваген", "volkswagen",
	"шкода", "skoda",
	"рено", "renault",
	"пежо", "peugeot",
	"ситроен", "citroen",
	"хендай", "hyundai",
	"киа", "kia",
	"ниссан", "nissan",
	"мазда", "mazda",
	"мицубиси", "mitsubishi",
	"субару", "subaru",
	"хонда", "honda",
	"ауди", "audi",
	"опель", "opel",
	"шевроле", "chevrolet",
	"форд", "ford",
	"фиат", "fiat",
	"дачиа", "dacia",
	"тесла", "tesla",
)

// Кириллические символы, которые визуально совпадают с латиницей.
// Применяем к обеим сторонам (запрос и текст объявления), чтобы «4с» и «4s» совпали.
var queryLookalikes = strings.NewReplacer(
	"а", "a",
	"е", "e",
	"о", "o",
	"р", "p",
	"с", "s",
	"х", "x",
	"у", "y",
	"і", "i",
	"ї", "i",
	"є", "e",
	"ґ", "g",
)

var queryNonAlnum = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// NormalizeQuery приводит запрос к единому виду:
// нижний регистр, синонимы, похожие кириллические буквы → латиница, только буквы/цифры.
func NormalizeQuery(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	s = querySynonyms.Replace(s)
	s = queryLookalikes.Replace(s)
	s = queryNonAlnum.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// QueryVariants возвращает варианты запроса для пост-фильтра.
func QueryVariants(s string) []string {
	canonical := NormalizeQuery(s)
	if canonical == "" {
		return nil
	}
	out := []string{canonical}
	if compact := strings.ReplaceAll(canonical, " ", ""); compact != canonical {
		out = append(out, compact)
	}
	return out
}

// OfferMatchesQuery проверяет, что в заголовке/описании есть все токены запроса.
// Для одного токена — точное совпадение токена (чтобы «4s» не матчил «14s»).
func OfferMatchesQuery(o *Offer, variants []string) bool {
	if len(variants) == 0 {
		return true
	}
	text := NormalizeQuery(o.Title + " " + stripHTMLTags(o.Description))
	if text == "" {
		return false
	}
	tokens := strings.Fields(text)
	compact := strings.ReplaceAll(text, " ", "")

	for _, v := range variants {
		vt := strings.Fields(v)
		if len(vt) == 0 {
			continue
		}
		if len(vt) == 1 {
			if stringInSlice(tokens, vt[0]) {
				return true
			}
			continue
		}
		all := true
		for _, tok := range vt {
			if !stringInSlice(tokens, tok) {
				all = false
				break
			}
		}
		if all {
			return true
		}
		if compactV := strings.ReplaceAll(v, " ", ""); compactV != "" && strings.Contains(compact, compactV) {
			return true
		}
	}
	return false
}

package main

import (
	"html"
	"strings"
	"unicode"
)

// Невидимые/нулевой ширины символы, которыми обходят фильтры.
var invisibleTextReplacer = strings.NewReplacer(
	"\u200b", "", // zero width space
	"\u200c", "", // zero width non-joiner
	"\u200d", "", // zero width joiner
	"\u2060", "", // word joiner
	"\ufeff", "", // BOM / zero width no-break space
	"\u00ad", "", // soft hyphen
	"\u180e", "", // Mongolian vowel separator
	"\u200e", "", // LRM
	"\u200f", "", // RLM
	"\u202a", "",
	"\u202b", "",
	"\u202c", "",
	"\u202d", "",
	"\u202e", "",
)

// sanitizeListingText — для однострочных полей: заголовок, цена, локация.
// Убирает невидимые символы, схлопывает все пробелы/переносы в один пробел
// и склеивает разряженные слова.
func sanitizeListingText(s string) string {
	s = prepareListingText(s)
	s = strings.Join(strings.Fields(s), " ")
	return glueSpacedLetters(s)
}

// sanitizeListingDescription — для описания. Сохраняет абзацы, но:
//   - убирает ведущие/хвостовые пробелы в каждой строке;
//   - схлопывает горизонтальные отступы до одного пробела;
//   - выкидывает пустые строки в начале/конце;
//   - оставляет максимум одну пустую строку между абзацами;
//   - склеивает разряженные слова.
func sanitizeListingDescription(s string) string {
	s = prepareListingText(s)
	if s == "" {
		return ""
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blankPending := false

	for _, line := range lines {
		line = normalizeLineWhitespace(line)
		line = glueSpacedLetters(line)
		line = strings.TrimSpace(line)

		if line == "" {
			if len(out) > 0 && !blankPending {
				out = append(out, "")
				blankPending = true
			}
			continue
		}

		out = append(out, line)
		blankPending = false
	}

	for len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}

	return strings.Join(out, "\n")
}

func prepareListingText(s string) string {
	if s == "" {
		return ""
	}
	s = html.UnescapeString(s)
	return invisibleTextReplacer.Replace(s)
}

// normalizeLineWhitespace убирает отступы и повторные пробелы внутри строки.
func normalizeLineWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	spacePending := false
	for _, r := range s {
		if r == '\n' {
			continue
		}
		if unicode.IsSpace(r) {
			if b.Len() > 0 {
				spacePending = true
			}
			continue
		}
		if spacePending {
			b.WriteRune(' ')
			spacePending = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

// glueSpacedLetters склеивает серии одиночных букв: "О Б Ь Я" -> "ОБЬЯ".
// Короткие легитимные серии вроде "я и ты" не трогает.
func glueSpacedLetters(s string) string {
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return ""
	}

	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); {
		j := i
		hasUpper := false
		for j < len(tokens) && isSingleLetterToken(tokens[j]) {
			if unicode.IsUpper([]rune(tokens[j])[0]) {
				hasUpper = true
			}
			j++
		}

		run := j - i
		if run >= 4 && (hasUpper || run >= 6) {
			out = append(out, strings.Join(tokens[i:j], ""))
			i = j
			continue
		}

		out = append(out, tokens[i])
		i++
	}

	return strings.Join(out, " ")
}

func isSingleLetterToken(s string) bool {
	r := []rune(s)
	return len(r) == 1 && unicode.IsLetter(r[0])
}

// isLikelySpamTitle отсекает заголовки-пустышки после нормализации.
func isLikelySpamTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	t = strings.NewReplacer("’", "'", "‘", "'", "`", "'").Replace(t)
	t = strings.Trim(t, " .,!?;:-—–_\"'()[]{}")
	if t == "" {
		return true
	}

	switch t {
	case "объявление", "обьявление", "обявление", "об'явление",
		"advertisement", "ad", "ads",
		"test", "тест", "проверка":
		return true
	}
	return false
}

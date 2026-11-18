package bot

import (
	"html"
	"strings"

	"pantheon/core/telegram"
)

// page builds one screen of the control panel as a small HTML document:
//
//	<h3> title
//	<blockquote> notice (optional, right under the title)
//	KPI strip, key/value tables, <h4> sections
//	<footer> hints
//
// The builder does not escape its arguments: every string that comes from
// OLX or from a user must be passed through esc() by the caller. Translated
// strings carry their own markup.
type page struct {
	b     strings.Builder
	rows  []string
	notes []string
}

type stat struct {
	label string
	value string
}

// esc makes user- and OLX-supplied text safe to put into a page.
func esc(s string) string { return html.EscapeString(s) }

func newPage(title string) *page {
	p := &page{}
	if title != "" {
		p.b.WriteString("<h3>" + title + "</h3>")
	}
	return p
}

func paragraphs(s string) string {
	var b strings.Builder
	for _, para := range strings.Split(strings.TrimSpace(s), "\n\n") {
		if para = strings.TrimSpace(para); para != "" {
			b.WriteString("<p>" + strings.ReplaceAll(para, "\n", "<br>") + "</p>")
		}
	}
	return b.String()
}

func (p *page) flush() {
	if len(p.rows) == 0 {
		return
	}
	p.b.WriteString("<table compact>")
	for _, r := range p.rows {
		p.b.WriteString(r)
	}
	p.b.WriteString("</table>")
	p.rows = nil
}

func (p *page) text(s string) *page {
	if s == "" {
		return p
	}
	p.flush()
	p.b.WriteString(paragraphs(s))
	return p
}

// kv adds a "label | value" row; consecutive rows form one table.
func (p *page) kv(label, value string) *page {
	if value == "" {
		return p
	}
	p.rows = append(p.rows, "<tr><td>"+label+"</td><td><b>"+value+"</b></td></tr>")
	return p
}

func (p *page) section(title string) *page {
	p.flush()
	p.b.WriteString("<h4>" + title + "</h4>")
	return p
}

// stats adds a KPI strip: a row of labels over a row of values.
func (p *page) stats(items ...stat) *page {
	if len(items) == 0 {
		return p
	}
	p.flush()
	p.b.WriteString("<table compact><tr>")
	for _, it := range items {
		p.b.WriteString(`<th align="center">` + it.label + "</th>")
	}
	p.b.WriteString("</tr><tr>")
	for _, it := range items {
		p.b.WriteString(`<td align="center"><b>` + it.value + "</b></td>")
	}
	p.b.WriteString("</tr></table>")
	return p
}

func (p *page) table(head []string, rows [][]string) *page {
	if len(rows) == 0 {
		return p
	}
	p.flush()
	p.b.WriteString("<table compact striped>")
	if len(head) > 0 {
		p.b.WriteString("<tr>")
		for _, h := range head {
			p.b.WriteString("<th>" + h + "</th>")
		}
		p.b.WriteString("</tr>")
	}
	for _, r := range rows {
		p.b.WriteString("<tr>")
		for _, c := range r {
			p.b.WriteString("<td>" + c + "</td>")
		}
		p.b.WriteString("</tr>")
	}
	p.b.WriteString("</table>")
	return p
}

func (p *page) list(items []string) *page {
	if len(items) == 0 {
		return p
	}
	p.flush()
	p.b.WriteString("<ul>")
	for _, it := range items {
		p.b.WriteString("<li>" + it + "</li>")
	}
	p.b.WriteString("</ul>")
	return p
}

func (p *page) quote(s string) *page {
	if s == "" {
		return p
	}
	p.flush()
	p.b.WriteString("<blockquote>" + strings.ReplaceAll(strings.TrimSpace(s), "\n", "<br>") + "</blockquote>")
	return p
}

// note adds a line to the grey footer of the page.
func (p *page) note(s string) *page {
	if s != "" {
		p.notes = append(p.notes, strings.ReplaceAll(strings.TrimSpace(s), "\n", "<br>"))
	}
	return p
}

func (p *page) String() string {
	p.flush()
	out := p.b.String()
	if len(p.notes) > 0 {
		out += "<footer>" + strings.Join(p.notes, "<br>") + "</footer>"
	}
	return out
}

// promptPage starts a page from a "Title\n\nBody" string.
func promptPage(text string) *page {
	title, body, _ := strings.Cut(strings.TrimSpace(text), "\n\n")
	return newPage(title).text(body)
}

// prompt turns a "Title\n\nBody" string into a finished page.
func prompt(text string) string { return promptPage(text).String() }

// withNotice puts a callout right under the title of a page.
func withNotice(notice, view string) string {
	if notice == "" {
		return view
	}
	callout := "<blockquote>" + strings.ReplaceAll(strings.TrimSpace(notice), "\n", "<br>") + "</blockquote>"
	if i := strings.Index(view, "</h3>"); i >= 0 && strings.HasPrefix(view, "<h3>") {
		return view[:i+5] + callout + view[i+5:]
	}
	if !telegram.IsPage(view) {
		view = paragraphs(view)
	}
	return callout + view
}

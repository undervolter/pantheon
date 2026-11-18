package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"pantheon/core/logx"
)

// A page is a small HTML document — headings, tables, lists, a footer — that
// the Bot API renders as a rich message. Every screen of the control panel is
// one page.

// ErrPageEditRejected is returned when the server refused to turn an existing
// message into the given page. The caller sends the page as a new message.
var ErrPageEditRejected = errors.New("rich page edit rejected")

var blockStarts = []string{"<h1", "<h2", "<h3", "<h4", "<h5", "<h6", "<p>", "<p ", "<table", "<blockquote", "<ul", "<ol", "<footer", "<details", "<hr", "<aside", "<pre"}

// IsPage reports whether text is already a page rather than a plain string.
func IsPage(text string) bool {
	t := strings.TrimSpace(text)
	for _, p := range blockStarts {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

// AsPage wraps plain text into paragraphs; a page is returned unchanged.
func AsPage(text string) string {
	if IsPage(text) {
		return text
	}
	var b strings.Builder
	for _, para := range strings.Split(strings.TrimSpace(text), "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		b.WriteString("<p>" + strings.ReplaceAll(para, "\n", "<br>") + "</p>")
	}
	return b.String()
}

func richPayload(page string) map[string]any {
	return map[string]any{"html": AsPage(page)}
}

// SendPage sends a page as a new message and returns its id.
func (c *Client) SendPage(ctx context.Context, chatID, page string, km *InlineKeyboardMarkup) (int, error) {
	if !c.plainPages.Load() {
		payload := map[string]any{"chat_id": chatID, "rich_message": richPayload(page)}
		if km != nil {
			payload["reply_markup"] = km
		}
		var msg TGMessage
		err := c.callUI(ctx, "sendRichMessage", payload, &msg)
		if err == nil {
			return msg.MessageID, nil
		}
		if !c.pageFallback(err) {
			return 0, err
		}
	}
	return c.SendHTML(ctx, chatID, RichToPlain(page), km, false)
}

// EditPage turns an existing message into the page.
func (c *Client) EditPage(ctx context.Context, chatID string, messageID int, page string, km *InlineKeyboardMarkup) error {
	if c.plainPages.Load() {
		return c.EditHTML(ctx, chatID, messageID, RichToPlain(page), km)
	}
	payload := map[string]any{"chat_id": chatID, "message_id": messageID, "rich_message": richPayload(page)}
	if km != nil {
		payload["reply_markup"] = km
	}
	err := c.callUI(ctx, "editMessageText", payload, nil)
	if err == nil || IsNotModified(err) {
		return nil
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusBadRequest {
		return fmt.Errorf("%w: %v", ErrPageEditRejected, err)
	}
	return err
}

// pageFallback decides whether a failed rich send should be repeated as plain
// HTML. Only a server that does not know the method, or that rejected this
// particular page, qualifies; a timeout or a flood limit does not, because the
// message may already have been delivered.
func (c *Client) pageFallback(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || IsDeadChat(err) {
		return false
	}
	switch apiErr.Code {
	case http.StatusNotFound:
		if !c.plainPages.Swap(true) {
			logx.Warnf("[TG] this Bot API server has no rich messages, menus switch to plain text")
		}
		return true
	case http.StatusBadRequest:
		logx.Warnf("[TG] rich menu rejected, sent as plain text: %s", apiErr.Description)
		return true
	}
	return false
}

var (
	reTable    = regexp.MustCompile(`(?s)<table([^>]*)>(.*?)</table>`)
	reRow      = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	reCell     = regexp.MustCompile(`(?s)<t([dh])[^>]*>(.*?)</t[dh]>`)
	reHeading  = regexp.MustCompile(`(?s)<h[1-6][^>]*>(.*?)</h[1-6]>`)
	reSummary  = regexp.MustCompile(`(?s)<details[^>]*>\s*<summary>(.*?)</summary>`)
	reFooter   = regexp.MustCompile(`(?s)<footer>(.*?)</footer>`)
	reBreak    = regexp.MustCompile(`<br\s*/?>`)
	reParaOpen = regexp.MustCompile(`<p[^>]*>`)
	reListOpen = regexp.MustCompile(`<[uo]l[^>]*>`)
	reItem     = regexp.MustCompile(`<li[^>]*>`)
	reBlank    = regexp.MustCompile(`\n{3,}`)
	reSpaces   = regexp.MustCompile(`[ \t]+\n`)
)

// tableToPlain flattens one table. striped marks a data table, which keeps
// one line per row even when it has a single row; without it a header row
// plus one value row is a KPI strip and reads best as "label: value".
func tableToPlain(table string, striped bool) string {
	type cell struct {
		head bool
		text string
	}
	var rows [][]cell
	for _, r := range reRow.FindAllStringSubmatch(table, -1) {
		var cells []cell
		for _, c := range reCell.FindAllStringSubmatch(r[1], -1) {
			cells = append(cells, cell{head: c[1] == "h", text: strings.TrimSpace(c[2])})
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}

	var header []string
	if len(rows) > 0 {
		all := true
		for _, c := range rows[0] {
			all = all && c.head
		}
		if all {
			for _, c := range rows[0] {
				header = append(header, c.text)
			}
			rows = rows[1:]
		}
	}

	var out []string
	if !striped && len(header) > 0 && len(rows) == 1 && len(rows[0]) == len(header) {
		for i, c := range rows[0] {
			switch {
			case c.text == "":
			case header[i] == "":
				out = append(out, c.text)
			default:
				out = append(out, header[i]+": "+c.text)
			}
		}
		return "\n" + strings.Join(out, "\n") + "\n"
	}
	for _, r := range rows {
		texts := make([]string, 0, len(r))
		for _, c := range r {
			if c.text != "" {
				texts = append(texts, c.text)
			}
		}
		if len(r) == 2 && len(texts) == 2 && len(header) == 0 {
			out = append(out, texts[0]+": "+texts[1])
			continue
		}
		out = append(out, strings.Join(texts, " · "))
	}
	return "\n" + strings.Join(out, "\n") + "\n"
}

// RichToPlain renders a page with the tags ordinary HTML messages support,
// for Bot API servers without rich messages.
func RichToPlain(page string) string {
	if !IsPage(page) {
		return page
	}
	s := reTable.ReplaceAllStringFunc(page, func(t string) string {
		m := reTable.FindStringSubmatch(t)
		return tableToPlain(m[2], strings.Contains(m[1], "striped"))
	})
	s = reSummary.ReplaceAllString(s, "\n\n<b>$1</b>\n")
	s = strings.ReplaceAll(s, "</details>", "\n")
	s = reHeading.ReplaceAllString(s, "\n\n<b>$1</b>\n")
	s = reFooter.ReplaceAllString(s, "\n\n<i>$1</i>\n")
	s = reBreak.ReplaceAllString(s, "\n")
	s = reParaOpen.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "</p>", "\n")
	s = reListOpen.ReplaceAllString(s, "\n")
	s = strings.NewReplacer("</ul>", "\n", "</ol>", "\n", "</li>", "\n").Replace(s)
	s = reItem.ReplaceAllString(s, "• ")
	s = strings.ReplaceAll(s, "<blockquote", "\n<blockquote")
	s = strings.ReplaceAll(s, "</blockquote>", "</blockquote>\n")
	s = reSpaces.ReplaceAllString(s, "\n")
	s = reBlank.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

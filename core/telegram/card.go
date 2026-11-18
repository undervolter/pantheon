package telegram

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/textutil"
)

// Size limits of a card. Listing text comes from sellers and has no upper
// bound, while a message that is too long is rejected for good and the alert
// is lost.
const (
	maxTextDescRunes = 450
	maxRichDescRunes = 3000
	maxRichParams    = 60
	// The limit of a text message is 4096; the margin covers what entities add.
	maxTextRunes = 3900
)

// CardOptions describe how one alert is rendered for one chat.
type CardOptions struct {
	Lang       i18n.Lang
	FilterName string
	Rich       bool // rich card with collage and table, otherwise plain HTML
	MaxPhotos  int  // 0 — no photos
	ShowParams bool
	MaxParams  int  // 0 — all
	Silent     bool // deliver without a sound (quiet hours)
}

// PriceWords returns the price vocabulary of a language.
func PriceWords(lang i18n.Lang) domain.PriceWords {
	return domain.PriceWords{
		UAH:            i18n.T(lang, "price.uah"),
		NegotiableNote: i18n.T(lang, "price.negotiable_note"),
		Negotiable:     i18n.T(lang, "price.negotiable"),
		Free:           i18n.T(lang, "price.free"),
		Exchange:       i18n.T(lang, "price.exchange"),
		Unspecified:    i18n.T(lang, "price.none"),
	}
}

// priceLine is the price sentence of a card. "Listed for …" only makes sense
// for an amount; a free or exchange listing gets its own wording.
func priceLine(o *domain.Offer, lang i18n.Lang) string {
	switch o.PriceKind() {
	case domain.PriceIsFree:
		return i18n.T(lang, "card.free")
	case domain.PriceExchange:
		return i18n.T(lang, "card.exchange")
	case domain.PriceArranged:
		return i18n.T(lang, "card.price_arranged")
	case domain.PriceNone:
		text := textutil.SanitizeListingText(o.FormatPrice(domain.PriceWords{}))
		if text == "" {
			return i18n.T(lang, "card.price_none")
		}
		return i18n.T(lang, "card.price", html.EscapeString(text))
	}
	return i18n.T(lang, "card.price", html.EscapeString(textutil.SanitizeListingText(o.FormatPrice(PriceWords(lang)))))
}

func cardTitle(o *domain.Offer, lang i18n.Lang) string {
	if title := textutil.SanitizeListingText(o.Title); title != "" {
		return title
	}
	return i18n.T(lang, "card.untitled")
}

func cardLocation(o *domain.Offer) string {
	if loc := textutil.SanitizeListingText(o.GetRegionCityDistrict()); loc != "" {
		return loc
	}
	return "Україна"
}

func contactLine(o *domain.Offer, lang i18n.Lang) string {
	if o.Contact.Phone {
		return i18n.T(lang, "card.contact_phone")
	}
	return i18n.T(lang, "card.contact_chat")
}

func cardDescription(o *domain.Offer, maxRunes int) string {
	desc := textutil.SanitizeListingDescription(textutil.StripHTMLTags(o.Description))
	if r := []rune(desc); len(r) > maxRunes {
		desc = strings.TrimSpace(string(r[:maxRunes])) + "…"
	}
	return desc
}

func isWebURL(u string) bool {
	return strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")
}

// OfferMarkup is the keyboard under an alert.
func OfferMarkup(o domain.Offer, lang i18n.Lang) *InlineKeyboardMarkup {
	if !isWebURL(o.URL) {
		return nil
	}
	return Markup(Row(URLBtn(i18n.T(lang, "btn.open"), o.URL)))
}

// BuildRichOffer renders the rich card:
//
//	photos (one image or a collage)
//	title
//	price sentence
//	description (expandable quote)
//	details (collapsible table)
//	location, publication time, contact line
//	footer: filter name and listing id
func BuildRichOffer(o domain.Offer, opt CardOptions, photos []string) string {
	var b strings.Builder
	lang := opt.Lang

	switch {
	case len(photos) == 1:
		b.WriteString(fmt.Sprintf("<img src=\"%s\"/>\n\n", html.EscapeString(photos[0])))
	case len(photos) > 1:
		b.WriteString("<tg-collage>\n")
		for _, u := range photos {
			b.WriteString(fmt.Sprintf("<img src=\"%s\"/>\n", html.EscapeString(u)))
		}
		b.WriteString("</tg-collage>\n\n")
	}

	b.WriteString("<h3>" + html.EscapeString(cardTitle(&o, lang)) + "</h3>\n")
	b.WriteString("<p>" + priceLine(&o, lang) + "</p>\n")

	if desc := cardDescription(&o, maxRichDescRunes); desc != "" {
		escaped := strings.ReplaceAll(html.EscapeString(desc), "\n", "<br>")
		b.WriteString("<blockquote expandable>" + escaped + "</blockquote>\n")
	}
	if opt.ShowParams {
		b.WriteString(buildParamsHTML(&o, opt.MaxParams, lang))
	}

	b.WriteString("<p>📍 " + html.EscapeString(cardLocation(&o)) + "</p>\n")
	if created, err := o.ParsedCreatedTime(); err == nil && !created.IsZero() {
		b.WriteString("<p>" + i18n.T(lang, "card.published", created.Format("15:04:05 02.01.2006")) + "</p>\n")
	}
	b.WriteString("<p>" + contactLine(&o, lang) + "</p>\n")

	filter := textutil.SanitizeListingText(opt.FilterName)
	if filter != "" || o.ID != 0 {
		footer := "OLX #" + strconv.FormatInt(o.ID, 10)
		if filter != "" {
			footer = i18n.T(lang, "card.filter_footer", html.EscapeString(filter))
			if o.ID != 0 {
				footer += " • OLX #" + strconv.FormatInt(o.ID, 10)
			}
		}
		b.WriteString("<footer>" + footer + "</footer>\n")
	}
	return b.String()
}

func buildParamsHTML(o *domain.Offer, maxParams int, lang i18n.Lang) string {
	params := o.DisplayParams()
	if len(params) == 0 {
		return ""
	}
	total := len(params)
	if maxParams <= 0 || maxParams > maxRichParams {
		maxParams = maxRichParams
	}
	if len(params) > maxParams {
		params = params[:maxParams]
	}

	var b strings.Builder
	b.WriteString("<details><summary>" + i18n.T(lang, "card.params", total) + "</summary>\n<table compact striped>\n")
	for _, p := range params {
		b.WriteString("<tr><td>" + html.EscapeString(textutil.TruncateRunes(p.Label, 40)) +
			"</td><td>" + html.EscapeString(textutil.TruncateRunes(p.Value, 90)) + "</td></tr>\n")
	}
	b.WriteString("</table>\n</details>\n")
	return b.String()
}

// BuildTextOffer renders the plain-HTML card. It always fits into one
// message: details are dropped from the end until it does.
func BuildTextOffer(o domain.Offer, opt CardOptions) string {
	params := []domain.ParamKV(nil)
	if opt.ShowParams {
		params = o.DisplayParams()
		if opt.MaxParams > 0 && len(params) > opt.MaxParams {
			params = params[:opt.MaxParams]
		}
	}
	for {
		text := buildTextOffer(&o, opt, params)
		if len([]rune(text)) <= maxTextRunes || len(params) == 0 {
			return text
		}
		params = params[:len(params)/2]
	}
}

func buildTextOffer(o *domain.Offer, opt CardOptions, params []domain.ParamKV) string {
	lang := opt.Lang
	var sb strings.Builder
	sb.WriteString(i18n.T(lang, "card.new") + "\n")
	sb.WriteString("<b>" + html.EscapeString(textutil.TruncateRunes(cardTitle(o, lang), 200)) + "</b>\n")
	sb.WriteString(priceLine(o, lang) + "\n")

	desc := cardDescription(o, maxTextDescRunes)
	if desc == "" {
		desc = i18n.T(lang, "card.no_description")
	}
	sb.WriteString("<blockquote expandable>" + html.EscapeString(desc) + "</blockquote>\n")

	if len(params) > 0 {
		sb.WriteString(i18n.T(lang, "card.params_title") + "\n")
		for _, p := range params {
			sb.WriteString(fmt.Sprintf("• %s: %s\n",
				html.EscapeString(textutil.TruncateRunes(p.Label, 40)), html.EscapeString(textutil.TruncateRunes(p.Value, 90))))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("📍 " + html.EscapeString(cardLocation(o)) + "\n")
	if created, err := o.ParsedCreatedTime(); err == nil && !created.IsZero() {
		sb.WriteString(i18n.T(lang, "card.published", created.Format("15:04:05 02.01.2006")) + "\n")
	}
	sb.WriteString(contactLine(o, lang) + "\n")
	if filter := textutil.SanitizeListingText(opt.FilterName); filter != "" {
		sb.WriteString(i18n.T(lang, "card.filter", html.EscapeString(textutil.TruncateRunes(filter, 80))) + "\n")
	}
	if isWebURL(o.URL) {
		sb.WriteString(fmt.Sprintf("🔗 <a href=\"%s\">%s</a>", html.EscapeString(o.URL), i18n.T(lang, "card.link")))
	}
	return strings.TrimSpace(sb.String())
}

func (c *Client) sendRich(ctx context.Context, chatID string, o domain.Offer, opt CardOptions, photos []string, km *InlineKeyboardMarkup) (int, error) {
	payload := map[string]any{
		"chat_id":      chatID,
		"rich_message": map[string]any{"html": BuildRichOffer(o, opt, photos)},
	}
	if opt.Silent {
		payload["disable_notification"] = true
	}
	if km != nil {
		payload["reply_markup"] = km
	}
	var msg TGMessage
	if err := c.Call(ctx, "sendRichMessage", payload, &msg); err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}

// richRejected reports that the server answered but would not take the rich
// card: it does not know the method, or it disliked this card. The plain card
// is the way out. Anything else — dead chat, flood limit, network trouble —
// would hit the plain card just the same and is left to the caller.
func richRejected(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || IsDeadChat(err) {
		return false
	}
	return apiErr.Code == http.StatusBadRequest || apiErr.Code == http.StatusNotFound
}

// SendOffer delivers one alert and returns the id of the card message.
//
// It makes at most one attempt per representation and never repeats a send
// whose outcome is unknown; retrying is the sender's decision.
func (c *Client) SendOffer(ctx context.Context, chatID string, o domain.Offer, opt CardOptions) (int, error) {
	km := OfferMarkup(o, opt.Lang)
	var photos []string
	if opt.MaxPhotos > 0 {
		photos = o.PhotoURLs(opt.MaxPhotos)
	}

	if opt.Rich && !c.plainPages.Load() {
		id, err := c.sendRich(ctx, chatID, o, opt, photos, km)
		if err != nil && len(photos) > 0 && isPhotoRelated(err) {
			logx.Warnf("[TG] photos of listing %d were not accepted (%v), sending the card without them", o.ID, err)
			id, err = c.sendRich(ctx, chatID, o, opt, nil, km)
		}
		if err == nil {
			return id, nil
		}
		if !richRejected(err) {
			return 0, err
		}
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			if !c.plainPages.Swap(true) {
				logx.Warnf("[TG] this Bot API server has no rich messages, alerts switch to plain cards")
			}
		} else {
			logx.Warnf("[TG] rich card of listing %d rejected (%v), sending the plain card", o.ID, err)
		}
	}

	id, err := c.SendHTML(ctx, chatID, BuildTextOffer(o, opt), km, opt.Silent)
	if err != nil {
		return 0, err
	}
	// The album is a bonus: the alert itself is already delivered.
	if len(photos) > 0 {
		if err := c.sendPhotos(ctx, chatID, photos); err != nil {
			logx.Warnf("[TG] photo album of listing %d was not sent to %s: %v", o.ID, chatID, err)
		}
	}
	return id, nil
}

type inputMediaPhoto struct {
	Type  string `json:"type"`
	Media string `json:"media"`
}

// sendPhotos sends a silent album that follows a plain card.
func (c *Client) sendPhotos(ctx context.Context, chatID string, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	if len(urls) == 1 {
		return c.Call(ctx, "sendPhoto", map[string]any{
			"chat_id":              chatID,
			"photo":                urls[0],
			"disable_notification": true,
		}, nil)
	}
	if len(urls) > 10 {
		urls = urls[:10]
	}
	media := make([]inputMediaPhoto, 0, len(urls))
	for _, u := range urls {
		media = append(media, inputMediaPhoto{Type: "photo", Media: u})
	}
	return c.Call(ctx, "sendMediaGroup", map[string]any{
		"chat_id":              chatID,
		"media":                media,
		"disable_notification": true,
	}, nil)
}

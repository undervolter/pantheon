package bot

import (
	"pantheon/core/olx"
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

func makeOfferMarkup(offer olx.Offer) *InlineKeyboardMarkup {
	return markup(row(InlineKeyboardButton{
		Text: "Перейти на лот",
		URL:  offer.URL,
	}))
}

// formatModerationDelay форматирует задержку модерации:
// меньше часа — мм:сс, иначе — чч:мм:сс.
func formatModerationDelay(d time.Duration) string {
	if d <= 0 {
		return "мгновенно"
	}
	d = d.Round(time.Second)
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	s := (d % time.Minute) / time.Second
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// buildRichOfferHTML собирает карточку объявления в строгом rich-формате:
//
// фото 1-10
//
// заголовок
// Продаётся за <цена> грн.
// описание (раскрываемая цитата)
// Область, Город - Район
// Опубликовано в чч:мм:сс дд.мм.гггг
// Только сообщения в OLX / Сообщения в OLX + номер
//
// кнопки: Перейти на лот | Позвонить продавцу
func buildRichOfferHTML(offer olx.Offer, filterName string, maxPhotos int, showParams bool, maxParams int) string {
	var b strings.Builder

	title := olx.SanitizeListingText(offer.Title)
	if title == "" {
		title = "Без названия"
	}

	// 1. Фото (до 10) — сверху карточки и в том же сообщении.
	urls := offer.PhotoURLs(maxPhotos)
	switch len(urls) {
	case 1:
		b.WriteString(fmt.Sprintf("<img src=\"%s\"/>\n\n", html.EscapeString(urls[0])))
	case 2, 3, 4, 5, 6, 7, 8, 9, 10:
		b.WriteString("<tg-collage>\n")
		for _, u := range urls {
			b.WriteString(fmt.Sprintf("<img src=\"%s\"/>\n", html.EscapeString(u)))
		}
		b.WriteString("</tg-collage>\n\n")
	}

	// 2. Заголовок объявления.
	b.WriteString("<h3>" + html.EscapeString(title) + "</h3>\n")

	// 3. Цена.
	price := olx.SanitizeListingText(offer.GetPrice())
	if price == "" {
		price = "Не вказана"
	}
	b.WriteString("<p>Продаётся за <b>" + html.EscapeString(price) + "</b></p>\n")

	// 4. Описание — раскрываемая цитата.
	desc := olx.SanitizeListingDescription(olx.StripHTMLTags(offer.Description))
	if desc != "" {
		escaped := html.EscapeString(desc)
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		b.WriteString("<blockquote expandable>" + escaped + "</blockquote>\n")
	}

	// 4b. Характеристики — сворачиваемая таблица.
	if showParams {
		b.WriteString(buildParamsHTML(offer, maxParams))
	}

	// 5. Местоположение: Область, Город - Район.
	loc := olx.SanitizeListingText(offer.GetRegionCityDistrict())
	if loc == "" {
		loc = "Україна"
	}
	b.WriteString("<p>📍 " + html.EscapeString(loc) + "</p>\n")

	// 6. Опубликовано.
	if created, err := offer.ParsedCreatedTime(); err == nil && !created.IsZero() {
		b.WriteString("<p>🕒 <b>Опубликовано в</b> " + created.Format("15:04:05 02.01.2006") + "</p>\n")
	}

	// 8. Контакты.
	contactLine := "💬 Только сообщения в OLX"
	if offer.Contact.Phone {
		contactLine = "💬 Сообщения в OLX + номер по запросу"
	}
	b.WriteString("<p>" + contactLine + "</p>\n")

	// 9. Служебная строка: фильтр и ID лота.
	filterStr := olx.SanitizeListingText(filterName)
	if filterStr != "" || offer.ID != 0 {
		footer := "Фильтр: " + html.EscapeString(filterStr)
		if filterStr == "" {
			footer = "OLX #" + strconv.FormatInt(offer.ID, 10)
		} else if offer.ID != 0 {
			footer += " • OLX #" + strconv.FormatInt(offer.ID, 10)
		}
		b.WriteString("<footer>" + footer + "</footer>\n")
	}

	return b.String()
}

// buildParamsHTML собирает сворачиваемую таблицу характеристик.
func buildParamsHTML(offer olx.Offer, maxParams int) string {
	params := offer.DisplayParams()
	if len(params) == 0 {
		return ""
	}
	total := len(params)
	if maxParams > 0 && len(params) > maxParams {
		params = params[:maxParams]
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<details><summary>Характеристики (%d)</summary>\n", total))
	b.WriteString("<table compact striped>\n")
	for _, p := range params {
		b.WriteString("<tr><td>" + html.EscapeString(olx.TruncateRunes(p.Label, 40)) +
			"</td><td>" + html.EscapeString(olx.TruncateRunes(p.Value, 90)) + "</td></tr>\n")
	}
	b.WriteString("</table>\n</details>\n")
	return b.String()
}

func (t *TelegramNotifier) sendRichOffer(chatID string, offer olx.Offer, filterName string, maxPhotos int, km *InlineKeyboardMarkup) error {
	if t.botToken == "" {
		return fmt.Errorf("telegram bot token пуст")
	}

	showParams := true
	maxParams := 0
	if t.cfg != nil {
		showParams = t.cfg.GetShowParams()
		maxParams = t.cfg.GetMaxParams()
	}

	payload := func(photos int) map[string]interface{} {
		p := map[string]interface{}{
			"chat_id": chatID,
			"rich_message": map[string]interface{}{
				"html": buildRichOfferHTML(offer, filterName, photos, showParams, maxParams),
			},
		}
		if km != nil {
			p["reply_markup"] = km
		}
		return p
	}

	err := t.postTelegramJSON("sendRichMessage", payload(maxPhotos))
	if err == nil {
		return nil
	}

	// Telegram не смог забрать фото (битый CDN/таймаут) — повторяем карточку без медиа.
	if maxPhotos >= 0 && isPhotoRelatedError(err) {
		fmt.Printf("[TG Rich] фото не приняты (%v), повторяю без фото\n", err)
		return t.postTelegramJSON("sendRichMessage", payload(-1))
	}
	return err
}

// isPhotoRelatedError — ошибки Telegram, связанные с загрузкой медиа.
func isPhotoRelatedError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "rich_message_photo") ||
		strings.Contains(s, "webpage_curl_failed") ||
		strings.Contains(s, "failed to get http url content") ||
		strings.Contains(s, "context deadline exceeded")
}

func testAlertOffer() olx.Offer {
	now := time.Now()
	price, _ := json.Marshal(map[string]interface{}{
		"value":      28000,
		"currency":   "UAH",
		"label":      "28 000 грн.",
		"arranged":   false,
		"budget":     false,
		"negotiable": false,
	})

	return olx.Offer{
		ID:              1160082,
		URL:             "https://www.olx.ua/d/uk/obyavlenie/prodam-iphone-15-pro-max-256-neverlock-ID116hSA.html",
		Title:           "iPhone 15 Pro Max 256GB Neverlock",
		Description:     "Продам свой iPhone 15 Pro Max на 256 ГБ в идеальном состоянии. С первого дня в чехле и с защитным стеклом. Без сколов, без ремонтов, батарея 98%. Полный заводской комплект: коробка, кабель, чек. Продажа при личной встрече в Киеве или отправка OLX Доставкой.",
		CreatedTime:     now.Add(-17 * time.Minute).Format(time.RFC3339),
		LastRefreshTime: now.Format(time.RFC3339),
		Location: olx.Location{
			City:     olx.City{ID: 268, Name: "Київ"},
			Region:   olx.Region{ID: 25, Name: "Київська область"},
			District: olx.District{ID: 11, Name: "Печерський"},
		},
		Params: []olx.Param{{
			Key:   "price",
			Name:  "Ціна",
			Type:  "price",
			Value: json.RawMessage(price),
		}},
		Photos: []olx.Photo{
			{Width: 750, Height: 1000, Link: "https://ireland.apollo.olxcdn.com/v1/files/0id9cnxtr6gn2-UA/image;s={width}x{height}"},
			{Width: 750, Height: 1000, Link: "https://ireland.apollo.olxcdn.com/v1/files/srpaym0pxest-UA/image;s={width}x{height}"},
			{Width: 750, Height: 1000, Link: "https://ireland.apollo.olxcdn.com/v1/files/a2rbfbzqxk2a3-UA/image;s={width}x{height}"},
		},
		Contact: olx.OfferContact{Phone: true, Chat: true},
		User:    olx.OfferUser{ID: 1, Name: "Продавец"},
	}
}

// SendAlertOffer отправляет rich-алерт по конкретному объявлению во все доверенные чаты.
func (t *TelegramNotifier) SendAlertOffer(offer olx.Offer, filterName string) {
	if t.botToken == "" {
		fmt.Println("[Test Alert] Токен бота не задан.")
		return
	}

	chats := t.chatIDs
	if t.cfg != nil {
		chats = t.cfg.GetTelegramChatIDs()
	}
	if len(chats) == 0 {
		fmt.Println("[Test Alert] Нет chat_id в конфиге.")
		return
	}

	km := makeOfferMarkup(offer)

	maxPhotos := 3
	if t.cfg != nil {
		maxPhotos = t.cfg.GetMaxPhotos()
	}

	for _, chatID := range chats {
		chatID = strings.TrimSpace(chatID)
		if chatID == "" {
			continue
		}

		if err := t.sendRichOffer(chatID, offer, filterName, maxPhotos, km); err != nil {
			fmt.Printf("[Test Alert] ❌ rich error для %s: %v\n", chatID, err)
			// Fallback: обычная карточка, чтобы тест не пропал.
			text := fmt.Sprintf("🧪 <b>Тестовый алерт (fallback)</b>\n<b>%s</b>\nПродаётся за <b>%s</b>\n📍 %s\n🔗 <a href=\"%s\">Перейти на лот</a>",
				html.EscapeString(offer.Title), html.EscapeString(offer.GetPrice()),
				html.EscapeString(offer.GetRegionCityDistrict()), html.EscapeString(offer.URL))
			if fallbackErr := t.sendMessage(chatID, text, km); fallbackErr != nil {
				fmt.Printf("[Test Alert] ❌ fallback тоже не ушёл в %s: %v\n", chatID, fallbackErr)
			} else {
				fmt.Printf("[Test Alert] ⚠️ отправлен fallback в %s\n", chatID)
			}
			continue
		}

		fmt.Printf("[Test Alert] ✅ rich-алерт отправлен в %s (%s)\n", chatID, offer.Title)
	}
}

// SendTestAlert отправляет тестовый rich-алерт на встроенном примере.
func (t *TelegramNotifier) SendTestAlert() {
	t.SendAlertOffer(testAlertOffer(), "Цифрові фотоапарати")
}

// SendTestAlertURL находит объявление по ссылке OLX и отправляет по нему алерт.
func (t *TelegramNotifier) SendTestAlertURL(client *olx.OLXClient, rawURL string) error {
	offer, err := client.FindOfferByURL(rawURL)
	if err != nil {
		return err
	}
	fmt.Printf("[Test Alert] Найдено объявление ID%d: %s\n", offer.ID, offer.Title)
	t.SendAlertOffer(offer, "Тест")
	return nil
}

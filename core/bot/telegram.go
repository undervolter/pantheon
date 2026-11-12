package bot

import (
	"pantheon/core/config"
	"pantheon/core/olx"
	"pantheon/core/checker"
	"pantheon/core/version"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
	Style        string `json:"style,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type TGUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type TGChat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type TGMessage struct {
	MessageID int    `json:"message_id"`
	From      TGUser `json:"from"`
	Chat      TGChat `json:"chat"`
	Text      string `json:"text"`
}

type CallbackQuery struct {
	ID      string     `json:"id"`
	From    TGUser     `json:"from"`
	Message *TGMessage `json:"message"`
	Data    string     `json:"data"`
}

type TGUpdate struct {
	UpdateID      int            `json:"update_id"`
	Message       *TGMessage     `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type TGUpdatesResponse struct {
	OK     bool       `json:"ok"`
	Result []TGUpdate `json:"result"`
}

type TelegramNotifier struct {
	botToken   string
	chatIDs    []string
	httpClient *http.Client
	cfg *config.Config
	ui         *BotUI
	sentCount  int64

	deadMu    sync.Mutex
	deadChats map[string]string // chat_id -> причина

	sendMu   sync.Mutex
	lastSend map[string]time.Time // chat_id -> время следующей разрешённой отправки
}

func NewTelegramNotifier(botToken string, chatIDs []string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatIDs:  chatIDs,
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
		deadChats: make(map[string]string),
		lastSend:  make(map[string]time.Time),
	}
}

func (t *TelegramNotifier) GetSentCount() int64 {
	return atomic.LoadInt64(&t.sentCount)
}

// Broadcast отправляет текст всем получателям (без привязки к фильтру).
func (t *TelegramNotifier) Broadcast(text string) {
	if t.botToken == "" {
		return
	}
	var chatIDs []string
	if t.cfg != nil {
		chatIDs = t.cfg.GetTelegramChatIDs()
	} else {
		chatIDs = t.chatIDs
	}
	for _, chatID := range chatIDs {
		chatID = strings.TrimSpace(chatID)
		if chatID == "" || t.isChatDead(chatID) {
			continue
		}
		_ = t.sendMessage(chatID, text, nil)
	}
}

func isDeadChatError(text string) bool {
	s := strings.ToLower(text)
	return strings.Contains(s, "user is deactivated") ||
		strings.Contains(s, "bot was blocked by the user") ||
		strings.Contains(s, "chat not found") ||
		strings.Contains(s, "bot can't initiate conversation") ||
		strings.Contains(s, "peer_id_invalid")
}

func (t *TelegramNotifier) isChatDead(chatID string) bool {
	t.deadMu.Lock()
	defer t.deadMu.Unlock()
	_, ok := t.deadChats[chatID]
	return ok
}

func (t *TelegramNotifier) markChatDead(chatID, reason string) {
	t.deadMu.Lock()
	if _, exists := t.deadChats[chatID]; exists {
		t.deadMu.Unlock()
		return
	}
	t.deadChats[chatID] = reason
	t.deadMu.Unlock()

	fmt.Printf("[TG] ❌ Чат %s отключён до перезапуска: %s\n", chatID, strings.TrimSpace(reason))
}

func (t *TelegramNotifier) reviveChat(chatID string) {
	t.deadMu.Lock()
	delete(t.deadChats, chatID)
	t.deadMu.Unlock()
}

func (t *TelegramNotifier) SetUI(ui *BotUI, cfg *config.Config) {
	t.ui = ui
	t.cfg = cfg
}

// filterChatIDsForOwner оставляет только чат-владелец фильтра.
// Если владелец не задан — возвращает всех получателей (общий фильтр).
// Если владельца нет в списке получателей — возвращает nil.
func filterChatIDsForOwner(chatIDs []string, ownerChatID string) []string {
	ownerChatID = strings.TrimSpace(ownerChatID)
	if ownerChatID == "" {
		return chatIDs
	}
	for _, id := range chatIDs {
		if strings.TrimSpace(id) == ownerChatID {
			return []string{ownerChatID}
		}
	}
	return nil
}

func (t *TelegramNotifier) NotifyNewOffer(offer olx.Offer, filterName, ownerChatID string) {
	if t.botToken == "" {
		return
	}

	var chatIDs []string
	if t.cfg != nil {
		chatIDs = t.cfg.GetTelegramChatIDs()
	} else {
		chatIDs = t.chatIDs
	}

	// Фильтр закреплён за конкретным чатом: алерты получает только он.
	// Пустой ownerChatID — старый/общий фильтр, шлём всем получателям.
	chatIDs = filterChatIDsForOwner(chatIDs, ownerChatID)

	if len(chatIDs) == 0 {
		return
	}

	showParams := true
	maxParams := 15
	if t.cfg != nil {
		showParams = t.cfg.GetShowParams()
		maxParams = t.cfg.GetMaxParams()
	}

	price := olx.SanitizeListingText(offer.GetPrice())
	if price == "" {
		price = "Не вказана"
	}
	desc := olx.SanitizeListingDescription(olx.StripHTMLTags(offer.Description))
	const maxDescLen = 450
	if len([]rune(desc)) > maxDescLen {
		desc = string([]rune(desc)[:maxDescLen]) + "..."
	}
	if desc == "" {
		desc = "Без опису"
	}

	title := olx.SanitizeListingText(offer.Title)
	if title == "" {
		title = "Без названия"
	}
	loc := olx.SanitizeListingText(offer.GetRegionCityDistrict())
	if loc == "" {
		loc = "Україна"
	}

	var sb strings.Builder
	sb.WriteString("🔥 <b>Новое объявление!</b>\n")
	sb.WriteString(fmt.Sprintf("<b>%s</b>\n", html.EscapeString(title)))
	sb.WriteString(fmt.Sprintf("Продаётся за <b>%s</b>\n", html.EscapeString(price)))
	sb.WriteString(fmt.Sprintf("<blockquote expandable>%s</blockquote>\n", html.EscapeString(desc)))

	if showParams {
		if params := offer.DisplayParams(); len(params) > 0 {
			if maxParams > 0 && len(params) > maxParams {
				params = params[:maxParams]
			}
			sb.WriteString("📋 <b>Характеристики:</b>\n")
			for _, p := range params {
				sb.WriteString(fmt.Sprintf("• %s: %s\n", html.EscapeString(p.Label), html.EscapeString(p.Value)))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("📍 %s\n", html.EscapeString(loc)))

	if created, err := offer.ParsedCreatedTime(); err == nil && !created.IsZero() {
		sb.WriteString(fmt.Sprintf("🕒 <b>Опубликовано в</b> %s\n", created.Format("15:04:05 02.01.2006")))
	}

	if offer.Contact.Phone {
		sb.WriteString("📞 <b>Связь:</b> телефон разрешён на OLX\n")
	} else {
		sb.WriteString("💬 <b>Связь:</b> только сообщения в OLX\n")
	}

	if filterName != "" {
		sb.WriteString(fmt.Sprintf("🎯 <b>Фильтр:</b> %s\n", html.EscapeString(filterName)))
	}
	sb.WriteString(fmt.Sprintf("🔗 <a href=\"%s\">Перейти на лот</a>", html.EscapeString(offer.URL)))

	text := sb.String()

	km := makeOfferMarkup(offer)

	sendPhotos := true
	maxPhotos := 3
	if t.cfg != nil {
		sendPhotos = t.cfg.GetSendPhotos()
		maxPhotos = t.cfg.GetMaxPhotos()
	}
	var photoURLs []string
	if sendPhotos {
		photoURLs = offer.PhotoURLs(maxPhotos)
	}

	richAlerts := true
	if t.cfg != nil {
		richAlerts = t.cfg.GetRichAlerts()
	}
	richPhotos := maxPhotos
	if !sendPhotos {
		richPhotos = -1
	}

	for _, chatID := range chatIDs {
		chatID = strings.TrimSpace(chatID)
		if chatID == "" || t.isChatDead(chatID) {
			continue
		}

		go func(cid string) {
			if t.isChatDead(cid) {
				return
			}

			if richAlerts {
				if err := t.sendRichOffer(cid, offer, filterName, richPhotos, km); err == nil {
					atomic.AddInt64(&t.sentCount, 1)
					return
				} else if isDeadChatError(err.Error()) {
					t.markChatDead(cid, err.Error())
					return
				} else {
					fmt.Printf("[TG Rich] fallback в %s: %v\n", cid, err)
				}
			}

			// Гарантированная доставка: сначала текст карточки, без фото.
			if t.isChatDead(cid) {
				return
			}
			if err := t.sendMessage(cid, text, km); err != nil {
				if !t.isChatDead(cid) {
					fmt.Printf("[TG Error] Не удалось отправить в %s: %v\n", cid, err)
				}
				return
			}
			atomic.AddInt64(&t.sentCount, 1)

			// Фото — необязательный бонус, ошибки не мешают доставке.
			if len(photoURLs) > 0 && !t.isChatDead(cid) {
				if err := t.sendOfferPhotos(cid, photoURLs); err != nil && !isDeadChatError(err.Error()) {
					fmt.Printf("[TG Photos] альбом не ушёл в %s: %v\n", cid, err)
				}
			}
		}(chatID)
	}
}

func (t *TelegramNotifier) SetupBotMenu() error {
	commands := []map[string]string{
		{"command": "start", "description": "🎮 Открыть пульт Pantheon"},
	}

	_ = t.postTelegramJSON("setMyCommands", map[string]interface{}{
		"commands": commands,
	})

	return t.postTelegramJSON("setChatMenuButton", map[string]interface{}{
		"menu_button": map[string]string{
			"type": "commands",
		},
	})
}

func (t *TelegramNotifier) StartPolling(checker *checker.Checker) {
	if t.botToken == "" {
		fmt.Println("[TG Bot] Токен Telegram не задан. Бот управления отключен.")
		return
	}

	_ = t.SetupBotMenu()
	fmt.Println("🤖 Интерактивный Telegram-пульт управления (кнопки) запущен...")

	// Сразу показываем инлайн-меню во всех привязанных чатах,
	// чтобы не искать /start после каждого запуска.
	if t.ui != nil && t.cfg != nil {
		for _, chatID := range t.cfg.GetTelegramChatIDs() {
			chatID = strings.TrimSpace(chatID)
			if chatID == "" {
				continue
			}
			go func(cid string) {
				text, km := t.ui.RenderMainMenu(cid)
				_ = t.sendMessage(cid, text, km)
			}(chatID)
		}
	}

	offset := 0
	for {
		updates, err := t.getUpdates(offset)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			if update.CallbackQuery != nil {
				t.handleCallback(update.CallbackQuery)
			} else if update.Message != nil {
				t.handleMessage(update.Message)
			}
		}
	}
}

func (t *TelegramNotifier) isAuthorized(chatID string) bool {
	if t.cfg == nil {
		return true
	}
	if t.cfg.IsAuthorized(chatID) {
		return true
	}

	// Группы/каналы не авто-привязываем: только явный /bind или конфиг.
	if strings.HasPrefix(chatID, "-") {
		return false
	}

	// Автоматически авторизуем тебя и твоего кента (до 2 пользователей)
	added, err := t.cfg.AutoAuthorizeChatID(chatID, 2)
	if err != nil {
		return false
	}
	if added {
		t.chatIDs = t.cfg.GetTelegramChatIDs()
		fmt.Printf("[TG Bot] 👤 Успешно привязан новый пользователь! chat_id: %s\n", chatID)
	}
	return added
}

func (t *TelegramNotifier) handleMessage(msg *TGMessage) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	text := strings.TrimSpace(msg.Text)
	isGroup := msg.Chat.Type == "group" || msg.Chat.Type == "supergroup"
	fromID := strconv.FormatInt(msg.From.ID, 10)

	if isGroup {
		if t.cfg == nil {
			return
		}
		// В группах слушаем только владельца: остальные участники не должны
		// дёргать фильтры и кнопки.
		if msg.From.ID == 0 || !t.cfg.IsAuthorized(fromID) {
			return
		}
	} else if !t.isAuthorized(chatID) {
		_ = t.sendMessage(chatID, "⛔ Доступ запрещен. Ваш ID не в списке доверенных.", nil)
		return
	}
	t.reviveChat(chatID)

	st := t.ui.getState(chatID)

	// /addchat и /delchat обрабатываем до switch: у них есть аргумент.
	if strings.HasPrefix(text, "/addchat") {
		if isGroup {
			_ = t.sendMessage(chatID, "В группе используй /bind.", nil)
			return
		}
		arg := strings.TrimSpace(strings.TrimPrefix(text, "/addchat"))
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil || id == 0 {
			_ = t.sendMessage(chatID, "Формат: <code>/addchat -1001234567890</code>\n\nID канала/группы можно узнать командой /chatid в этом чате.", nil)
			return
		}
		idStr := strconv.FormatInt(id, 10)
		added, err := t.cfg.AddTrustedChatID(idStr, 0)
		if err != nil {
			_ = t.sendMessage(chatID, fmt.Sprintf("❌ Не удалось сохранить: %v", err), nil)
			return
		}
		if !added {
			_ = t.sendMessage(chatID, "Этот чат уже в списке получателей.", nil)
			return
		}
		t.chatIDs = t.cfg.GetTelegramChatIDs()
		if err := t.sendMessage(idStr, "✅ <b>Pantheon подключён к этому чату.</b>", nil); err != nil {
			_ = t.sendMessage(chatID, fmt.Sprintf("⚠️ Чат добавлен, но тест не прошёл: %v\nПроверь, что бот добавлен и имеет право писать.", err), nil)
		} else {
			_ = t.sendMessage(chatID, fmt.Sprintf("✅ Чат <code>%s</code> добавлен, тестовое сообщение отправлено.", idStr), nil)
		}
		return
	}

	if strings.HasPrefix(text, "/delchat") {
		if isGroup {
			_ = t.sendMessage(chatID, "В группе используй /unbind.", nil)
			return
		}
		arg := strings.TrimSpace(strings.TrimPrefix(text, "/delchat"))
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil || id == 0 || id == msg.Chat.ID {
			_ = t.sendMessage(chatID, "Формат: <code>/delchat -1001234567890</code> (текущий чат удалить нельзя)", nil)
			return
		}
		removed, err := t.cfg.RemoveChatID(strconv.FormatInt(id, 10))
		if err != nil {
			_ = t.sendMessage(chatID, fmt.Sprintf("❌ Не удалось сохранить: %v", err), nil)
			return
		}
		t.chatIDs = t.cfg.GetTelegramChatIDs()
		if removed {
			_ = t.sendMessage(chatID, "🗑 Чат убран из получателей.", nil)
		} else {
			_ = t.sendMessage(chatID, "Такого чата нет в списке.", nil)
		}
		return
	}

	// 1. Команды из меню бота
	switch text {
	case "/chatid":
		kind := msg.Chat.Type
		if kind == "" {
			kind = "private"
		}
		reply := fmt.Sprintf("🆔 <b>chat_id:</b> <code>%s</code>\nТип: <b>%s</b>", chatID, html.EscapeString(kind))
		if msg.Chat.Title != "" {
			reply += "\nНазвание: " + html.EscapeString(msg.Chat.Title)
		}
		_ = t.sendMessage(chatID, reply, nil)
		return

	case "/bind":
		if !isGroup {
			_ = t.sendMessage(chatID, "Эту команду нужно отправить в группе, куда добавлен бот.", nil)
			return
		}
		if t.cfg == nil || !t.cfg.IsAuthorized(fromID) {
			_ = t.sendMessage(chatID, "⛔ Привязать чат может только владелец бота.", nil)
			return
		}
		added, err := t.cfg.AddTrustedChatID(chatID, 0)
		if err != nil {
			_ = t.sendMessage(chatID, fmt.Sprintf("❌ Не удалось сохранить: %v", err), nil)
			return
		}
		if added {
			t.chatIDs = t.cfg.GetTelegramChatIDs()
			_ = t.sendMessage(chatID, "✅ <b>Группа привязана.</b> Теперь сюда будут приходить алерты.", nil)
		} else {
			_ = t.sendMessage(chatID, "Эта группа уже в списке получателей.", nil)
		}
		return

	case "/unbind":
		if !isGroup {
			_ = t.sendMessage(chatID, "Эту команду нужно отправить в группе.", nil)
			return
		}
		if t.cfg == nil || !t.cfg.IsAuthorized(fromID) {
			_ = t.sendMessage(chatID, "⛔ Отвязать чат может только владелец бота.", nil)
			return
		}
		removed, err := t.cfg.RemoveChatID(chatID)
		if err != nil {
			_ = t.sendMessage(chatID, fmt.Sprintf("❌ Не удалось сохранить: %v", err), nil)
			return
		}
		t.chatIDs = t.cfg.GetTelegramChatIDs()
		if removed {
			_ = t.sendMessage(chatID, "🗑 Группа убрана из получателей алертов.", nil)
		} else {
			_ = t.sendMessage(chatID, "Эта группа и так не была привязана.", nil)
		}
		return

	case "/chats":
		if t.cfg == nil {
			return
		}
		ids := t.cfg.GetTelegramChatIDs()
		var sb strings.Builder
		sb.WriteString("📤 <b>Получатели алертов:</b>\n")
		for i, id := range ids {
			sb.WriteString(fmt.Sprintf("%d. <code>%s</code>\n", i+1, html.EscapeString(id)))
		}
		sb.WriteString("\n<b>Группа:</b> /bind в группе или /addchat &lt;id&gt;\n<b>Удалить:</b> /unbind в группе или /delchat &lt;id&gt;")
		_ = t.sendMessage(chatID, sb.String(), nil)
		return

	case "/start", "/menu":
		t.ui.resetState(chatID)
		menuText, km := t.ui.RenderMainMenu(chatID)
		_ = t.sendMessage(chatID, menuText, km)
		return

	case "/filters":
		t.ui.resetState(chatID)
		listText, km := t.ui.RenderFiltersList(chatID)
		_ = t.sendMessage(chatID, listText, km)
		return

	case "/add":
		st.Draft = config.SearchConfig{}
		st.SelectedBrands = map[string]string{}
		st.Step = StepNone
		catText, km := t.ui.RenderCategoryStep()
		_ = t.sendMessage(chatID, catText, km)
		return

	case "/check":
		if t.ui.checker != nil {
			go t.ui.checker.CheckNow()
		}
		_ = t.sendMessage(chatID, "⚡ <b>Внеочередная проверка запущена!</b>", nil)
		return

	case "/status":
		t.ui.resetState(chatID)
		menuText, km := t.ui.RenderMainMenu(chatID)
		_ = t.sendMessage(chatID, menuText, km)
		return

	case "/version":
		text := fmt.Sprintf("🐉 <b>pantheon open-core</b> v%s", version.AppVersion)
		_ = t.sendMessage(chatID, text, nil)
		return
	}

	// 2. Шаг поиска города/области (autocomplete)
	if st.Step == StepWaitingCity {
		t.handleCityQuery(chatID, text)
		return
	}

	// 3. Распознавание ссылок OLX
	if strings.Contains(text, "olx.ua") && (strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")) {
		sc, err := olx.ParseOLXURL(text)
		if err == nil {
			st.Step = StepNone
			st.Draft = *sc
			confText, km := t.ui.RenderConfirmStep(st.Draft)
			_ = t.sendMessage(chatID, "🔗 <b>Распознана ссылка OLX!</b>\n\n"+confText, km)
			return
		}
	}

	// 3. Шаг ввода поискового запроса
	if st.Step == StepWaitingQuery {
		st.Draft.Query = strings.TrimSpace(text)
		st.Draft.QueryVariants = olx.QueryVariants(text)
		if bd := brandDisplay(st.Draft); bd != "" {
			st.Draft.Name = bd + " " + st.Draft.Query
		} else {
			st.Draft.Name = st.Draft.Query
		}
		st.Step = StepNone
		variantsText, km := t.ui.RenderVariantsStep(st.Draft)
		_ = t.sendMessage(chatID, variantsText, km)
		return
	}

	// 3b. Шаг добавления своего варианта запроса
	if st.Step == StepWaitingVariant {
		norm := olx.NormalizeQuery(text)
		if norm == "" {
			_ = t.sendMessage(chatID, "❌ Пустой вариант. Напиши текст, например: <code>iphone 4s</code> или <code>4s</code>.", nil)
			return
		}
		if olx.StringInSlice(st.Draft.QueryVariants, norm) {
			_ = t.sendMessage(chatID, fmt.Sprintf(
				"ℹ️ Вариант «%s» уже покрыт формой <code>%s</code> — отдельно добавлять не нужно.",
				html.EscapeString(strings.TrimSpace(text)), html.EscapeString(norm)), nil)
		} else {
			st.Draft.QueryVariants = append(st.Draft.QueryVariants, norm)
		}
		st.Step = StepNone
		view, km := t.ui.RenderVariantsStep(st.Draft)
		_ = t.sendMessage(chatID, view, km)
		return
	}

	// 4. Шаг ввода цены: одно число (максимум) или диапазон (от–до).
	if st.Step == StepWaitingCustomPrice || st.Step == StepWaitingPriceRange {
		minP, maxP, ok := parsePriceRange(text)
		if !ok || (minP <= 0 && maxP <= 0) {
			_ = t.sendMessage(chatID, "❌ Не понял цену. Примеры: <code>25000</code>, <code>5000-20000</code>, <code>от 5000 до 20000</code>.", nil)
			return
		}
		if maxP > 0 && minP > maxP {
			minP, maxP = maxP, minP
		}
		st.Draft.MinPrice = minP
		st.Draft.MaxPrice = maxP
		st.Step = StepNone
		cityText, km := t.ui.RenderOwnerStep(st.Draft)
		_ = t.sendMessage(chatID, cityText, km)
		return
	}

	// 4b. Шаг ввода значения числового фильтра (год, пробег, объём и т.п.).
	if st.Step == StepWaitingFilterValue {
		defs := t.ui.categoryFilters(st.Draft.Category, st.Draft.BrandParam)
		if st.FilterIdx < 0 || st.FilterIdx >= len(defs) {
			st.Step = StepNone
			_ = t.sendMessage(chatID, "❌ Фильтр не найден, начни заново.", nil)
			return
		}
		def := defs[st.FilterIdx]
		minP, maxP, ok := parsePriceRange(text)
		if !ok || (minP <= 0 && maxP <= 0) {
			_ = t.sendMessage(chatID, "❌ Не понял значение. Примеры: <code>2000-2010</code>, <code>от 100</code>, <code>1.6-2.0</code>.", nil)
			return
		}
		if maxP > 0 && minP > maxP {
			minP, maxP = maxP, minP
		}
		if st.Draft.Ranges == nil {
			st.Draft.Ranges = map[string]config.FilterRange{}
		}
		st.Draft.Ranges[def.Key] = config.FilterRange{From: minP, To: maxP}
		st.Step = StepNone
		view, km := t.ui.RenderFilterValues(st.Draft, st.FilterIdx, 0)
		_ = t.sendMessage(chatID, view, km)
		return
	}


	// По умолчанию выдаем главное меню
	menuText, km := t.ui.RenderMainMenu(chatID)
	_ = t.sendMessage(chatID, menuText, km)
}

func (t *TelegramNotifier) handleCityQuery(chatID, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		_ = t.sendMessage(chatID, "🏙 Напиши название города или области текстом.", nil)
		return
	}

	results, err := t.ui.checker.Client().SearchLocations(query)
	if err != nil {
		_ = t.sendMessage(chatID, fmt.Sprintf("❌ Не удалось получить список городов: %v", err), nil)
		return
	}
	if len(results) == 0 {
		_ = t.sendMessage(chatID, "❌ Ничего не найдено. Попробуй другое название или напиши область.", nil)
		return
	}

	text, km := t.ui.RenderCityResults(query, results)
	_ = t.sendMessage(chatID, text, km)
}

// applyCitySelection применяет город к черновику и, если у города есть районы,
// показывает шаг выбора района. Иначе сразу ведёт на подтверждение.
func (t *TelegramNotifier) applyCitySelection(chatID string, msgID, cityID, regionID int, cityName string) {
	st := t.ui.getState(chatID)
	st.Draft.CityID = cityID
	st.Draft.RegionID = regionID
	st.Draft.CityName = cityName
	st.Draft.DistrictID = 0
	st.Draft.DistrictName = ""

	if cityID > 0 {
		districts, err := t.ui.checker.Client().FetchCityDistricts(cityID)
		if err == nil && len(districts) > 0 {
			st.Step = StepNone
			text, km := t.ui.RenderDistrictStep(strings.TrimPrefix(cityName, "📍 "), districts, cityID, regionID)
			_ = t.editMessageText(chatID, msgID, text, km)
			return
		}
	}

	st.Step = StepNone
	text, km := t.ui.RenderConfirmStep(st.Draft)
	_ = t.editMessageText(chatID, msgID, text, km)
}

// ---------- мультивыбор брендов ----------

func selectedBrandSet(st *UserWizardState) map[string]bool {
	res := make(map[string]bool, len(st.SelectedBrands))
	for id := range st.SelectedBrands {
		res[id] = true
	}
	return res
}

func seedBrandSelection(st *UserWizardState) {
	if st.SelectedBrands == nil {
		st.SelectedBrands = map[string]string{}
	}
	if len(st.SelectedBrands) > 0 {
		return
	}
	for i, id := range st.Draft.BrandValues {
		name := ""
		if i < len(st.Draft.BrandNames) {
			name = st.Draft.BrandNames[i]
		}
		st.SelectedBrands[id] = name
	}
	if st.Draft.BrandValue != "" {
		st.SelectedBrands[st.Draft.BrandValue] = st.Draft.BrandName
	}
}

func setBrandDraftFromSelection(st *UserWizardState, catID int) {
	st.Draft.BrandParam = ""
	st.Draft.BrandValue = ""
	st.Draft.BrandName = ""
	st.Draft.BrandValues = nil
	st.Draft.BrandNames = nil

	info := olx.GetCategoryBrands(catID)
	if info == nil {
		return
	}
	st.Draft.BrandParam = info.ParamName
	for _, b := range info.Brands {
		if _, ok := st.SelectedBrands[b.ID]; ok {
			st.Draft.BrandValues = append(st.Draft.BrandValues, b.ID)
			st.Draft.BrandNames = append(st.Draft.BrandNames, b.Name)
		}
	}
	if len(st.Draft.BrandValues) == 0 {
		st.Draft.BrandParam = ""
		return
	}
	if len(st.Draft.BrandValues) == 1 {
		st.Draft.BrandValue = st.Draft.BrandValues[0]
		st.Draft.BrandName = st.Draft.BrandNames[0]
	}
}

// startBrandStep сбрасывает прошлый выбор и показывает бренды категории.
func (t *TelegramNotifier) startBrandStep(chatID string, msgID, catID int) {
	st := t.ui.getState(chatID)
	st.SelectedBrands = map[string]string{}
	st.Draft.BrandParam = ""
	st.Draft.BrandValue = ""
	st.Draft.BrandName = ""
	st.Draft.BrandValues = nil
	st.Draft.BrandNames = nil

	if info := olx.GetCategoryBrands(catID); info != nil && len(info.Brands) > 0 {
		text, km := t.ui.RenderBrandStep(catID, 0, selectedBrandSet(st))
		_ = t.editMessageText(chatID, msgID, text, km)
		return
	}

	st.Step = StepWaitingQuery
	text, km := t.ui.RenderQueryStep(catID, st.Draft.CatName, "")
	_ = t.editMessageText(chatID, msgID, text, km)
}

func (t *TelegramNotifier) finishBrandStep(chatID string, msgID, catID int) {
	st := t.ui.getState(chatID)
	setBrandDraftFromSelection(st, catID)
	st.Step = StepWaitingQuery
	text, km := t.ui.RenderQueryStep(st.Draft.Category, st.Draft.CatName, brandDisplay(st.Draft))
	_ = t.editMessageText(chatID, msgID, text, km)
}

func (t *TelegramNotifier) handleCallback(cq *CallbackQuery) {
	authID := strconv.FormatInt(cq.From.ID, 10)
	if !t.isAuthorized(authID) {
		_ = t.answerCallbackAlert(cq.ID, "⛔ Доступ запрещен!")
		return
	}
	chatID := authID
	if cq.Message != nil && cq.Message.Chat.ID != 0 {
		chatID = strconv.FormatInt(cq.Message.Chat.ID, 10)
	}
	t.reviveChat(chatID)

	answered := false
	answerToast := func(text string) {
		if !answered {
			answered = true
			_ = t.answerCallbackQuery(cq.ID, text)
		}
	}
	answerAlert := func(text string) {
		if !answered {
			answered = true
			_ = t.answerCallbackAlert(cq.ID, text)
		}
	}
	defer func() {
		if !answered {
			_ = t.answerCallbackQuery(cq.ID, "")
		}
	}()

	if cq.Message == nil {
		// Callback без сообщения (inline-режим) — редактировать нечего.
		return
	}

	msgID := cq.Message.MessageID
	data := cq.Data
	st := t.ui.getState(chatID)

	switch {
	case data == "menu_main":
		t.ui.resetState(chatID)
		text, km := t.ui.RenderMainMenu(chatID)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "menu_list":
		t.ui.resetState(chatID)
		text, km := t.ui.RenderFiltersList(chatID)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "f_view:"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "f_view:"))
		if _, ok := t.ui.searchForChat(chatID, idx); !ok {
			answerAlert("⛔ Этот фильтр принадлежит другому пользователю")
			break
		}
		text, km := t.ui.RenderFilterDetail(chatID, idx)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "f_toggle:"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "f_toggle:"))
		if _, ok := t.ui.searchForChat(chatID, idx); !ok {
			answerAlert("⛔ Этот фильтр принадлежит другому пользователю")
			break
		}
		state, err := t.cfg.ToggleSearch(idx)
		if err == nil {
			msg := "⏸ Фильтр на паузе"
			if state {
				msg = "▶️ Фильтр активен"
			}
			answerToast(msg)
		}
		text, km := t.ui.RenderFilterDetail(chatID, idx)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "f_del:"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "f_del:"))
		if _, ok := t.ui.searchForChat(chatID, idx); !ok {
			answerAlert("⛔ Этот фильтр принадлежит другому пользователю")
			break
		}
		err := t.cfg.RemoveSearch(idx)
		if err == nil {
			answerToast("🗑 Фильтр удален!")
		} else {
			answerToast("❌ Ошибка удаления")
		}
		text, km := t.ui.RenderFiltersList(chatID)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "f_claim:"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "f_claim:"))
		s, ok := t.ui.searchForChat(chatID, idx)
		if !ok || s.OwnerChatID != "" {
			answerAlert("⚠️ Фильтр уже закреплён за другим пользователем")
			break
		}
		if err := t.cfg.SetSearchOwner(idx, chatID); err != nil {
			answerAlert(fmt.Sprintf("❌ Не удалось сохранить: %v", err))
			break
		}
		answerToast("👤 Фильтр закреплён за этим чатом")
		text, km := t.ui.RenderFilterDetail(chatID, idx)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "f_share:"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "f_share:"))
		s, ok := t.ui.searchForChat(chatID, idx)
		if !ok || s.OwnerChatID != chatID {
			answerAlert("⛔ Это не твой фильтр")
			break
		}
		if err := t.cfg.SetSearchOwner(idx, ""); err != nil {
			answerAlert(fmt.Sprintf("❌ Не удалось сохранить: %v", err))
			break
		}
		answerToast("👥 Фильтр снова общий")
		text, km := t.ui.RenderFilterDetail(chatID, idx)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "add_start":
		st.Draft = config.SearchConfig{}
		st.SelectedBrands = map[string]string{}
		st.Step = StepNone
		text, km := t.ui.RenderCategoryStep()
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_cat:"):
		catID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_cat:"))
		if catID == 0 {
			// Любая категория
			st.Draft.Category = 0
			st.Draft.CatName = "Любая"
			st.Draft.BrandParam = ""
			st.Draft.BrandValue = ""
			st.Draft.BrandName = ""
			st.Draft.BrandValues = nil
			st.Draft.BrandNames = nil
			st.SelectedBrands = map[string]string{}
			st.Step = StepWaitingQuery
			text, km := t.ui.RenderQueryStep(0, "Любая", "")
			_ = t.editMessageText(chatID, msgID, text, km)
		} else {
			st.ParentCat = catID
			c := olx.FindCategory(catID)
			if c != nil && len(c.Children) > 0 {
				text, km := t.ui.RenderSubcategoryStep(catID)
				_ = t.editMessageText(chatID, msgID, text, km)
			} else {
				st.Draft.Category = catID
				if c != nil {
					st.Draft.CatName = c.Name
				}
				t.startBrandStep(chatID, msgID, catID)
			}
		}

	case strings.HasPrefix(data, "add_cat_select:"):
		catID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_cat_select:"))
		st.Draft.Category = catID
		if c := olx.FindCategory(catID); c != nil {
			st.Draft.CatName = c.Name
		}
		t.startBrandStep(chatID, msgID, catID)

	case strings.HasPrefix(data, "add_subcat:"):
		subID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_subcat:"))

		// Если у подкатегории есть свои подкатегории — идём глубже,
		// а не притворяемся, что это лист.
		if c := olx.FindCategory(subID); c != nil && len(c.Children) > 0 {
			st.ParentCat = subID
			text, km := t.ui.RenderSubcategoryStep(subID)
			_ = t.editMessageText(chatID, msgID, text, km)
			return
		}

		st.Draft.Category = subID
		if c := olx.FindCategory(subID); c != nil {
			st.Draft.CatName = c.Name
		}
		t.startBrandStep(chatID, msgID, subID)

	case strings.HasPrefix(data, "add_cat_back:"):
		parentID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_cat_back:"))
		st.Step = StepNone
		st.ParentCat = parentID
		if parentID > 0 {
			text, km := t.ui.RenderSubcategoryStep(parentID)
			_ = t.editMessageText(chatID, msgID, text, km)
		} else {
			text, km := t.ui.RenderCategoryStep()
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "add_brand_toggle:"):
		parts := strings.Split(strings.TrimPrefix(data, "add_brand_toggle:"), ":")
		if len(parts) == 3 {
			catID, _ := strconv.Atoi(parts[0])
			page, _ := strconv.Atoi(parts[1])
			brandID := parts[2]

			seedBrandSelection(st)
			if _, ok := st.SelectedBrands[brandID]; ok {
				delete(st.SelectedBrands, brandID)
			} else {
				name := brandID
				if b := olx.FindBrand(catID, brandID); b != nil {
					name = b.Name
				}
				st.SelectedBrands[brandID] = name
			}

			text, km := t.ui.RenderBrandStep(catID, page, selectedBrandSet(st))
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "add_brand_done:"):
		catID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_brand_done:"))
		t.finishBrandStep(chatID, msgID, catID)

	case strings.HasPrefix(data, "add_brand_page:"):
		parts := strings.Split(strings.TrimPrefix(data, "add_brand_page:"), ":")
		if len(parts) == 2 {
			catID, _ := strconv.Atoi(parts[0])
			page, _ := strconv.Atoi(parts[1])
			seedBrandSelection(st)
			text, km := t.ui.RenderBrandStep(catID, page, selectedBrandSet(st))
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "add_brand_skip:"):
		catID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_brand_skip:"))
		st.SelectedBrands = map[string]string{}
		t.finishBrandStep(chatID, msgID, catID)

	case strings.HasPrefix(data, "add_brand:"):
		// Совместимость со старыми сообщениями: одиночный выбор.
		parts := strings.SplitN(strings.TrimPrefix(data, "add_brand:"), ":", 2)
		if len(parts) == 2 {
			catID, _ := strconv.Atoi(parts[0])
			brandID := parts[1]
			st.Draft.Category = catID
			if c := olx.FindCategory(catID); c != nil {
				st.Draft.CatName = c.Name
			}
			name := brandID
			if b := olx.FindBrand(catID, brandID); b != nil {
				name = b.Name
			}
			st.SelectedBrands = map[string]string{brandID: name}
			t.finishBrandStep(chatID, msgID, catID)
		}

	case data == "noop":
		// No action for page indicator button

	case data == "add_query_skip":
		st.Draft.Query = ""
		st.Draft.QueryVariants = nil
		if bd := brandDisplay(st.Draft); bd != "" {
			st.Draft.Name = bd
		}
		st.Step = StepNone
		text, km := t.ui.RenderPriceStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "variant_add":
		st.Step = StepWaitingVariant
		prompt := "🔤 <b>Напиши вариант запроса</b>\n\nНапример: <code>iphone 4s</code>, <code>айфон 4с</code>, <code>4s</code>.\nБот добавит его к текущему фильтру."
		_ = t.editMessageText(chatID, msgID, prompt, markup(row(btn("⬅️ Назад", "variants_back"))))

	case strings.HasPrefix(data, "variant_del:"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "variant_del:"))
		if idx >= 0 && idx < len(st.Draft.QueryVariants) {
			st.Draft.QueryVariants = append(st.Draft.QueryVariants[:idx], st.Draft.QueryVariants[idx+1:]...)
		}
		text, km := t.ui.RenderVariantsStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "variants_back":
		st.Step = StepNone
		text, km := t.ui.RenderVariantsStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "variants_done":
		st.Step = StepNone
		text, km := t.ui.RenderPriceStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_price:"):
		pVal, _ := strconv.ParseFloat(strings.TrimPrefix(data, "add_price:"), 64)
		st.Draft.MinPrice = 0
		st.Draft.MaxPrice = pVal
		st.Step = StepNone
		text, km := t.ui.RenderOwnerStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_owner:"):
		switch strings.TrimPrefix(data, "add_owner:") {
		case "private":
			st.Draft.OwnerType = "private"
		case "business":
			st.Draft.OwnerType = "business"
		default:
			st.Draft.OwnerType = ""
		}
		st.Step = StepNone
		text, km := t.ui.RenderCityStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "add_price_custom":
		st.Step = StepWaitingCustomPrice
		prompt := "💰 <b>Введи цену в гривнах</b>\n\nМожно одно число (максимум): <code>28500</code>\nИли диапазон: <code>5000-20000</code>"
		_ = t.editMessageText(chatID, msgID, prompt, markup(row(btn("❌ Отмена", "menu_main"))))

	case data == "add_price_range":
		st.Step = StepWaitingPriceRange
		prompt := "↔️ <b>Введи диапазон цены</b>\n\n" +
			"<code>5000-20000</code> — от 5 000 до 20 000 грн\n" +
			"<code>от 5000</code> — только нижняя граница\n" +
			"<code>до 20000</code> — только верхняя\n" +
			"<code>5000</code> — до 5 000 грн"
		_ = t.editMessageText(chatID, msgID, prompt, markup(row(btn("❌ Отмена", "menu_main"))))

	case data == "add_city_manual":
		st.Step = StepWaitingCity
		// Заранее греем индекс городов в фоне, чтобы поиск с опечатками
		// не заставлял ждать 25 запросов к OLX.
		go t.ui.checker.Client().WarmGeoIndex()
		prompt := "🏙 <b>Напиши город или область</b>\n\nНапример: <code>Кременчук</code>, <code>Бровари</code> или <code>Полтавська область</code>:"
		_ = t.editMessageText(chatID, msgID, prompt, markup(row(btn("❌ Отмена", "menu_main"))))

	case data == "add_city_back":
		st.Step = StepNone
		text, km := t.ui.RenderCityStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_city_pick:"):
		parts := strings.Split(strings.TrimPrefix(data, "add_city_pick:"), ":")
		if len(parts) == 2 {
			cityID, _ := strconv.Atoi(parts[0])
			regionID, _ := strconv.Atoi(parts[1])

			name := ""
			if c := olx.FindCity(cityID); c != nil {
				name = strings.TrimPrefix(c.Name, "📍 ")
			} else if city, err := t.ui.checker.Client().FetchCity(cityID); err == nil {
				name = city.Name
				if regionID == 0 {
					regionID = city.RegionID
				}
			}
			if name == "" {
				name = fmt.Sprintf("город %d", cityID)
			}
			t.applyCitySelection(chatID, msgID, cityID, regionID, "📍 "+name)
			return
		}
		st.Step = StepNone
		text, km := t.ui.RenderConfirmStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_region_pick:"):
		regionID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_region_pick:"))
		if regionID > 0 {
			st.Draft.CityID = 0
			st.Draft.RegionID = regionID
			st.Draft.DistrictID = 0
			st.Draft.DistrictName = ""

			name := ""
			if region, err := t.ui.checker.Client().FetchRegion(regionID); err == nil {
				name = region.Name
			}
			if name == "" {
				name = fmt.Sprintf("область %d", regionID)
			}
			st.Draft.CityName = "🌍 " + name
		}
		st.Step = StepNone
		text, km := t.ui.RenderConfirmStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_city_whole:"):
		parts := strings.Split(strings.TrimPrefix(data, "add_city_whole:"), ":")
		if len(parts) == 2 {
			cityID, _ := strconv.Atoi(parts[0])
			regionID, _ := strconv.Atoi(parts[1])
			st.Draft.CityID = cityID
			st.Draft.RegionID = regionID
			st.Draft.DistrictID = 0
			st.Draft.DistrictName = ""
		}
		st.Step = StepNone
		text, km := t.ui.RenderConfirmStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_district_pick:"):
		parts := strings.Split(strings.TrimPrefix(data, "add_district_pick:"), ":")
		if len(parts) == 3 {
			districtID, _ := strconv.Atoi(parts[0])
			cityID, _ := strconv.Atoi(parts[1])
			regionID, _ := strconv.Atoi(parts[2])

			st.Draft.DistrictID = districtID
			st.Draft.CityID = cityID
			st.Draft.RegionID = regionID

			name := ""
			if districts, err := t.ui.checker.Client().FetchCityDistricts(cityID); err == nil {
				for _, d := range districts {
					if d.ID == districtID {
						name = d.Name
						break
					}
				}
			}
			st.Draft.DistrictName = name
		}
		st.Step = StepNone
		text, km := t.ui.RenderConfirmStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "add_city:"):
		parts := strings.Split(strings.TrimPrefix(data, "add_city:"), ":")
		if len(parts) == 2 {
			cityID, _ := strconv.Atoi(parts[0])
			regionID, _ := strconv.Atoi(parts[1])

			name := "Вся Украина"
			if cityID > 0 {
				if c := olx.FindCity(cityID); c != nil {
					name = c.Name
				} else if city, err := t.ui.checker.Client().FetchCity(cityID); err == nil {
					name = "📍 " + city.Name
					if regionID == 0 {
						regionID = city.RegionID
					}
				}
			}
			t.applyCitySelection(chatID, msgID, cityID, regionID, name)
			return
		}
		st.Step = StepNone
		text, km := t.ui.RenderConfirmStep(st.Draft)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "add_finish":
		if st.Draft.Name == "" {
			if st.Draft.Query != "" {
				st.Draft.Name = st.Draft.Query
			} else if bd := brandDisplay(st.Draft); bd != "" {
				st.Draft.Name = bd
			} else if st.Draft.CatName != "" {
				st.Draft.Name = st.Draft.CatName
			} else {
				st.Draft.Name = "Новый поиск"
			}
		}
		// Уведомления по фильтру получает только чат, в котором его создали.
		st.Draft.OwnerChatID = chatID
		_ = t.cfg.AddSearch(st.Draft)
		t.ui.resetState(chatID)
		successMsg := fmt.Sprintf("✅ <b>Фильтр '%s' успешно создан и запущен!</b>\nЧекер уже мониторит свежие лоты.", st.Draft.Name)
		_ = t.sendMessage(chatID, successMsg, nil)
		text, km := t.ui.RenderMainMenu(chatID)
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "menu_settings":
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "check_now":
		go t.ui.checker.CheckNow()
		answerToast("⚡ Внеочередная проверка запущена")

	case data == "toggle_photos":
		newState := !t.cfg.GetSendPhotos()
		if err := t.cfg.SetSendPhotos(newState); err != nil {
			answerAlert(fmt.Sprintf("❌ Ошибка: %v", err))
			break
		}
		if newState {
			answerToast(fmt.Sprintf("📷 Фото включены: до %d шт.", t.cfg.GetMaxPhotos()))
		} else {
			answerToast("📷 Фото выключены: только текст")
		}
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "toggle_skip_bumped":
		newState := !t.cfg.GetSkipBumpedOld()
		if err := t.cfg.SetSkipBumpedOld(newState); err != nil {
			answerAlert(fmt.Sprintf("❌ Ошибка: %v", err))
			break
		}
		if newState {
			answerToast("♻️ Пропуск старых поднятых лотов")
		} else {
			answerToast("♻️ Поднятые старые лоты будут приходить")
		}
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "cycle_alert_age":
		steps := []int{0, 5, 10, 15, 20, 30, 60}
		cur := t.cfg.GetMaxAlertAgeMin()
		next := 5
		for i, v := range steps {
			if v == cur {
				next = steps[(i+1)%len(steps)]
				break
			}
		}
		if err := t.cfg.SetMaxAlertAgeMin(next); err != nil {
			answerAlert(fmt.Sprintf("❌ Ошибка: %v", err))
			break
		}
		if next == 0 {
			answerToast("⏳ Лимит возраста лота снят")
		} else {
			answerToast(fmt.Sprintf("⏳ Лоты не старше %d мин", next))
		}
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "toggle_params":
		newState := !t.cfg.GetShowParams()
		if err := t.cfg.SetShowParams(newState); err != nil {
			answerAlert(fmt.Sprintf("❌ Ошибка: %v", err))
			break
		}
		if newState {
			answerToast(fmt.Sprintf("📋 Характеристики: до %d полей", t.cfg.GetMaxParams()))
		} else {
			answerToast("📋 Характеристики выключены")
		}
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "cycle_fetch_limit":
		steps := []int{20, 35, 50}
		_, cur := t.cfg.PerfSettings()
		next := 20
		for i, v := range steps {
			if v == cur {
				next = steps[(i+1)%len(steps)]
				break
			}
		}
		if err := t.cfg.SetFetchLimit(next); err != nil {
			answerAlert(fmt.Sprintf("❌ Ошибка: %v", err))
			break
		}
		answerToast(fmt.Sprintf("📦 Лотов за цикл: %d", next))
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "filters_open:"):
		st.FilterListPage = 0
		text, km := t.ui.RenderFiltersStep(st.Draft, 0)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "filters_list:"), strings.HasPrefix(data, "filters_page:"):
		parts := strings.Split(data, ":")
		page := 0
		if len(parts) > 1 {
			page, _ = strconv.Atoi(parts[len(parts)-1])
		}
		st.FilterListPage = page
		text, km := t.ui.RenderFiltersStep(st.Draft, page)
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "filter_open:"):
		parts := strings.Split(strings.TrimPrefix(data, "filter_open:"), ":")
		if len(parts) >= 3 {
			idx, _ := strconv.Atoi(parts[1])
			listPage, _ := strconv.Atoi(parts[2])
			st.FilterIdx = idx
			st.FilterListPage = listPage
			st.FilterValPage = 0
			text, km := t.ui.RenderFilterValues(st.Draft, idx, 0)
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "filter_page:"):
		parts := strings.Split(strings.TrimPrefix(data, "filter_page:"), ":")
		if len(parts) >= 3 {
			idx, _ := strconv.Atoi(parts[1])
			page, _ := strconv.Atoi(parts[2])
			st.FilterIdx = idx
			st.FilterValPage = page
			text, km := t.ui.RenderFilterValues(st.Draft, idx, page)
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "filter_toggle:"):
		parts := strings.Split(strings.TrimPrefix(data, "filter_toggle:"), ":")
		if len(parts) >= 4 {
			idx, _ := strconv.Atoi(parts[1])
			vIdx, _ := strconv.Atoi(parts[2])
			page, _ := strconv.Atoi(parts[3])
			defs := t.ui.categoryFilters(st.Draft.Category, st.Draft.BrandParam)
			if idx >= 0 && idx < len(defs) {
				def := defs[idx]
				if vIdx >= 0 && vIdx < len(def.Values) {
					v := def.Values[vIdx].Value
					if st.Draft.Filters == nil {
						st.Draft.Filters = map[string][]string{}
					}
					if def.Type == "checkboxes" {
						cur := st.Draft.Filters[def.Key]
						if olx.StringInSlice(cur, v) {
							out := cur[:0]
							for _, x := range cur {
								if x != v {
									out = append(out, x)
								}
							}
							cur = out
						} else {
							cur = append(cur, v)
						}
						if len(cur) == 0 {
							delete(st.Draft.Filters, def.Key)
						} else {
							st.Draft.Filters[def.Key] = cur
						}
					} else {
						if olx.StringInSlice(st.Draft.Filters[def.Key], v) {
							delete(st.Draft.Filters, def.Key)
						} else {
							st.Draft.Filters[def.Key] = []string{v}
						}
					}
				}
			}
			st.FilterIdx = idx
			text, km := t.ui.RenderFilterValues(st.Draft, idx, page)
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "filter_clear:"):
		parts := strings.Split(strings.TrimPrefix(data, "filter_clear:"), ":")
		if len(parts) >= 2 {
			idx, _ := strconv.Atoi(parts[1])
			defs := t.ui.categoryFilters(st.Draft.Category, st.Draft.BrandParam)
			if idx >= 0 && idx < len(defs) {
				delete(st.Draft.Filters, defs[idx].Key)
				delete(st.Draft.Ranges, defs[idx].Key)
			}
			text, km := t.ui.RenderFilterValues(st.Draft, idx, 0)
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "filter_bool:"):
		parts := strings.Split(strings.TrimPrefix(data, "filter_bool:"), ":")
		if len(parts) >= 2 {
			idx, _ := strconv.Atoi(parts[1])
			defs := t.ui.categoryFilters(st.Draft.Category, st.Draft.BrandParam)
			if idx >= 0 && idx < len(defs) {
				def := defs[idx]
				if st.Draft.Filters == nil {
					st.Draft.Filters = map[string][]string{}
				}
				if len(st.Draft.Filters[def.Key]) > 0 {
					delete(st.Draft.Filters, def.Key)
				} else {
					st.Draft.Filters[def.Key] = []string{"1"}
				}
			}
			text, km := t.ui.RenderFilterValues(st.Draft, idx, 0)
			_ = t.editMessageText(chatID, msgID, text, km)
		}

	case strings.HasPrefix(data, "filter_input:"):
		parts := strings.Split(strings.TrimPrefix(data, "filter_input:"), ":")
		if len(parts) >= 2 {
			idx, _ := strconv.Atoi(parts[1])
			defs := t.ui.categoryFilters(st.Draft.Category, st.Draft.BrandParam)
			if idx >= 0 && idx < len(defs) {
				st.Step = StepWaitingFilterValue
				st.FilterCat = st.Draft.Category
				st.FilterIdx = idx
				unit := ""
				if defs[idx].Unit != "" {
					unit = " (" + defs[idx].Unit + ")"
				}
				prompt := fmt.Sprintf("✏️ <b>%s</b>%s\n\nВведи число или диапазон, например: <code>2000-2010</code> или <code>от 100</code>.",
					html.EscapeString(defs[idx].Label), html.EscapeString(unit))
				_ = t.editMessageText(chatID, msgID, prompt,
					markup(row(btn("⬅️ Назад", fmt.Sprintf("filter_open:%d:%d:%d", st.Draft.Category, idx, st.FilterListPage)))))
			}
		}

	case strings.HasPrefix(data, "filters_back:"):
		text, km := t.ui.RenderQueryStep(st.Draft.Category, st.Draft.CatName, brandDisplay(st.Draft))
		_ = t.editMessageText(chatID, msgID, text, km)

	case data == "list_chats":
		if t.cfg == nil {
			break
		}
		ids := t.cfg.GetTelegramChatIDs()
		var sb strings.Builder
		sb.WriteString("📤 <b>Получатели алертов:</b>\n")
		for i, id := range ids {
			sb.WriteString(fmt.Sprintf("%d. <code>%s</code>\n", i+1, html.EscapeString(id)))
		}
		sb.WriteString("\n<b>Добавить группу:</b> кнопка «Привязать этот чат» внутри группы\n")
		sb.WriteString("<b>Канал или группа по ID:</b> /addchat &lt;id&gt;\n")
		sb.WriteString("<b>Удалить:</b> /unbind в группе или /delchat &lt;id&gt;")
		_ = t.sendMessage(chatID, sb.String(), nil)

	case data == "show_chatid":
		kind := "private"
		title := ""
		if cq.Message != nil {
			if cq.Message.Chat.Type != "" {
				kind = cq.Message.Chat.Type
			}
			title = cq.Message.Chat.Title
		}
		reply := fmt.Sprintf("🆔 chat_id: %s\nТип: %s", chatID, kind)
		if title != "" {
			reply += "\nНазвание: " + title
		}
		answerAlert(reply)

	case data == "bind_current":
		if t.cfg == nil {
			break
		}
		if !strings.HasPrefix(chatID, "-") {
			answerAlert("Это личный чат — он уже получатель алертов. Кнопка нужна внутри группы или канала.")
			break
		}
		added, err := t.cfg.AddTrustedChatID(chatID, 0)
		if err != nil {
			answerAlert(fmt.Sprintf("❌ Не удалось сохранить: %v", err))
			break
		}
		if added {
			t.chatIDs = t.cfg.GetTelegramChatIDs()
			answerToast("✅ Этот чат привязан к алертам")
		} else {
			answerToast("Этот чат уже в списке получателей")
		}

	case strings.HasPrefix(data, "set_photos:"):
		n, _ := strconv.Atoi(strings.TrimPrefix(data, "set_photos:"))
		if n > 0 {
			_ = t.cfg.SetMaxPhotos(n)
			_ = t.cfg.SetSendPhotos(true)
			answerToast(fmt.Sprintf("📷 Фото: до %d шт.", n))
		}
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)

	case strings.HasPrefix(data, "set_interval:"):
		sec, _ := strconv.Atoi(strings.TrimPrefix(data, "set_interval:"))
		_ = t.cfg.SetInterval(sec)
		answerToast(fmt.Sprintf("⏱️ Интервал: %d сек", sec))
		text, km := t.ui.RenderSettings()
		_ = t.editMessageText(chatID, msgID, text, km)
	}
}

// Telegram API network calls
func (t *TelegramNotifier) sendRawMessage(chatID, text, parseMode string, km *InlineKeyboardMarkup) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := map[string]interface{}{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": false,
	}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	if km != nil {
		payload["reply_markup"] = km
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		text := string(respBody)
		if isDeadChatError(text) {
			t.markChatDead(chatID, text)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, text)
	}
	return nil
}

func (t *TelegramNotifier) sendMessage(chatID, text string, km *InlineKeyboardMarkup) error {
	err := t.sendRawMessage(chatID, text, "HTML", km)
	if err != nil {
		if isDeadChatError(err.Error()) {
			return err
		}
		// Fallback: пробуем отправить как чистый текст, если Telegram ругается на HTML сущности
		plainText := olx.StripHTMLTags(text)
		if plainErr := t.sendRawMessage(chatID, plainText, "", km); plainErr == nil {
			return nil
		}
		return err
	}
	return nil
}

type inputMediaPhoto struct {
	Type  string `json:"type"`
	Media string `json:"media"`
}

func (t *TelegramNotifier) postTelegramJSON(method string, payload interface{}) error {
	if t.botToken == "" {
		return nil
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.botToken, method)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	chatID := payloadChatID(payload)
	var lastErr error

	for attempt := 0; attempt < maxSendAttempts; attempt++ {
		if chatID != "" {
			t.waitChat(chatID)
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := t.httpClient.Do(req)
		if err != nil {
			return err
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		lastErr = fmt.Errorf("%s HTTP %d: %s", method, resp.StatusCode, string(respBody))
		if resp.StatusCode != http.StatusTooManyRequests {
			return lastErr
		}

		// 429: ждём retry_after и пробуем ещё раз, не сваливаясь в fallback.
		retry := parseRetryAfter(respBody)
		if chatID != "" {
			t.noteRateLimit(chatID, retry)
		} else {
			time.Sleep(retry)
		}
	}
	return lastErr
}

const (
	telegramChatGap  = 1100 * time.Millisecond
	telegramGroupGap = 3 * time.Second
	maxSendAttempts  = 4
	defaultRetryWait = 2 * time.Second
)

func payloadChatID(payload interface{}) string {
	if m, ok := payload.(map[string]interface{}); ok {
		if v, ok := m["chat_id"]; ok {
			return fmt.Sprint(v)
		}
	}
	return ""
}

// waitChat резервирует слот отправки в чат, чтобы не ловить 429 на пачках.
func (t *TelegramNotifier) waitChat(chatID string) {
	gap := telegramChatGap
	if strings.HasPrefix(chatID, "-") {
		// В группах/каналах лимит жёстче: ~20 сообщений в минуту.
		gap = telegramGroupGap
	}

	t.sendMu.Lock()
	now := time.Now()
	sendAt := now
	if last, ok := t.lastSend[chatID]; ok {
		if earliest := last.Add(gap); earliest.After(sendAt) {
			sendAt = earliest
		}
	}
	t.lastSend[chatID] = sendAt
	t.sendMu.Unlock()

	if d := time.Until(sendAt); d > 0 {
		time.Sleep(d)
	}
}

func (t *TelegramNotifier) noteRateLimit(chatID string, retry time.Duration) {
	until := time.Now().Add(retry)
	t.sendMu.Lock()
	if cur := t.lastSend[chatID]; until.After(cur) {
		t.lastSend[chatID] = until
	}
	t.sendMu.Unlock()
}

func parseRetryAfter(body []byte) time.Duration {
	var payload struct {
		RetryAfter int `json:"retry_after"`
		Parameters struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		n := payload.Parameters.RetryAfter
		if n == 0 {
			n = payload.RetryAfter
		}
		if n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return defaultRetryWait
}

// sendOfferPhotos отправляет тихий альбом (без звука), чтобы не дублировать
// уведомление: основная карточка идёт отдельным сообщением с кнопкой.
func (t *TelegramNotifier) sendOfferPhotos(chatID string, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	if len(urls) == 1 {
		err := t.postTelegramJSON("sendPhoto", map[string]interface{}{
			"chat_id":              chatID,
			"photo":                urls[0],
			"disable_notification": true,
		})
		if err != nil && isDeadChatError(err.Error()) {
			t.markChatDead(chatID, err.Error())
		}
		return err
	}

	if len(urls) > 10 {
		urls = urls[:10]
	}
	media := make([]inputMediaPhoto, 0, len(urls))
	for _, u := range urls {
		media = append(media, inputMediaPhoto{Type: "photo", Media: u})
	}
	err := t.postTelegramJSON("sendMediaGroup", map[string]interface{}{
		"chat_id":              chatID,
		"media":                media,
		"disable_notification": true,
	})
	if err != nil && isDeadChatError(err.Error()) {
		t.markChatDead(chatID, err.Error())
	}
	return err
}

func (t *TelegramNotifier) editMessageText(chatID string, messageID int, text string, km *InlineKeyboardMarkup) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", t.botToken)

	payload := map[string]interface{}{
		"chat_id":                  chatID,
		"message_id":               messageID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}
	if km != nil {
		payload["reply_markup"] = km
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusBadRequest && strings.Contains(string(respBody), "message is not modified") {
			return nil
		}
		// Пробуем fallback без HTML если ошибка парсинга
		delete(payload, "parse_mode")
		payload["text"] = olx.StripHTMLTags(text)
		if fbBody, err := json.Marshal(payload); err == nil {
			if fbReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(fbBody)); err == nil {
				fbReq.Header.Set("Content-Type", "application/json")
				if fbResp, err := t.httpClient.Do(fbReq); err == nil {
					defer fbResp.Body.Close()
					if fbResp.StatusCode == http.StatusOK {
						return nil
					}
				}
			}
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (t *TelegramNotifier) answerCallbackQuery(callbackID, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", t.botToken)

	payload := map[string]interface{}{
		"callback_query_id": callbackID,
	}
	if text != "" {
		payload["text"] = text
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (t *TelegramNotifier) answerCallbackAlert(callbackID, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", t.botToken)

	payload := map[string]interface{}{
		"callback_query_id": callbackID,
		"text":              text,
		"show_alert":        true,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (t *TelegramNotifier) getUpdates(offset int) ([]TGUpdate, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", t.botToken)

	payload := map[string]interface{}{
		"offset":          offset,
		"timeout":         25,
		"allowed_updates": []string{"message", "callback_query"},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tgResp TGUpdatesResponse
	if err := json.Unmarshal(respBody, &tgResp); err != nil {
		return nil, err
	}

	return tgResp.Result, nil
}


const (
	TelegramChatGap  = telegramChatGap
	DefaultRetryWait = defaultRetryWait
)

func FilterChatIDsForOwner(chatIDs []string, ownerChatID string) []string {
	return filterChatIDsForOwner(chatIDs, ownerChatID)
}
func IsDeadChatError(text string) bool {
	return isDeadChatError(text)
}
func IsPhotoRelatedError(err error) bool {
	return isPhotoRelatedError(err)
}
func ParseRetryAfter(body []byte) time.Duration {
	return parseRetryAfter(body)
}
func (t *TelegramNotifier) WaitChat(chatID string) {
	t.waitChat(chatID)
}
func (t *TelegramNotifier) MarkChatDead(chatID, reason string) {
	t.markChatDead(chatID, reason)
}
func (t *TelegramNotifier) ReviveChat(chatID string) {
	t.reviveChat(chatID)
}
func (t *TelegramNotifier) IsChatDead(chatID string) bool {
	return t.isChatDead(chatID)
}

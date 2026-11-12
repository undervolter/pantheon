package main

import (
	"fmt"
	"html"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type WizardStep int

const (
	StepNone WizardStep = iota
	StepWaitingQuery
	StepWaitingVariant
	StepWaitingCustomPrice
	StepWaitingPriceRange
	StepWaitingFilterValue
	StepWaitingCity
)

type UserWizardState struct {
	Step           WizardStep
	Draft          SearchConfig
	ParentCat      int
	SelectedBrands map[string]string // brandID -> name
	FilterCat      int
	FilterIdx      int
	FilterListPage int
	FilterValPage  int
}

type BotUI struct {
	cfg     *Config
	checker *Checker
	states  map[string]*UserWizardState
	stateMu sync.Mutex
}

func NewBotUI(cfg *Config, checker *Checker) *BotUI {
	return &BotUI{
		cfg:     cfg,
		checker: checker,
		states:  make(map[string]*UserWizardState),
	}
}

func (ui *BotUI) getState(chatID string) *UserWizardState {
	ui.stateMu.Lock()
	defer ui.stateMu.Unlock()
	st, ok := ui.states[chatID]
	if !ok {
		st = &UserWizardState{Step: StepNone}
		ui.states[chatID] = st
	}
	return st
}

func (ui *BotUI) resetState(chatID string) {
	ui.stateMu.Lock()
	defer ui.stateMu.Unlock()
	delete(ui.states, chatID)
}

// Inline keyboard building helpers
func btn(text, data string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, CallbackData: data}
}

func row(buttons ...InlineKeyboardButton) []InlineKeyboardButton {
	return buttons
}

func markup(rows ...[]InlineKeyboardButton) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: rows}
}

// 1. Главное меню
func (ui *BotUI) RenderMainMenu(chatID string) (string, *InlineKeyboardMarkup) {
	stats := ui.checker.GetStats()

	ownTotal, ownActive := 0, 0
	for _, s := range ui.cfg.GetSearches() {
		if s.OwnerChatID != "" && s.OwnerChatID != chatID {
			continue
		}
		ownTotal++
		if !s.Disabled {
			ownActive++
		}
	}

	statusText := "⏸ Нет активных фильтров — добавь первый"
	if ownActive > 0 {
		statusText = "🟢 В работе"
	}

	text := fmt.Sprintf(
		"⠀⠀           ⠀       ⠀<b>p a n t h e o n | Парсер OLX</b> <i>v"+appVersion+"</i>\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📊 <b>Статус:</b> %s\n"+
			"🎯 <b>Твоих фильтров:</b> %d из %d\n"+
			"📦 <b>В кэше (просмотрено):</b> %d\n"+
			"🔔 <b>Доставлено алертов:</b> %d\n"+
			"⏱ <b>Интервал:</b> %d сек.\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Управляй фильтрами и параметрами кнопками ниже:",
		statusText, ownActive, ownTotal, stats.SeenCount, stats.SentCount, stats.Interval,
	)

	km := markup(
		row(btn("➕ Добавить фильтр", "add_start"), btn("📋 Мои фильтры", "menu_list")),
		row(btn("⚙️ Настройки", "menu_settings"), btn("🔄 Обновить статус", "menu_main")),
	)

	return text, km
}

// searchForChat возвращает фильтр по индексу, если он доступен этому чату.
// Свои фильтры видит только владелец, общие (без владельца) — все.
func (ui *BotUI) searchForChat(chatID string, index int) (SearchConfig, bool) {
	searches := ui.cfg.GetSearches()
	if index < 0 || index >= len(searches) {
		return SearchConfig{}, false
	}
	s := searches[index]
	if s.OwnerChatID != "" && s.OwnerChatID != chatID {
		return SearchConfig{}, false
	}
	return s, true
}

// 2. Список фильтров текущего пользователя.
func (ui *BotUI) RenderFiltersList(chatID string) (string, *InlineKeyboardMarkup) {
	searches := ui.cfg.GetSearches()

	var sb strings.Builder
	sb.WriteString("📋 <b>Твои поисковые задачи:</b>\n")
	sb.WriteString("<i>Нажми на нужный фильтр для паузы, удаления или закрепления:</i>\n\n")

	var rows [][]InlineKeyboardButton
	shown := 0

	for i, s := range searches {
		// Чужие фильтры не показываем: каждый управляет только своими.
		if s.OwnerChatID != "" && s.OwnerChatID != chatID {
			continue
		}
		shown++

		statusIcon := "🟢"
		if s.Disabled {
			statusIcon = "🔴"
		}
		if s.OwnerChatID == "" {
			// Общий фильтр: алерты идут всем получателям.
			statusIcon = "🌐" + statusIcon
		}

		priceInfo := "любая цена"
		switch {
		case s.MinPrice > 0 && s.MaxPrice > 0:
			priceInfo = fmt.Sprintf("%.0f–%.0f грн", s.MinPrice, s.MaxPrice)
		case s.MaxPrice > 0:
			priceInfo = fmt.Sprintf("до %.0f грн", s.MaxPrice)
		case s.MinPrice > 0:
			priceInfo = fmt.Sprintf("от %.0f грн", s.MinPrice)
		}

		label := fmt.Sprintf("%s %s (%s)", statusIcon, s.Name, priceInfo)
		rows = append(rows, row(btn(truncateRunes(label, 60), fmt.Sprintf("f_view:%d", i))))
	}

	if shown == 0 {
		text := "📋 <b>У тебя пока нет своих фильтров.</b>\n\n" +
			"Нажми кнопку ниже, чтобы создать первый. Уведомления по нему будет получать только этот чат:"
		km := markup(
			row(btn("➕ Добавить фильтр", "add_start")),
			row(btn("⬅️ Главное меню", "menu_main")),
		)
		return text, km
	}

	rows = append(rows, row(btn("➕ Добавить фильтр", "add_start"), btn("⬅️ Главное меню", "menu_main")))

	return sb.String(), markup(rows...)
}

// 3. Детальный вид фильтра
func (ui *BotUI) RenderFilterDetail(chatID string, index int) (string, *InlineKeyboardMarkup) {
	s, ok := ui.searchForChat(chatID, index)
	if !ok {
		return "❌ Фильтр не найден или принадлежит другому пользователю",
			markup(row(btn("⬅️ К списку", "menu_list")))
	}

	statusText := "🟢 Активен (мониторится)"
	toggleBtnText := "⏸ Поставить на паузу"
	if s.Disabled {
		statusText = "🔴 На паузе"
		toggleBtnText = "▶️ Возобновить"
	}

	priceStr := "Без ограничений"
	if s.MinPrice > 0 && s.MaxPrice > 0 {
		priceStr = fmt.Sprintf("от %.0f до %.0f грн", s.MinPrice, s.MaxPrice)
	} else if s.MaxPrice > 0 {
		priceStr = fmt.Sprintf("до %.0f грн", s.MaxPrice)
	} else if s.MinPrice > 0 {
		priceStr = fmt.Sprintf("от %.0f грн", s.MinPrice)
	}

	catStr := "Все категории"
	if s.CatName != "" {
		catStr = html.EscapeString(s.CatName)
	} else if s.Category > 0 {
		if c := FindCategory(s.Category); c != nil {
			catStr = html.EscapeString(c.Name)
		}
	}

	cityStr := "Вся Украина"
	if s.CityName != "" {
		cityStr = html.EscapeString(s.CityName)
	} else if s.CityID > 0 {
		if c := FindCity(s.CityID); c != nil {
			cityStr = html.EscapeString(c.Name)
		}
	}

	queryStr := html.EscapeString(s.Query)
	if queryStr == "" {
		queryStr = "<i>(все объявления в категории)</i>"
	}

	brandStr := ""
	if bd := brandDisplay(s); bd != "" {
		brandStr = fmt.Sprintf("🏷 <b>Бренд:</b> %s\n", html.EscapeString(bd))
	}

	districtStr := ""
	if s.DistrictName != "" {
		districtStr = fmt.Sprintf("🏘 <b>Район:</b> %s\n", html.EscapeString(s.DistrictName))
	}

	ownerStr := ""
	switch s.OwnerType {
	case "private":
		ownerStr = "👤 <b>Продавец:</b> только частные\n"
	case "business":
		ownerStr = "🏢 <b>Продавец:</b> только бизнес\n"
	}
	filtersStr := filtersSummary(s)
	variantsStr := ""
	if len(s.QueryVariants) > 0 {
		variantsStr = "🔤 <b>Варианты:</b> " + html.EscapeString(truncateRunes(strings.Join(s.QueryVariants, ", "), 180)) + "\n"
	}

	recipientsStr := "🌐 <b>Получатели:</b> все (общий фильтр)\n"
	if s.OwnerChatID != "" {
		recipientsStr = "👤 <b>Получатели:</b> только этот чат\n"
	}

	text := fmt.Sprintf(
		"🎯 <b>Фильтр: %s</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"🔍 <b>Запрос:</b> %s\n"+
			"📁 <b>Категория:</b> %s\n"+
			"%s"+
			"💰 <b>Цена:</b> %s\n"+
			"📍 <b>Локация:</b> %s\n"+
			"%s"+
			"%s"+
			"%s"+
			"%s"+
			"%s"+
			"⚡ <b>Статус:</b> %s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━",
		html.EscapeString(s.Name), queryStr, catStr, brandStr, priceStr, cityStr, districtStr, ownerStr, filtersStr, variantsStr, recipientsStr,
		statusText,
	)

	var rows [][]InlineKeyboardButton
	rows = append(rows, row(btn(toggleBtnText, fmt.Sprintf("f_toggle:%d", index))))
	switch {
	case s.OwnerChatID == "":
		rows = append(rows, row(btn("👤 Закрепить за мной", fmt.Sprintf("f_claim:%d", index))))
	case s.OwnerChatID == chatID:
		rows = append(rows, row(btn("👥 Сделать общим", fmt.Sprintf("f_share:%d", index))))
	}
	rows = append(rows, row(btn("🗑 Удалить фильтр", fmt.Sprintf("f_del:%d", index))))
	rows = append(rows, row(btn("⬅️ К списку фильтров", "menu_list")))

	return text, markup(rows...)
}

// 4. Шаг 1: Выбор категории
func (ui *BotUI) RenderCategoryStep() (string, *InlineKeyboardMarkup) {
	text := "📁 <b>Шаг 1 из 4: Выбери категорию</b>\n\nВыбери нужный раздел OLX или нажми 'Любая категория':"

	var rows [][]InlineKeyboardButton
	var currentRow []InlineKeyboardButton

	for _, c := range TopCategories {
		currentRow = append(currentRow, btn(fmt.Sprintf("%s %s", c.Icon, c.Name), fmt.Sprintf("add_cat:%d", c.ID)))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	rows = append(rows, row(btn("🌐 Любая категория (пропустить)", "add_cat:0")))
	rows = append(rows, row(btn("❌ Отмена", "menu_main")))

	return text, markup(rows...)
}

// Шаг 1b: Подкатегории
func (ui *BotUI) RenderSubcategoryStep(parentID int) (string, *InlineKeyboardMarkup) {
	parent := FindCategory(parentID)
	if parent == nil || len(parent.Children) == 0 {
		if info := GetCategoryBrands(parentID); info != nil && len(info.Brands) > 0 {
			return ui.RenderBrandStep(parentID, 0, nil)
		}
		return ui.RenderQueryStep(parentID, "", "")
	}

	text := fmt.Sprintf(
		"📁 <b>%s</b>\n\nВыбери подкатегорию или выбери весь раздел целиком:",
		html.EscapeString(CategoryBreadcrumb(parentID)),
	)

	var rows [][]InlineKeyboardButton
	var currentRow []InlineKeyboardButton

	for _, sub := range parent.Children {
		currentRow = append(currentRow, btn(truncateRunes(sub.Name, 40), fmt.Sprintf("add_subcat:%d", sub.ID)))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	rows = append(rows, row(btn(fmt.Sprintf("🎯 Выбрать всю '%s'", truncateRunes(parent.Name, 35)), fmt.Sprintf("add_cat_select:%d", parentID))))

	backData := "add_start"
	backText := "⬅️ Назад к категориям"
	if parent.ParentID > 0 {
		backData = fmt.Sprintf("add_cat_back:%d", parent.ParentID)
		backText = "⬅️ Назад"
	}
	rows = append(rows, row(btn(backText, backData)))

	return text, markup(rows...)
}

// 5. Шаг 2: Выбор бренда (мультивыбор)
func (ui *BotUI) RenderBrandStep(catID int, page int, selected map[string]bool) (string, *InlineKeyboardMarkup) {
	info := GetCategoryBrands(catID)
	if info == nil || len(info.Brands) == 0 {
		catName := ""
		if c := FindCategory(catID); c != nil {
			catName = c.Name
		}
		return ui.RenderQueryStep(catID, catName, "")
	}

	catName := ""
	if c := FindCategory(catID); c != nil {
		catName = c.Name
	}

	const pageSize = 10
	totalBrands := len(info.Brands)
	totalPages := (totalBrands + pageSize - 1) / pageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * pageSize
	end := start + pageSize
	if end > totalBrands {
		end = totalBrands
	}

	label := info.Label
	if label == "" {
		label = "Бренд"
	}

	selectedCount := 0
	for _, b := range info.Brands {
		if selected[b.ID] {
			selectedCount++
		}
	}

	text := fmt.Sprintf(
		"🏷 <b>%s</b>\n"+
			"Категория: <b>%s</b>\n\n"+
			"Можно выбрать несколько брендов. Отмеченные — ✅.\n"+
			"Выбрано: <b>%d</b>.",
		html.EscapeString(label), html.EscapeString(catName), selectedCount,
	)

	var rows [][]InlineKeyboardButton
	var currentRow []InlineKeyboardButton

	for _, b := range info.Brands[start:end] {
		prefix := "☐ "
		if selected[b.ID] {
			prefix = "✅ "
		}
		currentRow = append(currentRow, btn(
			truncateRunes(prefix+b.Name, 55),
			fmt.Sprintf("add_brand_toggle:%d:%d:%s", catID, page, b.ID),
		))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	// Pagination row if multiple pages
	if totalPages > 1 {
		var pageNav []InlineKeyboardButton
		if page > 0 {
			pageNav = append(pageNav, btn("⬅️ Назад", fmt.Sprintf("add_brand_page:%d:%d", catID, page-1)))
		}
		pageNav = append(pageNav, btn(fmt.Sprintf("📄 %d/%d", page+1, totalPages), "noop"))
		if page < totalPages-1 {
			pageNav = append(pageNav, btn("Вперед ➡️", fmt.Sprintf("add_brand_page:%d:%d", catID, page+1)))
		}
		rows = append(rows, pageNav)
	}

	rows = append(rows, row(
		btn(fmt.Sprintf("✅ Готово (%d)", selectedCount), fmt.Sprintf("add_brand_done:%d", catID)),
		btn("🌐 Без бренда", fmt.Sprintf("add_brand_skip:%d", catID)),
	))
	rows = append(rows, row(btn("⬅️ Назад к категориям", "add_start")))

	return text, markup(rows...)
}

// 6. Шаг 3: Ввод текста поиска
func (ui *BotUI) RenderQueryStep(catID int, catName string, brandName string) (string, *InlineKeyboardMarkup) {
	if catName == "" && catID > 0 {
		if c := FindCategory(catID); c != nil {
			catName = c.Name
		}
	}
	if catName == "" {
		catName = "Любая"
	}

	brandInfo := ""
	if brandName != "" {
		brandInfo = fmt.Sprintf(" | Бренд: <b>%s</b>", html.EscapeString(brandName))
	}

	text := fmt.Sprintf(
		"🔍 <b>Поисковый запрос</b>\n"+
			"Категория: <b>%s</b>%s\n\n"+
			"✍️ <b>Напиши текстом в чат</b> поисковый запрос (например: <code>15 Pro</code>, <code>RTX 4080</code>, <code>PlayStation 5</code>).\n\n"+
			"Или нажми кнопку ниже, чтобы ловить <b>все новые объявления</b>:",
		html.EscapeString(catName), brandInfo,
	)

	km := markup(
		row(btn("⏩ Пропустить (ловить всё подряд)", "add_query_skip")),
		row(btn("❌ Отмена", "menu_main")),
	)

	if catID > 0 && HasFilterCategory(catID) {
		km = markup(
			row(btn("⚙️ Доп. фильтры", fmt.Sprintf("filters_open:%d", catID))),
			row(btn("⏩ Пропустить (ловить всё подряд)", "add_query_skip")),
			row(btn("❌ Отмена", "menu_main")),
		)
	}

	return text, km
}

// RenderVariantsStep — управление вариантами запроса (пользователь добавляет сам).
func (ui *BotUI) RenderVariantsStep(draft SearchConfig) (string, *InlineKeyboardMarkup) {
	text := fmt.Sprintf(
		"🔤 <b>Варианты запроса</b>\n\nЗапрос: <b>%s</b>\n\n"+
			"Бот пришлёт лот, если в заголовке или описании есть все слова хотя бы одного варианта.\n"+
			"Например: <code>iphone 4s</code>, <code>айфон 4с</code>, <code>4s</code>.\n\n"+
			"<i>Варианты нормализуются: «Айфон» = «iphone», «4с» = «4s».</i>\n\n",
		html.EscapeString(draft.Query),
	)
	if len(draft.QueryVariants) == 0 {
		text += "<i>Пока вариантов нет — будет использован сам запрос.</i>"
	}

	var rows [][]InlineKeyboardButton
	for i, v := range draft.QueryVariants {
		rows = append(rows, row(btn("🗑 "+truncateRunes(v, 45), fmt.Sprintf("variant_del:%d", i))))
	}
	rows = append(rows, row(btn("➕ Добавить вариант", "variant_add")))
	rows = append(rows, row(btn("✅ Готово", "variants_done")))

	return text, markup(rows...)
}

// categoryFilters — фильтры категории без служебных (цена, владелец, сортировка)
// и без бренда, если он уже выбран отдельным шагом.
func (ui *BotUI) categoryFilters(catID int, brandParam string) []FilterDef {
	all := FiltersForCategory(catID)
	out := make([]FilterDef, 0, len(all))
	for _, d := range all {
		if isReservedFilterKey(d.Key) {
			continue
		}
		if brandParam != "" && d.Key == brandParam {
			continue
		}
		out = append(out, d)
	}
	return out
}

func filterSelectedCount(d SearchConfig, def FilterDef) int {
	if len(d.Filters[def.Key]) > 0 {
		return len(d.Filters[def.Key])
	}
	if r, ok := d.Ranges[def.Key]; ok && (r.From > 0 || r.To > 0) {
		return 1
	}
	return 0
}

// RenderFiltersStep — список дополнительных фильтров категории.
func (ui *BotUI) RenderFiltersStep(draft SearchConfig, page int) (string, *InlineKeyboardMarkup) {
	defs := ui.categoryFilters(draft.Category, draft.BrandParam)
	if len(defs) == 0 {
		return "⚙️ Для этой категории нет дополнительных фильтров.",
			markup(row(btn("⬅️ К запросу", fmt.Sprintf("filters_back:%d", draft.Category))))
	}

	const pageSize = 8
	totalPages := (len(defs) + pageSize - 1) / pageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * pageSize
	end := start + pageSize
	if end > len(defs) {
		end = len(defs)
	}

	selected := 0
	for _, d := range defs {
		if filterSelectedCount(draft, d) > 0 {
			selected++
		}
	}

	catName := draft.CatName
	if catName == "" && draft.Category > 0 {
		if c := FindCategory(draft.Category); c != nil {
			catName = c.Name
		}
	}
	text := fmt.Sprintf(
		"⚙️ <b>Дополнительные фильтры</b>\nКатегория: <b>%s</b>\nВыбрано: <b>%d</b> из %d.\n\nНажми фильтр, чтобы настроить:",
		html.EscapeString(catName), selected, len(defs),
	)

	var rows [][]InlineKeyboardButton
	var currentRow []InlineKeyboardButton
	for i := start; i < end; i++ {
		d := defs[i]
		label := d.Label
		if n := filterSelectedCount(draft, d); n > 0 {
			label += " ✅"
			if n > 1 {
				label += fmt.Sprintf("%d", n)
			}
		}
		currentRow = append(currentRow, btn(truncateRunes(label, 30), fmt.Sprintf("filter_open:%d:%d:%d", draft.Category, i, page)))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	if totalPages > 1 {
		var nav []InlineKeyboardButton
		if page > 0 {
			nav = append(nav, btn("⬅️", fmt.Sprintf("filters_page:%d:%d", draft.Category, page-1)))
		}
		nav = append(nav, btn(fmt.Sprintf("📄 %d/%d", page+1, totalPages), "noop"))
		if page < totalPages-1 {
			nav = append(nav, btn("➡️", fmt.Sprintf("filters_page:%d:%d", draft.Category, page+1)))
		}
		rows = append(rows, nav)
	}

	rows = append(rows, row(btn("⬅️ К запросу", fmt.Sprintf("filters_back:%d", draft.Category))))
	return text, markup(rows...)
}

// RenderFilterValues — значения выбранного фильтра.
func (ui *BotUI) RenderFilterValues(draft SearchConfig, filterIdx, page int) (string, *InlineKeyboardMarkup) {
	defs := ui.categoryFilters(draft.Category, draft.BrandParam)
	if filterIdx < 0 || filterIdx >= len(defs) {
		return ui.RenderFiltersStep(draft, 0)
	}
	def := defs[filterIdx]

	backBtn := btn("⬅️ К фильтрам", fmt.Sprintf("filters_list:%d:%d", draft.Category, 0))

	// Булевый фильтр (checkbox без значений): просто вкл/выкл.
	if strings.HasPrefix(def.Key, "filter_bool_") || (def.Type == "checkbox" && len(def.Values) == 0) {
		on := len(draft.Filters[def.Key]) > 0
		state := "выключен"
		toggle := "✅ Включить"
		if on {
			state = "включён"
			toggle = "❌ Выключить"
		}
		text := fmt.Sprintf("⚙️ <b>%s</b>\nСейчас: <b>%s</b>", html.EscapeString(def.Label), state)
		km := markup(
			row(btn(toggle, fmt.Sprintf("filter_bool:%d:%d", draft.Category, filterIdx))),
			row(backBtn),
		)
		return text, km
	}

	// Числовой фильтр (input/float).
	if def.Type == "input" || strings.HasPrefix(def.Key, "filter_float_") {
		cur := "не задан"
		if r, ok := draft.Ranges[def.Key]; ok && (r.From > 0 || r.To > 0) {
			cur = rangeLabel(r, def.Unit)
		}
		text := fmt.Sprintf(
			"⚙️ <b>%s</b>\nТип: число%s\nСейчас: <b>%s</b>\n\nВведи значение или диапазон.",
			html.EscapeString(def.Label), unitSuffix(def.Unit), html.EscapeString(cur),
		)
		km := markup(
			row(btn("✏️ Ввести значение", fmt.Sprintf("filter_input:%d:%d", draft.Category, filterIdx))),
			row(btn("🗑 Сбросить", fmt.Sprintf("filter_clear:%d:%d", draft.Category, filterIdx))),
			row(backBtn),
		)
		return text, km
	}

	values := def.Values
	const pageSize = 8
	totalPages := (len(values) + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * pageSize
	end := start + pageSize
	if end > len(values) {
		end = len(values)
	}

	selected := draft.Filters[def.Key]
	var selLabels []string
	for _, v := range values {
		if stringInSlice(selected, v.Value) {
			selLabels = append(selLabels, v.Label)
		}
	}
	cur := "не задан"
	if len(selLabels) > 0 {
		cur = strings.Join(selLabels, ", ")
	}

	multi := def.Type == "checkboxes"
	typeStr := "один вариант"
	if multi {
		typeStr = "можно несколько"
	}
	text := fmt.Sprintf(
		"⚙️ <b>%s</b>\nТип: %s\nСейчас: <b>%s</b>",
		html.EscapeString(def.Label), typeStr, html.EscapeString(truncateRunes(cur, 120)),
	)

	var rows [][]InlineKeyboardButton
	var currentRow []InlineKeyboardButton
	for i := start; i < end; i++ {
		v := values[i]
		label := v.Label
		if stringInSlice(selected, v.Value) {
			label = "✅ " + label
		}
		currentRow = append(currentRow, btn(truncateRunes(label, 32), fmt.Sprintf("filter_toggle:%d:%d:%d:%d", draft.Category, filterIdx, i, page)))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	if totalPages > 1 {
		var nav []InlineKeyboardButton
		if page > 0 {
			nav = append(nav, btn("⬅️", fmt.Sprintf("filter_page:%d:%d:%d", draft.Category, filterIdx, page-1)))
		}
		nav = append(nav, btn(fmt.Sprintf("📄 %d/%d", page+1, totalPages), "noop"))
		if page < totalPages-1 {
			nav = append(nav, btn("➡️", fmt.Sprintf("filter_page:%d:%d:%d", draft.Category, filterIdx, page+1)))
		}
		rows = append(rows, nav)
	}

	rows = append(rows, row(btn("🗑 Сбросить", fmt.Sprintf("filter_clear:%d:%d", draft.Category, filterIdx))))
	rows = append(rows, row(backBtn))
	return text, markup(rows...)
}

func unitSuffix(unit string) string {
	if unit == "" {
		return ""
	}
	return " (" + unit + ")"
}

func rangeLabel(r FilterRange, unit string) string {
	suffix := ""
	if unit != "" {
		suffix = " " + unit
	}
	switch {
	case r.From > 0 && r.To > 0:
		return fmt.Sprintf("от %s до %s%s", trimFloat(r.From), trimFloat(r.To), suffix)
	case r.From > 0:
		return fmt.Sprintf("от %s%s", trimFloat(r.From), suffix)
	case r.To > 0:
		return fmt.Sprintf("до %s%s", trimFloat(r.To), suffix)
	}
	return "не задан"
}

func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// filtersSummary — строка выбранных фильтров для подтверждения.
func filtersSummary(draft SearchConfig) string {
	defs := FiltersForCategory(draft.Category)
	var parts []string
	for _, d := range defs {
		if vals := draft.Filters[d.Key]; len(vals) > 0 {
			labels := make([]string, 0, len(vals))
			if strings.HasPrefix(d.Key, "filter_bool_") {
				labels = append(labels, "включено")
			} else {
				for _, v := range vals {
					label := v
					for _, fv := range d.Values {
						if fv.Value == v {
							label = fv.Label
							break
						}
					}
					labels = append(labels, label)
				}
			}
			parts = append(parts, d.Label+": "+strings.Join(labels, ", "))
		}
		if r, ok := draft.Ranges[d.Key]; ok && (r.From > 0 || r.To > 0) {
			parts = append(parts, d.Label+": "+rangeLabel(r, d.Unit))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "⚙️ <b>Фильтры:</b> " + html.EscapeString(strings.Join(parts, "; ")) + "\n"
}

// 7. Шаг 4: Выбор цены
func (ui *BotUI) RenderPriceStep(draft SearchConfig) (string, *InlineKeyboardMarkup) {
	q := draft.Query
	if q == "" {
		q = "все лоты"
	}

	brandInfo := ""
	if bd := brandDisplay(draft); bd != "" {
		brandInfo = fmt.Sprintf(" | Бренд: <b>%s</b>", html.EscapeString(bd))
	}

	text := fmt.Sprintf(
		"💰 <b>Ограничение цены</b>\n"+
			"Категория: <b>%s</b>%s | Запрос: <b>%s</b>\n\n"+
			"Выбери верхний порог кнопкой, введи число или задай диапазон <b>от–до</b>:",
		html.EscapeString(draft.CatName), brandInfo, html.EscapeString(q),
	)

	km := markup(
		row(btn("До 5 000 грн", "add_price:5000"), btn("До 10 000 грн", "add_price:10000")),
		row(btn("До 20 000 грн", "add_price:20000"), btn("До 30 000 грн", "add_price:30000")),
		row(btn("До 50 000 грн", "add_price:50000"), btn("♾ Без ограничений", "add_price:0")),
		row(btn("✏️ Ввести свою цену", "add_price_custom")),
		row(btn("↔️ Диапазон (от–до)", "add_price_range")),
		row(btn("❌ Отмена", "menu_main")),
	)

	return text, km
}

// RenderOwnerStep — выбор типа продавца (все / частные / бизнес).
func (ui *BotUI) RenderOwnerStep(draft SearchConfig) (string, *InlineKeyboardMarkup) {
	text := "👤 <b>Кто продаёт?</b>\n\n" +
		"Можно отсечь магазины и бизнес-аккаунты, либо наоборот — смотреть только их."
	km := markup(
		row(btn("👥 Все продавцы", "add_owner:all")),
		row(btn("🙋 Только частные", "add_owner:private"), btn("🏢 Только бизнес", "add_owner:business")),
		row(btn("❌ Отмена", "menu_main")),
	)
	return text, km
}

// 8. Шаг 5: Выбор города
func (ui *BotUI) RenderCityStep(draft SearchConfig) (string, *InlineKeyboardMarkup) {
	text := "📍 <b>Локация поиска</b>\n\nВ каком регионе искать объявления?"

	var rows [][]InlineKeyboardButton
	rows = append(rows, row(btn("🇺🇦 Вся Україна", "add_city:0:0")))

	var currentRow []InlineKeyboardButton
	for i := 1; i < len(TopCities); i++ {
		c := TopCities[i]
		currentRow = append(currentRow, btn(c.Name, fmt.Sprintf("add_city:%d:%d", c.CityID, c.RegionID)))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	rows = append(rows, row(btn("🏙 Другой город / область", "add_city_manual")))
	rows = append(rows, row(btn("❌ Отмена", "menu_main")))

	return text, markup(rows...)
}

// RenderCityResults показывает результаты автокомплита городов/областей.
func (ui *BotUI) RenderCityResults(query string, results []GeoLocationSuggestion) (string, *InlineKeyboardMarkup) {
	text := fmt.Sprintf(
		"🏙 <b>Результаты по запросу:</b> %s\n\n"+
			"Выбери город или всю область:",
		html.EscapeString(query),
	)

	var rows [][]InlineKeyboardButton
	seenRegions := map[int]bool{}

	for i, r := range results {
		if i >= 8 {
			break
		}
		if r.City.ID == 0 {
			// Результат «вся область» без конкретного города.
			if r.Region.ID > 0 && !seenRegions[r.Region.ID] {
				seenRegions[r.Region.ID] = true
				rows = append(rows, row(btn(
					"🌍 Вся "+truncateRunes(r.Region.Name, 45),
					fmt.Sprintf("add_region_pick:%d", r.Region.ID),
				)))
			}
			continue
		}

		label := r.City.Name
		if r.Region.Name != "" {
			label = fmt.Sprintf("%s (%s)", r.City.Name, r.Region.Name)
		}
		rows = append(rows, row(btn(
			truncateRunes(label, 60),
			fmt.Sprintf("add_city_pick:%d:%d", r.City.ID, r.Region.ID),
		)))

		// Для каждой области — отдельная кнопка «вся область»,
		// чтобы захватывать и подгорода вроде Кременчуга.
		if r.Region.ID > 0 && !seenRegions[r.Region.ID] {
			seenRegions[r.Region.ID] = true
			rows = append(rows, row(btn(
				"🌍 Вся "+truncateRunes(r.Region.Name, 45),
				fmt.Sprintf("add_region_pick:%d", r.Region.ID),
			)))
		}
	}

	rows = append(rows, row(btn("⬅️ Назад к городам", "add_city_back"), btn("❌ Отмена", "menu_main")))

	return text, markup(rows...)
}

// RenderDistrictStep показывает районы выбранного города.
func (ui *BotUI) RenderDistrictStep(cityName string, districts []GeoDistrict, cityID, regionID int) (string, *InlineKeyboardMarkup) {
	text := fmt.Sprintf(
		"🏘 <b>Район города %s</b>\n\nВыбери район или ищи по всему городу:",
		html.EscapeString(cityName),
	)

	var rows [][]InlineKeyboardButton
	var currentRow []InlineKeyboardButton

	for _, d := range districts {
		currentRow = append(currentRow, btn(
			truncateRunes(d.Name, 30),
			fmt.Sprintf("add_district_pick:%d:%d:%d", d.ID, cityID, regionID),
		))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	rows = append(rows, row(btn("🏙 Весь город", fmt.Sprintf("add_city_whole:%d:%d", cityID, regionID))))
	rows = append(rows, row(btn("⬅️ Назад к городам", "add_city_back"), btn("❌ Отмена", "menu_main")))

	return text, markup(rows...)
}

// 9. Подтверждение и запуск
func (ui *BotUI) RenderConfirmStep(draft SearchConfig) (string, *InlineKeyboardMarkup) {
	priceStr := "Без ограничений"
	switch {
	case draft.MinPrice > 0 && draft.MaxPrice > 0:
		priceStr = fmt.Sprintf("от %.0f до %.0f грн", draft.MinPrice, draft.MaxPrice)
	case draft.MaxPrice > 0:
		priceStr = fmt.Sprintf("до %.0f грн", draft.MaxPrice)
	case draft.MinPrice > 0:
		priceStr = fmt.Sprintf("от %.0f грн", draft.MinPrice)
	}

	queryStr := html.EscapeString(draft.Query)
	if queryStr == "" {
		queryStr = "<i>(все объявления)</i>"
	}

	cityStr := html.EscapeString(draft.CityName)
	if cityStr == "" {
		cityStr = "Вся Украина"
	}

	catStr := html.EscapeString(draft.CatName)
	if catStr == "" {
		catStr = "Любая"
	}

	brandStr := ""
	if bd := brandDisplay(draft); bd != "" {
		brandStr = fmt.Sprintf("🏷 <b>Бренд:</b> %s\n", html.EscapeString(bd))
	}

	districtStr := ""
	if draft.DistrictName != "" {
		districtStr = fmt.Sprintf("🏘 <b>Район:</b> %s\n", html.EscapeString(draft.DistrictName))
	}

	ownerStr := ""
	switch draft.OwnerType {
	case "private":
		ownerStr = "👤 <b>Продавец:</b> только частные\n"
	case "business":
		ownerStr = "🏢 <b>Продавец:</b> только бизнес\n"
	}
	filtersStr := filtersSummary(draft)
	variantsStr := ""
	if len(draft.QueryVariants) > 0 {
		variantsStr = "🔤 <b>Варианты:</b> " + html.EscapeString(truncateRunes(strings.Join(draft.QueryVariants, ", "), 180)) + "\n"
	}

	name := draft.Name
	if name == "" {
		if draft.Query != "" {
			name = draft.Query
		} else if bd := brandDisplay(draft); bd != "" {
			name = bd
		} else {
			name = draft.CatName
		}
	}
	if name == "" {
		name = "Новый поиск"
	}

	text := fmt.Sprintf(
		"🚀 <b>Всё готово к запуску!</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"🎯 <b>Название:</b> %s\n"+
			"🔍 <b>Запрос:</b> %s\n"+
			"📁 <b>Категория:</b> %s\n"+
			"%s"+
			"💰 <b>Цена:</b> %s\n"+
			"📍 <b>Локация:</b> %s\n"+
			"%s"+
			"%s"+
			"%s"+
			"%s"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Нажми кнопку ниже, чтобы запустить монитор:",
		html.EscapeString(name), queryStr, catStr, brandStr, priceStr, cityStr, districtStr, ownerStr, filtersStr, variantsStr,
	)

	km := markup(
		row(btn("🚀 Запустить мониторинг!", "add_finish")),
		row(btn("❌ Отмена", "menu_main")),
	)

	return text, km
}

// 10. Настройки
func (ui *BotUI) RenderSettings() (string, *InlineKeyboardMarkup) {
	stats := ui.checker.GetStats()

	sendPhotos := ui.cfg.GetSendPhotos()
	maxPhotos := ui.cfg.GetMaxPhotos()
	photoMode := "выключены"
	if sendPhotos {
		photoMode = fmt.Sprintf("включены (до %d фото)", maxPhotos)
	}

	skipBumped := ui.cfg.GetSkipBumpedOld()
	bumpedMode := "по дате создания (поднятые старые скрываются)"
	if !skipBumped {
		bumpedMode = "по дате обновления (поднятые старые приходят)"
	}

	maxAge := ui.cfg.GetMaxAlertAgeMin()
	ageMode := "без ограничений"
	if maxAge > 0 {
		ageMode = fmt.Sprintf("%d мин", maxAge)
	}

	showParams := ui.cfg.GetShowParams()
	maxParams := ui.cfg.GetMaxParams()
	paramsMode := "выключены"
	if showParams {
		paramsMode = fmt.Sprintf("включены (до %d)", maxParams)
	}

	_, fetchLimit := ui.cfg.PerfSettings()
	chatCount := len(ui.cfg.GetTelegramChatIDs())

	text := fmt.Sprintf(
		"⚙️ <b>Настройки системы</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"⏱ <b>Интервал проверки:</b> %d сек.\n"+
			"📦 <b>Лотов за цикл:</b> %d\n"+
			"📤 <b>Получатели алертов:</b> %d\n"+
			"📷 <b>Фото в алертах:</b> %s\n"+
			"♻️ <b>Новизна лота:</b> %s\n"+
			"⏳ <b>Макс. возраст лота:</b> %s\n"+
			"📋 <b>Характеристики:</b> %s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"<i>Pantheon Open-Core Edition v%s</i>",
		stats.Interval, fetchLimit, chatCount, photoMode, bumpedMode, ageMode, paramsMode, appVersion,
	)

	photoBtn := "📷 Фото: ВЫКЛ"
	if sendPhotos {
		photoBtn = fmt.Sprintf("📷 Фото: ВКЛ (%d)", maxPhotos)
	}

	bumpedBtn := "♻️ Скрывать поднятые старые: ВКЛ"
	if !skipBumped {
		bumpedBtn = "♻️ Скрывать поднятые старые: ВЫКЛ"
	}

	ageBtn := "⏳ Возраст лота: без лимита"
	if maxAge > 0 {
		ageBtn = fmt.Sprintf("⏳ Возраст лота: %d мин", maxAge)
	}

	paramsBtn := "📋 Характеристики: ВЫКЛ"
	if showParams {
		paramsBtn = fmt.Sprintf("📋 Характеристики: ВКЛ (%d)", maxParams)
	}

	fetchBtn := fmt.Sprintf("📦 Выдача: %d", fetchLimit)

	km := markup(
		row(btn("⏱ 5 сек", "set_interval:5"), btn("⏱ 10 сек", "set_interval:10")),
		row(btn("⏱ 15 сек", "set_interval:15"), btn("⏱ 30 сек", "set_interval:30")),
		row(btn("⚡ Проверить сейчас", "check_now")),
		row(btn(fetchBtn, "cycle_fetch_limit")),
		row(btn("🆔 chat_id", "show_chatid"), btn("📤 Получатели", "list_chats")),
		row(btn("✅ Привязать этот чат", "bind_current")),
		row(btn(photoBtn, "toggle_photos")),
		row(btn("📷 1", "set_photos:1"), btn("📷 3", "set_photos:3"), btn("📷 5", "set_photos:5")),
		row(btn(bumpedBtn, "toggle_skip_bumped")),
		row(btn(ageBtn, "cycle_alert_age")),
		row(btn(paramsBtn, "toggle_params")),
		row(btn("⬅️ Главное меню", "menu_main")),
	)

	return text, km
}

func onOff(v bool) string {
	if v {
		return "ВКЛ"
	}
	return "ВЫКЛ"
}

// Парсер ссылок OLX
func ParseOLXURL(rawURL string) (*SearchConfig, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	sc := &SearchConfig{}

	// Query
	if query := q.Get("search[query]"); query != "" {
		sc.Query = query
	} else if query := q.Get("q"); query != "" {
		sc.Query = query
	}

	// Price
	if maxP := q.Get("search[filter_float_price:to]"); maxP != "" {
		if p, err := strconv.ParseFloat(maxP, 64); err == nil {
			sc.MaxPrice = p
		}
	}
	if minP := q.Get("search[filter_float_price:from]"); minP != "" {
		if p, err := strconv.ParseFloat(minP, 64); err == nil {
			sc.MinPrice = p
		}
	}

	// City / region из ссылки
	if v := firstNonEmpty(q.Get("search[city_id]"), q.Get("city_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.CityID = id
			if city := FindCity(id); city != nil {
				sc.CityName = city.Name
				sc.RegionID = city.RegionID
			}
		}
	}
	if v := firstNonEmpty(q.Get("search[region_id]"), q.Get("region_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.RegionID = id
		}
	}
	if v := firstNonEmpty(q.Get("search[category_id]"), q.Get("category_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.Category = id
			if cat := FindCategory(id); cat != nil {
				sc.CatName = cat.Name
			}
		}
	}

	if v := firstNonEmpty(q.Get("search[district_id]"), q.Get("district_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.DistrictID = id
		}
	}

	// Тип продавца: приватна особа / бізнес.
	if v := firstNonEmpty(q.Get("owner_type"), q.Get("search[owner_type]"), q.Get("private_business"), q.Get("search[private_business]")); v != "" {
		switch v {
		case "private", "business":
			sc.OwnerType = v
		}
	}

	// Остальные фильтры из ссылки: filter_enum_*[N], filter_float_*[from/to].
	for k, vals := range q {
		base, sub := parseFilterParam(k)
		if base == "" || isReservedFilterKey(base) || base == sc.BrandParam {
			continue
		}
		switch {
		case strings.HasPrefix(base, "filter_float_") || strings.HasPrefix(base, "filter_int_"):
			if sc.Ranges == nil {
				sc.Ranges = map[string]FilterRange{}
			}
			r := sc.Ranges[base]
			for _, v := range vals {
				f, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(v), ",", "."), 64)
				if err != nil {
					continue
				}
				if sub == "to" {
					r.To = f
				} else {
					r.From = f
				}
			}
			sc.Ranges[base] = r
		case strings.HasPrefix(base, "filter_enum_") || strings.HasPrefix(base, "filter_bool_"):
			if sc.Filters == nil {
				sc.Filters = map[string][]string{}
			}
			if strings.HasPrefix(base, "filter_bool_") {
				if len(vals) > 0 {
					sc.Filters[base] = []string{"1"}
				}
				continue
			}
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v == "" || stringInSlice(sc.Filters[base], v) {
					continue
				}
				sc.Filters[base] = append(sc.Filters[base], v)
			}
		}
	}

	// Brand filters: filter_enum_*[0], filter_enum_*[1], search[filter_enum_*][0]...
	brandVals := map[string][]string{}
	for k, vals := range q {
		cleanK := strings.TrimPrefix(k, "search[")
		cleanK = strings.TrimSuffix(cleanK, "]")
		if idx := strings.Index(cleanK, "]"); idx >= 0 {
			cleanK = cleanK[:idx]
		}
		cleanK = strings.TrimSuffix(cleanK, "]")
		if !strings.HasPrefix(cleanK, "filter_enum_") {
			continue
		}
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v != "" {
				brandVals[cleanK] = append(brandVals[cleanK], v)
			}
		}
	}

	if len(brandVals) > 0 {
		chosen := ""
		if sc.Category > 0 {
			if info := GetCategoryBrands(sc.Category); info != nil {
				if _, ok := brandVals[info.ParamName]; ok {
					chosen = info.ParamName
				}
			}
		}
		if chosen == "" {
			for p := range brandVals {
				chosen = p
				break
			}
		}

		sc.BrandParam = chosen
		seen := map[string]bool{}
		for _, id := range brandVals[chosen] {
			if seen[id] {
				continue
			}
			seen[id] = true
			sc.BrandValues = append(sc.BrandValues, id)
		}

		// Стабильный порядок: как в списке брендов категории (Canon, Nikon, ...).
		order := map[string]int{}
		nextOrder := 0
		for _, cinfo := range CategoryBrands {
			if cinfo.ParamName != chosen {
				continue
			}
			for _, b := range cinfo.Brands {
				if _, ok := order[b.ID]; !ok {
					order[b.ID] = nextOrder
					nextOrder++
				}
			}
		}
		sort.SliceStable(sc.BrandValues, func(i, j int) bool {
			oi, iok := order[sc.BrandValues[i]]
			oj, jok := order[sc.BrandValues[j]]
			if iok && jok {
				return oi < oj
			}
			if iok != jok {
				return iok
			}
			return sc.BrandValues[i] < sc.BrandValues[j]
		})

		nameByID := map[string]string{}
		for _, cinfo := range CategoryBrands {
			if cinfo.ParamName == chosen {
				for _, b := range cinfo.Brands {
					nameByID[b.ID] = b.Name
				}
			}
		}
		for _, id := range sc.BrandValues {
			name := nameByID[id]
			if name == "" {
				name = id
			}
			sc.BrandNames = append(sc.BrandNames, name)
		}

		if len(sc.BrandValues) == 1 {
			sc.BrandValue = sc.BrandValues[0]
			sc.BrandName = sc.BrandNames[0]
		}
	}

	// Try extracting slug/query from path if empty
	// e.g. /d/uk/elektronika/telefony-i-aksesuary/q-iphone-15/
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "q-") {
			sc.Query = strings.TrimPrefix(part, "q-")
			sc.Query = strings.ReplaceAll(sc.Query, "-", " ")
		}
		if cat := FindCategoryBySlug(part); cat != nil {
			sc.Category = cat.ID
			sc.CatName = cat.Name
		}
	}

	if sc.Query != "" {
		sc.QueryVariants = QueryVariants(sc.Query)
	}

	if sc.Query != "" {
		sc.Name = sc.Query
	} else if sc.BrandName != "" {
		sc.Name = sc.BrandName
	} else if sc.CatName != "" {
		sc.Name = sc.CatName
	} else {
		sc.Name = "Поиск по ссылке"
	}

	return sc, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// isReservedFilterKey — фильтры, которые обрабатываются отдельными полями.
func isReservedFilterKey(key string) bool {
	switch key {
	case "filter_float_price", "filter_enum_price", "owner_type", "sort_by", "price":
		return true
	}
	return false
}

// parseFilterParam разбирает параметр вида
// "search[filter_enum_color][0]", "filter_float_motor_year:from" и т.п.
func parseFilterParam(raw string) (base, sub string) {
	s := raw
	if strings.HasPrefix(s, "search[") {
		s = s[len("search["):]
		if i := strings.Index(s, "]"); i >= 0 {
			base = s[:i]
			rest := s[i+1:]
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				sub = rest[1 : len(rest)-1]
			}
			return base, sub
		}
	}
	if i := strings.Index(s, "["); i > 0 && strings.HasSuffix(s, "]") {
		return s[:i], s[i+1 : len(s)-1]
	}
	if i := strings.Index(s, ":"); i > 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

func stringInSlice(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

// parsePriceRange разбирает пользовательский ввод диапазона цены:
// "5000-20000", "5 000 – 20 000", "от 5000 до 20000", "до 20000", "от 5000", "5000".
func parsePriceRange(raw string) (min, max float64, ok bool) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.NewReplacer("–", "-", "—", "-", "−", "-", "\u00a0", " ").Replace(s)
	for _, w := range []string{"грн", "uah", "₴", "гривен", "гривень"} {
		s = strings.ReplaceAll(s, w, " ")
	}

	parseNum := func(part string) (float64, bool) {
		var b strings.Builder
		for _, r := range part {
			switch {
			case unicode.IsDigit(r) || r == '.':
				b.WriteRune(r)
			case r == ',':
				b.WriteRune('.')
			}
		}
		if b.Len() == 0 {
			return 0, false
		}
		v, err := strconv.ParseFloat(b.String(), 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}

	// Диапазон через дефис: "5000-20000", "-20000", "5000-".
	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		lo, lok := parseNum(parts[0])
		hi, hik := parseNum(parts[1])
		if !lok && !hik {
			return 0, 0, false
		}
		return lo, hi, true
	}

	hasFrom := strings.Contains(s, "от") || strings.Contains(s, "від")
	hasTo := strings.Contains(s, "до")

	// Склеиваем разряды: "от 5 000 до 20 000" -> "от5000до20000".
	s = strings.ReplaceAll(s, " ", "")

	var nums []float64
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		if v, err := strconv.ParseFloat(cur.String(), 64); err == nil {
			nums = append(nums, v)
		}
		cur.Reset()
	}
	for _, r := range s {
		switch {
		case unicode.IsDigit(r) || r == '.':
			cur.WriteRune(r)
		case r == ',':
			cur.WriteRune('.')
		default:
			flush()
		}
	}
	flush()

	if len(nums) == 0 {
		return 0, 0, false
	}
	if len(nums) == 1 {
		if hasFrom && !hasTo {
			return nums[0], 0, true
		}
		return 0, nums[0], true
	}
	return nums[0], nums[1], true
}

// brandDisplay возвращает строку выбранных брендов для карточек/шагов.
func brandDisplay(draft SearchConfig) string {
	if len(draft.BrandNames) > 0 {
		return strings.Join(draft.BrandNames, ", ")
	}
	return draft.BrandName
}

package bot

import (
	"strings"
	"testing"
	"time"

	"pantheon/core/catalog"
	"pantheon/core/checker"
	"pantheon/core/config"
	"pantheon/core/domain"
	"pantheon/core/olx"
)

func TestRenderCityResults(t *testing.T) {
	ui := &BotUI{}
	results := []olx.GeoLocationSuggestion{
		{
			City:   olx.GeoCity{ID: 221, Name: "Кременчук"},
			Region: olx.GeoRegion{ID: 15, Name: "Полтавська область"},
		},
		{
			City:   olx.GeoCity{ID: 284, Name: "Полтава"},
			Region: olx.GeoRegion{ID: 15, Name: "Полтавська область"},
		},
	}

	_, km := ui.RenderCityResults("Кременчук", results)
	if km == nil {
		t.Fatal("no keyboard")
	}

	var data []string
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			data = append(data, b.CallbackData)
		}
	}

	hasCity1 := false
	hasCity2 := false
	hasRegion := 0
	for _, d := range data {
		switch d {
		case "add_city_pick:221:15":
			hasCity1 = true
		case "add_city_pick:284:15":
			hasCity2 = true
		case "add_region_pick:15":
			hasRegion++
		}
	}
	if !hasCity1 || !hasCity2 {
		t.Errorf("city buttons missing: %v", data)
	}
	if hasRegion != 1 {
		t.Errorf("region button must be unique, got %d: %v", hasRegion, data)
	}
}

func TestRenderCityStepHasManualButton(t *testing.T) {
	ui := &BotUI{}
	_, km := ui.RenderCityStep(domain.Search{})
	if km == nil {
		t.Fatal("no keyboard")
	}
	found := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_city_manual" {
				found = true
			}
		}
	}
	if !found {
		t.Error("city step has no manual city button")
	}
}

func TestRenderCityResultsRegionOnly(t *testing.T) {
	ui := &BotUI{}
	results := []olx.GeoLocationSuggestion{
		{Region: olx.GeoRegion{ID: 15, Name: "Полтавська область"}},
	}
	_, km := ui.RenderCityResults("Полтавська", results)
	if km == nil {
		t.Fatal("no keyboard")
	}
	found := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_region_pick:15" {
				found = true
			}
		}
	}
	if !found {
		t.Error("region-only result must produce an add_region_pick button")
	}
}

func TestRenderPriceStepHasRangeButton(t *testing.T) {
	ui := &BotUI{}
	_, km := ui.RenderPriceStep(domain.Search{})
	if km == nil {
		t.Fatal("no keyboard")
	}
	found := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_price_range" {
				found = true
			}
		}
	}
	if !found {
		t.Error("price step has no range button")
	}
}

func TestParsePriceRange(t *testing.T) {
	checks := []struct {
		in       string
		min, max float64
		ok       bool
	}{
		{"5000-20000", 5000, 20000, true},
		{"5 000 – 20 000", 5000, 20000, true},
		{"от 5000 до 20000", 5000, 20000, true},
		{"від 5000 до 20000", 5000, 20000, true},
		{"до 20000", 0, 20000, true},
		{"от 5000", 5000, 0, true},
		{"5000", 0, 5000, true},
		{"1000-2000 грн", 1000, 2000, true},
		{"-20000", 0, 20000, true},
		{"5000-", 5000, 0, true},
		{"грн", 0, 0, false},
		{"abc", 0, 0, false},
	}

	for _, c := range checks {
		gotMin, gotMax, ok := parsePriceRange(c.in)
		if ok != c.ok || gotMin != c.min || gotMax != c.max {
			t.Errorf("parsePriceRange(%q) = (%v, %v, %v), want (%v, %v, %v)",
				c.in, gotMin, gotMax, ok, c.min, c.max, c.ok)
		}
	}
}

func TestRenderDistrictStep(t *testing.T) {
	ui := &BotUI{}
	districts := []olx.GeoDistrict{
		{ID: 11, Name: "Печерський"},
		{ID: 13, Name: "Подільський"},
	}
	_, km := ui.RenderDistrictStep("Київ", districts, 268, 25)
	if km == nil {
		t.Fatal("no keyboard")
	}
	foundDistrict := false
	foundWhole := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_district_pick:11:268:25" {
				foundDistrict = true
			}
			if b.CallbackData == "add_city_whole:268:25" {
				foundWhole = true
			}
		}
	}
	if !foundDistrict {
		t.Error("district button missing")
	}
	if !foundWhole {
		t.Error("whole-city button missing")
	}
}

func TestRenderSettingsCategories(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := &config.Config{
		IntervalSeconds: 15,
		FetchLimit:      35,
	}
	chk := checker.NewChecker(cfg, nil, nil)
	defer chk.Close()
	ui := NewBotUI(cfg, chk)

	// 1. Main settings
	text, km := ui.RenderSettings()
	if !strings.Contains(text, "Настройки системы") {
		t.Errorf("expected header in settings text, got: %s", text)
	}
	hasSpeed, hasAlerts, hasChats := false, false, false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "settings_speed" {
				hasSpeed = true
			}
			if b.CallbackData == "settings_alerts" {
				hasAlerts = true
			}
			if b.CallbackData == "settings_chats" {
				hasChats = true
			}
		}
	}
	if !hasSpeed || !hasAlerts || !hasChats {
		t.Errorf("missing category buttons: speed=%v, alerts=%v, chats=%v", hasSpeed, hasAlerts, hasChats)
	}

	// 2. Settings Speed
	textSpeed, kmSpeed := ui.RenderSettingsSpeed()
	if !strings.Contains(textSpeed, "Скорость и парсер") {
		t.Errorf("expected header in speed text, got: %s", textSpeed)
	}
	if kmSpeed == nil {
		t.Fatal("no keyboard in speed")
	}

	// 3. Settings Alerts
	textAlerts, kmAlerts := ui.RenderSettingsAlerts()
	if !strings.Contains(textAlerts, "Оформление и фильтры") {
		t.Errorf("expected header in alerts text, got: %s", textAlerts)
	}
	if kmAlerts == nil {
		t.Fatal("no keyboard in alerts")
	}

	// 4. Settings Chats
	textChats, kmChats := ui.RenderSettingsChats("123456")
	if !strings.Contains(textChats, "Чаты и доступ") {
		t.Errorf("expected header in chats text, got: %s", textChats)
	}
	if kmChats == nil {
		t.Fatal("no keyboard in chats")
	}
}

func TestDeepCategoryKeyboards(t *testing.T) {
	ui := &BotUI{}

	// Внутри "Фото/Відео" должна быть кнопка "Цифрові фотоапарати".
	_, km := ui.RenderSubcategoryStep(429)
	if !hasCallback(km, "add_subcat:73") {
		t.Error("subcategory step 429 has no add_subcat:73 button")
	}
	if !hasCallback(km, "add_cat_back:37") {
		t.Error("subcategory step 429 has no back-to-37 button")
	}

	// Внутри "Електроніка" должна быть кнопка "Фото/Відео".
	_, km = ui.RenderSubcategoryStep(37)
	if !hasCallback(km, "add_subcat:429") {
		t.Error("subcategory step 37 has no add_subcat:429 button")
	}

	// Для листа 73 должны показываться бренды.
	_, km = ui.RenderBrandStep(73, 0, nil)
	if !hasCallback(km, "add_brand_toggle:73:0:2580") {
		t.Error("brand step 73 has no Canon toggle button")
	}
}

func hasCallback(km *InlineKeyboardMarkup, data string) bool {
	if km == nil {
		return false
	}
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == data {
				return true
			}
		}
	}
	return false
}

func TestCategoryFiltersSkipsReserved(t *testing.T) {
	defs := (&BotUI{}).categoryFilters(1115, "")
	if len(defs) == 0 {
		t.Fatal("no filters for category 1115")
	}
	for _, d := range defs {
		if catalog.IsReservedFilterKey(d.Key) {
			t.Errorf("reserved filter %q leaked into UI", d.Key)
		}
	}
	hasColor := false
	for _, d := range defs {
		if d.Key == "filter_enum_color" {
			hasColor = true
		}
	}
	if !hasColor {
		t.Error("color filter missing for category 1115")
	}
}

func TestFiltersSummary(t *testing.T) {
	draft := domain.Search{
		Category: 1115,
		Filters: map[string][]string{
			"filter_enum_color": {"4"},
		},
		Ranges: map[string]domain.FilterRange{
			"filter_float_motor_year": {From: 2000, To: 2010},
		},
	}
	got := filtersSummary(draft)
	if !strings.Contains(got, "олір") && !strings.Contains(got, "олiр") {
		t.Errorf("summary has no color label: %q", got)
	}
	if !strings.Contains(got, "2000") || !strings.Contains(got, "2010") {
		t.Errorf("summary has no year range: %q", got)
	}
}

func TestRangeLabel(t *testing.T) {
	cases := []struct {
		r    domain.FilterRange
		unit string
		want string
	}{
		{domain.FilterRange{From: 2000, To: 2010}, "", "от 2000 до 2010"},
		{domain.FilterRange{From: 100}, "тис.км", "от 100 тис.км"},
		{domain.FilterRange{To: 1.6}, "л", "до 1.6 л"},
	}
	for _, c := range cases {
		if got := rangeLabel(c.r, c.unit); got != c.want {
			t.Errorf("rangeLabel(%+v, %q) = %q, want %q", c.r, c.unit, got, c.want)
		}
	}
}

func TestRentCategoryFilterKeys(t *testing.T) {
	defs := (&BotUI{}).categoryFilters(1760, "")
	has := map[string]bool{}
	for _, d := range defs {
		has[d.Key] = true
	}
	for _, key := range []string{"filter_enum_number_of_rooms_string", "filter_float_total_area", "filter_bool_commission"} {
		if !has[key] {
			t.Errorf("rent filters missing %q", key)
		}
	}
}

func TestFilterChatIDsForOwner(t *testing.T) {
	all := []string{"111", "222"}

	if got := filterChatIDsForOwner(all, ""); len(got) != 2 {
		t.Fatalf("shared filter must go to all chats, got %v", got)
	}

	if got := filterChatIDsForOwner(all, "222"); len(got) != 1 || got[0] != "222" {
		t.Fatalf("owned filter must go only to owner, got %v", got)
	}

	if got := filterChatIDsForOwner(all, "999"); got != nil {
		t.Fatalf("owner missing from recipients must disable delivery, got %v", got)
	}
}

func TestSearchForChatAccess(t *testing.T) {
	cfg := &config.Config{Searches: []domain.Search{
		{Name: "my", OwnerChatID: "111"},
		{Name: "other", OwnerChatID: "222"},
		{Name: "shared"},
	}}
	ui := NewBotUI(cfg, nil)

	if _, ok := ui.searchForChat("111", 0); !ok {
		t.Error("owner must access own filter")
	}
	if _, ok := ui.searchForChat("111", 1); ok {
		t.Error("user must not access another user's filter")
	}
	if _, ok := ui.searchForChat("111", 2); !ok {
		t.Error("shared filter must be accessible to everyone")
	}
}

func TestRenderFiltersListHidesOtherUsers(t *testing.T) {
	cfg := &config.Config{Searches: []domain.Search{
		{Name: "мой телефон", OwnerChatID: "111"},
		{Name: "чужая недвижка", OwnerChatID: "222"},
		{Name: "общий фильтр"},
	}}
	ui := NewBotUI(cfg, nil)

	_, km := ui.RenderFiltersList("111")
	seen := map[string]bool{}
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			seen[b.CallbackData] = true
		}
	}

	if !seen["f_view:0"] {
		t.Error("own filter must be visible")
	}
	if seen["f_view:1"] {
		t.Error("another user's filter must be hidden")
	}
	if !seen["f_view:2"] {
		t.Error("shared filter must be visible to everyone")
	}
}

func TestDeadChatDetection(t *testing.T) {
	dead := []string{
		`HTTP 403: {"ok":false,"error_code":403,"description":"Forbidden: user is deactivated"}`,
		`HTTP 403: {"ok":false,"error_code":403,"description":"Forbidden: bot was blocked by the user"}`,
		`HTTP 400: {"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`,
		`Forbidden: bot can't initiate conversation with a user`,
	}
	for _, s := range dead {
		if !isDeadChatError(s) {
			t.Errorf("isDeadChatError(%q) = false, want true", s)
		}
	}

	alive := []string{
		`HTTP 400: Bad Request: message is too long`,
		`HTTP 429: Too Many Requests: retry after 5`,
		``,
	}
	for _, s := range alive {
		if isDeadChatError(s) {
			t.Errorf("isDeadChatError(%q) = true, want false", s)
		}
	}
}

func TestIsPhotoRelatedError(t *testing.T) {
	yes := []string{
		`sendRichMessage HTTP 400: {"description":"Bad Request: RICH_MESSAGE_PHOTO_NO_MEDIA_FOUND"}`,
		`sendMediaGroup HTTP 400: WEBPAGE_CURL_FAILED`,
		`sendPhoto HTTP 400: Bad Request: failed to get HTTP URL content`,
		`Post "https://api.telegram.org/...": context deadline exceeded`,
	}
	for _, s := range yes {
		if !isPhotoRelatedError(errString(s)) {
			t.Errorf("isPhotoRelatedError(%q) = false, want true", s)
		}
	}
	no := []string{
		`Bad Request: chat not found`,
		`Bad Request: message is too long`,
	}
	for _, s := range no {
		if isPhotoRelatedError(errString(s)) {
			t.Errorf("isPhotoRelatedError(%q) = true, want false", s)
		}
	}
	if isPhotoRelatedError(nil) {
		t.Error("nil must not be a photo error")
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		body string
		want time.Duration
	}{
		{`{"ok":false,"error_code":429,"description":"Too Many Requests: retry after 7","parameters":{"retry_after":7}}`, 7 * time.Second},
		{`{"ok":false,"error_code":429,"retry_after":3}`, 3 * time.Second},
		{`{"ok":false,"error_code":429}`, defaultRetryWait},
		{`not json`, defaultRetryWait},
	}
	for _, c := range cases {
		if got := parseRetryAfter([]byte(c.body)); got != c.want {
			t.Errorf("parseRetryAfter(%q) = %s, want %s", c.body, got, c.want)
		}
	}
}

func TestWaitChatReservesSlots(t *testing.T) {
	n := NewTelegramNotifier("", nil)
	start := time.Now()
	n.waitChat("1")
	n.waitChat("1")
	elapsed := time.Since(start)
	if elapsed < telegramChatGap-100*time.Millisecond {
		t.Errorf("second send to same chat waited only %s, want >= %s", elapsed, telegramChatGap)
	}

	start = time.Now()
	n.waitChat("2")
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("first send to a fresh chat waited %s, want ~0", elapsed)
	}
}

func TestMarkAndReviveChat(t *testing.T) {
	n := NewTelegramNotifier("", nil)
	n.markChatDead("123", "test")
	if !n.isChatDead("123") {
		t.Fatal("chat must be dead")
	}
	n.reviveChat("123")
	if n.isChatDead("123") {
		t.Fatal("chat must be alive after revive")
	}
}

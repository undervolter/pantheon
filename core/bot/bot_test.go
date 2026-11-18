package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"pantheon/core/catalog"
	"pantheon/core/config"
	"pantheon/core/domain"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/telegram/tgtest"
)

var ctx = context.Background()

// ---------- harness ----------

type fakeGeo struct {
	mu        sync.Mutex
	results   map[string][]olx.GeoLocationSuggestion
	districts map[int][]olx.GeoDistrict
	slugs     map[string]olx.GeoCity
	block     chan struct{} // SearchLocations waits on it when set
	searches  int
}

func (g *fakeGeo) SearchLocations(query string) ([]olx.GeoLocationSuggestion, error) {
	g.mu.Lock()
	g.searches++
	block := g.block
	res := g.results[strings.ToLower(query)]
	g.mu.Unlock()
	if block != nil {
		<-block
	}
	return res, nil
}

func (g *fakeGeo) FetchCityDistricts(cityID int) ([]olx.GeoDistrict, error) {
	return g.districts[cityID], nil
}

func (g *fakeGeo) FetchCity(cityID int) (*olx.GeoCity, error) {
	return &olx.GeoCity{ID: cityID, Name: fmt.Sprintf("Місто %d", cityID), RegionID: 7}, nil
}

func (g *fakeGeo) FetchRegion(regionID int) (*olx.GeoRegion, error) {
	return &olx.GeoRegion{ID: regionID, Name: "Полтавська область"}, nil
}

func (g *fakeGeo) ResolveSlug(slug string) (olx.GeoCity, olx.GeoRegion, bool) {
	city, ok := g.slugs[slug]
	return city, olx.GeoRegion{ID: city.RegionID, Name: "Київська область"}, ok
}

func (g *fakeGeo) WarmGeoIndex() {}

type fakeScanner struct {
	mu       sync.Mutex
	triggers int
	status   ScanStatus
}

func (s *fakeScanner) TriggerCheck() { s.mu.Lock(); s.triggers++; s.mu.Unlock() }
func (s *fakeScanner) Status() ScanStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

const (
	admin = 111
	user  = 222
	group = -100500
)

type harness struct {
	t    *testing.T
	cfg  *config.Config
	db   *store.Store
	tg   *tgtest.Server
	geo  *fakeGeo
	scan *fakeScanner
	b    *Bot
	log  *bytes.Buffer
	seq  int
}

// newHarness builds a bot with admin 111, user 222 and group -100500 trusted,
// unless a config is given.
func newHarness(t *testing.T, configJSON string) *harness {
	t.Helper()
	dir := t.TempDir()
	if configJSON == "" {
		configJSON = `{"telegram_bot_token":"` + tgtest.Token + `","telegram_chat_ids":["111","222","-100500"],"admin_chat_ids":["111"],"language":"ru"}`
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(configJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(dir, "pantheon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	h := &harness{
		t: t, cfg: cfg, db: db, tg: tgtest.New(t), log: &bytes.Buffer{},
		geo:  &fakeGeo{results: map[string][]olx.GeoLocationSuggestion{}, districts: map[int][]olx.GeoDistrict{}, slugs: map[string]olx.GeoCity{}},
		scan: &fakeScanner{status: ScanStatus{Started: true}},
	}
	t.Cleanup(logx.SetOutput(h.log))
	h.b = New(telegram.New(tgtest.Token).WithBaseURL(h.tg.URL), cfg, db, h.geo, h.scan)
	h.b.username = "pantheon_test_bot"
	return h
}

func chatType(chatID, userID int64) string {
	if chatID == userID {
		return "private"
	}
	return "supergroup"
}

// say delivers a text message from userID in chatID.
func (h *harness) say(chatID, userID int64, text string) {
	h.t.Helper()
	h.seq++
	h.b.onMessage(ctx, &telegram.TGMessage{
		MessageID: 9000 + h.seq,
		From:      telegram.TGUser{ID: userID, LanguageCode: "ru"},
		Chat:      telegram.TGChat{ID: chatID, Type: chatType(chatID, userID)},
		Text:      text,
	})
}

// press presses a button of the newest bot message.
func (h *harness) press(chatID, userID int64, data string) {
	h.t.Helper()
	h.pressOn(chatID, userID, h.tg.LastMessageID(), data)
}

func (h *harness) pressOn(chatID, userID int64, msgID int, data string) {
	h.t.Helper()
	h.seq++
	h.b.onCallback(ctx, &telegram.CallbackQuery{
		ID:      fmt.Sprintf("cb%d", h.seq),
		From:    telegram.TGUser{ID: userID, LanguageCode: "ru"},
		Message: &telegram.TGMessage{MessageID: msgID, Chat: telegram.TGChat{ID: chatID, Type: chatType(chatID, userID)}},
		Data:    data,
	})
}

// screen returns the newest message or edit the bot produced.
func (h *harness) screen() tgtest.Call {
	h.t.Helper()
	sent := h.tg.Sent()
	if len(sent) == 0 {
		h.t.Fatal("the bot showed nothing")
	}
	return sent[len(sent)-1]
}

func (h *harness) text() string { return h.screen().Text() }

// toast returns the text and kind of the last button answer.
func (h *harness) toast() (text string, modal bool) {
	h.t.Helper()
	call, ok := h.tg.Last("answerCallbackQuery")
	if !ok {
		h.t.Fatal("the button was not answered")
	}
	text, _ = call.Payload["text"].(string)
	modal, _ = call.Payload["show_alert"].(bool)
	return text, modal
}

func (h *harness) filters() []domain.Search {
	h.t.Helper()
	list, err := h.db.Searches(ctx)
	if err != nil {
		h.t.Fatal(err)
	}
	return list
}

func (h *harness) addFilter(s domain.Search) domain.Search {
	h.t.Helper()
	out, err := h.db.AddSearch(ctx, s)
	if err != nil {
		h.t.Fatal(err)
	}
	return out
}

func has(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func mustContain(t *testing.T, text string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(text, w) {
			t.Errorf("screen lacks %q:\n%s", w, text)
		}
	}
}

// state of a user in a private chat.
func (h *harness) state(chatID, userID int64) *state {
	return h.b.states.get(stateKey(fmt.Sprint(chatID), fmt.Sprint(userID)))
}

// ---------- access ----------

// The hole of 0.1.x: a button press ran the same check that auto-bound
// unknown users, so any member of a group with the menu in it became trusted.
func TestStrangerPressingAButtonIsDeniedAndNotBound(t *testing.T) {
	h := newHarness(t, "")
	h.pressOn(group, 999, 55, "menu_main")

	text, modal := h.toast()
	if !modal || !strings.Contains(text, "Нет доступа") {
		t.Errorf("answer = %q modal=%v, want a denial", text, modal)
	}
	if h.cfg.IsTrusted("999") {
		t.Fatal("pressing a button must never make a user trusted")
	}
	if n := len(h.tg.Sent()); n != 0 {
		t.Errorf("a stranger's press must not draw anything, got %d screens", n)
	}
}

func TestTrustedUserInAnUnlinkedGroupIsDenied(t *testing.T) {
	h := newHarness(t, "")
	h.pressOn(-777, user, 55, "menu_main")
	if _, modal := h.toast(); !modal {
		t.Error("a button in a chat that is not linked must be refused")
	}
	if n := len(h.tg.Sent()); n != 0 {
		t.Errorf("got %d screens", n)
	}
}

func TestStrangerGetsOneThrottledReply(t *testing.T) {
	h := newHarness(t, "")
	h.say(999, 999, "hello")
	h.say(999, 999, "/help")
	if n := len(h.tg.Sent()); n != 0 {
		t.Fatalf("a stranger is answered only on /start, got %d replies", n)
	}

	h.say(999, 999, "/start")
	mustContain(t, h.text(), "Доступ закрыт", "<code>999</code>")
	h.say(999, 999, "/start")
	h.say(999, 999, "/start")
	if n := len(h.tg.Sent()); n != 1 {
		t.Errorf("replies = %d, want one per hour", n)
	}
	if h.cfg.IsTrusted("999") {
		t.Fatal("talking to the bot must never make a user trusted")
	}
}

func TestPairingBindsTheFirstAdminOnly(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["your_chat_id"]}`)
	h.b.preparePairing()
	code := h.b.pair.code
	if len(code) != pairingCodeLen {
		t.Fatalf("no pairing code was issued: %q", code)
	}
	if !strings.Contains(h.log.String(), "/start "+code) {
		t.Errorf("the code must be printed to the console:\n%s", h.log.String())
	}

	h.say(500, 500, "/start wrong-code")
	mustContain(t, h.text(), "Код привязки не подошёл")
	if h.cfg.IsTrusted("500") {
		t.Fatal("a wrong code must not bind")
	}

	// The code works only in a private chat: in a group it would be read by
	// every member.
	h.say(-900, 600, "/start "+code)
	if h.cfg.IsTrusted("600") || h.cfg.IsTrusted("-900") {
		t.Fatal("pairing must not work from a group")
	}

	h.say(700, 700, "/start "+code)
	if !h.cfg.IsAdmin("700") || !h.cfg.IsTrusted("700") {
		t.Fatal("the right code binds the admin")
	}
	mustContain(t, h.text(), "ты администратор", "🐉 Pantheon")

	// The code is single-use.
	h.say(800, 800, "/start "+code)
	if h.cfg.IsTrusted("800") {
		t.Fatal("the code must not work twice")
	}
	if h.b.pairingOpen() {
		t.Error("pairing must be closed once there is an admin")
	}
}

func TestPairingCodeRotatesAfterWrongGuesses(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`"}`)
	h.b.preparePairing()
	first := h.b.pair.code
	for i := 0; i < maxPairingFailures; i++ {
		h.say(int64(1000+i), int64(1000+i), "/start guess")
	}
	if h.b.pair.code == first || h.b.pair.code == "" {
		t.Fatalf("the code must be replaced after %d wrong guesses", maxPairingFailures)
	}
	h.say(2000, 2000, "/start "+first)
	if h.cfg.IsTrusted("2000") {
		t.Fatal("the old code must be dead")
	}
}

func TestConfigWithUsersNeedsNoPairing(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["-100500","333","444"]}`)
	h.b.preparePairing()
	if h.b.pairingOpen() {
		t.Fatal("a config that lists a user gets that user as the admin")
	}
	if !h.cfg.IsAdmin("333") || h.cfg.IsAdmin("444") {
		t.Fatalf("admins = %v, want the first user", h.cfg.Admins())
	}
}

func TestAdminOnlyButtonsAndCommands(t *testing.T) {
	h := newHarness(t, "")
	before := h.cfg.GetInterval()

	h.pressOn(user, user, 55, "set_interval")
	if text, modal := h.toast(); !modal || !strings.Contains(text, "только администратору") {
		t.Errorf("answer = %q modal=%v", text, modal)
	}
	if h.cfg.GetInterval() != before {
		t.Fatal("a non-admin changed a bot-wide setting")
	}

	h.pressOn(admin, admin, 56, "set_interval")
	if h.cfg.GetInterval() == before {
		t.Fatal("the admin must be able to change it")
	}

	for _, cmd := range []string{"/delchat 111", "/addchat 555", "/chats"} {
		h.tg.Reset()
		h.say(user, user, cmd)
		mustContain(t, h.text(), "только администратору")
	}
	if !h.cfg.IsTrusted("111") || h.cfg.IsTrusted("555") {
		t.Fatal("a non-admin changed the chat list")
	}
}

func TestLastAdminCannotBeRemoved(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/delchat 111")
	mustContain(t, h.text(), "Нельзя убрать единственного администратора")
	if !h.cfg.IsTrusted("111") {
		t.Fatal("the admin was removed")
	}
}

// Removing a chat pauses its filters first: left running they would poll OLX
// for nobody and be invisible in every menu.
func TestRemovingAChatPausesItsFilters(t *testing.T) {
	h := newHarness(t, "")
	theirs := h.addFilter(domain.Search{Name: "theirs", Query: "a", OwnerChatID: "222"})
	mine := h.addFilter(domain.Search{Name: "mine", Query: "b", OwnerChatID: "111"})

	h.say(admin, admin, "/delchat 222")
	mustContain(t, h.text(), "<code>222</code>", "1")
	if h.cfg.IsTrusted("222") {
		t.Fatal("the chat is still trusted")
	}
	if got, _ := h.db.Search(ctx, theirs.ID); !got.Disabled {
		t.Error("the removed chat's filter must be paused")
	}
	if got, _ := h.db.Search(ctx, mine.ID); got.Disabled {
		t.Error("other filters must not be touched")
	}
}

func TestBindAndUnbindAGroup(t *testing.T) {
	h := newHarness(t, "")
	// A command addressed to this bot by name, as Telegram sends it in groups.
	h.say(-4242, admin, "/bind@pantheon_test_bot")
	if !h.cfg.IsTrusted("-4242") {
		t.Fatal("/bind@bot by the admin must link the group")
	}
	mustContain(t, h.text(), "Группа привязана")

	h.say(-4343, user, "/bind")
	if h.cfg.IsTrusted("-4343") {
		t.Fatal("only an admin links a group")
	}

	h.say(-4242, admin, "/unbind")
	if h.cfg.IsTrusted("-4242") {
		t.Fatal("/unbind must unlink the group")
	}
}

func TestParseCommand(t *testing.T) {
	cases := []struct {
		text, cmd, args string
		ok              bool
	}{
		{"/start", "start", "", true},
		{"/START", "start", "", true},
		{"/bind@pantheon_test_bot", "bind", "", true},
		{"/bind@Pantheon_Test_Bot now", "bind", "now", true},
		{"/bind@other_bot", "", "", false},
		{"/addchat  -100123 ", "addchat", "-100123", true},
		{"/addchatxyz", "addchatxyz", "", true},
		{"hello", "", "", false},
		{"/", "", "", false},
	}
	for _, c := range cases {
		cmd, args, ok := parseCommand(c.text, "pantheon_test_bot")
		if cmd != c.cmd || args != c.args || ok != c.ok {
			t.Errorf("parseCommand(%q) = %q, %q, %v; want %q, %q, %v", c.text, cmd, args, ok, c.cmd, c.args, c.ok)
		}
	}
}

func TestOrdinaryChatterInAGroupIsIgnored(t *testing.T) {
	h := newHarness(t, "")
	h.say(group, admin, "привет всем")
	h.say(group, 999, "/start")
	if n := len(h.tg.Calls("")); n != 0 {
		t.Fatalf("the bot reacted to group chatter with %d calls", n)
	}
	h.say(admin, admin, "привет")
	mustContain(t, h.text(), "🐉 Pantheon")
}

// ---------- menus ----------

func TestMainMenu(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/start")
	mustContain(t, h.text(), "<h3>🐉 Pantheon</h3>", "Фильтров пока нет")
	buttons := h.screen().Buttons()
	for _, want := range []string{"add_new", "menu_list", "history_show", "menu_settings", "chat_pause", "lang_menu", "help_show"} {
		if !has(buttons, want) {
			t.Errorf("main menu lacks %s: %v", want, buttons)
		}
	}

	h.addFilter(domain.Search{Name: "a", Query: "a", OwnerChatID: "111"})
	h.addFilter(domain.Search{Name: "b", Query: "b", OwnerChatID: "111", Disabled: true})
	h.addFilter(domain.Search{Name: "чужой", Query: "c", OwnerChatID: "222"})
	h.press(admin, admin, "menu_main")
	mustContain(t, h.text(), "Сканер работает", "1 / 2", "🎯 Фильтры")
	if strings.Contains(h.text(), "чужой") {
		t.Error("another chat's filter leaked into the menu")
	}
}

func TestScannerLineReportsABlock(t *testing.T) {
	h := newHarness(t, "")
	h.addFilter(domain.Search{Name: "a", Query: "a", OwnerChatID: "111"})
	h.scan.status = ScanStatus{Started: true, BlockedUntil: time.Now().Add(10 * time.Minute)}
	h.say(admin, admin, "/start")
	mustContain(t, h.text(), "OLX ограничил запросы")
}

func TestLanguageSwitch(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/start")
	h.press(admin, admin, "lang:uk")
	if h.cfg.LangFor("111") != "uk" {
		t.Fatalf("language = %q", h.cfg.LangFor("111"))
	}
	mustContain(t, h.text(), "Фільтрів поки немає")
	if text, _ := h.toast(); !strings.Contains(text, "Українська") {
		t.Errorf("toast = %q", text)
	}
	// Another chat keeps its own language.
	h.say(user, user, "/start")
	mustContain(t, h.text(), "Фильтров пока нет")
}

func TestPauseAndResumeChat(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/start")
	h.press(admin, admin, "chat_pause")
	if !h.cfg.IsPaused("111") {
		t.Fatal("the chat must be paused")
	}
	mustContain(t, h.text(), "Алерты в этот чат на паузе")
	if !has(h.screen().Buttons(), "chat_resume") {
		t.Error("the button must turn into resume")
	}
	h.say(admin, admin, "/resume")
	if h.cfg.IsPaused("111") {
		t.Fatal("/resume must lift the pause")
	}
}

func TestSettingsCycleAndQuietHours(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/settings")
	mustContain(t, h.text(), "⚙️ Настройки", "🔔 Алерты", "🛠 Система")

	h.press(admin, admin, "set_quiet")
	if from, to, ok := h.cfg.GetQuietHours("111"); !ok || from != 22 || to != 8 {
		t.Fatalf("quiet hours = %d-%d %v, want the first preset", from, to, ok)
	}
	mustContain(t, h.text(), "22:00–08:00")
	for i := 0; i < len(quietPresets); i++ {
		h.press(admin, admin, "set_quiet")
	}
	if _, _, ok := h.cfg.GetQuietHours("111"); ok {
		t.Fatal("the ring of presets must end with off")
	}

	h.press(admin, admin, "set_style")
	if h.cfg.GetRichAlerts() {
		t.Fatal("the card style must toggle")
	}
	h.press(admin, admin, "set_age")
	if got := h.cfg.GetMaxAlertAgeMin(); got != 45 {
		t.Errorf("age after one press from 30 = %d, want 45", got)
	}
	h.press(admin, admin, "check_now")
	if h.scan.triggers != 1 {
		t.Errorf("check now must nudge the scanner, triggers = %d", h.scan.triggers)
	}

	// A plain user sees the screen but only his own switches.
	h.say(user, user, "/settings")
	buttons := h.screen().Buttons()
	if has(buttons, "set_interval") || !has(buttons, "set_quiet") {
		t.Errorf("user settings buttons = %v", buttons)
	}
	if strings.Contains(h.text(), "🛠 Система") {
		t.Error("the system section is for the admin")
	}
}

func TestNextIntRing(t *testing.T) {
	steps := []int{0, 20, 30, 45, 60, 120}
	for cur, want := range map[int]int{0: 20, 20: 30, 120: 0, 10: 20, 25: 30, 500: 0} {
		if got := nextInt(steps, cur); got != want {
			t.Errorf("nextInt(%d) = %d, want %d", cur, got, want)
		}
	}
}

// ---------- filters ----------

func TestFiltersListAndCard(t *testing.T) {
	h := newHarness(t, "")
	mine := h.addFilter(domain.Search{Name: "Телефоны", Query: "iphone 15", MaxPrice: 30000, Category: 44, CatName: "x",
		CityID: 268, CityName: "Київ", OwnerChatID: "111", ExcludeWords: []string{"чохол"}})
	free := h.addFilter(domain.Search{Name: "Даром", Category: 44, PriceMode: domain.PriceFree, OwnerChatID: "111"})
	shared := h.addFilter(domain.Search{Name: "Общий", Query: "x"})
	h.addFilter(domain.Search{Name: "Чужой", Query: "y", OwnerChatID: "222"})

	h.say(admin, admin, "/filters")
	mustContain(t, h.text(), "Телефоны", "до 30 000 грн", "бесплатно", "Общий", "👥 — общий фильтр")
	if strings.Contains(h.text(), "Чужой") {
		t.Error("another chat's filter is listed")
	}
	buttons := h.screen().Buttons()
	for _, s := range []domain.Search{mine, free, shared} {
		if !has(buttons, "f_view:"+s.ID) {
			t.Errorf("no button for %q: %v", s.Name, buttons)
		}
	}

	h.press(admin, admin, "f_view:"+mine.ID)
	mustContain(t, h.text(), "🎯 Телефоны", "🟢 Активен", "только этот чат", "iphone 15", "до 30 000 грн", "📍 Київ", "чохол")
	h.press(admin, admin, "f_view:"+free.ID)
	mustContain(t, h.text(), "только бесплатные")
}

// Filters are addressed by id. With the slice index of 0.1.x a stale button
// acted on whatever filter had moved into that position.
func TestStaleFilterButtonCannotHitAnotherFilter(t *testing.T) {
	h := newHarness(t, "")
	a := h.addFilter(domain.Search{Name: "a", Query: "a", OwnerChatID: "111"})
	b := h.addFilter(domain.Search{Name: "b", Query: "b", OwnerChatID: "111"})
	c := h.addFilter(domain.Search{Name: "c", Query: "c", OwnerChatID: "111"})
	h.say(admin, admin, "/filters")
	old := h.tg.LastMessageID()

	// Delete the first filter through its confirmation.
	h.press(admin, admin, "f_del:"+a.ID)
	mustContain(t, h.text(), "Удалить фильтр «a»?")
	if len(h.filters()) != 3 {
		t.Fatal("asking for confirmation must not delete anything")
	}
	h.press(admin, admin, "f_delok:"+a.ID)
	if len(h.filters()) != 2 {
		t.Fatal("the filter was not deleted")
	}

	// The same buttons pressed again on the old message.
	h.pressOn(admin, admin, old, "f_delok:"+a.ID)
	if text, modal := h.toast(); !modal || !strings.Contains(text, "Фильтр не найден") {
		t.Errorf("answer = %q modal=%v", text, modal)
	}
	left := h.filters()
	if len(left) != 2 || left[0].ID != b.ID || left[1].ID != c.ID {
		t.Fatalf("a stale button deleted another filter: %+v", left)
	}
}

func TestFilterOfAnotherChatIsOutOfReach(t *testing.T) {
	h := newHarness(t, "")
	theirs := h.addFilter(domain.Search{Name: "theirs", Query: "a", OwnerChatID: "222"})
	for _, data := range []string{"f_view:", "f_toggle:", "f_delok:"} {
		h.pressOn(admin, admin, 55, data+theirs.ID)
		if text, modal := h.toast(); !modal || !strings.Contains(text, "Фильтр не найден") {
			t.Errorf("%s: answer = %q modal=%v", data, text, modal)
		}
	}
	if got, _ := h.db.Search(ctx, theirs.ID); got.Disabled || got.ID == "" {
		t.Fatal("another chat's filter was changed")
	}
}

func TestSharedFilterIsManagedByAdminOnly(t *testing.T) {
	h := newHarness(t, "")
	shared := h.addFilter(domain.Search{Name: "shared", Query: "a"})

	h.pressOn(user, user, 55, "f_view:"+shared.ID)
	if buttons := h.screen().Buttons(); has(buttons, "f_del:"+shared.ID) || has(buttons, "f_toggle:"+shared.ID) {
		t.Errorf("a plain user sees management buttons of a shared filter: %v", buttons)
	}
	h.pressOn(user, user, 55, "f_delok:"+shared.ID)
	if _, modal := h.toast(); !modal {
		t.Error("a plain user must be refused")
	}
	if len(h.filters()) != 1 {
		t.Fatal("a plain user deleted a shared filter")
	}

	h.pressOn(admin, admin, 56, "f_claim:"+shared.ID)
	if got, _ := h.db.Search(ctx, shared.ID); got.OwnerChatID != "111" {
		t.Fatalf("owner = %q after claim", got.OwnerChatID)
	}
	h.pressOn(admin, admin, 56, "f_share:"+shared.ID)
	if got, _ := h.db.Search(ctx, shared.ID); !got.IsShared() {
		t.Fatal("the filter must be shared again")
	}
}

func TestPauseSnoozeAndResumeMoveTheAgeFloor(t *testing.T) {
	h := newHarness(t, "")
	s := h.addFilter(domain.Search{Name: "a", Query: "a", OwnerChatID: "111", CreatedAt: time.Now().Add(-48 * time.Hour).Unix()})
	enabledAt := s.EnabledAt

	h.pressOn(admin, admin, 55, "f_toggle:"+s.ID)
	if got, _ := h.db.Search(ctx, s.ID); !got.Disabled {
		t.Fatal("the filter must be paused")
	}
	h.pressOn(admin, admin, 55, "f_toggle:"+s.ID)
	got, _ := h.db.Search(ctx, s.ID)
	if got.Disabled || got.EnabledAt <= enabledAt {
		t.Fatalf("a resumed filter must record when it was resumed: %+v", got)
	}

	h.pressOn(admin, admin, 55, "f_snoozemenu:"+s.ID)
	mustContain(t, h.text(), "Отложить «a»")
	h.pressOn(admin, admin, 55, "f_snooze:"+s.ID+":3")
	got, _ = h.db.Search(ctx, s.ID)
	if until := time.Unix(got.SnoozedUntil, 0); time.Until(until) < 170*time.Minute || time.Until(until) > 181*time.Minute {
		t.Fatalf("snoozed until %s, want about 3h from now", until)
	}
	mustContain(t, h.text(), "😴 Спит до")
	h.pressOn(admin, admin, 55, "f_snooze:"+s.ID+":0")
	if got, _ = h.db.Search(ctx, s.ID); got.SnoozedUntil != 0 {
		t.Fatal("waking up must clear the snooze")
	}
}

func TestRealEstateFilterFromAnOldInstallCanOnlyBeDeleted(t *testing.T) {
	h := newHarness(t, "")
	if n, _, err := h.db.ImportLegacySearches(ctx, []domain.Search{{Name: "Квартиры", Category: 1758, OwnerChatID: "111"}}, time.Now()); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	flat := h.filters()[0]

	h.pressOn(admin, admin, 55, "f_view:"+flat.ID)
	mustContain(t, h.text(), "Раздел не поддерживается", "Недвижимость бот не обслуживает")
	buttons := h.screen().Buttons()
	if has(buttons, "f_toggle:"+flat.ID) || !has(buttons, "f_del:"+flat.ID) {
		t.Errorf("buttons = %v; such a filter can only be deleted", buttons)
	}
	h.pressOn(admin, admin, 55, "f_toggle:"+flat.ID)
	if got, _ := h.db.Search(ctx, flat.ID); !got.Disabled {
		t.Fatal("a real-estate filter must not be switched on")
	}
}

// ---------- wizard ----------

func TestWizardBuildsAFilter(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/start")
	h.press(admin, admin, "add_new")
	mustContain(t, h.text(), "📁 Категория", "шаг 1 из 5", "●○○○○")

	h.press(admin, admin, "add_cat:0")
	mustContain(t, h.text(), "🔍 Что ищем?", "шаг 2 из 5")
	h.say(admin, admin, "  iPhone   15  ")
	mustContain(t, h.text(), "Варианты запроса · iPhone 15", "<code>iphone 15</code>")
	h.press(admin, admin, "variants_done")
	mustContain(t, h.text(), "💰 Цена", "шаг 3 из 5")
	h.press(admin, admin, "add_price:5000")
	mustContain(t, h.text(), "👤 Продавец", "шаг 4 из 5")
	h.press(admin, admin, "add_owner:private")
	mustContain(t, h.text(), "📍 Где искать?", "шаг 5 из 5")
	h.press(admin, admin, "add_city:0:0")
	mustContain(t, h.text(), "🚀 Почти готово", "iPhone 15", "до 5 000 грн", "только частные", "вся Украина")

	h.press(admin, admin, "add_finish")
	list := h.filters()
	if len(list) != 1 {
		t.Fatalf("filters = %d, want 1", len(list))
	}
	f := list[0]
	if f.Query != "iPhone 15" || f.Name != "iPhone 15" || f.MaxPrice != 5000 || f.OwnerType != "private" || f.OwnerChatID != "111" || f.Disabled {
		t.Fatalf("stored filter = %+v", f)
	}
	mustContain(t, h.text(), "Фильтр «iPhone 15» запущен", "🎯 iPhone 15")
	if h.scan.triggers != 1 {
		t.Errorf("a new filter must nudge the scanner, triggers = %d", h.scan.triggers)
	}
	if h.state(admin, admin).Active {
		t.Error("the wizard must be over")
	}
}

// 0.1.x: "Launch" on an old message, after the draft was reset, created a
// filter named «Новый поиск» with no condition — every new listing on OLX.
func TestStaleLaunchButtonCreatesNothing(t *testing.T) {
	h := newHarness(t, "")
	h.pressOn(admin, admin, 55, "add_finish")
	if text, _ := h.toast(); !strings.Contains(text, "Сессия создания фильтра истекла") {
		t.Errorf("toast = %q", text)
	}
	if n := len(h.filters()); n != 0 {
		t.Fatalf("a stale button created %d filters", n)
	}

	// The same after a filter was really launched: the second tap is stale.
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.say(admin, admin, "ps5")
	h.press(admin, admin, "variants_done")
	h.press(admin, admin, "add_price:0")
	h.press(admin, admin, "add_owner:all")
	h.press(admin, admin, "add_city:0:0")
	h.press(admin, admin, "add_finish")
	h.press(admin, admin, "add_finish")
	if n := len(h.filters()); n != 1 {
		t.Fatalf("a double tap on Launch gave %d filters, want 1", n)
	}
}

func TestWizardRefusesAFilterWithoutAnyCondition(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.press(admin, admin, "add_query_skip")
	h.press(admin, admin, "add_price:0")
	h.press(admin, admin, "add_owner:all")
	h.press(admin, admin, "add_city:268:25")
	h.press(admin, admin, "add_finish")

	if text, modal := h.toast(); !modal || !strings.Contains(text, "иначе фильтр ловил бы весь OLX") {
		t.Errorf("answer = %q modal=%v", text, modal)
	}
	if n := len(h.filters()); n != 0 {
		t.Fatalf("an unconstrained filter was stored: %d", n)
	}
	if !h.state(admin, admin).Active {
		t.Error("the draft must stay so it can be fixed")
	}
}

// The wizard dead end of 0.1.x: after a numeric extra filter was typed, the
// query step no longer accepted text.
func TestQueryIsStillAcceptedAfterANumericExtraFilter(t *testing.T) {
	h := newHarness(t, "")
	const cars = 108
	idx := -1
	for i, d := range extraFilters(cars, "") {
		if isNumberFilter(d) {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Skip("the cars category has no numeric filter in this catalog")
	}

	h.say(admin, admin, "/add")
	h.press(admin, admin, fmt.Sprintf("add_cat_select:%d", cars))
	if strings.Contains(h.text(), "🏷") {
		h.press(admin, admin, fmt.Sprintf("add_brand_skip:%d", cars))
	}
	mustContain(t, h.text(), "🔍 Что ищем?")

	h.press(admin, admin, "filters_open")
	h.press(admin, admin, fmt.Sprintf("filter_open:%d:0", idx))
	h.press(admin, admin, fmt.Sprintf("filter_input:%d", idx))
	h.say(admin, admin, "2010-2015")
	mustContain(t, h.text(), "2010", "2015")
	h.press(admin, admin, "filters_page:0")
	h.press(admin, admin, "extras_done")
	mustContain(t, h.text(), "🔍 Что ищем?")

	h.say(admin, admin, "golf")
	mustContain(t, h.text(), "Варианты запроса · golf")
	st := h.state(admin, admin)
	if st.Draft.Query != "golf" || len(st.Draft.Ranges) != 1 {
		t.Fatalf("draft = %+v", st.Draft)
	}

	// "Back" from the number prompt must not leave the bot waiting for a number.
	h.press(admin, admin, "wiz_query")
	h.press(admin, admin, "filters_open")
	h.press(admin, admin, fmt.Sprintf("filter_open:%d:0", idx))
	h.press(admin, admin, fmt.Sprintf("filter_input:%d", idx))
	h.press(admin, admin, fmt.Sprintf("filter_open:%d:0", idx))
	if st.Step != stepNone {
		t.Fatalf("step = %d after pressing Back, want none", st.Step)
	}
}

func TestPriceStepAcceptsZeroAsFreeOnly(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat_select:44")
	if strings.Contains(h.text(), "🏷") {
		h.press(admin, admin, "add_brand_skip:44")
	}
	h.press(admin, admin, "add_query_skip")
	mustContain(t, h.text(), "«До N» включает и бесплатные")
	if !has(h.screen().Buttons(), "add_price_free") || !strings.Contains(strings.Join(h.screen().ButtonTexts(), "|"), "🆓 Бесплатно") {
		t.Fatalf("the price step must offer free-only in any category: %v", h.screen().ButtonTexts())
	}

	h.press(admin, admin, "add_price_range")
	h.say(admin, admin, "abc")
	mustContain(t, h.text(), "Не понял цену")
	h.say(admin, admin, "0")
	if !h.state(admin, admin).Draft.FreeOnly() {
		t.Fatal(`typing "0" must switch the filter to free-only`)
	}
	mustContain(t, h.text(), "👤 Продавец")

	// The button does the same, and a range clears the mode again.
	h.press(admin, admin, "wiz_price")
	h.press(admin, admin, "add_price:5000")
	if d := h.state(admin, admin).Draft; d.FreeOnly() || d.MaxPrice != 5000 {
		t.Fatalf("a price bound must replace free-only: %+v", d)
	}
	h.press(admin, admin, "wiz_price")
	h.press(admin, admin, "add_price_free")
	h.press(admin, admin, "add_owner:all")
	h.press(admin, admin, "add_city:0:0")
	mustContain(t, h.text(), "только бесплатные")
	h.press(admin, admin, "add_finish")
	if list := h.filters(); len(list) != 1 || !list[0].FreeOnly() || list[0].MaxPrice != 0 {
		t.Fatalf("stored = %+v", list)
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
		{"from 5000", 5000, 0, true},
		{"5000", 0, 5000, true},
		{"1000-2000 грн", 1000, 2000, true},
		{"-20000", 0, 20000, true},
		{"5000-", 5000, 0, true},
		{"0", 0, 0, true},
		{"0-0", 0, 0, true},
		{"грн", 0, 0, false},
		{"abc", 0, 0, false},
	}
	for _, c := range checks {
		gotMin, gotMax, ok := parsePriceRange(c.in)
		if ok != c.ok || gotMin != c.min || gotMax != c.max {
			t.Errorf("parsePriceRange(%q) = (%v, %v, %v), want (%v, %v, %v)", c.in, gotMin, gotMax, ok, c.min, c.max, c.ok)
		}
	}
}

// Once typed input has moved the wizard to a new message, the buttons of the
// earlier screen must not act on the draft.
func TestButtonOfAnEarlierWizardScreenIsIgnored(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	queryScreen := h.tg.LastMessageID()
	h.say(admin, admin, "rtx 4080")
	if h.tg.LastMessageID() == queryScreen {
		t.Fatal("typed input must move the wizard to a new message")
	}
	if edit, ok := h.tg.Last("editMessageReplyMarkup"); !ok || fmt.Sprint(edit.Payload["message_id"]) != fmt.Sprint(queryScreen) {
		t.Errorf("the buttons of the previous screen must be removed: %+v", edit.Payload)
	}

	h.pressOn(admin, admin, queryScreen, "add_query_skip")
	if text, _ := h.toast(); !strings.Contains(text, "устарела") {
		t.Errorf("toast = %q", text)
	}
	if got := h.state(admin, admin).Draft.Query; got != "rtx 4080" {
		t.Fatalf("a stale Skip button wiped the query: %q", got)
	}
}

func TestVariantsAndExclusions(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.say(admin, admin, "айфон 4с")
	st := h.state(admin, admin)

	h.press(admin, admin, "variant_add")
	h.say(admin, admin, "4s")
	if !has(st.Draft.QueryVariants, "4s") {
		t.Fatalf("variants = %v", st.Draft.QueryVariants)
	}
	h.press(admin, admin, "variant_add")
	h.say(admin, admin, "iPhone 4S")
	mustContain(t, h.tg.Sent()[len(h.tg.Sent())-2].Text(), "уже покрыт")
	h.press(admin, admin, "variant_del:"+itemTag("4s"))
	if has(st.Draft.QueryVariants, "4s") {
		t.Fatal("the variant was not removed")
	}

	h.press(admin, admin, "excl_open:"+exclBackVariant)
	mustContain(t, h.text(), "➖ Слова-исключения", "Пока пусто")
	h.press(admin, admin, "excl_add")
	h.say(admin, admin, "Чохол, коробка;  запчаст\nчохол")
	if got := strings.Join(st.Draft.ExcludeWords, "|"); got != "чохол|коробка|запчаст" {
		t.Fatalf("exclusions = %q", got)
	}
	h.press(admin, admin, "excl_del:"+itemTag("коробка"))
	h.press(admin, admin, "excl_done")
	mustContain(t, h.text(), "Варианты запроса")
	if len(st.Draft.ExcludeWords) != 2 {
		t.Fatalf("exclusions = %v", st.Draft.ExcludeWords)
	}
}

func TestLocationWithDistricts(t *testing.T) {
	h := newHarness(t, "")
	h.geo.districts[268] = []olx.GeoDistrict{{ID: 11, Name: "Печерський"}, {ID: 13, Name: "Подільський"}}
	h.geo.results["кременчук"] = []olx.GeoLocationSuggestion{{
		City: olx.GeoCity{ID: 221, Name: "Кременчук"}, Region: olx.GeoRegion{ID: 15, Name: "Полтавська область"},
	}}

	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.say(admin, admin, "x")
	h.press(admin, admin, "variants_done")
	h.press(admin, admin, "add_price:0")
	h.press(admin, admin, "add_owner:all")

	h.press(admin, admin, "add_city:268:25")
	mustContain(t, h.text(), "🏘 Київ")
	if !has(h.screen().Buttons(), "add_district_pick:11:268:25") {
		t.Fatalf("district buttons = %v", h.screen().Buttons())
	}
	h.press(admin, admin, "add_district_pick:11:268:25")
	st := h.state(admin, admin)
	if st.Draft.CityName != "Київ" || st.Draft.DistrictID != 11 || st.Draft.DistrictName != "Печерський" {
		t.Fatalf("draft location = %+v", st.Draft)
	}
	mustContain(t, h.text(), "📍 Київ, Печерський р-н")

	// Another city, found by typing its name.
	h.press(admin, admin, "wiz_where")
	h.press(admin, admin, "add_city_manual")
	h.say(admin, admin, "Кременчук")
	buttons := h.screen().Buttons()
	if !has(buttons, "add_city_pick:221:15") || !has(buttons, "add_region_pick:15") {
		t.Fatalf("city results = %v", buttons)
	}
	h.press(admin, admin, "add_region_pick:15")
	if st.Draft.CityID != 0 || st.Draft.RegionID != 15 || st.Draft.DistrictID != 0 {
		t.Fatalf("draft location = %+v", st.Draft)
	}
	mustContain(t, h.text(), "🌍 Полтавська область")
}

func TestTwoUsersInAGroupHaveSeparateDrafts(t *testing.T) {
	h := newHarness(t, "")
	h.say(group, admin, "/add")
	h.press(group, admin, "add_cat:0")
	h.say(group, user, "/add")
	h.press(group, user, "add_cat:0")
	h.say(group, user, "samsung")

	a, b := h.state(group, admin), h.state(group, user)
	if a == b {
		t.Fatal("two users of one group share a draft")
	}
	if a.Draft.Query != "" || b.Draft.Query != "samsung" {
		t.Fatalf("drafts mixed up: %q / %q", a.Draft.Query, b.Draft.Query)
	}
}

func TestFilterLimit(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["111"],"admin_chat_ids":["111"],"max_filters_per_chat":1}`)
	h.addFilter(domain.Search{Name: "a", Query: "a", OwnerChatID: "111"})
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.say(admin, admin, "b")
	h.press(admin, admin, "variants_done")
	h.press(admin, admin, "add_price:0")
	h.press(admin, admin, "add_owner:all")
	h.press(admin, admin, "add_city:0:0")
	h.press(admin, admin, "add_finish")
	if text, modal := h.toast(); !modal || !strings.Contains(text, "это предел") {
		t.Errorf("answer = %q modal=%v", text, modal)
	}
	if n := len(h.filters()); n != 1 {
		t.Fatalf("filters = %d, the limit is 1", n)
	}
}

// ---------- real estate ----------

func TestRealEstateIsNotOffered(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	buttons := h.screen().Buttons()
	if has(buttons, fmt.Sprintf("add_cat:%d", catalog.RealEstateCategoryID)) {
		t.Fatal("real estate is offered among the categories")
	}
	if !has(buttons, "add_cat:37") {
		t.Fatalf("other categories must be there: %v", buttons)
	}
	if strings.Contains(strings.Join(h.screen().ButtonTexts(), "|"), "Нерухомість") {
		t.Fatal("real estate is offered by name")
	}

	// A button from an old build, or a forged one, gets nowhere.
	for _, data := range []string{"add_cat:1", "add_subcat:1758", "add_cat_select:1758", "add_cat_back:1"} {
		h.press(admin, admin, data)
		if cat := h.state(admin, admin).Draft.Category; cat != 0 {
			t.Fatalf("%s put real-estate category %d into the draft", data, cat)
		}
	}
}

func TestRealEstateLinkIsRefused(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "https://www.olx.ua/uk/nedvizhimost/kvartiry/prodazha-kvartir/kiev/")
	mustContain(t, h.text(), "Недвижимость бот не обслуживает")
	if h.state(admin, admin).Active {
		t.Fatal("no draft may be started from a real-estate link")
	}
}

// ---------- OLX links ----------

func TestOLXLinkBecomesADraft(t *testing.T) {
	h := newHarness(t, "")
	h.geo.slugs["kiev"] = olx.GeoCity{ID: 268, Name: "Київ", RegionID: 25}
	h.say(admin, admin, "https://www.olx.ua/uk/transport/legkovye-avtomobili/kiev/q-golf/?currency=USD"+
		"&search%5Bfilter_float_motor_year%3Afrom%5D=2015&search%5Bfilter_float_price%3Ato%5D=9000"+
		"&search%5Bfilter_enum_price%5D%5B0%5D=exchange")

	mustContain(t, h.text(), "Ссылка OLX распознана", "🚀 Почти готово", "golf", "📍 Київ",
		"«Обмін» из ссылки не поддерживается", "Цены в ссылке заданы в USD")
	st := h.state(admin, admin)
	if !st.Active || st.Draft.CityID != 268 || st.Draft.MaxPrice != 0 {
		t.Fatalf("draft = %+v", st.Draft)
	}
	if r := st.Draft.Ranges["filter_float_motor_year"]; r.From != 2015 {
		t.Fatalf("the year range from the link was lost: %+v", st.Draft.Ranges)
	}

	h.press(admin, admin, "add_finish")
	if list := h.filters(); len(list) != 1 || list[0].Ranges["filter_float_motor_year"].From != 2015 || list[0].OwnerChatID != "111" {
		t.Fatalf("stored = %+v", list)
	}
}

func TestLinkProblemsAreExplained(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "https://www.olx.ua/d/uk/obyavlenie/prodam-iphone-ID11HsSR.html")
	mustContain(t, h.text(), "ссылка на объявление, а не на поиск")

	h.tg.Reset()
	h.say(admin, admin, "https://www.olx.ua/uk/list/")
	mustContain(t, h.text(), "иначе фильтр ловил бы весь OLX")

	h.tg.Reset()
	h.say(admin, admin, "https://www.olx.ua/uk/elektronika/odessa-unknown/")
	mustContain(t, h.text(), "Город из ссылки не распознан")

	// A link that only mentions olx.ua is ordinary text.
	h.tg.Reset()
	h.say(admin, admin, "https://evil.example/?x=olx.ua")
	mustContain(t, h.text(), "🐉 Pantheon")
	if n := len(h.filters()); n != 0 {
		t.Fatalf("filters = %d", n)
	}
}

// ---------- escaping ----------

// Listing titles come from sellers and filter names from users; neither may
// reach a page as markup.
func TestUserAndListingTextIsEscapedEverywhere(t *testing.T) {
	h := newHarness(t, "")
	const evil = `<b>&"x"</b>`
	s := h.addFilter(domain.Search{Name: evil, Query: evil, OwnerChatID: "111", ExcludeWords: []string{"<i>"}})
	queued, _, err := h.db.CommitPage(ctx, store.Commit{SearchID: s.ID, Alerts: []store.Reservation{{
		OfferID: 1, ChatID: "111", Title: `<script>alert(1)</script>`, URL: `https://olx.ua/x?a="b"`, Price: "<1>", Payload: "{}",
	}}})
	if err != nil || len(queued) != 1 {
		t.Fatal(err)
	}
	if err := h.db.MarkAlertSent(ctx, queued[0].ID, time.Now()); err != nil {
		t.Fatal(err)
	}

	screens := map[string]func(){
		"list":    func() { h.say(admin, admin, "/filters") },
		"card":    func() { h.pressOn(admin, admin, 55, "f_view:"+s.ID) },
		"delete":  func() { h.pressOn(admin, admin, 55, "f_del:"+s.ID) },
		"snooze":  func() { h.pressOn(admin, admin, 55, "f_snoozemenu:"+s.ID) },
		"history": func() { h.say(admin, admin, "/history") },
	}
	for name, show := range screens {
		h.tg.Reset()
		show()
		page := h.text()
		// The page has its own <b> and <i>; what must not appear is the raw
		// text of the name, the title, the price or the link.
		for _, raw := range []string{`&"x"`, "<script>", `a="b"`, "<1>", "<b><i></b>"} {
			if strings.Contains(page, raw) {
				t.Errorf("%s: raw %q reached the page:\n%s", name, raw, page)
			}
		}
	}
	h.tg.Reset()
	h.say(admin, admin, "/history")
	mustContain(t, h.text(), "&lt;script&gt;", "&lt;1&gt;")
}

// ---------- export / import ----------

func TestExportAndImportRoundTrip(t *testing.T) {
	h := newHarness(t, "")
	h.addFilter(domain.Search{Name: "Телефоны", Query: "iphone", MaxPrice: 9000, OwnerChatID: "111", SnoozedUntil: time.Now().Add(time.Hour).Unix()})
	h.addFilter(domain.Search{Name: "Даром", Category: 44, PriceMode: domain.PriceFree, OwnerChatID: "111"})

	h.say(admin, admin, "/export")
	doc, ok := h.tg.Last("sendDocument")
	if !ok {
		t.Fatal("no file was sent")
	}
	raw, _ := doc.Payload["document"].(string)
	var file filterExport
	if err := json.Unmarshal([]byte(raw), &file); err != nil || len(file.Filters) != 2 || file.Version != exportVersion {
		t.Fatalf("export = %s (%v)", raw, err)
	}
	for _, f := range file.Filters {
		if f.ID != "" || f.OwnerChatID != "" || f.SnoozedUntil != 0 {
			t.Errorf("the file must not carry ids, owners or runtime state: %+v", f)
		}
	}

	// Imported into another chat: both are new there.
	id := h.tg.AddFile("filters.json", []byte(raw))
	h.b.onMessage(ctx, &telegram.TGMessage{From: telegram.TGUser{ID: user}, Chat: telegram.TGChat{ID: user, Type: "private"},
		Document: &telegram.TGDocument{FileID: id, FileName: "filters.json", FileSize: int64(len(raw))}})
	mustContain(t, h.text(), "📥 Импорт фильтров", "Телефоны", "Даром")
	h.press(user, user, "import_ok")

	mine := 0
	for _, f := range h.filters() {
		if f.OwnerChatID == "222" {
			mine++
			if f.ID == "" || f.SnoozedUntil != 0 {
				t.Errorf("imported filter = %+v", f)
			}
		}
	}
	if mine != 2 {
		t.Fatalf("imported %d filters, want 2", mine)
	}

	// The same file again: nothing new.
	h.tg.Reset()
	h.b.onMessage(ctx, &telegram.TGMessage{From: telegram.TGUser{ID: user}, Chat: telegram.TGChat{ID: user, Type: "private"},
		Document: &telegram.TGDocument{FileID: id, FileName: "filters.json", FileSize: int64(len(raw))}})
	mustContain(t, h.text(), "уже есть в этом чате")
}

// An import file is user input, and its keys end up in requests to OLX.
func TestImportRejectsBadFilters(t *testing.T) {
	h := newHarness(t, "")
	file := `{"version":1,"filters":[
		{"name":"ok","query":"iphone"},
		{"name":"injected","query":"x","filters":{"limit":["1"],"sort_by":["price:asc"]}},
		{"name":"brand","query":"x","brand_param":"category_id","brand_values":["1"]},
		{"name":"flat","category_id":1758},
		{"name":"empty"},
		{"name":"stolen","query":"y","id":"deadbeef","owner_chat_id":"999","disabled":false}
	]}`
	id := h.tg.AddFile("f.json", []byte(file))
	h.b.onMessage(ctx, &telegram.TGMessage{From: telegram.TGUser{ID: admin}, Chat: telegram.TGChat{ID: admin, Type: "private"},
		Document: &telegram.TGDocument{FileID: id, FileName: "f.json", FileSize: int64(len(file))}})
	mustContain(t, h.text(), "Пропущены фильтры недвижимости и фильтры с недопустимыми параметрами")
	h.press(admin, admin, "import_ok")

	list := h.filters()
	if len(list) != 2 {
		t.Fatalf("imported %d filters, want the 2 valid ones: %+v", len(list), list)
	}
	for _, f := range list {
		if f.OwnerChatID != "111" || f.ID == "deadbeef" {
			t.Errorf("an imported filter must belong to the importing chat with a fresh id: %+v", f)
		}
		if len(f.Filters) != 0 || f.BrandParam != "" {
			t.Errorf("a foreign parameter was imported: %+v", f)
		}
	}
}

func TestImportRefusesOversizedAndBrokenFiles(t *testing.T) {
	h := newHarness(t, "")
	h.b.onMessage(ctx, &telegram.TGMessage{From: telegram.TGUser{ID: admin}, Chat: telegram.TGChat{ID: admin, Type: "private"},
		Document: &telegram.TGDocument{FileID: "x", FileName: "big.json", FileSize: maxImportBytes + 1}})
	mustContain(t, h.text(), "Файл слишком большой")

	id := h.tg.AddFile("broken.json", []byte("{not json"))
	h.tg.Reset()
	h.b.onMessage(ctx, &telegram.TGMessage{From: telegram.TGUser{ID: admin}, Chat: telegram.TGChat{ID: admin, Type: "private"},
		Document: &telegram.TGDocument{FileID: id, FileName: "broken.json", FileSize: 9}})
	mustContain(t, h.text(), "Не получилось прочитать файл")

	// A stale confirm button imports nothing.
	h.pressOn(admin, admin, 55, "import_ok")
	if n := len(h.filters()); n != 0 {
		t.Fatalf("filters = %d", n)
	}
}

// ---------- queues ----------

// One chat waiting on OLX must not hold up another.
func TestSlowChatDoesNotBlockAnother(t *testing.T) {
	h := newHarness(t, "")
	h.geo.block = make(chan struct{})

	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.say(admin, admin, "x")
	h.press(admin, admin, "variants_done")
	h.press(admin, admin, "add_price:0")
	h.press(admin, admin, "add_owner:all")
	h.press(admin, admin, "add_city_manual")
	h.tg.Reset()

	slow := telegram.TGUpdate{UpdateID: 1, Message: &telegram.TGMessage{From: telegram.TGUser{ID: admin}, Chat: telegram.TGChat{ID: admin, Type: "private"}, Text: "Кременчук"}}
	queued := telegram.TGUpdate{UpdateID: 2, Message: &telegram.TGMessage{From: telegram.TGUser{ID: admin}, Chat: telegram.TGChat{ID: admin, Type: "private"}, Text: "/version"}}
	fast := telegram.TGUpdate{UpdateID: 3, Message: &telegram.TGMessage{From: telegram.TGUser{ID: user}, Chat: telegram.TGChat{ID: user, Type: "private"}, Text: "/version"}}
	h.b.enqueue(ctx, slow)
	h.b.enqueue(ctx, queued)
	h.b.enqueue(ctx, fast)

	deadline := time.After(5 * time.Second)
	for {
		sent := h.tg.Calls("sendMessage")
		if len(sent) == 1 && sent[0].ChatID() == "222" {
			break
		}
		if len(sent) > 1 {
			t.Fatalf("the blocked chat was answered out of turn: %d messages", len(sent))
		}
		select {
		case <-deadline:
			t.Fatal("the second chat was not served while the first one waited on OLX")
		case <-time.After(10 * time.Millisecond):
		}
	}

	close(h.geo.block)
	deadline = time.After(5 * time.Second)
	for len(h.tg.Calls("sendMessage")) < 3 {
		select {
		case <-deadline:
			t.Fatalf("the first chat's own updates were not finished in order: %d", len(h.tg.Calls("sendMessage")))
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestPanicInAHandlerDoesNotKillTheBot(t *testing.T) {
	h := newHarness(t, "")
	h.b.geo = nil
	h.b.db = nil // every handler that touches the store will now panic
	h.b.dispatch(ctx, telegram.TGUpdate{UpdateID: 7, Message: &telegram.TGMessage{
		From: telegram.TGUser{ID: admin}, Chat: telegram.TGChat{ID: admin, Type: "private"}, Text: "/filters"}})
	if !strings.Contains(h.log.String(), "panic while handling update 7") {
		t.Fatalf("the panic must be logged, got:\n%s", h.log.String())
	}
}

// ---------- callback data ----------

// Telegram rejects a keyboard with callback data over 64 bytes, which would
// take the whole screen down with it.
func TestCallbackDataFitsTheLimit(t *testing.T) {
	longest := ""
	check := func(data string) {
		if len(data) > len(longest) {
			longest = data
		}
	}
	for catID, info := range catalog.CategoryBrands {
		pages := (len(info.Brands) + brandsPerPage - 1) / brandsPerPage
		for _, br := range info.Brands {
			check(fmt.Sprintf("add_brand_toggle:%d:%d:%s", catID, pages, br.ID))
		}
	}
	check("add_district_pick:99999:99999:99")
	check("f_snooze:deadbeef:24")
	check("variant_del:" + itemTag("x"))
	check("filter_toggle:999:999:99")
	if len(longest) > 64 {
		t.Fatalf("callback data of %d bytes exceeds the limit: %s", len(longest), longest)
	}
}

func TestHelpAndHistoryAndVersion(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/help")
	mustContain(t, h.text(), "❔ Справка", "/filters", "/addchat &lt;id&gt;")
	h.say(user, user, "/help")
	if strings.Contains(h.text(), "/addchat") {
		t.Error("chat management is shown to the admin only")
	}
	h.say(user, user, "/history")
	mustContain(t, h.text(), "Алертов в этот чат пока не было")
	h.say(user, user, "/version")
	mustContain(t, h.text(), "Pantheon")
	h.say(user, user, "/nonsense")
	mustContain(t, h.text(), "Не знаю такой команды")
	h.say(user, user, "/cancel")
	mustContain(t, h.text(), "Отменено", "🐉 Pantheon")
}

func TestAnnounceOncePerVersionToPrivateChatsOnly(t *testing.T) {
	h := newHarness(t, "")
	h.b.announce(ctx)
	deadline := time.After(5 * time.Second)
	for len(h.tg.Sent()) < 2 {
		select {
		case <-deadline:
			t.Fatalf("menus sent = %d, want one per private chat", len(h.tg.Sent()))
		case <-time.After(10 * time.Millisecond):
		}
	}
	time.Sleep(50 * time.Millisecond)
	for _, c := range h.tg.Sent() {
		if c.ChatID() == fmt.Sprint(group) {
			t.Fatal("the menu must not be posted to a group on start")
		}
	}
	// A restart on the same version announces nothing: no spam under Restart=always.
	h.tg.Reset()
	h.b.announce(ctx)
	time.Sleep(50 * time.Millisecond)
	if n := len(h.tg.Sent()); n != 0 {
		t.Fatalf("announced again on the same version: %d", n)
	}
}

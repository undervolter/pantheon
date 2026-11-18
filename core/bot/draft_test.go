package bot

import (
	"strings"
	"testing"

	"pantheon/core/telegram"
	"pantheon/core/telegram/tgtest"
)

// sendFile delivers a document from userID in chatID and returns the id of
// the message the bot answered with.
func (h *harness) sendFile(chatID, userID int64, name, body string) int {
	h.t.Helper()
	id := h.tg.AddFile(name, []byte(body))
	h.b.onMessage(ctx, &telegram.TGMessage{
		From:     telegram.TGUser{ID: userID},
		Chat:     telegram.TGChat{ID: chatID, Type: chatType(chatID, userID)},
		Document: &telegram.TGDocument{FileID: id, FileName: name, FileSize: int64(len(body))},
	})
	return h.tg.LastMessageID()
}

// A page that is not part of the wizard — the menu answering stray text,
// /help, /history — used to become the draft's "home": the real wizard screen
// lost its buttons and the draft could not be reached any more.
func TestOtherPagesDoNotTakeTheDraftAway(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	h.press(admin, admin, "add_query_skip") // the price screen: buttons only
	wizardMsg := h.tg.LastMessageID()
	mustContain(t, h.text(), "💰 Цена")

	for _, text := range []string{"20000", "/help", "/history", "/lang", "/version"} {
		h.say(admin, admin, text)
	}

	st := h.state(admin, admin)
	if !st.Active || st.MsgID != wizardMsg {
		t.Fatalf("the draft moved from message %d to %d (active=%v)", wizardMsg, st.MsgID, st.Active)
	}
	if _, ok := h.tg.Last("editMessageReplyMarkup"); ok {
		t.Fatal("the wizard screen lost its buttons")
	}

	h.pressOn(admin, admin, wizardMsg, "add_price:5000")
	mustContain(t, h.text(), "👤 Продавец")
	if st.Draft.MaxPrice != 5000 {
		t.Fatalf("the wizard did not carry on: %+v", st.Draft)
	}
}

// In a group everyone sees everyone's messages. A draft still belongs to the
// person building it: another member's tap used to turn its screen into a
// menu, and two drafts could end up sharing one message.
func TestAnotherMemberCannotDriveOrReplaceADraft(t *testing.T) {
	h := newHarness(t, "")
	h.say(group, admin, "/add")
	draftMsg := h.tg.LastMessageID()
	h.tg.Reset()

	for _, data := range []string{"add_cat:37", "add_finish", "menu_main", "add_new", "menu_list"} {
		h.pressOn(group, user, draftMsg, data)
		if text, modal := h.toast(); !modal || !strings.Contains(text, "чужой черновик") {
			t.Errorf("%s: answer = %q modal=%v, want a refusal", data, text, modal)
		}
	}
	if n := len(h.tg.Sent()); n != 0 {
		t.Fatalf("the draft's message was redrawn %d times by another member", n)
	}
	owner := h.state(group, admin)
	if !owner.Active || owner.MsgID != draftMsg || owner.Draft.Category != 0 {
		t.Fatalf("the owner's draft was touched: %+v", owner)
	}
	if h.state(group, user).Active {
		t.Fatal("the other member got a draft on the same message")
	}

	// Its owner carries on as before.
	h.pressOn(group, admin, draftMsg, "add_cat:0")
	mustContain(t, h.text(), "🔍 Что ищем?")

	// Once the draft is gone the message is everybody's again.
	h.pressOn(group, admin, draftMsg, "menu_main")
	h.pressOn(group, user, draftMsg, "menu_list")
	mustContain(t, h.text(), "🎯 Фильтры")
}

// The confirmation buttons carried no reference to their file: "Import" under
// an earlier file imported the filters of the latest one.
func TestImportConfirmationAnswersForItsOwnFileOnly(t *testing.T) {
	h := newHarness(t, "")
	first := h.sendFile(admin, admin, "a.json", `{"version":1,"filters":[{"name":"из первого","query":"a"}]}`)
	second := h.sendFile(admin, admin, "b.json", `{"version":1,"filters":[{"name":"из второго","query":"b"}]}`)
	if first == second {
		t.Fatal("each file gets its own confirmation message")
	}

	h.pressOn(admin, admin, first, "import_ok")
	if text, _ := h.toast(); !strings.Contains(text, "устарела") {
		t.Errorf("toast = %q", text)
	}
	if n := len(h.filters()); n != 0 {
		t.Fatalf("the buttons under the first file imported %d filters of the second", n)
	}
	h.pressOn(admin, admin, first, "import_cancel")
	if len(h.state(admin, admin).Import) != 1 {
		t.Fatal("Cancel under the first file cancelled the second")
	}

	h.pressOn(admin, admin, second, "import_ok")
	list := h.filters()
	if len(list) != 1 || list[0].Name != "из второго" || list[0].OwnerChatID != "111" {
		t.Fatalf("imported = %+v", list)
	}
	// A second tap finds nothing waiting.
	h.pressOn(admin, admin, second, "import_ok")
	if n := len(h.filters()); n != 1 {
		t.Fatalf("a double tap imported twice: %d filters", n)
	}
}

func TestImportOfAnotherMemberIsNotConfirmedForThem(t *testing.T) {
	h := newHarness(t, "")
	msg := h.sendFile(group, admin, "a.json", `{"version":1,"filters":[{"name":"x","query":"a"}]}`)
	h.tg.Reset()

	for _, data := range []string{"import_ok", "import_cancel"} {
		h.pressOn(group, user, msg, data)
		if text, modal := h.toast(); !modal || !strings.Contains(text, "чужой черновик") {
			t.Errorf("%s: answer = %q modal=%v", data, text, modal)
		}
	}
	if n := len(h.filters()); n != 0 || len(h.state(group, admin).Import) != 1 {
		t.Fatalf("another member answered for the file: %d filters", n)
	}
	h.pressOn(group, admin, msg, "import_ok")
	if n := len(h.filters()); n != 1 {
		t.Fatalf("the sender's own confirmation gave %d filters", n)
	}
}

// A search link pasted while the bot waits for a query used to become the
// query itself, cut off at a hundred characters.
func TestLinkPastedAtATextStepIsALink(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/add")
	h.press(admin, admin, "add_cat:0")
	mustContain(t, h.text(), "🔍 Что ищем?")

	h.say(admin, admin, "https://www.olx.ua/uk/list/q-golf/")
	mustContain(t, h.text(), "Ссылка OLX распознана", "🚀 Почти готово")
	if q := h.state(admin, admin).Draft.Query; q != "golf" {
		t.Fatalf("query = %q", q)
	}
}

// If the first wizard screen never reached the chat, the draft has no message
// of its own, and no button anywhere may act on it.
func TestDraftWithoutAScreenAcceptsNoButtons(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/start")
	menu := h.tg.LastMessageID()

	h.tg.Queue("sendRichMessage", tgtest.Fail(500, "Internal Server Error"))
	h.say(admin, admin, "/add")
	st := h.state(admin, admin)
	if !st.Active || st.MsgID != 0 {
		t.Fatalf("precondition: active=%v msg=%d, want a draft with no screen", st.Active, st.MsgID)
	}

	for _, data := range []string{"add_cat:37", "add_finish"} {
		h.pressOn(admin, admin, menu, data)
		if text, _ := h.toast(); !strings.Contains(text, "устарела") {
			t.Errorf("%s: toast = %q", data, text)
		}
	}
	if st.Draft.Category != 0 || len(h.filters()) != 0 {
		t.Fatalf("a button of an unrelated message acted on the draft: %+v", st.Draft)
	}
}

// A stranger could keep replacing the pairing code — five wrong guesses and
// the one printed for the owner was dead — and make the bot answer each time.
func TestOneChatCannotBurnThePairingCode(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`"}`)
	h.b.preparePairing()
	code := h.b.pair.code

	for i := 0; i < 3*maxPairingFailures; i++ {
		h.say(900, 900, "/start guess")
	}
	if h.b.pair.code != code {
		t.Fatal("one chat replaced the pairing code by guessing")
	}
	if n := len(h.tg.Sent()); n != 1 {
		t.Errorf("replies to the guesses = %d, want one", n)
	}

	h.say(700, 700, "/start "+code)
	if !h.cfg.IsAdmin("700") {
		t.Fatal("the code printed for the owner must still work")
	}
}

func TestChatIDsAreStoredTheWayTelegramSendsThem(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/addchat +555")
	if !h.cfg.IsTrusted("555") {
		t.Fatalf("chats = %v: +555 must be stored as 555", h.cfg.GetTelegramChatIDs())
	}
	h.say(admin, admin, "/addchat 0555")
	n := 0
	for _, id := range h.cfg.GetTelegramChatIDs() {
		if strings.TrimLeft(id, "+0") == "555" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("chats = %v: the same chat was added twice", h.cfg.GetTelegramChatIDs())
	}
	h.say(admin, admin, "/delchat 0555")
	if h.cfg.IsTrusted("555") {
		t.Fatal("/delchat must find the chat however its id is typed")
	}
}

func TestCommandForAnotherBotIsNotOurs(t *testing.T) {
	// Before the bot knows its own name it cannot tell, so it stays out.
	if _, _, ok := parseCommand("/start@some_other_bot", ""); ok {
		t.Error("a command addressed to a bot by name was accepted without knowing our own name")
	}
	if cmd, _, ok := parseCommand("/start", ""); !ok || cmd != "start" {
		t.Error("a plain command needs no name")
	}
}

func TestForgedPriceButtonCannotStoreInfinity(t *testing.T) {
	h := newHarness(t, "")
	for _, value := range []string{"inf", "NaN", "1e300", "-5"} {
		h.say(admin, admin, "/add")
		h.press(admin, admin, "add_cat:0")
		h.press(admin, admin, "add_query_skip")
		h.press(admin, admin, "add_price:"+value)
		if got := h.state(admin, admin).Draft.MaxPrice; got != 0 {
			t.Errorf("add_price:%s stored %v", value, got)
		}
	}
	for _, typed := range []string{"99999999999999999999", "5000-99999999999999999999"} {
		if _, hi, ok := parsePriceRange(typed); ok && hi > 1e12 {
			t.Errorf("parsePriceRange(%q) accepted %v", typed, hi)
		}
	}
}

func TestHelpShowsCheckToAdminsOnly(t *testing.T) {
	h := newHarness(t, "")
	h.say(admin, admin, "/help")
	mustContain(t, h.text(), "/check")
	h.say(user, user, "/help")
	if strings.Contains(h.text(), "/check") {
		t.Error("the help offers /check to a user who is refused when using it")
	}
}

// On an install upgraded from 0.1.x the first chat in the list is whoever
// wrote to the bot first. Making it the admin must not happen silently.
func TestAdminTakenFromTheChatListIsAnnounced(t *testing.T) {
	h := newHarness(t, `{"telegram_bot_token":"`+tgtest.Token+`","telegram_chat_ids":["-100500","333","444"]}`)
	h.b.preparePairing()
	if !strings.Contains(h.log.String(), "Admin of this bot: chat 333") {
		t.Fatalf("the promotion must be printed to the console:\n%s", h.log.String())
	}
	// Later starts have nothing to announce.
	h.log.Reset()
	h.b.preparePairing()
	if strings.Contains(h.log.String(), "Admin of this bot") {
		t.Error("announced again although the admin was already set")
	}
}

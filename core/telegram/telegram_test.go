package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/telegram/tgtest"
)

var ctx = context.Background()

func newClient(t *testing.T) (*Client, *tgtest.Server) {
	t.Helper()
	srv := tgtest.New(t)
	return New(tgtest.Token).WithBaseURL(srv.URL), srv
}

func sampleOffer(t *testing.T) domain.Offer {
	t.Helper()
	var o domain.Offer
	raw := `{
		"id": 941268305,
		"url": "https://www.olx.ua/d/uk/obyavlenie/iphone-ID11HsSR.html",
		"title": "iPhone 15 Pro <Max> & co",
		"description": "Стан ідеальний.<br>Комплект повний.",
		"created_time": "2026-11-18T00:31:16+02:00",
		"last_refresh_time": "2026-11-18T00:35:17+02:00",
		"params": [
			{"key":"price","name":"Ціна","type":"price","value":{"value":28000,"type":"price","currency":"UAH"}},
			{"key":"state","name":"Стан","type":"select","value":{"key":"used","label":"Вживане"}}
		],
		"photos": [
			{"width":3024,"height":4032,"link":"https://cdn.example/a;s={width}x{height}"},
			{"width":750,"height":1000,"link":"https://cdn.example/b;s={width}x{height}"}
		],
		"location": {"city":{"id":268,"name":"Київ"},"region":{"id":25,"name":"Київська область"}},
		"contact": {"phone": true}
	}`
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatal(err)
	}
	return o
}

func withPrice(t *testing.T, o domain.Offer, value string) domain.Offer {
	t.Helper()
	var p domain.Param
	if err := json.Unmarshal([]byte(`{"key":"price","name":"Ціна","type":"price","value":`+value+`}`), &p); err != nil {
		t.Fatal(err)
	}
	o.Params = []domain.Param{p}
	return o
}

// ---------- transport ----------

func TestNoTokenInErrorsOrLogs(t *testing.T) {
	var buf bytes.Buffer
	t.Cleanup(logx.SetOutput(&buf))

	// A port nobody listens on: the transport error carries the full URL.
	c := New(tgtest.Token).WithBaseURL("http://127.0.0.1:1")
	_, err := c.SendHTML(ctx, "1", "hi", nil, false)
	if err == nil {
		t.Fatal("expected a transport error")
	}
	if strings.Contains(err.Error(), tgtest.Token) {
		t.Fatalf("token leaked into the error: %v", err)
	}
	logx.Errorf("send failed: %v", err)
	if strings.Contains(buf.String(), tgtest.Token) {
		t.Fatalf("token leaked into the log: %s", buf.String())
	}
	if !IsRetryable(err) {
		t.Error("a network failure must be retryable")
	}
}

func TestAPIErrorClassification(t *testing.T) {
	c, srv := newClient(t)

	srv.Queue("sendMessage", tgtest.FloodWait(7))
	_, err := c.SendHTML(ctx, "1", "x", nil, false)
	if wait, ok := IsRateLimited(err); !ok || wait != 7*time.Second {
		t.Fatalf("IsRateLimited = %s, %v; want 7s", wait, ok)
	}
	if !IsRetryable(err) {
		t.Error("429 must be retryable")
	}

	srv.Queue("sendMessage", tgtest.Fail(403, "Forbidden: bot was blocked by the user"))
	_, err = c.SendHTML(ctx, "1", "x", nil, false)
	if !IsDeadChat(err) || IsRetryable(err) {
		t.Errorf("blocked bot: dead=%v retryable=%v", IsDeadChat(err), IsRetryable(err))
	}

	srv.Queue("sendMessage", tgtest.Fail(502, "Bad Gateway"))
	_, err = c.SendHTML(ctx, "1", "x", nil, false)
	if !IsRetryable(err) {
		t.Error("5xx must be retryable")
	}

	srv.Queue("sendMessage", tgtest.Fail(400, "Bad Request: message is too long"))
	_, err = c.SendHTML(ctx, "1", "x", nil, false)
	if IsRetryable(err) || IsDeadChat(err) {
		t.Error("a 400 is final and is not a dead chat")
	}
	if IsRetryable(context.Canceled) {
		t.Error("a cancelled request must not be retried")
	}
}

func TestUICallWaitsOutAShortFloodLimit(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("sendMessage", tgtest.FloodWait(1))
	start := time.Now()
	if _, err := c.SendHTML(ctx, "1", "menu", nil, false); err != nil {
		t.Fatalf("a short flood limit must be retried once: %v", err)
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Error("the retry must wait for retry_after")
	}
	if n := len(srv.Calls("sendMessage")); n != 2 {
		t.Errorf("sendMessage calls = %d, want 2", n)
	}

	srv.Reset()
	srv.Queue("sendMessage", tgtest.FloodWait(30))
	if _, err := c.SendHTML(ctx, "1", "menu", nil, false); err == nil {
		t.Fatal("a long flood limit is returned to the caller")
	}
	if n := len(srv.Calls("sendMessage")); n != 1 {
		t.Errorf("a long flood limit must not be retried, calls = %d", n)
	}
}

func TestPlainFallbackOnlyForMarkupErrors(t *testing.T) {
	c, srv := newClient(t)

	srv.Queue("sendMessage", tgtest.Fail(400, "Bad Request: can't parse entities: Unsupported start tag \"x\""))
	if _, err := c.SendHTML(ctx, "1", "<x>hello</x> <b>world</b>", nil, false); err != nil {
		t.Fatalf("markup error must fall back to plain text: %v", err)
	}
	calls := srv.Calls("sendMessage")
	if len(calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(calls))
	}
	if calls[1].Payload["parse_mode"] != nil || strings.Contains(calls[1].Text(), "<b>") {
		t.Errorf("the fallback must be plain text, got %q", calls[1].Raw)
	}

	// A flood limit, a server error or a blocked bot must NOT trigger a second
	// "plain" request: that doubled the load and duplicated messages.
	for _, reply := range []tgtest.Reply{
		tgtest.FloodWait(30),
		tgtest.Fail(502, "Bad Gateway"),
		tgtest.Fail(403, "Forbidden: bot was blocked by the user"),
		tgtest.Fail(400, "Bad Request: message is too long"),
	} {
		srv.Reset()
		srv.Queue("sendMessage", reply)
		if _, err := c.SendHTML(ctx, "1", "<b>x</b>", nil, false); err == nil {
			t.Errorf("status %d must be returned", reply.Status)
		}
		if n := len(srv.Calls("sendMessage")); n != 1 {
			t.Errorf("status %d: calls = %d, want exactly 1", reply.Status, n)
		}
	}
}

func TestEditNotModifiedIsSuccess(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("editMessageText", tgtest.Fail(400, "Bad Request: message is not modified"))
	if err := c.EditHTML(ctx, "1", 5, "same", nil); err != nil {
		t.Fatalf("not modified must be success: %v", err)
	}
	srv.Queue("editMessageText", tgtest.Fail(400, "Bad Request: message is not modified"))
	if err := c.EditPage(ctx, "1", 5, "<h3>same</h3>", nil); err != nil {
		t.Fatalf("not modified must be success for pages: %v", err)
	}
}

func TestLimiterReservesSlots(t *testing.T) {
	l := NewLimiter()
	now := time.Unix(1_800_000_000, 0)
	l.now = func() time.Time { return now }

	if w := l.Reserve("1"); w != 0 {
		t.Errorf("first slot wait = %s, want 0", w)
	}
	if w := l.Reserve("1"); w != ChatGap {
		t.Errorf("second slot wait = %s, want %s", w, ChatGap)
	}
	if w := l.Reserve("1"); w != 2*ChatGap {
		t.Errorf("third slot wait = %s, want %s", w, 2*ChatGap)
	}
	if w := l.Reserve("2"); w != 0 {
		t.Errorf("another chat must not wait, got %s", w)
	}
	_ = l.Reserve("-100")
	if w := l.Reserve("-100"); w != GroupGap {
		t.Errorf("group gap = %s, want %s", w, GroupGap)
	}

	l.Penalize("3", 30*time.Second)
	if w := l.Reserve("3"); w != 30*time.Second {
		t.Errorf("penalised chat wait = %s, want 30s", w)
	}
}

func TestDocumentRoundTrip(t *testing.T) {
	c, srv := newClient(t)
	if err := c.SendDocument(ctx, "1", "filters.json", []byte(`{"a":1}`), "<b>3</b> filters"); err != nil {
		t.Fatal(err)
	}
	call, _ := srv.Last("sendDocument")
	if call.Payload["document_name"] != "filters.json" || call.Payload["document"] != `{"a":1}` || call.ChatID() != "1" {
		t.Errorf("unexpected upload: %+v", call.Payload)
	}

	id := srv.AddFile("export.json", []byte(`[1,2,3]`))
	data, err := c.DownloadFile(ctx, id, 1024)
	if err != nil || string(data) != `[1,2,3]` {
		t.Fatalf("DownloadFile = %q, %v", data, err)
	}
	big := srv.AddFile("big.json", bytes.Repeat([]byte("x"), 2048))
	if _, err := c.DownloadFile(ctx, big, 1024); err == nil {
		t.Fatal("a file over the limit must be refused")
	}
}

// ---------- pages ----------

func TestRichToPlain(t *testing.T) {
	page := `<h3>🎯 Фильтры</h3><blockquote>✅ Создан</blockquote>` +
		`<table compact><tr><th align="center">Фильтры</th><th align="center">Сегодня</th></tr><tr><td align="center"><b>2</b></td><td align="center"><b>7</b></td></tr></table>` +
		`<p>Строка один<br>Строка два</p>` +
		`<table compact><tr><td>Запрос</td><td><b>iphone</b></td></tr><tr><td>Цена</td><td><b>до 5000</b></td></tr></table>` +
		`<table compact striped><tr><th></th><th>Фильтр</th><th>Цена</th></tr><tr><td>🟢</td><td>A</td><td>—</td></tr></table>` +
		`<ul><li>один</li><li>два</li></ul><footer>v0.2.0<br>подвал</footer>`
	got := RichToPlain(page)

	for _, want := range []string{
		"<b>🎯 Фильтры</b>", "<blockquote>✅ Создан</blockquote>",
		"Фильтры: <b>2</b>", "Сегодня: <b>7</b>",
		"Строка один\nСтрока два",
		"Запрос: <b>iphone</b>", "Цена: <b>до 5000</b>",
		"🟢 · A · —", "• один", "• два", "<i>v0.2.0\nподвал</i>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("plain page lacks %q:\n%s", want, got)
		}
	}
	for _, tag := range []string{"<table", "<tr", "<td", "<th", "<h3", "<p>", "<ul", "<li", "<footer"} {
		if strings.Contains(got, tag) {
			t.Errorf("plain page still has %s:\n%s", tag, got)
		}
	}
	if RichToPlain("just text") != "just text" {
		t.Error("plain text must pass through unchanged")
	}
}

func TestPagesFallBackWhenServerHasNoRichMessages(t *testing.T) {
	var buf bytes.Buffer
	t.Cleanup(logx.SetOutput(&buf))
	c, srv := newClient(t)
	srv.On("sendRichMessage", func(tgtest.Call) *tgtest.Reply { r := tgtest.Fail(404, "Not Found"); return &r })

	id, err := c.SendPage(ctx, "1", "<h3>Меню</h3><p>текст</p>", Markup(Row(Btn("a", "b"))))
	if err != nil || id == 0 {
		t.Fatalf("SendPage = %d, %v", id, err)
	}
	last, _ := srv.Last("sendMessage")
	if !strings.Contains(last.Text(), "<b>Меню</b>") || len(last.Buttons()) != 1 {
		t.Errorf("fallback message = %q buttons=%v", last.Text(), last.Buttons())
	}

	// The switch is remembered: no more rich attempts, edits go plain too.
	srv.Reset()
	if _, err := c.SendPage(ctx, "1", "<h3>Ещё</h3>", nil); err != nil {
		t.Fatal(err)
	}
	if err := c.EditPage(ctx, "1", id, "<h3>Правка</h3>", nil); err != nil {
		t.Fatal(err)
	}
	if n := len(srv.Calls("sendRichMessage")); n != 0 {
		t.Errorf("rich was tried again %d times after a 404", n)
	}
	edit, _ := srv.Last("editMessageText")
	if edit.Payload["rich_message"] != nil || !strings.Contains(edit.Text(), "<b>Правка</b>") {
		t.Errorf("edit after the switch must be plain: %s", edit.Raw)
	}
}

func TestPageSendDoesNotFallBackOnUnknownOutcome(t *testing.T) {
	c, srv := newClient(t)
	for _, reply := range []tgtest.Reply{tgtest.FloodWait(30), tgtest.Fail(502, "Bad Gateway"), tgtest.Fail(403, "Forbidden: bot was blocked by the user")} {
		srv.Reset()
		srv.Queue("sendRichMessage", reply)
		if _, err := c.SendPage(ctx, "1", "<h3>x</h3>", nil); err == nil {
			t.Errorf("status %d must be returned", reply.Status)
		}
		if n := len(srv.Calls("sendMessage")); n != 0 {
			t.Errorf("status %d must not produce a plain duplicate, sendMessage calls = %d", reply.Status, n)
		}
	}
}

func TestEditPageRejectedIsReported(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("editMessageText", tgtest.Fail(400, "Bad Request: message can't be edited"))
	err := c.EditPage(ctx, "1", 5, "<h3>x</h3>", nil)
	if !errors.Is(err, ErrPageEditRejected) {
		t.Fatalf("err = %v, want ErrPageEditRejected", err)
	}
}

// ---------- alert card ----------

func TestRichCardEscapesListingText(t *testing.T) {
	o := sampleOffer(t)
	card := BuildRichOffer(o, CardOptions{Lang: i18n.RU, FilterName: "<i>мой</i> & фильтр", ShowParams: true}, o.PhotoURLs(2))

	if !strings.Contains(card, "iPhone 15 Pro &lt;Max&gt; &amp; co") {
		t.Errorf("title must be escaped:\n%s", card)
	}
	if strings.Contains(card, "<Max>") || strings.Contains(card, "<i>мой</i>") {
		t.Errorf("raw listing or filter markup leaked into the card:\n%s", card)
	}
	if !strings.Contains(card, "<tg-collage>") || !strings.Contains(card, "s=960x1280") {
		t.Errorf("photos must be a collage of capped images:\n%s", card)
	}
	if !strings.Contains(card, "Продаётся за <b>28 000 грн. (~$618)</b>") {
		t.Errorf("price line missing:\n%s", card)
	}
	if !strings.Contains(card, "Стан ідеальний.<br>Комплект повний.") {
		t.Errorf("description must keep its line break:\n%s", card)
	}
	if !strings.Contains(card, "<footer>Фильтр: &lt;i&gt;мой&lt;/i&gt; &amp; фильтр • OLX #941268305</footer>") {
		t.Errorf("footer:\n%s", card)
	}
}

func TestCardWordingForZeroPriceListings(t *testing.T) {
	base := sampleOffer(t)
	cases := []struct {
		name, value string
		want        map[i18n.Lang]string
	}{
		{"free", `{"value":0,"type":"free","budget":false,"label":"Безкоштовно"}`, map[i18n.Lang]string{
			i18n.RU: "🆓 <b>Отдают бесплатно</b>", i18n.UK: "🆓 <b>Віддають безкоштовно</b>", i18n.EN: "🆓 <b>Given away for free</b>"}},
		{"exchange", `{"value":0,"type":"exchange","label":"Обмін"}`, map[i18n.Lang]string{
			i18n.RU: "🔄 <b>Обмен</b>", i18n.UK: "🔄 <b>Обмін</b>", i18n.EN: "🔄 <b>Exchange</b>"}},
		{"arranged", `{"value":0,"arranged":true,"label":"Договірна"}`, map[i18n.Lang]string{
			i18n.RU: "Цена: <b>договорная</b>", i18n.UK: "Ціна: <b>договірна</b>", i18n.EN: "Price: <b>negotiable</b>"}},
	}
	for _, c := range cases {
		o := withPrice(t, base, c.value)
		for lang, want := range c.want {
			for kind, card := range map[string]string{
				"rich": BuildRichOffer(o, CardOptions{Lang: lang}, nil),
				"text": BuildTextOffer(o, CardOptions{Lang: lang}),
			} {
				if !strings.Contains(card, want) {
					t.Errorf("%s/%s/%s: want %q in\n%s", c.name, lang, kind, want, card)
				}
				// The sentence this replaces: "Продаётся за Безкоштовно".
				if strings.Contains(card, "Продаётся за") || strings.Contains(card, "Продається за") || strings.Contains(card, "Listed for") {
					t.Errorf("%s/%s/%s: a zero-price listing is not \"listed for\" anything:\n%s", c.name, lang, kind, card)
				}
			}
		}
	}

	none := base
	none.Params = nil
	if card := BuildTextOffer(none, CardOptions{Lang: i18n.RU}); !strings.Contains(card, "Цена не указана") {
		t.Errorf("a listing without a price param:\n%s", card)
	}
}

func TestCardsStayWithinLimitsOnAHugeListing(t *testing.T) {
	o := sampleOffer(t)
	o.Title = strings.Repeat("Дуже довгий заголовок ", 40)
	o.Description = strings.Repeat("опис ", 1800) // 9000 runes
	for i := 0; i < 80; i++ {
		var p domain.Param
		_ = json.Unmarshal([]byte(`{"key":"k`+strings.Repeat("x", i%7)+`","name":"`+strings.Repeat("Характеристика ", 4)+
			`","type":"select","value":{"label":"`+strings.Repeat("значення ", 15)+`"}}`), &p)
		o.Params = append(o.Params, p)
	}

	text := BuildTextOffer(o, CardOptions{Lang: i18n.RU, ShowParams: true, FilterName: "f"})
	if n := len([]rune(text)); n > 4096 {
		t.Fatalf("text card is %d runes, Telegram rejects more than 4096", n)
	}
	if !strings.Contains(text, "🔗 <a href=") {
		t.Error("the link must survive trimming")
	}

	rich := BuildRichOffer(o, CardOptions{Lang: i18n.RU, ShowParams: true}, nil)
	if strings.Count(rich, "<tr>") > maxRichParams {
		t.Errorf("rich card has %d detail rows, cap is %d", strings.Count(rich, "<tr>"), maxRichParams)
	}
	if n := len([]rune(rich)); n > maxRichDescRunes+maxRichParams*200+2000 {
		t.Errorf("rich card is %d runes, the description cap did not apply", n)
	}
}

func TestSendOfferRich(t *testing.T) {
	c, srv := newClient(t)
	o := sampleOffer(t)
	id, err := c.SendOffer(ctx, "-100500", o, CardOptions{Lang: i18n.UK, Rich: true, MaxPhotos: 5, Silent: true, FilterName: "Телефони"})
	if err != nil || id == 0 {
		t.Fatalf("SendOffer = %d, %v", id, err)
	}
	call, _ := srv.Last("sendRichMessage")
	if call.ChatID() != "-100500" || call.Payload["disable_notification"] != true {
		t.Errorf("unexpected payload: %s", call.Raw)
	}
	if got := call.Buttons(); len(got) != 1 || got[0] != o.URL {
		t.Errorf("buttons = %v, want the listing link", got)
	}
	if !strings.Contains(call.Text(), "Продається за") {
		t.Errorf("card must be in the chat language:\n%s", call.Text())
	}
	if n := len(srv.Calls("sendMessage")); n != 0 {
		t.Errorf("a delivered rich card must not be followed by a plain one, sendMessage calls = %d", n)
	}
}

func TestSendOfferRetriesWithoutPhotosOnce(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("sendRichMessage", tgtest.Fail(400, "Bad Request: RICH_MESSAGE_PHOTO_NO_MEDIA_FOUND"))
	if _, err := c.SendOffer(ctx, "1", sampleOffer(t), CardOptions{Lang: i18n.RU, Rich: true, MaxPhotos: 5}); err != nil {
		t.Fatal(err)
	}
	calls := srv.Calls("sendRichMessage")
	if len(calls) != 2 {
		t.Fatalf("rich attempts = %d, want 2", len(calls))
	}
	if !strings.Contains(calls[0].Text(), "<img") || strings.Contains(calls[1].Text(), "<img") {
		t.Error("the second attempt must be the same card without photos")
	}
}

func TestSendOfferFallsBackToPlainCard(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("sendRichMessage", tgtest.Fail(400, "Bad Request: RICH_MESSAGE_INVALID"))
	id, err := c.SendOffer(ctx, "1", sampleOffer(t), CardOptions{Lang: i18n.RU, Rich: true, MaxPhotos: 2})
	if err != nil || id == 0 {
		t.Fatalf("SendOffer = %d, %v", id, err)
	}
	msg, ok := srv.Last("sendMessage")
	if !ok || !strings.Contains(msg.Text(), "🔥 <b>Новое объявление!</b>") {
		t.Fatalf("expected the plain card, got %q", msg.Text())
	}
	if album, ok := srv.Last("sendMediaGroup"); !ok || album.Payload["disable_notification"] != true {
		t.Error("the plain card is followed by a silent album")
	}
}

// A timeout, a flood limit and a dead chat leave the outcome to the caller:
// sending a second representation here is what used to duplicate alerts.
func TestSendOfferDoesNotDuplicateOnUnknownOutcome(t *testing.T) {
	for _, reply := range []tgtest.Reply{
		tgtest.FloodWait(30),
		tgtest.Fail(502, "Bad Gateway"),
		tgtest.Fail(403, "Forbidden: bot was blocked by the user"),
	} {
		c, srv := newClient(t)
		srv.Queue("sendRichMessage", reply)
		if _, err := c.SendOffer(ctx, "1", sampleOffer(t), CardOptions{Lang: i18n.RU, Rich: true, MaxPhotos: 3}); err == nil {
			t.Errorf("status %d must be returned to the sender", reply.Status)
		}
		if n := len(srv.Calls("")); n != 1 {
			t.Errorf("status %d: %d requests made, want exactly 1", reply.Status, n)
		}
	}
}

func TestSendOfferPlainWhenRichDisabled(t *testing.T) {
	c, srv := newClient(t)
	if _, err := c.SendOffer(ctx, "1", sampleOffer(t), CardOptions{Lang: i18n.EN, Rich: false}); err != nil {
		t.Fatal(err)
	}
	if n := len(srv.Calls("sendRichMessage")); n != 0 {
		t.Errorf("rich must not be attempted when disabled, calls = %d", n)
	}
	msg, _ := srv.Last("sendMessage")
	if !strings.Contains(msg.Text(), "New listing!") || !strings.Contains(msg.Text(), "Listed for") {
		t.Errorf("plain card in English:\n%s", msg.Text())
	}
	if n := len(srv.Calls("sendMediaGroup")) + len(srv.Calls("sendPhoto")); n != 0 {
		t.Error("no photos were requested")
	}
}

func TestUnauthorizedAndConflict(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("getUpdates", tgtest.Fail(401, "Unauthorized"))
	_, err := c.GetUpdates(ctx, 0)
	if !IsUnauthorized(err) {
		t.Errorf("401 must be reported as a bad token: %v", err)
	}
	srv.Queue("getUpdates", tgtest.Fail(http.StatusConflict, "Conflict: terminated by other getUpdates request"))
	_, err = c.GetUpdates(ctx, 0)
	if !IsConflict(err) {
		t.Errorf("409 must be reported as a second instance: %v", err)
	}
}

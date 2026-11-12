package test

import (
	"testing"
	"time"

	"pantheon/core/bot"
)

func TestDeadChatDetection(t *testing.T) {
	dead := []string{
		`HTTP 403: {"ok":false,"error_code":403,"description":"Forbidden: user is deactivated"}`,
		`HTTP 403: {"ok":false,"error_code":403,"description":"Forbidden: bot was blocked by the user"}`,
		`HTTP 400: {"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`,
		`Forbidden: bot can't initiate conversation with a user`,
	}
	for _, s := range dead {
		if !bot.IsDeadChatError(s) {
			t.Errorf("isDeadChatError(%q) = false, want true", s)
		}
	}

	alive := []string{
		`HTTP 400: Bad Request: message is too long`,
		`HTTP 429: Too Many Requests: retry after 5`,
		``,
	}
	for _, s := range alive {
		if bot.IsDeadChatError(s) {
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
		if !bot.IsPhotoRelatedError(errString(s)) {
			t.Errorf("isPhotoRelatedError(%q) = false, want true", s)
		}
	}
	no := []string{
		`Bad Request: chat not found`,
		`Bad Request: message is too long`,
	}
	for _, s := range no {
		if bot.IsPhotoRelatedError(errString(s)) {
			t.Errorf("isPhotoRelatedError(%q) = true, want false", s)
		}
	}
	if bot.IsPhotoRelatedError(nil) {
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
		{`{"ok":false,"error_code":429}`, bot.DefaultRetryWait},
		{`not json`, bot.DefaultRetryWait},
	}
	for _, c := range cases {
		if got := bot.ParseRetryAfter([]byte(c.body)); got != c.want {
			t.Errorf("parseRetryAfter(%q) = %s, want %s", c.body, got, c.want)
		}
	}
}

func TestWaitChatReservesSlots(t *testing.T) {
	n := bot.NewTelegramNotifier("", nil)
	start := time.Now()
	n.WaitChat("1")
	n.WaitChat("1")
	elapsed := time.Since(start)
	if elapsed < bot.TelegramChatGap-100*time.Millisecond {
		t.Errorf("second send to same chat waited only %s, want >= %s", elapsed, bot.TelegramChatGap)
	}

	start = time.Now()
	n.WaitChat("2")
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("first send to a fresh chat waited %s, want ~0", elapsed)
	}
}

func TestMarkAndReviveChat(t *testing.T) {
	n := bot.NewTelegramNotifier("", nil)
	n.MarkChatDead("123", "test")
	if !n.IsChatDead("123") {
		t.Fatal("chat must be dead")
	}
	n.ReviveChat("123")
	if n.IsChatDead("123") {
		t.Fatal("chat must be alive after revive")
	}
}

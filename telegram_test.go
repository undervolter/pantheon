package main

import (
	"testing"
	"time"
)

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

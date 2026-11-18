package config

import (
	"strings"
	"testing"
)

// "+123" and "0123" pass as numbers but are not what Telegram sends. Stored
// as typed, such an entry never matched the real chat: it looked added and
// stayed locked out.
func TestChatIDsAreCanonical(t *testing.T) {
	for in, want := range map[string]string{"123": "123", "+123": "123", "0123": "123", " -100500 ": "-100500", "-0100500": "-100500"} {
		if got, ok := CanonicalChatID(in); !ok || got != want {
			t.Errorf("CanonicalChatID(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "0", "+0", "abc", "12a", "your_chat_id", "1.5"} {
		if got, ok := CanonicalChatID(bad); ok {
			t.Errorf("CanonicalChatID(%q) = %q, want a refusal", bad, got)
		}
	}

	cfg := load(t, `{"telegram_bot_token":"123456789:AAtesttesttesttesttesttesttesttest","telegram_chat_ids":["+123","0123","123"," -100500 "],"admin_chat_ids":["0123"]}`)
	if got := strings.Join(cfg.GetTelegramChatIDs(), ","); got != "123,-100500" {
		t.Fatalf("chats after load = %q, want each chat once in its canonical form", got)
	}
	if !cfg.IsAdmin("123") {
		t.Fatalf("admins = %v", cfg.Admins())
	}

	if added, err := cfg.AddTrustedChatID("+77"); err != nil || !added || !cfg.IsTrusted("77") {
		t.Fatalf("AddTrustedChatID(+77): added=%v err=%v chats=%v", added, err, cfg.GetTelegramChatIDs())
	}
	if added, _ := cfg.AddTrustedChatID("077"); added {
		t.Fatal("the same chat was added twice under another spelling")
	}
	if removed, err := cfg.RemoveChatID("077"); err != nil || !removed || cfg.IsTrusted("77") {
		t.Fatalf("RemoveChatID(077): removed=%v err=%v", removed, err)
	}
}

package telegram

import (
	"testing"

	"pantheon/core/telegram/tgtest"
)

// Telegram answered ok: the message is in the chat. A result the client
// cannot read used to come back as an ordinary error, which the alert sender
// treats as "try again" — a second copy of a card that had been delivered.
func TestAcceptedSendIsNotFailedByAnUnreadableResult(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("sendMessage", tgtest.Reply{Body: `{"ok":true,"result":"not a message"}`})
	id, err := c.SendHTML(ctx, "111", "hi", nil, false)
	if err != nil {
		t.Fatalf("an accepted send was reported as failed: %v", err)
	}
	if id != 0 {
		t.Errorf("message id = %d, want 0 for an unreadable result", id)
	}
	if n := len(srv.Calls("sendMessage")); n != 1 {
		t.Errorf("requests = %d, want no second attempt", n)
	}
}

// For calls whose whole point is the result, an unreadable one is an error:
// polling must back off instead of spinning on an empty list.
func TestUnreadableResultOfAQueryIsAnError(t *testing.T) {
	c, srv := newClient(t)
	srv.Queue("getUpdates", tgtest.Reply{Body: `{"ok":true,"result":"junk"}`})
	if _, err := c.GetUpdates(ctx, 0); err == nil {
		t.Fatal("getUpdates must report a result it cannot read")
	}
	srv.Queue("getMe", tgtest.Reply{Body: `{"ok":true,"result":[1,2]}`})
	if _, err := c.GetMe(ctx); err == nil {
		t.Fatal("getMe must report a result it cannot read")
	}
}

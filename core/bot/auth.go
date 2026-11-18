package bot

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"strings"
	"sync"
	"time"

	"pantheon/core/logx"
)

// Who may do what:
//
//   - A trusted chat (telegram_chat_ids) uses the control panel and manages
//     its own filters, language, pause and quiet hours.
//   - An admin (admin_chat_ids) also manages the chat list, the bot-wide
//     settings and the shared filters.
//   - Nobody becomes trusted by talking to the bot or by pressing a button.
//     The first admin is bound with a one-time code printed to the console;
//     everyone else is added by an admin.

const (
	pairingCodeLen     = 8
	maxPairingFailures = 5
	strangerReplyEvery = time.Hour
	pairingGuessEvery  = time.Minute
	maxStrangersKept   = 2048
)

// pairing holds the one-time code that binds the first admin.
type pairing struct {
	mu       sync.Mutex
	code     string
	failures int
}

func newPairingCode() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic("bot: crypto/rand failed: " + err.Error())
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:]))[:pairingCodeLen]
}

// preparePairing makes sure the bot has an admin, or starts waiting for one.
// A config that lists user ids gets its first user as the admin, once; an
// empty one gets a pairing code.
func (b *Bot) preparePairing() {
	had := len(b.cfg.Admins()) > 0
	has, err := b.cfg.EnsureAdmin()
	if err != nil {
		logx.Errorf("could not store the admin: %v", err)
	}
	if !has {
		b.issuePairingCode()
		return
	}
	if !had {
		// Said out loud: on an install upgraded from 0.1.x the first chat in
		// the list is whoever wrote to the bot first, which is not
		// necessarily its owner.
		logx.Print("")
		logx.Print("  Admin of this bot: chat " + strings.Join(b.cfg.Admins(), ", ") + " — the first user in telegram_chat_ids.")
		logx.Print("  The admin manages chats, bot-wide settings and shared filters. If that is not you,")
		logx.Print("  stop the bot, put your own id into admin_chat_ids in the config and start it again.")
		logx.Print("")
	}
}

func (b *Bot) issuePairingCode() {
	b.pair.mu.Lock()
	b.pair.code = newPairingCode()
	b.pair.failures = 0
	code := b.pair.code
	b.pair.mu.Unlock()

	logx.Print("")
	logx.Print("  This bot has no admin yet. To become one, send it this command in a private chat:")
	logx.Print("      /start " + code)
	if b.username != "" {
		logx.Print("  or open  https://t.me/" + b.username + "?start=" + code)
	}
	logx.Print("")
}

// tryPair checks a pairing code. The comparison is constant-time, and after
// a few wrong guesses the code is replaced, so it cannot be brute-forced
// through the chat.
func (b *Bot) tryPair(candidate string) bool {
	b.pair.mu.Lock()
	code := b.pair.code
	if code == "" {
		b.pair.mu.Unlock()
		return false
	}
	ok := subtle.ConstantTimeCompare([]byte(strings.ToLower(strings.TrimSpace(candidate))), []byte(code)) == 1
	if ok {
		b.pair.code = ""
		b.pair.mu.Unlock()
		return true
	}
	b.pair.failures++
	rotate := b.pair.failures >= maxPairingFailures
	b.pair.mu.Unlock()
	if rotate {
		logx.Warnf("too many wrong pairing codes, issuing a new one")
		b.issuePairingCode()
	}
	return false
}

// pairingOpen reports whether the bot is still waiting for its first admin.
func (b *Bot) pairingOpen() bool {
	b.pair.mu.Lock()
	defer b.pair.mu.Unlock()
	return b.pair.code != ""
}

// throttle limits how often the bot answers a chat it does not know.
type throttle struct {
	mu   sync.Mutex
	last map[string]time.Time
}

func (th *throttle) allow(chatID string, now time.Time, every time.Duration) bool {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.last == nil {
		th.last = map[string]time.Time{}
	}
	if at, ok := th.last[chatID]; ok && now.Sub(at) < every {
		return false
	}
	// The map is bounded: a stream of new chat ids cannot grow it forever.
	if len(th.last) >= maxStrangersKept {
		for id, at := range th.last {
			if now.Sub(at) >= every {
				delete(th.last, id)
			}
		}
		if len(th.last) >= maxStrangersKept {
			return false
		}
	}
	th.last[chatID] = now
	return true
}

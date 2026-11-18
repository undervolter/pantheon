package telegram

import (
	"strings"
	"sync"
	"time"
)

// Gaps between alerts into one chat. Telegram allows about one message a
// second to a user and about twenty a minute to a group or channel.
const (
	ChatGap  = 1100 * time.Millisecond
	GroupGap = 3 * time.Second
)

// Limiter spaces out alerts per chat. It belongs to the alert sender only:
// menu traffic never waits in this line.
type Limiter struct {
	mu       sync.Mutex
	next     map[string]time.Time
	now      func() time.Time
	chatGap  time.Duration
	groupGap time.Duration
}

func NewLimiter() *Limiter {
	return NewLimiterWith(ChatGap, GroupGap)
}

// NewLimiterWith builds a limiter with its own gaps.
func NewLimiterWith(chatGap, groupGap time.Duration) *Limiter {
	return &Limiter{next: make(map[string]time.Time), now: time.Now, chatGap: chatGap, groupGap: groupGap}
}

func (l *Limiter) gap(chatID string) time.Duration {
	if strings.HasPrefix(chatID, "-") {
		return l.groupGap
	}
	return l.chatGap
}

// Reserve books the next free send slot for the chat and returns how long the
// caller has to wait for it.
func (l *Limiter) Reserve(chatID string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	at := now
	if next, ok := l.next[chatID]; ok && next.After(at) {
		at = next
	}
	l.next[chatID] = at.Add(l.gap(chatID))
	return at.Sub(now)
}

// Penalize holds the chat back for d after a flood limit.
func (l *Limiter) Penalize(chatID string, d time.Duration) {
	until := l.now().Add(d)
	l.mu.Lock()
	if cur := l.next[chatID]; until.After(cur) {
		l.next[chatID] = until
	}
	l.mu.Unlock()
}

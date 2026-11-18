// Package bot is the Telegram control panel: menus, the filter wizard,
// settings, and who is allowed to use them.
package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"pantheon/core/config"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/version"
)

const (
	// A chat that floods the bot cannot grow its backlog without bound.
	maxQueuedPerChat = 256
	// One update may involve a few OLX geo requests; anything longer is stuck.
	updateTimeout  = 90 * time.Second
	maxPollBackoff = 30 * time.Second
)

// Geo is the part of the OLX client the wizard needs to pick a location.
type Geo interface {
	SearchLocations(query string) ([]olx.GeoLocationSuggestion, error)
	FetchCityDistricts(cityID int) ([]olx.GeoDistrict, error)
	FetchCity(cityID int) (*olx.GeoCity, error)
	FetchRegion(regionID int) (*olx.GeoRegion, error)
	ResolveSlug(slug string) (olx.GeoCity, olx.GeoRegion, bool)
	WarmGeoIndex()
}

// ScanStatus is what the main menu shows about the scanner.
type ScanStatus struct {
	// Started is set once the first poll cycle has finished.
	Started bool
	// BlockedUntil is set while OLX is refusing requests.
	BlockedUntil time.Time
}

// Scanner is the part of the checker the control panel talks to.
type Scanner interface {
	// TriggerCheck asks for a poll cycle right now.
	TriggerCheck()
	Status() ScanStatus
}

type Bot struct {
	tg   *telegram.Client
	cfg  *config.Config
	db   *store.Store
	geo  Geo
	scan Scanner

	username  string
	pair      pairing
	states    states
	queues    chatQueues
	strangers throttle
	now       func() time.Time
}

func New(tg *telegram.Client, cfg *config.Config, db *store.Store, geo Geo, scan Scanner) *Bot {
	return &Bot{tg: tg, cfg: cfg, db: db, geo: geo, scan: scan, now: time.Now}
}

// menuCommands is what the "/" button of the chat offers, in this order.
var menuCommands = []string{"start", "filters", "add", "history", "settings", "pause", "resume", "lang", "export", "help"}

func (b *Bot) setupCommands(ctx context.Context) {
	for i, lang := range append([]i18n.Lang{i18n.EN}, i18n.All...) {
		commands := make([]telegram.BotCommand, 0, len(menuCommands))
		for _, c := range menuCommands {
			commands = append(commands, telegram.BotCommand{Command: c, Description: i18n.T(lang, "bc."+c)})
		}
		code := ""
		if i > 0 {
			code = string(lang)
		}
		if err := b.tg.SetCommands(ctx, commands, code); err != nil {
			logx.Warnf("[TG] setMyCommands(%s): %v", lang, err)
		}
	}
	if err := b.tg.ShowCommandsButton(ctx); err != nil {
		logx.Warnf("[TG] setChatMenuButton: %v", err)
	}
}

// Run polls Telegram until ctx is cancelled. It returns an error only when
// polling cannot work at all (a rejected token).
func (b *Bot) Run(ctx context.Context) error {
	me, err := b.tg.GetMe(ctx)
	if err != nil {
		if telegram.IsUnauthorized(err) {
			return fmt.Errorf("Telegram rejected the bot token (%v): check telegram_bot_token", err)
		}
		logx.Warnf("[TG] getMe: %v", err)
	}
	b.username = me.Username

	if err := b.tg.DeleteWebhook(ctx); err != nil && ctx.Err() == nil {
		logx.Warnf("[TG] deleteWebhook: %v", err)
	}
	b.setupCommands(ctx)
	b.preparePairing()
	logx.Infof("🤖 Telegram control panel is up (@%s)", b.username)
	b.announce(ctx)

	offset := 0
	backoff := time.Second
	for ctx.Err() == nil {
		updates, err := b.tg.GetUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			switch {
			case telegram.IsUnauthorized(err):
				return fmt.Errorf("Telegram rejected the bot token (%v): check telegram_bot_token", err)
			case telegram.IsConflict(err):
				logx.Errorf("[TG] another instance is polling this bot, or a webhook is set (%v); retry in %s", err, backoff)
			default:
				logx.Warnf("[TG] getUpdates: %v (retry in %s)", err, backoff)
			}
			select {
			case <-ctx.Done():
			case <-time.After(backoff):
			}
			if backoff < maxPollBackoff {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second

		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}
			b.enqueue(ctx, u)
		}
	}
	return nil
}

// announce shows the main menu in every private chat once per version, so a
// new release is noticed without a restart turning into spam: under
// Restart=always the old "menu on every start" flooded the chats.
func (b *Bot) announce(ctx context.Context) {
	fresh, err := b.cfg.MarkAnnounced(version.AppVersion)
	if err != nil || !fresh {
		return
	}
	for _, chatID := range b.cfg.GetTelegramChatIDs() {
		if config.IsGroupID(chatID) {
			continue
		}
		go func(chatID string) {
			t := b.newTurn(ctx, chatID, chatID)
			t.send(b.viewMain(t))
		}(chatID)
	}
}

// ---------- per-chat queues ----------

// chatQueues holds the updates waiting behind the one a chat is busy with.
// A chat has an entry only while its worker is running.
type chatQueues struct {
	mu      sync.Mutex
	pending map[string][]telegram.TGUpdate
}

func updateChat(u telegram.TGUpdate) string {
	switch {
	case u.Message != nil:
		return strconv.FormatInt(u.Message.Chat.ID, 10)
	case u.CallbackQuery != nil && u.CallbackQuery.Message != nil:
		return strconv.FormatInt(u.CallbackQuery.Message.Chat.ID, 10)
	case u.CallbackQuery != nil:
		return "from:" + strconv.FormatInt(u.CallbackQuery.From.ID, 10)
	}
	return ""
}

// enqueue takes an update off the polling loop. Updates of one chat are
// handled in the order they arrived, because its wizard state is not shared;
// different chats do not wait for each other, so one slow OLX lookup cannot
// stall the bot for everybody.
func (b *Bot) enqueue(ctx context.Context, u telegram.TGUpdate) {
	chat := updateChat(u)
	if chat == "" {
		return
	}
	b.queues.mu.Lock()
	if b.queues.pending == nil {
		b.queues.pending = make(map[string][]telegram.TGUpdate)
	}
	backlog, busy := b.queues.pending[chat]
	if !busy {
		b.queues.pending[chat] = nil
		b.queues.mu.Unlock()
		go b.drain(ctx, chat, u)
		return
	}
	if len(backlog) < maxQueuedPerChat {
		b.queues.pending[chat] = append(backlog, u)
	}
	b.queues.mu.Unlock()
}

func (b *Bot) drain(ctx context.Context, chat string, u telegram.TGUpdate) {
	for {
		b.dispatch(ctx, u)
		b.queues.mu.Lock()
		backlog := b.queues.pending[chat]
		if len(backlog) == 0 {
			delete(b.queues.pending, chat)
			b.queues.mu.Unlock()
			return
		}
		u, b.queues.pending[chat] = backlog[0], backlog[1:]
		b.queues.mu.Unlock()
	}
}

func (b *Bot) dispatch(parent context.Context, u telegram.TGUpdate) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("[TG] panic while handling update %d: %v", u.UpdateID, r)
		}
	}()
	ctx, cancel := context.WithTimeout(parent, updateTimeout)
	defer cancel()
	switch {
	case u.CallbackQuery != nil:
		b.onCallback(ctx, u.CallbackQuery)
	case u.Message != nil:
		b.onMessage(ctx, u.Message)
	}
}

// ---------- one handled update ----------

// turn carries everything a handler needs to answer one update.
type turn struct {
	b      *Bot
	ctx    context.Context
	chatID string
	userID string
	lang   i18n.Lang
	st     *state
	admin  bool
	group  bool

	// Set for a pressed button.
	cb       *telegram.CallbackQuery
	msgID    int
	answered bool
}

func (b *Bot) lang(chatID string) i18n.Lang {
	return i18n.Of(b.cfg.LangFor(chatID))
}

func (b *Bot) newTurn(ctx context.Context, chatID, userID string) *turn {
	return &turn{
		b:      b,
		ctx:    ctx,
		chatID: chatID,
		userID: userID,
		lang:   b.lang(chatID),
		st:     b.states.get(stateKey(chatID, userID)),
		admin:  b.cfg.IsAdmin(userID),
		group:  chatID != userID,
	}
}

func (t *turn) tr(key string, args ...any) string {
	return i18n.T(t.lang, key, args...)
}

// resetState drops the wizard draft and any pending text input.
func (t *turn) resetState() {
	t.st = t.b.states.reset(stateKey(t.chatID, t.userID))
}

// send posts a page as a new message.
func (t *turn) send(view string, km *markup) {
	id, err := t.b.tg.SendPage(t.ctx, t.chatID, view, km)
	if err != nil {
		if t.ctx.Err() == nil {
			logx.Warnf("[TG] send to %s: %v", t.chatID, err)
		}
		return
	}
	t.moveWizard(id)
}

// show replaces the message the button was pressed on; without a message to
// replace (a typed command, an edit the server refused) it sends a new one.
func (t *turn) show(view string, km *markup) {
	if t.msgID > 0 {
		err := t.b.tg.EditPage(t.ctx, t.chatID, t.msgID, view, km)
		if err == nil {
			t.moveWizard(t.msgID)
			return
		}
		if !errors.Is(err, telegram.ErrPageEditRejected) && t.ctx.Err() == nil {
			logx.Warnf("[TG] edit in %s: %v", t.chatID, err)
		}
	}
	t.send(view, km)
}

// moveWizard records which message the wizard now lives in. When typed input
// moved it to a fresh message, the buttons of the previous one are removed so
// nobody presses a stale step.
func (t *turn) moveWizard(msgID int) {
	if !t.st.Active || msgID == 0 || t.st.MsgID == msgID {
		return
	}
	old := t.st.MsgID
	t.st.MsgID = msgID
	if old != 0 {
		_ = t.b.tg.Call(t.ctx, "editMessageReplyMarkup", map[string]any{
			"chat_id": t.chatID, "message_id": old, "reply_markup": map[string]any{"inline_keyboard": [][]button{}},
		}, nil)
	}
}

// say sends a one-off line (an error, a confirmation) as its own message.
func (t *turn) say(key string, args ...any) {
	if _, err := t.b.tg.SendHTML(t.ctx, t.chatID, t.tr(key, args...), nil, false); err != nil && t.ctx.Err() == nil {
		logx.Warnf("[TG] send to %s: %v", t.chatID, err)
	}
}

// toast answers a pressed button with a short notice; each press is answered once.
func (t *turn) toast(text string) { t.answer(text, false) }

// alert answers a pressed button with a modal the user has to dismiss.
func (t *turn) alert(text string) { t.answer(text, true) }

func (t *turn) answer(text string, modal bool) {
	if t.cb == nil || t.answered {
		return
	}
	t.answered = true
	if err := t.b.tg.AnswerCallback(t.ctx, t.cb.ID, text, modal); err != nil && t.ctx.Err() == nil {
		// An answer that arrives too late is rejected; the menu still updates.
		if !strings.Contains(err.Error(), "query is too old") {
			logx.Warnf("[TG] answerCallbackQuery: %v", err)
		}
	}
}

// saveFailed reports a failed write to the user.
func (t *turn) saveFailed(err error) {
	logx.Errorf("save failed for chat %s: %v", t.chatID, err)
	msg := t.tr("common.save_failed", esc(err.Error()))
	if t.cb != nil {
		t.alert(i18n.T(t.lang, "common.save_failed", err.Error()))
		return
	}
	_, _ = t.b.tg.SendHTML(t.ctx, t.chatID, msg, nil, false)
}

// parseCommand splits "/cmd@bot args" and drops commands addressed to
// another bot of the same group.
func parseCommand(text, botUsername string) (cmd, args string, ok bool) {
	if !strings.HasPrefix(text, "/") {
		return "", "", false
	}
	head, rest, _ := strings.Cut(text, " ")
	head = strings.ToLower(strings.TrimPrefix(head, "/"))
	if name, target, found := strings.Cut(head, "@"); found {
		if botUsername != "" && !strings.EqualFold(target, botUsername) {
			return "", "", false
		}
		head = name
	}
	if head == "" {
		return "", "", false
	}
	return head, strings.TrimSpace(rest), true
}

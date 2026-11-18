package bot

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"pantheon/core/config"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/telegram"
	"pantheon/core/version"
)

// onMessage handles a message: a command, a typed answer to a wizard step, a
// pasted OLX link or an export file.
func (b *Bot) onMessage(ctx context.Context, msg *telegram.TGMessage) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	userID := strconv.FormatInt(msg.From.ID, 10)
	group := msg.Chat.Type != "" && msg.Chat.Type != "private"
	text := strings.TrimSpace(msg.Text)
	cmd, args, isCmd := parseCommand(text, b.username)

	if !group {
		if !b.cfg.IsTrusted(chatID) {
			b.onStranger(ctx, chatID, cmd, args, isCmd, msg.From.LanguageCode)
			return
		}
	} else {
		// In a group only trusted users are heard, and only inside a linked
		// group. The exceptions let an admin link the group in the first place.
		if msg.From.ID == 0 || !b.cfg.IsTrusted(userID) {
			return
		}
		if !b.cfg.IsTrusted(chatID) && !(isCmd && (cmd == "bind" || cmd == "chatid")) {
			return
		}
	}

	t := b.newTurn(ctx, chatID, userID)
	b.detectLang(t, msg.From.LanguageCode)

	switch {
	case isCmd:
		b.onCommand(t, cmd, args, msg)
	case msg.Document != nil:
		b.onDocument(t, msg.Document)
	case text == "":
	case t.st.Step != stepNone:
		b.onStepText(t, text)
	case olx.IsOLXURL(text):
		b.onURL(t, text)
	case !group:
		// Ordinary chatter in a group is nobody's business; in a private chat
		// the menu is the most useful answer to anything else.
		t.send(b.viewMain(t))
	}
}

// detectLang gives a private chat the language of its Telegram client the
// first time the bot hears from it.
func (b *Bot) detectLang(t *turn, code string) {
	if t.group || b.cfg.HasLang(t.chatID) {
		return
	}
	lang, ok := i18n.Parse(code)
	if !ok {
		return
	}
	if err := b.cfg.SetLang(t.chatID, string(lang)); err != nil {
		logx.Warnf("store language of %s: %v", t.chatID, err)
	}
	t.lang = lang
}

// onStranger answers a private chat the bot does not trust. The only things
// it reacts to are a pairing attempt and /start, and the latter not more
// often than once an hour — the bot must not become a way to make it talk.
func (b *Bot) onStranger(ctx context.Context, chatID, cmd, args string, isCmd bool, langCode string) {
	if !isCmd || cmd != "start" {
		return
	}
	lang, _ := i18n.Parse(langCode)

	if args != "" && b.pairingOpen() {
		if !b.tryPair(args) {
			_, _ = b.tg.SendHTML(ctx, chatID, i18n.T(lang, "acc.pair_bad"), nil, false)
			return
		}
		bound, err := b.cfg.BindFirstAdmin(chatID)
		if err != nil || !bound {
			logx.Errorf("pairing of %s failed: bound=%v err=%v", chatID, bound, err)
			return
		}
		logx.Infof("👤 chat %s is now the admin of this bot", chatID)
		t := b.newTurn(ctx, chatID, chatID)
		b.detectLang(t, langCode)
		view, km := b.viewMain(t)
		t.send(withNotice(t.tr("acc.paired"), view), km)
		return
	}

	if b.strangers.allow(chatID, b.now(), strangerReplyEvery) {
		_, _ = b.tg.SendHTML(ctx, chatID, i18n.T(lang, "acc.stranger", esc(chatID)), nil, false)
	}
}

func (b *Bot) onCommand(t *turn, cmd, args string, msg *telegram.TGMessage) {
	switch cmd {
	case "start", "menu":
		t.resetState()
		t.send(b.viewMain(t))

	case "filters":
		t.resetState()
		t.send(b.viewFilters(t, 0))

	case "add", "new":
		b.startWizard(t)
		t.send(b.viewCategory(t))

	case "settings":
		t.resetState()
		t.send(b.viewSettings(t))

	case "lang", "language":
		t.send(b.viewLang(t))

	case "help":
		t.send(b.viewHelp(t))

	case "history", "recent":
		t.send(b.viewHistory(t))

	case "cancel":
		t.resetState()
		view, km := b.viewMain(t)
		t.send(withNotice(t.tr("cmd.cancelled"), view), km)

	case "pause":
		if err := b.cfg.SetPaused(t.chatID, true); err != nil {
			t.saveFailed(err)
			return
		}
		t.say("cmd.paused")

	case "resume":
		if err := b.cfg.SetPaused(t.chatID, false); err != nil {
			t.saveFailed(err)
			return
		}
		t.say("cmd.resumed")

	case "export":
		b.exportTo(t)

	case "chatid":
		kind := msg.Chat.Type
		if kind == "" {
			kind = "private"
		}
		t.say("cmd.chatid", esc(t.chatID), esc(kind))

	case "version":
		t.say("cmd.version", version.AppVersion)

	case "check":
		if b.scan != nil {
			b.scan.TriggerCheck()
		}
		t.say("set.check_queued")

	case "chats", "bind", "unbind", "addchat", "delchat":
		if !t.admin {
			t.say("common.admin_only")
			return
		}
		b.onChatCommand(t, cmd, args)

	default:
		t.say("cmd.unknown")
	}
}

// onChatCommand manages the list of trusted chats. Admin only.
func (b *Bot) onChatCommand(t *turn, cmd, args string) {
	switch cmd {
	case "chats":
		t.send(b.viewChats(t))

	case "bind":
		if !t.group {
			t.say("cmd.group_only")
			return
		}
		added, err := b.cfg.AddTrustedChatID(t.chatID)
		switch {
		case err != nil:
			t.saveFailed(err)
		case added:
			logx.Infof("💬 group %s linked by %s", t.chatID, t.userID)
			t.say("cmd.bound")
		default:
			t.say("cmd.chat_exists")
		}

	case "unbind":
		if !t.group {
			t.say("cmd.group_only")
			return
		}
		if !b.cfg.IsTrusted(t.chatID) {
			t.say("cmd.not_bound")
			return
		}
		paused, err := b.removeChat(t, t.chatID)
		if err != nil {
			t.saveFailed(err)
			return
		}
		t.say("cmd.unbound", paused)

	case "addchat":
		if t.group {
			t.say("cmd.use_bind")
			return
		}
		id := strings.TrimSpace(args)
		if !config.ValidChatID(id) {
			t.say("cmd.addchat_usage")
			return
		}
		added, err := b.cfg.AddTrustedChatID(id)
		if err != nil {
			t.saveFailed(err)
			return
		}
		if !added {
			t.say("cmd.chat_exists")
			return
		}
		logx.Infof("💬 chat %s added by %s", id, t.userID)
		// A user who never started the bot, or a channel without the bot in
		// it, cannot be written to; say so right away.
		if _, err := b.tg.SendHTML(t.ctx, id, i18n.T(b.lang(id), "cmd.connected"), nil, false); err != nil {
			t.say("cmd.chat_added_mute", esc(id))
			return
		}
		t.say("cmd.chat_added", esc(id))

	case "delchat":
		if t.group {
			t.say("cmd.use_bind")
			return
		}
		id := strings.TrimSpace(args)
		if !config.ValidChatID(id) {
			t.say("cmd.delchat_usage")
			return
		}
		if !b.cfg.IsTrusted(id) {
			t.say("cmd.chat_missing")
			return
		}
		paused, err := b.removeChat(t, id)
		switch {
		case errors.Is(err, config.ErrLastAdmin):
			t.say("cmd.last_admin")
		case err != nil:
			t.saveFailed(err)
		default:
			t.say("cmd.chat_removed", esc(id), paused)
		}
	}
}

// removeChat takes a chat off the trusted list. Its filters are paused first:
// a filter whose owner is gone would keep polling OLX for nobody and would be
// invisible in every menu. What is still queued for the chat is dropped.
func (b *Bot) removeChat(t *turn, chatID string) (pausedFilters int, err error) {
	if b.cfg.IsAdmin(chatID) && len(b.cfg.Admins()) == 1 {
		return 0, config.ErrLastAdmin
	}
	if pausedFilters, err = b.db.DisableSearchesOf(t.ctx, chatID); err != nil {
		return 0, err
	}
	if _, err = b.cfg.RemoveChatID(chatID); err != nil {
		return pausedFilters, err
	}
	if _, err := b.db.FailPendingOf(t.ctx, chatID, "chat removed"); err != nil {
		logx.Warnf("drop queued alerts of %s: %v", chatID, err)
	}
	logx.Infof("💬 chat %s removed by %s (%d filters paused)", chatID, t.userID, pausedFilters)
	return pausedFilters, nil
}

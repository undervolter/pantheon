package bot

import (
	"context"
	"strconv"
	"strings"

	"pantheon/core/i18n"
	"pantheon/core/telegram"
)

// onCallback handles a pressed inline button.
//
// A button can be pressed by anyone who can see the message: every member of
// a group the menu was posted to. Pressing one never grants access — the user
// must already be trusted, and so must the chat.
func (b *Bot) onCallback(ctx context.Context, q *telegram.CallbackQuery) {
	userID := strconv.FormatInt(q.From.ID, 10)
	if q.Message == nil || q.Message.Chat.ID == 0 {
		// Nothing to edit (the message is too old to be included).
		_ = b.tg.AnswerCallback(ctx, q.ID, "", false)
		return
	}
	chatID := strconv.FormatInt(q.Message.Chat.ID, 10)
	group := chatID != userID

	if !b.cfg.IsTrusted(userID) || (group && !b.cfg.IsTrusted(chatID)) {
		_ = b.tg.AnswerCallback(ctx, q.ID, i18n.T(b.lang(chatID), "common.denied"), true)
		return
	}

	t := b.newTurn(ctx, chatID, userID)
	t.cb, t.msgID = q, q.Message.MessageID
	// Whatever happens below, the spinner on the button is closed once.
	defer t.answer("", false)

	name, rest, _ := strings.Cut(q.Data, ":")
	var args []string
	if rest != "" {
		args = strings.Split(rest, ":")
	}
	if name == telegram.NoopData {
		return
	}
	if adminCallbacks[name] && !t.admin {
		t.alert(t.tr("common.admin_only"))
		return
	}

	if wizardCallbacks[name] {
		if !t.st.Active {
			// The draft is gone: the bot was restarted or the filter was
			// already launched. Starting over is the only honest answer —
			// carrying on would launch an empty filter.
			t.toast(t.tr("wz.expired"))
			t.show(b.viewMain(t))
			return
		}
		if t.st.MsgID != 0 && t.st.MsgID != t.msgID {
			// A button of an earlier wizard screen. The draft has moved on;
			// acting on it from here would undo what was typed since.
			t.toast(t.tr("common.stale_keep"))
			return
		}
		t.st.Step = stepNone
		b.wizardCallback(t, name, args)
		return
	}

	// A pressed button cancels whatever text the bot was waiting for.
	t.st.Step = stepNone

	switch {
	case strings.HasPrefix(name, "f_"):
		b.filterCallback(t, name, args)
	case strings.HasPrefix(name, "set_"), name == "check_now", name == "menu_chats":
		b.settingsCallback(t, name)
	default:
		b.menuCallback(t, name, args)
	}
}

func (b *Bot) menuCallback(t *turn, name string, args []string) {
	switch name {
	case "menu_main":
		t.resetState()
		t.show(b.viewMain(t))

	case "menu_list":
		t.show(b.viewFilters(t, argInt(args, 0)))

	case "add_new":
		b.startWizard(t)
		t.show(b.viewCategory(t))

	case "history_show":
		t.show(b.viewHistory(t))

	case "menu_settings":
		t.show(b.viewSettings(t))

	case "lang_menu":
		t.show(b.viewLang(t))

	case "lang":
		lang, ok := i18n.Parse(firstArg(args))
		if !ok {
			t.show(b.viewLang(t))
			return
		}
		if err := b.cfg.SetLang(t.chatID, string(lang)); err != nil {
			t.saveFailed(err)
			return
		}
		t.lang = lang
		t.toast(t.tr("lang.set", lang.Name()))
		t.show(b.viewMain(t))

	case "help_show":
		t.show(b.viewHelp(t))

	case "chat_pause", "chat_resume":
		pause := name == "chat_pause"
		if err := b.cfg.SetPaused(t.chatID, pause); err != nil {
			t.saveFailed(err)
			return
		}
		if pause {
			t.toast(t.tr("cmd.paused"))
		} else {
			t.toast(t.tr("cmd.resumed"))
		}
		t.show(b.viewMain(t))

	case "export":
		b.exportTo(t)

	case "import_ok":
		b.confirmImport(t)

	case "import_cancel":
		t.st.Import = nil
		t.show(b.viewMain(t))

	default:
		// A button from a build that no longer exists.
		t.toast(t.tr("common.stale"))
		t.resetState()
		t.show(b.viewMain(t))
	}
}

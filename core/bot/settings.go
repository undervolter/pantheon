package bot

import (
	"fmt"
	"strconv"

	"pantheon/core/config"
	"pantheon/core/telegram"
)

// Values the cycle buttons step through.
var (
	photoSteps    = []int{0, 1, 3, 5, 10}
	ageSteps      = []int{0, 20, 30, 45, 60, 120}
	intervalSteps = []int{5, 10, 15, 30, 60}
	fetchSteps    = []int{20, 35, 50}
	quietPresets  = [][2]int{{22, 8}, {23, 8}, {0, 8}, {23, 7}}
)

// adminCallbacks change the bot for everyone, so only an admin may press them.
var adminCallbacks = map[string]bool{
	"set_style": true, "set_photos": true, "set_params": true, "set_age": true, "set_bumped": true,
	"set_interval": true, "set_fetch": true, "set_backfill": true, "menu_chats": true, "check_now": true,
}

// nextInt returns the value after cur in steps, wrapping around. A value
// that is not among the steps (set by hand in the config) continues from the
// first step above it.
func nextInt(steps []int, cur int) int {
	for i, v := range steps {
		if v == cur {
			return steps[(i+1)%len(steps)]
		}
	}
	for _, v := range steps {
		if v > cur {
			return v
		}
	}
	return steps[0]
}

func (t *turn) onOff(v bool) string {
	if v {
		return t.tr("common.on")
	}
	return t.tr("common.off")
}

func (t *turn) photosText() string {
	if !t.b.cfg.GetSendPhotos() {
		return t.tr("common.off")
	}
	return t.tr("set.photos_upto", t.b.cfg.GetMaxPhotos())
}

func (t *turn) ageText() string {
	if age := t.b.cfg.GetMaxAlertAgeMin(); age > 0 {
		return t.tr("dur.m", age)
	}
	return t.tr("set.age_none")
}

func (t *turn) styleText() string {
	if t.b.cfg.GetRichAlerts() {
		return t.tr("set.style_rich")
	}
	return t.tr("set.style_text")
}

func (t *turn) bumpedText() string {
	if t.b.cfg.GetSkipBumpedOld() {
		return t.tr("set.bumped_hide")
	}
	return t.tr("set.bumped_show")
}

func (t *turn) quietText() string {
	if from, to, ok := t.b.cfg.GetQuietHours(t.chatID); ok {
		return t.tr("set.quiet_range", from, to)
	}
	return t.tr("common.off")
}

func (b *Bot) viewSettings(t *turn) (string, *markup) {
	cfg := b.cfg
	p := newPage(t.tr("set.title"))

	p.section(t.tr("set.section_alerts"))
	p.kv(t.tr("set.style"), t.styleText())
	p.kv(t.tr("set.photos"), t.photosText())
	p.kv(t.tr("set.params"), t.onOff(cfg.GetShowParams()))
	p.kv(t.tr("set.age"), t.ageText())
	p.kv(t.tr("set.bumped"), t.bumpedText())
	p.kv(t.tr("set.quiet"), t.quietText())

	if !t.admin {
		p.note(t.tr("set.user_note"))
		km := telegram.Markup(
			telegram.Row(telegram.Btn(t.tr("set.btn_quiet"), "set_quiet"), telegram.Btn(t.tr("btn.lang"), "lang_menu")),
			telegram.Row(btnBack(t.lang, "menu_main")),
		)
		return p.String(), km
	}

	_, fetch := cfg.PerfSettings()
	p.section(t.tr("set.section_system"))
	p.kv(t.tr("set.interval"), t.tr("dur.s", cfg.GetInterval()))
	p.kv(t.tr("set.fetch"), strconv.Itoa(fetch))
	p.kv(t.tr("set.backfill"), t.onOff(cfg.GetBackfillEnabled()))
	p.kv(t.tr("set.chats"), strconv.Itoa(len(cfg.GetTelegramChatIDs())))
	p.note(t.tr("set.note"))
	if age := cfg.GetMaxAlertAgeMin(); age > 0 && age < 20 {
		p.note(t.tr("set.age_hint"))
	}

	km := telegram.Markup(
		telegram.Row(telegram.Btn(t.tr("set.btn_style"), "set_style"), telegram.Btn(t.tr("set.btn_photos"), "set_photos")),
		telegram.Row(toggleBtn(cfg.GetShowParams(), t.tr("set.btn_params"), "set_params"), telegram.Btn(t.tr("set.btn_age"), "set_age")),
		telegram.Row(toggleBtn(cfg.GetSkipBumpedOld(), t.tr("set.btn_bumped"), "set_bumped"), telegram.Btn(t.tr("set.btn_quiet"), "set_quiet")),
		telegram.Row(telegram.Btn(t.tr("set.btn_interval"), "set_interval"), telegram.Btn(t.tr("set.btn_fetch"), "set_fetch")),
		telegram.Row(toggleBtn(cfg.GetBackfillEnabled(), t.tr("set.btn_backfill"), "set_backfill"), telegram.Btn(t.tr("set.btn_chats"), "menu_chats")),
		telegram.Row(telegram.Btn(t.tr("set.btn_check"), "check_now")),
		telegram.Row(btnBack(t.lang, "menu_main")),
	)
	return p.String(), km
}

// changed confirms a setting with a "<label>: <value>" toast, or reports the
// failed write.
func (t *turn) changed(err error, labelKey, value string) bool {
	if err != nil {
		t.saveFailed(err)
		return false
	}
	t.toast(t.tr("common.changed", t.tr(labelKey), value))
	return true
}

func (b *Bot) settingsCallback(t *turn, name string) {
	cfg := b.cfg
	switch name {
	case "set_style":
		t.changed(cfg.SetRichAlerts(!cfg.GetRichAlerts()), "set.style", t.styleText())

	case "set_photos":
		cur := 0
		if cfg.GetSendPhotos() {
			cur = cfg.GetMaxPhotos()
		}
		next := nextInt(photoSteps, cur)
		var err error
		if next == 0 {
			err = cfg.SetSendPhotos(false)
		} else if err = cfg.SetMaxPhotos(next); err == nil {
			err = cfg.SetSendPhotos(true)
		}
		t.changed(err, "set.photos", t.photosText())

	case "set_params":
		t.changed(cfg.SetShowParams(!cfg.GetShowParams()), "set.params", t.onOff(cfg.GetShowParams()))

	case "set_age":
		t.changed(cfg.SetMaxAlertAgeMin(nextInt(ageSteps, cfg.GetMaxAlertAgeMin())), "set.age", t.ageText())

	case "set_bumped":
		t.changed(cfg.SetSkipBumpedOld(!cfg.GetSkipBumpedOld()), "set.bumped", t.bumpedText())

	case "set_quiet":
		// Presets in a ring, ending with "off".
		from, to, on := cfg.GetQuietHours(t.chatID)
		next := 0
		if on {
			next = len(quietPresets)
			for i, p := range quietPresets {
				if p[0] == from && p[1] == to {
					next = i + 1
					break
				}
			}
		}
		var err error
		if next >= len(quietPresets) {
			err = cfg.SetQuietHours(t.chatID, 0, 0, false)
		} else {
			err = cfg.SetQuietHours(t.chatID, quietPresets[next][0], quietPresets[next][1], true)
		}
		t.changed(err, "set.quiet", t.quietText())

	case "set_interval":
		t.changed(cfg.SetInterval(nextInt(intervalSteps, cfg.GetInterval())), "set.interval", t.tr("dur.s", cfg.GetInterval()))

	case "set_fetch":
		_, cur := cfg.PerfSettings()
		err := cfg.SetFetchLimit(nextInt(fetchSteps, cur))
		_, now := cfg.PerfSettings()
		t.changed(err, "set.fetch", strconv.Itoa(now))

	case "set_backfill":
		t.changed(cfg.SetBackfillEnabled(!cfg.GetBackfillEnabled()), "set.backfill", t.onOff(cfg.GetBackfillEnabled()))

	case "check_now":
		if b.scan != nil {
			b.scan.TriggerCheck()
		}
		t.toast(t.tr("set.check_queued"))
		return

	case "menu_chats":
		t.show(b.viewChats(t))
		return
	}
	t.show(b.viewSettings(t))
}

func (b *Bot) viewChats(t *turn) (string, *markup) {
	p := newPage(t.tr("ch.title")).text(t.tr("ch.intro"))
	var rows [][]string
	for _, id := range b.cfg.GetTelegramChatIDs() {
		role := t.tr("ch.role_user")
		switch {
		case config.IsGroupID(id):
			role = t.tr("ch.role_group")
		case b.cfg.IsAdmin(id):
			role = t.tr("ch.role_admin")
		}
		cell := fmt.Sprintf("<code>%s</code>", esc(id))
		if id == t.chatID {
			cell = t.tr("ch.this", cell)
		}
		rows = append(rows, []string{cell, role})
	}
	p.table([]string{t.tr("ch.col_chat"), t.tr("ch.col_role")}, rows)
	p.section(t.tr("ch.howto"))
	p.list([]string{t.tr("ch.help_bind"), t.tr("ch.help_unbind"), t.tr("ch.help_add"), t.tr("ch.help_del"), t.tr("ch.help_id")})
	return p.String(), telegram.Markup(telegram.Row(btnBack(t.lang, "menu_settings")))
}

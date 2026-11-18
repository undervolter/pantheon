package bot

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/textutil"
)

const filtersPerPage = 8

var snoozeHours = []int{1, 3, 8, 24}

// filterIcon is the one-glance state of a filter in lists.
func filterIcon(s domain.Search, now time.Time) string {
	icon := "🟢"
	switch {
	case s.Hidden():
		icon = "🚫"
	case s.Disabled:
		icon = "⏸"
	case s.Snoozed(now):
		icon = "😴"
	}
	if s.IsShared() {
		icon += "👥"
	}
	return icon
}

func (b *Bot) viewFilters(t *turn, page int) (string, *markup) {
	filters := b.visibleFilters(t)
	p := newPage(t.tr("fl.title"))
	if len(filters) == 0 {
		p.text(t.tr("fl.empty"))
		return p.String(), telegram.Markup(
			telegram.Row(telegram.Primary(t.tr("btn.new_filter"), "add_new")),
			telegram.Row(btnBack(t.lang, "menu_main")),
		)
	}

	now := b.now()
	active, hasShared := 0, false
	for _, s := range filters {
		if s.Active(now) {
			active++
		}
		hasShared = hasShared || s.IsShared()
	}
	p.text(t.tr("fl.counts", active, len(filters)-active))

	c := b.counters(t)
	start, end, page, pages := pageBounds(len(filters), filtersPerPage, page)
	var rows [][]string
	var buttons [][]button
	for _, s := range filters[start:end] {
		name := displayName(t.lang, s)
		icon := filterIcon(s, now)
		rows = append(rows, []string{
			icon, esc(textutil.TruncateRunes(name, 28)), priceCell(t.lang, s),
			strconv.Itoa(c.dayBy[s.ID]), strconv.Itoa(c.weekBy[s.ID]),
		})
		buttons = append(buttons, telegram.Row(telegram.Btn(textutil.TruncateRunes(icon+" "+name, 60), "f_view:"+s.ID)))
	}
	p.table([]string{"", t.tr("fl.col_filter"), t.tr("fl.col_price"), t.tr("fl.col_today"), t.tr("fl.col_week")}, rows)
	p.note(t.tr("fl.hint"))
	if hasShared {
		p.note(t.tr("fl.shared_hint"))
	}
	p.note(t.tr("fl.import_hint"))

	if nav := pagerRow(page, pages, func(n int) string { return fmt.Sprintf("menu_list:%d", n) }); nav != nil {
		buttons = append(buttons, nav)
	}
	buttons = append(buttons,
		telegram.Row(telegram.Primary(t.tr("btn.new_filter"), "add_new"), telegram.Btn(t.tr("btn.export"), "export")),
		telegram.Row(btnBack(t.lang, "menu_main")),
	)
	return p.String(), telegram.Markup(buttons...)
}

// findFilter returns a filter this chat may see.
func (b *Bot) findFilter(t *turn, id string) (domain.Search, bool) {
	s, err := b.db.Search(t.ctx, id)
	if err != nil {
		return domain.Search{}, false
	}
	if !s.IsShared() && s.OwnerChatID != t.chatID {
		return domain.Search{}, false
	}
	return s, true
}

// canManage reports whether the user may change a filter. A shared filter
// alerts every chat, so only an admin touches it.
func canManage(t *turn, s domain.Search) bool {
	if s.IsShared() {
		return t.admin
	}
	return s.OwnerChatID == t.chatID
}

func (b *Bot) viewFilter(t *turn, s domain.Search, notice string) (string, *markup) {
	now := b.now()
	loc := b.cfg.Location()

	status := t.tr("fd.status_active")
	switch {
	case s.Hidden():
		status = t.tr("fd.status_hidden")
	case s.Disabled:
		status = t.tr("fd.status_paused")
	case s.Snoozed(now):
		status = t.tr("fd.status_snoozed", time.Unix(s.SnoozedUntil, 0).In(loc).Format("02.01 15:04"))
	}
	owner := t.tr("fd.owner_private")
	if s.IsShared() {
		owner = t.tr("fd.owner_shared")
	}

	p := newPage("🎯 " + esc(displayName(t.lang, s))).text(status + " · " + owner)
	if s.Hidden() {
		p.quote(t.tr("fd.hidden_note"))
	}
	c := b.counters(t)
	p.stats(
		stat{t.tr("menu.kpi_today"), strconv.Itoa(c.dayBy[s.ID])},
		stat{t.tr("menu.kpi_week"), strconv.Itoa(c.weekBy[s.ID])},
	)
	p.section(t.tr("fd.section_params"))
	summary(p, t.lang, s)
	if s.CreatedAt > 0 {
		p.note(t.tr("fd.created", time.Unix(s.CreatedAt, 0).In(loc).Format("02.01.2006")))
	}

	var rows [][]button
	manage := canManage(t, s)
	if manage && !s.Hidden() {
		toggle := telegram.Btn(t.tr("fd.btn_pause"), "f_toggle:"+s.ID)
		if s.Disabled {
			toggle = telegram.Success(t.tr("fd.btn_resume"), "f_toggle:"+s.ID)
		}
		snooze := telegram.Btn(t.tr("fd.btn_snooze"), "f_snoozemenu:"+s.ID)
		if s.Snoozed(now) {
			snooze = telegram.Success(t.tr("fd.btn_wake"), "f_snooze:"+s.ID+":0")
		}
		rows = append(rows, telegram.Row(toggle, snooze))
		if t.admin {
			if s.IsShared() {
				rows = append(rows, telegram.Row(telegram.Btn(t.tr("fd.btn_claim"), "f_claim:"+s.ID)))
			} else {
				rows = append(rows, telegram.Row(telegram.Btn(t.tr("fd.btn_share"), "f_share:"+s.ID)))
			}
		}
	}
	last := telegram.Row(btnBack(t.lang, "menu_list"))
	if manage {
		last = append(last, telegram.Danger(t.tr("fd.btn_delete"), "f_del:"+s.ID))
	}
	rows = append(rows, last)
	return withNotice(notice, p.String()), telegram.Markup(rows...)
}

func (b *Bot) viewSnooze(t *turn, s domain.Search) (string, *markup) {
	var buttons []button
	for _, h := range snoozeHours {
		buttons = append(buttons, telegram.Btn(t.tr("fd.snooze_hours", h), fmt.Sprintf("f_snooze:%s:%d", s.ID, h)))
	}
	rows := grid(buttons, 2)
	rows = append(rows, telegram.Row(btnBack(t.lang, "f_view:"+s.ID)))
	return prompt(t.tr("fd.snooze_prompt", esc(displayName(t.lang, s)))), telegram.Markup(rows...)
}

func (b *Bot) viewDeleteConfirm(t *turn, s domain.Search) (string, *markup) {
	km := telegram.Markup(telegram.Row(
		telegram.Btn(t.tr("btn.cancel"), "f_view:"+s.ID),
		telegram.Danger(t.tr("fd.btn_delete_yes"), "f_delok:"+s.ID),
	))
	return prompt(t.tr("fd.delete_prompt", esc(displayName(t.lang, s)))), km
}

// filterCallback handles every f_* button. The filter is looked up by its id
// on each press: a button on an old message can only ever act on the filter
// it was drawn for, or on nothing if that filter is gone.
func (b *Bot) filterCallback(t *turn, name string, args []string) {
	if len(args) == 0 {
		t.toast(t.tr("common.stale"))
		t.show(b.viewFilters(t, 0))
		return
	}
	s, ok := b.findFilter(t, args[0])
	if !ok {
		t.alert(t.tr("fd.not_found"))
		t.show(b.viewFilters(t, 0))
		return
	}
	if name == "f_view" {
		t.show(b.viewFilter(t, s, ""))
		return
	}
	if !canManage(t, s) {
		t.alert(t.tr("common.admin_only"))
		return
	}

	update := func(fn func(*domain.Search)) (domain.Search, bool) {
		out, err := b.db.UpdateSearch(t.ctx, s.ID, func(sc *domain.Search) error { fn(sc); return nil })
		if errors.Is(err, store.ErrNotFound) {
			t.alert(t.tr("fd.not_found"))
			t.show(b.viewFilters(t, 0))
			return domain.Search{}, false
		}
		if err != nil {
			t.saveFailed(err)
			return domain.Search{}, false
		}
		return out, true
	}
	now := b.now()

	switch name {
	case "f_toggle":
		if s.Hidden() {
			t.alert(t.tr("fd.hidden_note"))
			return
		}
		out, ok := update(func(sc *domain.Search) {
			sc.Disabled = !sc.Disabled
			if !sc.Disabled {
				// The age floor moves to now: a resumed filter must not pour
				// out everything that was posted while it was paused.
				sc.EnabledAt = now.Unix()
				sc.SnoozedUntil = 0
			}
		})
		if !ok {
			return
		}
		if out.Disabled {
			t.toast(t.tr("fd.toast_paused"))
		} else {
			t.toast(t.tr("fd.toast_resumed"))
		}
		t.show(b.viewFilter(t, out, ""))

	case "f_snoozemenu":
		t.show(b.viewSnooze(t, s))

	case "f_snooze":
		hours := 0
		if len(args) > 1 {
			hours, _ = strconv.Atoi(args[1])
		}
		if hours < 0 || hours > 24*7 {
			hours = 0
		}
		out, ok := update(func(sc *domain.Search) {
			if hours == 0 {
				sc.SnoozedUntil = 0
				sc.EnabledAt = now.Unix()
				return
			}
			sc.SnoozedUntil = now.Add(time.Duration(hours) * time.Hour).Unix()
		})
		if !ok {
			return
		}
		if hours == 0 {
			t.toast(t.tr("fd.toast_woken"))
		} else {
			t.toast(t.tr("fd.toast_snoozed", hours))
		}
		t.show(b.viewFilter(t, out, ""))

	case "f_claim", "f_share":
		if !t.admin {
			t.alert(t.tr("common.admin_only"))
			return
		}
		out, ok := update(func(sc *domain.Search) {
			if name == "f_claim" {
				sc.OwnerChatID = t.chatID
			} else {
				sc.OwnerChatID = ""
			}
		})
		if !ok {
			return
		}
		if name == "f_claim" {
			t.toast(t.tr("fd.toast_claimed"))
		} else {
			t.toast(t.tr("fd.toast_shared"))
		}
		t.show(b.viewFilter(t, out, ""))

	case "f_del":
		t.show(b.viewDeleteConfirm(t, s))

	case "f_delok":
		if _, err := b.db.DeleteSearch(t.ctx, s.ID); err != nil {
			t.saveFailed(err)
			return
		}
		t.toast(t.tr("fd.toast_deleted"))
		t.show(b.viewFilters(t, 0))

	default:
		t.toast(t.tr("common.stale"))
		t.show(b.viewFilter(t, s, ""))
	}
}

// defaultFilterName picks a name for a filter the user did not name.
func defaultFilterName(lang i18n.Lang, s domain.Search) string {
	s.Name = ""
	return textutil.TruncateRunes(displayName(lang, s), domain.MaxNameRunes)
}

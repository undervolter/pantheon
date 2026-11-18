package bot

import (
	"fmt"
	"strconv"
	"time"

	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/textutil"
	"pantheon/core/version"
)

const historySize = 10

// visibleFilters returns the filters this chat sees: its own and the shared
// ones.
func (b *Bot) visibleFilters(t *turn) []domain.Search {
	all, err := b.db.Searches(t.ctx)
	if err != nil {
		logx.Errorf("read filters: %v", err)
		return nil
	}
	out := all[:0:0]
	for _, s := range all {
		if s.IsShared() || s.OwnerChatID == t.chatID {
			out = append(out, s)
		}
	}
	return out
}

// counters are the delivered alerts of a chat: today and over the last week.
type counters struct {
	day, week     int
	dayBy, weekBy map[string]int
}

func (b *Bot) counters(t *turn) counters {
	now := b.now().In(b.cfg.Location())
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var c counters
	var err error
	if c.day, c.dayBy, err = b.db.AlertCounts(t.ctx, t.chatID, midnight); err != nil {
		logx.Errorf("read alert counters: %v", err)
	}
	if c.week, c.weekBy, err = b.db.AlertCounts(t.ctx, t.chatID, now.Add(-7*24*time.Hour)); err != nil {
		logx.Errorf("read alert counters: %v", err)
	}
	return c
}

func (b *Bot) scanLine(t *turn, activeFilters int) string {
	status := ScanStatus{}
	if b.scan != nil {
		status = b.scan.Status()
	}
	switch {
	case status.BlockedUntil.After(b.now()):
		return t.tr("scan.blocked", status.BlockedUntil.In(b.cfg.Location()).Format("15:04"))
	case activeFilters == 0:
		return t.tr("scan.idle")
	case !status.Started:
		return t.tr("scan.starting")
	}
	return t.tr("scan.running", b.cfg.GetInterval())
}

func (b *Bot) viewMain(t *turn) (string, *markup) {
	filters := b.visibleFilters(t)
	now := b.now()
	active := 0
	for _, s := range filters {
		if s.Active(now) {
			active++
		}
	}

	p := newPage("🐉 Pantheon").text(b.scanLine(t, active))
	paused := b.cfg.IsPaused(t.chatID)
	if paused {
		p.quote(t.tr("menu.paused"))
	}

	if len(filters) == 0 {
		p.text(t.tr("menu.empty"))
		p.list([]string{t.tr("menu.empty_new"), t.tr("menu.empty_link"), t.tr("menu.empty_file")})
	} else {
		c := b.counters(t)
		shown := strconv.Itoa(active)
		if active != len(filters) {
			shown = fmt.Sprintf("%d / %d", active, len(filters))
		}
		p.stats(
			stat{t.tr("menu.kpi_filters"), shown},
			stat{t.tr("menu.kpi_today"), strconv.Itoa(c.day)},
			stat{t.tr("menu.kpi_week"), strconv.Itoa(c.week)},
		)
	}

	footer := t.tr("menu.version", version.AppVersion)
	if last, err := b.db.LastSentAt(t.ctx, t.chatID); err == nil && !last.IsZero() {
		footer += " · " + t.tr("menu.last_alert", ago(t.lang, now.Sub(last)))
	}
	p.note(footer)

	pauseBtn := telegram.Btn(t.tr("btn.pause"), "chat_pause")
	if paused {
		pauseBtn = telegram.Success(t.tr("btn.resume"), "chat_resume")
	}
	km := telegram.Markup(
		telegram.Row(telegram.Primary(t.tr("btn.new_filter"), "add_new"), telegram.Btn(t.tr("btn.filters", len(filters)), "menu_list")),
		telegram.Row(telegram.Btn(t.tr("btn.history"), "history_show"), telegram.Btn(t.tr("btn.settings"), "menu_settings")),
		telegram.Row(pauseBtn, telegram.Btn(t.tr("btn.lang"), "lang_menu"), telegram.Btn(t.tr("btn.help"), "help_show")),
	)
	return p.String(), km
}

func (b *Bot) viewHistory(t *turn) (string, *markup) {
	p := newPage(t.tr("hist.title"))
	alerts, err := b.db.RecentAlerts(t.ctx, t.chatID, historySize)
	if err != nil {
		logx.Errorf("read history: %v", err)
	}
	if len(alerts) == 0 {
		p.text(t.tr("hist.empty"))
	} else {
		loc := b.cfg.Location()
		today := b.now().In(loc).Format("02.01")
		rows := make([][]string, 0, len(alerts))
		for _, a := range alerts {
			rows = append(rows, []string{historyTime(a, loc, today), historyTitle(a), esc(a.Price)})
		}
		p.table([]string{t.tr("hist.col_time"), t.tr("hist.col_listing"), t.tr("hist.col_price")}, rows)
		p.note(t.tr("hist.note"))
	}
	km := telegram.Markup(telegram.Row(btnBack(t.lang, "menu_main"), telegram.Btn(t.tr("btn.refresh"), "history_show")))
	return p.String(), km
}

func historyTime(a store.Alert, loc *time.Location, today string) string {
	at := a.SentAt.In(loc)
	if at.Format("02.01") == today {
		return at.Format("15:04")
	}
	return at.Format("02.01 15:04")
}

// historyTitle links the listing and names the filter that caught it. Both
// the title (from a seller) and the filter name (from a user) are escaped.
func historyTitle(a store.Alert) string {
	title := textutil.TruncateRunes(textutil.SanitizeListingText(a.Title), 60)
	if title == "" {
		title = "OLX #" + strconv.FormatInt(a.OfferID, 10)
	}
	cell := esc(title)
	if isWebURL(a.URL) {
		cell = `<a href="` + esc(a.URL) + `">` + cell + `</a>`
	}
	if a.SearchName != "" {
		cell += " · <i>" + esc(textutil.TruncateRunes(a.SearchName, 28)) + "</i>"
	}
	return cell
}

func isWebURL(u string) bool {
	return len(u) > 8 && (u[:8] == "https://" || u[:7] == "http://")
}

func (b *Bot) viewLang(t *turn) (string, *markup) {
	p := newPage(t.tr("lang.title")).text(t.tr("lang.intro"))
	var rows [][]button
	for _, l := range i18n.All {
		label := l.Flag() + "  " + l.Name()
		btn := telegram.Btn(label, "lang:"+string(l))
		if l == t.lang {
			btn = telegram.Success("✓  "+l.Name(), "lang:"+string(l))
		}
		rows = append(rows, telegram.Row(btn))
	}
	rows = append(rows, telegram.Row(btnBack(t.lang, "menu_main")))
	return p.String(), telegram.Markup(rows...)
}

func (b *Bot) viewHelp(t *turn) (string, *markup) {
	p := newPage(t.tr("help.title")).text(t.tr("help.intro"))

	p.section(t.tr("help.sec_commands"))
	var rows [][]string
	for _, c := range menuCommands {
		rows = append(rows, []string{"/" + c, t.tr("bc." + c)})
	}
	rows = append(rows,
		[]string{"/cancel", t.tr("bc.cancel")},
		[]string{"/check", t.tr("help.cmd_check")},
		[]string{"/chatid", t.tr("help.cmd_chatid")},
		[]string{"/version", t.tr("help.cmd_version")},
	)
	p.table(nil, rows)

	if t.admin {
		p.section(t.tr("help.sec_chats"))
		p.table(nil, [][]string{
			{"/chats", t.tr("help.cmd_chats")},
			{"/bind", t.tr("help.cmd_bind")},
			{"/unbind", t.tr("help.cmd_unbind")},
			{"/addchat &lt;id&gt;", t.tr("help.cmd_addchat")},
			{"/delchat &lt;id&gt;", t.tr("help.cmd_delchat")},
		})
	}
	p.note(t.tr("help.note"))
	return p.String(), telegram.Markup(telegram.Row(btnBack(t.lang, "menu_main")))
}

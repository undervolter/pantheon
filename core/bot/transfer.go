package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pantheon/core/domain"
	"pantheon/core/telegram"
)

// Filters travel between chats and installations as a JSON file:
//
//	{"version": 1, "exported_at": "...", "filters": [ {filter}, ... ]}
//
// A bare array of filters is accepted too. The format is shared with the
// closed edition; fields one side does not know are ignored by the other.

const (
	exportVersion    = 1
	maxImportBytes   = 1 << 20
	maxImportFilters = 200
)

var errImportEmpty = errors.New("no filters in the file")

type filterExport struct {
	Version    int             `json:"version"`
	ExportedAt time.Time       `json:"exported_at"`
	Filters    []domain.Search `json:"filters"`
}

// exportFilters serialises what this chat sees, without ids, owners and
// runtime marks: the file describes rules, not this installation's state.
func (b *Bot) exportFilters(t *turn) ([]byte, int, error) {
	list := b.visibleFilters(t)
	out := make([]domain.Search, 0, len(list))
	for _, s := range list {
		c := s.Clone()
		c.ID, c.OwnerChatID, c.SnoozedUntil, c.EnabledAt = "", "", 0, 0
		out = append(out, c)
	}
	data, err := json.MarshalIndent(filterExport{Version: exportVersion, ExportedAt: b.now().UTC(), Filters: out}, "", "  ")
	return data, len(out), err
}

// parseImport reads an export file. Every filter is normalised and
// validated the same way a filter from the wizard is: a file is user input,
// and its keys end up in requests to OLX. Filters that do not pass are
// counted, not imported.
func parseImport(data []byte) (valid []domain.Search, skipped int, err error) {
	var list []domain.Search
	var wrapped filterExport
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Filters != nil {
		list = wrapped.Filters
	} else if err := json.Unmarshal(data, &list); err != nil {
		return nil, 0, fmt.Errorf("not a Pantheon filter export: %w", err)
	}
	if len(list) == 0 {
		return nil, 0, errImportEmpty
	}
	if len(list) > maxImportFilters {
		return nil, 0, fmt.Errorf("too many filters (%d, at most %d)", len(list), maxImportFilters)
	}
	for _, s := range list {
		s.ID, s.OwnerChatID, s.SnoozedUntil, s.EnabledAt, s.CreatedAt = "", "", 0, 0, 0
		s.Normalize()
		if s.Validate() != nil {
			skipped++
			continue
		}
		valid = append(valid, s)
	}
	return valid, skipped, nil
}

// splitDuplicates drops the filters this chat already has.
func (b *Bot) splitDuplicates(t *turn, list []domain.Search) (fresh []domain.Search, dupes int) {
	seen := map[string]bool{}
	for _, s := range b.visibleFilters(t) {
		seen[s.RuleKey()] = true
	}
	for _, s := range list {
		key := s.RuleKey()
		if seen[key] {
			dupes++
			continue
		}
		seen[key] = true
		fresh = append(fresh, s)
	}
	return fresh, dupes
}

func (b *Bot) viewImportConfirm(t *turn, fresh []domain.Search, dupes, skipped int) (string, *markup) {
	p := newPage(t.tr("xfer.title")).stats(
		stat{t.tr("xfer.kpi_new"), strconv.Itoa(len(fresh))},
		stat{t.tr("xfer.kpi_dupes"), strconv.Itoa(dupes)},
		stat{t.tr("xfer.kpi_skipped"), strconv.Itoa(skipped)},
	)
	var names []string
	for i, s := range fresh {
		if i >= 10 {
			names = append(names, t.tr("xfer.more", len(fresh)-i))
			break
		}
		names = append(names, esc(displayName(t.lang, s)))
	}
	p.list(names)
	if dupes > 0 {
		p.note(t.tr("xfer.dupes"))
	}
	if skipped > 0 {
		p.note(t.tr("xfer.skipped"))
	}
	km := telegram.Markup(telegram.Row(
		telegram.Btn(t.tr("btn.cancel"), "import_cancel"),
		telegram.Success(t.tr("xfer.btn_import", len(fresh)), "import_ok"),
	))
	return p.String(), km
}

func isJSONDocument(d *telegram.TGDocument) bool {
	return d != nil && (strings.HasSuffix(strings.ToLower(d.FileName), ".json") || d.MimeType == "application/json")
}

func (b *Bot) exportTo(t *turn) {
	data, n, err := b.exportFilters(t)
	if err != nil {
		t.say("xfer.bad", esc(err.Error()))
		return
	}
	if n == 0 {
		t.say("xfer.nothing")
		return
	}
	name := "pantheon-filters-" + b.now().Format("2006-01-02") + ".json"
	if err := b.tg.SendDocument(t.ctx, t.chatID, name, data, t.tr("xfer.exported", n)); err != nil {
		t.say("xfer.bad", esc(err.Error()))
	}
}

func (b *Bot) onDocument(t *turn, doc *telegram.TGDocument) {
	if !isJSONDocument(doc) {
		return
	}
	if doc.FileSize > maxImportBytes {
		t.say("xfer.too_large")
		return
	}
	data, err := b.tg.DownloadFile(t.ctx, doc.FileID, maxImportBytes)
	if err != nil {
		t.say("xfer.bad", esc(err.Error()))
		return
	}
	valid, skipped, err := parseImport(data)
	if err != nil {
		t.say("xfer.bad", esc(err.Error()))
		return
	}
	if len(valid) == 0 {
		t.say("xfer.none_valid")
		return
	}
	fresh, dupes := b.splitDuplicates(t, valid)
	if len(fresh) == 0 {
		t.say("xfer.all_dupes", dupes)
		return
	}
	t.resetState()
	t.st.Import, t.st.ImportSkipped = fresh, skipped
	t.send(b.viewImportConfirm(t, fresh, dupes, skipped))
}

func (b *Bot) confirmImport(t *turn) {
	pending := t.st.Import
	t.st.Import = nil
	if len(pending) == 0 {
		t.toast(t.tr("common.stale"))
		t.show(b.viewMain(t))
		return
	}
	// Checked again: filters may have been added since the file was sent.
	fresh, _ := b.splitDuplicates(t, pending)
	added := 0
	for _, s := range fresh {
		if limit, over := b.overLimit(t, 1); over {
			t.alert(t.tr("wz.err_limit", limit))
			break
		}
		s.OwnerChatID = t.chatID
		s.Name = defaultFilterNameKeep(t, s)
		if _, err := b.db.AddSearch(t.ctx, s); err != nil {
			t.saveFailed(err)
			break
		}
		added++
	}
	if added > 0 {
		t.toast(t.tr("xfer.imported", added))
		if b.scan != nil {
			b.scan.TriggerCheck()
		}
	}
	t.show(b.viewFilters(t, 0))
}

// defaultFilterNameKeep keeps the name a filter came with and invents one
// only when it has none.
func defaultFilterNameKeep(t *turn, s domain.Search) string {
	if strings.TrimSpace(s.Name) != "" {
		return s.Name
	}
	return defaultFilterName(t.lang, s)
}

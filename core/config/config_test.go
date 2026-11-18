package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"pantheon/core/domain"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func load(t *testing.T, body string) *Config {
	t.Helper()
	cfg, err := LoadConfig(writeConfig(t, body))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg := load(t, `{"interval_seconds":5}`)
	if !cfg.SendPhotos || cfg.MaxPhotos != 3 {
		t.Errorf("photos default = %v/%d, want true/3", cfg.SendPhotos, cfg.MaxPhotos)
	}
	if !cfg.GetANSIHyperlinks() {
		t.Error("ansi_hyperlinks must default to true")
	}
	if !cfg.GetBackfillEnabled() {
		t.Error("backfill must default to true")
	}
	if cfg.GetMaxAlertAgeMin() != 30 {
		t.Errorf("max_alert_age_min default = %d, want 30", cfg.GetMaxAlertAgeMin())
	}

	cfg = load(t, `{"send_photos":false,"max_photos":5,"ansi_hyperlinks":false}`)
	if cfg.SendPhotos || cfg.MaxPhotos != 5 || cfg.GetANSIHyperlinks() {
		t.Errorf("explicit values must be respected: %v/%d/%v", cfg.SendPhotos, cfg.MaxPhotos, cfg.GetANSIHyperlinks())
	}
}

func TestLoadConfigClampsInterval(t *testing.T) {
	if got := load(t, `{"interval_seconds":1}`).GetInterval(); got != 3 {
		t.Errorf("interval = %d, want 3 (the floor)", got)
	}
}

func TestBackfillLegacyKeyIsFolded(t *testing.T) {
	cfg := load(t, `{"disable_backfill":true}`)
	if cfg.GetBackfillEnabled() {
		t.Error("disable_backfill=true must disable backfill")
	}
	if cfg.DisableBackfill || cfg.Backfill == nil || *cfg.Backfill {
		t.Errorf("legacy key must be folded into backfill=false, got disable=%v backfill=%v", cfg.DisableBackfill, cfg.Backfill)
	}

	cfg = load(t, `{"backfill":false}`)
	if cfg.GetBackfillEnabled() {
		t.Error("backfill=false must disable backfill")
	}
	if err := cfg.SetBackfillEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !cfg.GetBackfillEnabled() {
		t.Error("SetBackfillEnabled(true) failed")
	}
}

// Zero and false must survive a restart: with omitempty they were dropped
// from the file and turned back into the defaults on the next load.
func TestZeroSettingsSurviveReload(t *testing.T) {
	cfg := load(t, `{}`)
	if err := cfg.SetMaxAlertAgeMin(0); err != nil {
		t.Fatal(err)
	}
	cfg.mu.Lock()
	cfg.ANSIHyperlinks = false
	err := cfg.saveLocked()
	cfg.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	again, err := LoadConfig(cfg.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if again.GetMaxAlertAgeMin() != 0 {
		t.Errorf("max_alert_age_min = %d after reload, want 0 (no limit)", again.GetMaxAlertAgeMin())
	}
	if again.GetANSIHyperlinks() {
		t.Error("ansi_hyperlinks=false was lost on reload")
	}
}

func TestPlaceholderChatIDsAreDropped(t *testing.T) {
	cfg := load(t, `{"telegram_chat_ids":["your_chat_id","YOUR_CHAT_ID_HERE",""," 111 ","111","0","-100500"]}`)
	got := cfg.GetTelegramChatIDs()
	if len(got) != 2 || got[0] != "111" || got[1] != "-100500" {
		t.Fatalf("chat ids = %v, want [111 -100500]", got)
	}
}

func TestSaveIsAtomicAndPrivate(t *testing.T) {
	cfg := load(t, `{"telegram_bot_token":"1:x"}`)
	if err := cfg.SetInterval(7); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg.FilePath() + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temporary file must not be left behind: %v", err)
	}
	again, err := LoadConfig(cfg.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if again.GetInterval() != 7 {
		t.Errorf("interval = %d, want 7", again.GetInterval())
	}
}

func TestConcurrentSavesKeepValidJSON(t *testing.T) {
	cfg := load(t, `{"telegram_bot_token":"1:x"}`)
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 3 {
			case 0:
				_ = cfg.SetInterval(5 + i)
			case 1:
				_ = cfg.SetMaxPhotos(1 + i%10)
			default:
				_ = cfg.Save()
			}
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(cfg.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	var probe map[string]any
	if err := json.Unmarshal(data, &probe); err != nil {
		t.Fatalf("config is not valid JSON after concurrent saves: %v\n%s", err, data)
	}
}

func TestTokenFromEnvIsNotPersisted(t *testing.T) {
	t.Setenv(TokenEnv, "999:ENV-TOKEN")
	cfg := load(t, `{"telegram_bot_token":"1:file"}`)
	if cfg.Token() != "999:ENV-TOKEN" {
		t.Fatalf("Token() = %q, want the env value", cfg.Token())
	}
	if err := cfg.SetInterval(9); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(cfg.FilePath())
	if strings.Contains(string(data), "ENV-TOKEN") {
		t.Errorf("env token leaked into the file:\n%s", data)
	}
}

func TestValidateToken(t *testing.T) {
	if err := load(t, `{}`).Validate(); err == nil {
		t.Error("empty token must be rejected")
	}
	if err := load(t, `{"telegram_bot_token":"YOUR_BOT_TOKEN"}`).Validate(); err == nil {
		t.Error("placeholder token must be rejected")
	}
	if err := load(t, `{"telegram_bot_token":"123:AAxyz"}`).Validate(); err != nil {
		t.Errorf("valid token rejected: %v", err)
	}
}

func TestEnsureAdminIsPersistedOnce(t *testing.T) {
	cfg := load(t, `{"telegram_chat_ids":["-100500","111","222"]}`)
	ok, err := cfg.EnsureAdmin()
	if err != nil || !ok {
		t.Fatalf("EnsureAdmin = %v, %v", ok, err)
	}
	if !cfg.IsAdmin("111") || cfg.IsAdmin("222") || cfg.IsAdmin("-100500") {
		t.Fatalf("admins = %v, want only the first user 111", cfg.Admins())
	}

	// Admin rights are stored, not derived from the list order: removing a
	// chat in front of the admin must not move them to somebody else.
	if _, err := cfg.RemoveChatID("-100500"); err != nil {
		t.Fatal(err)
	}
	again, err := LoadConfig(cfg.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if !again.IsAdmin("111") || again.IsAdmin("222") {
		t.Fatalf("admins after reload = %v", again.Admins())
	}
}

func TestEnsureAdminNeedsAUser(t *testing.T) {
	cfg := load(t, `{"telegram_chat_ids":["-100500"]}`)
	if ok, _ := cfg.EnsureAdmin(); ok {
		t.Fatal("a group cannot become the admin")
	}
	if bound, err := cfg.BindFirstAdmin("-100500"); err == nil || bound {
		t.Fatal("pairing must refuse a group id")
	}
	bound, err := cfg.BindFirstAdmin("777")
	if err != nil || !bound {
		t.Fatalf("BindFirstAdmin = %v, %v", bound, err)
	}
	if !cfg.IsAdmin("777") || !cfg.IsTrusted("777") {
		t.Fatal("the paired user must be trusted and admin")
	}
	if bound, _ := cfg.BindFirstAdmin("888"); bound {
		t.Fatal("pairing must work only while there is no admin")
	}
}

func TestLastAdminCannotBeRemoved(t *testing.T) {
	cfg := load(t, `{"telegram_chat_ids":["111","222"],"admin_chat_ids":["111"]}`)
	if _, err := cfg.RemoveChatID("111"); !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("err = %v, want ErrLastAdmin", err)
	}
	if !cfg.IsTrusted("111") {
		t.Fatal("the admin must still be trusted after the refused removal")
	}
}

func TestRemoveChatDropsItsState(t *testing.T) {
	cfg := load(t, `{"telegram_chat_ids":["111","222"],"admin_chat_ids":["111"]}`)
	_ = cfg.SetLang("222", "uk")
	_ = cfg.SetPaused("222", true)
	_ = cfg.SetQuietHours("222", 22, 8, true)

	removed, err := cfg.RemoveChatID("222")
	if err != nil || !removed {
		t.Fatalf("RemoveChatID = %v, %v", removed, err)
	}
	if cfg.IsTrusted("222") || cfg.IsPaused("222") || cfg.HasLang("222") {
		t.Error("removed chat left state behind")
	}
	if _, _, ok := cfg.GetQuietHours("222"); ok {
		t.Error("quiet hours of the removed chat must be gone")
	}
}

func TestQuietHoursAcrossMidnight(t *testing.T) {
	cfg := load(t, `{"timezone":"Europe/Kyiv"}`)
	if err := cfg.SetQuietHours("1", 22, 8, true); err != nil {
		t.Fatal(err)
	}
	kyiv := cfg.Location()
	at := func(h int) time.Time { return time.Date(2026, 11, 18, h, 30, 0, 0, kyiv) }

	for h, want := range map[int]bool{21: false, 22: true, 23: true, 0: true, 7: true, 8: false, 12: false} {
		if got := cfg.InQuietHours("1", at(h)); got != want {
			t.Errorf("InQuietHours at %02d:30 = %v, want %v", h, got, want)
		}
	}
	if cfg.InQuietHours("2", at(23)) {
		t.Error("a chat without quiet hours is never quiet")
	}

	// The same instant expressed in UTC must give the same answer: the bot's
	// host usually runs in UTC while the user lives in Kyiv time.
	if !cfg.InQuietHours("1", at(23).UTC()) {
		t.Error("quiet hours must be evaluated in the configured zone, not the host zone")
	}

	if err := cfg.SetQuietHours("1", 0, 0, false); err != nil {
		t.Fatal(err)
	}
	if cfg.InQuietHours("1", at(23)) {
		t.Error("quiet hours must be off after disabling")
	}
}

func TestParseQuietHours(t *testing.T) {
	if f, to, ok := ParseQuietHours("22-8"); !ok || f != 22 || to != 8 {
		t.Errorf("ParseQuietHours(22-8) = %d %d %v", f, to, ok)
	}
	for _, bad := range []string{"", "22", "25-8", "8-8", "a-b", "-1-5"} {
		if _, _, ok := ParseQuietHours(bad); ok {
			t.Errorf("ParseQuietHours(%q) must fail", bad)
		}
	}
}

func TestLanguagePerChat(t *testing.T) {
	cfg := load(t, `{"language":"uk"}`)
	if cfg.LangFor("1") != "uk" || cfg.HasLang("1") {
		t.Fatal("a chat without a choice gets the default language")
	}
	if err := cfg.SetLang("1", "en"); err != nil {
		t.Fatal(err)
	}
	again, _ := LoadConfig(cfg.FilePath())
	if again.LangFor("1") != "en" || !again.HasLang("1") || again.LangFor("2") != "uk" {
		t.Errorf("languages after reload: 1=%q 2=%q", again.LangFor("1"), again.LangFor("2"))
	}
}

func TestMarkAnnounced(t *testing.T) {
	cfg := load(t, `{}`)
	if fresh, _ := cfg.MarkAnnounced("0.2.0"); !fresh {
		t.Fatal("first announce must be fresh")
	}
	if fresh, _ := cfg.MarkAnnounced("0.2.0"); fresh {
		t.Fatal("the same version must not be announced twice")
	}
	if fresh, _ := cfg.MarkAnnounced("0.2.1"); !fresh {
		t.Fatal("a new version is announced again")
	}
}

func TestLimits(t *testing.T) {
	cfg := load(t, `{}`)
	if cfg.MaxFilters() != 50 || cfg.SeenRetention() != 7*24*time.Hour {
		t.Errorf("defaults: filters=%d retention=%s", cfg.MaxFilters(), cfg.SeenRetention())
	}
	cfg = load(t, `{"max_filters_per_chat":0,"seen_retention_days":-1}`)
	if cfg.MaxFilters() != 0 || cfg.SeenRetention() != 0 {
		t.Errorf("0 and -1 mean unlimited: filters=%d retention=%s", cfg.MaxFilters(), cfg.SeenRetention())
	}
}

func TestDataPathDefaultsToConfigDir(t *testing.T) {
	cfg := load(t, `{}`)
	want := filepath.Join(filepath.Dir(cfg.FilePath()), "pantheon.db")
	if got := cfg.DataPath("pantheon.db"); got != want {
		t.Errorf("DataPath = %q, want %q", got, want)
	}
	custom := t.TempDir()
	cfg.DataDir = custom
	if got := cfg.DataPath("pantheon.db"); got != filepath.Join(custom, "pantheon.db") {
		t.Errorf("DataPath with data_dir = %q", got)
	}
}

func TestDropLegacySearchesKeepsABackup(t *testing.T) {
	original := `{"telegram_bot_token":"1:x","searches":[{"name":"phones","query":"iphone"}]}`
	cfg := load(t, original)
	if got := cfg.LegacySearches(); len(got) != 1 || got[0].Name != "phones" {
		t.Fatalf("LegacySearches = %+v", got)
	}
	if err := cfg.DropLegacySearches(); err != nil {
		t.Fatal(err)
	}

	bak, err := os.ReadFile(cfg.FilePath() + ".bak")
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if string(bak) != original {
		t.Errorf("backup must be the untouched original, got:\n%s", bak)
	}
	now, _ := os.ReadFile(cfg.FilePath())
	if strings.Contains(string(now), "searches") {
		t.Errorf("searches must leave the config file:\n%s", now)
	}

	// A second drop must not overwrite the backup with the already migrated file.
	cfg.mu.Lock()
	cfg.Searches = []domain.Search{{Name: "later"}}
	cfg.mu.Unlock()
	if err := cfg.DropLegacySearches(); err != nil {
		t.Fatal(err)
	}
	bak2, _ := os.ReadFile(cfg.FilePath() + ".bak")
	if string(bak2) != original {
		t.Error("an existing backup must never be overwritten")
	}
}

func TestAutoAuthorizeIgnoresPlaceholders(t *testing.T) {
	cfg := &Config{TelegramChatIDs: []string{"your_chat_id", "YOUR_CHAT_ID_HERE", ""}}
	cfg.SetFilePath(filepath.Join(t.TempDir(), "config.json"))

	for _, id := range []string{"111", "222"} {
		added, err := cfg.AutoAuthorizeChatID(id, 2)
		if err != nil || !added {
			t.Fatalf("AutoAuthorizeChatID(%s) = %v, %v; placeholders must not take a slot", id, added, err)
		}
	}
	if added, err := cfg.AutoAuthorizeChatID("333", 2); err == nil || added {
		t.Fatalf("third user must be refused, got added=%v err=%v", added, err)
	}
}

func TestSetSearchOwnerPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{
		Searches: []domain.Search{{Name: "phones"}, {Name: "realty"}},
	}
	cfg.SetFilePath(path)

	if err := cfg.SetSearchOwner(1, "222"); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Searches[1].OwnerChatID != "222" || reloaded.Searches[0].OwnerChatID != "" {
		t.Fatalf("owners after reload: %q / %q", reloaded.Searches[0].OwnerChatID, reloaded.Searches[1].OwnerChatID)
	}
}

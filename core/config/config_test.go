package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pantheon/core/domain"
)

func TestLoadConfigPhotoDefaults(t *testing.T) {
	dir := t.TempDir()

	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	cfg, err := LoadConfig(write("default.json", `{"interval_seconds":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.SendPhotos {
		t.Error("send_photos must default to true")
	}
	if cfg.MaxPhotos != 3 {
		t.Errorf("MaxPhotos = %d, want 3", cfg.MaxPhotos)
	}

	cfg, err = LoadConfig(write("off.json", `{"send_photos":false,"max_photos":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SendPhotos {
		t.Error("explicit send_photos=false must be respected")
	}
	if cfg.MaxPhotos != 5 {
		t.Errorf("MaxPhotos = %d, want 5", cfg.MaxPhotos)
	}
}

func TestLoadConfigANSIHyperlinksDefaults(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	cfg, err := LoadConfig(write("default.json", `{"interval_seconds":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GetANSIHyperlinks() {
		t.Error("ansi_hyperlinks must default to true")
	}

	cfg, err = LoadConfig(write("off.json", `{"ansi_hyperlinks":false}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GetANSIHyperlinks() {
		t.Error("explicit ansi_hyperlinks=false must be respected")
	}
}

func TestLoadConfigBackfillDefaults(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	cfg, err := LoadConfig(write("default.json", `{"interval_seconds":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GetBackfillEnabled() {
		t.Error("backfill must default to true")
	}
	if cfg.GetDisableBackfill() {
		t.Error("disable_backfill must default to false")
	}

	// Test disable_backfill: true
	cfg, err = LoadConfig(write("off1.json", `{"disable_backfill":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GetBackfillEnabled() {
		t.Error("disable_backfill=true must disable backfill")
	}

	// Test backfill: false
	cfg, err = LoadConfig(write("off2.json", `{"backfill":false}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GetBackfillEnabled() {
		t.Error("backfill=false must disable backfill")
	}

	// Test setter
	if err := cfg.SetBackfillEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !cfg.GetBackfillEnabled() {
		t.Error("SetBackfillEnabled(true) failed")
	}
}

func TestAutoAuthorizeIgnoresPlaceholders(t *testing.T) {
	cfg := &Config{TelegramChatIDs: []string{"your_chat_id", "YOUR_CHAT_ID_HERE", ""}}
	cfg.SetFilePath(filepath.Join(t.TempDir(), "json"))

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

func TestSaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "json")
	cfg := &Config{TelegramBotToken: "t", IntervalSeconds: 10}
	cfg.SetFilePath(path)

	if err := cfg.SetInterval(7); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temporary file must not be left behind: %v", err)
	}
	reloaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.GetInterval() != 7 {
		t.Errorf("interval = %d, want 7", reloaded.GetInterval())
	}
}

func TestLoadConfigClampsInterval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "json")
	if err := os.WriteFile(path, []byte(`{"interval_seconds":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GetInterval() != 3 {
		t.Errorf("interval = %d, want 3 (the floor)", cfg.GetInterval())
	}
	if cfg.GetMaxAlertAgeMin() != 30 {
		t.Errorf("max_alert_age_min default = %d, want 30", cfg.GetMaxAlertAgeMin())
	}
}

func TestSetSearchOwnerPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "json")
	cfg := &Config{
		Searches: []domain.Search{{Name: "phones"}, {Name: "realty"}},
	}
	cfg.SetFilePath(path)

	if err := cfg.SetSearchOwner(1, "222"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"owner_chat_id": "222"`) {
		t.Fatalf("owner_chat_id not saved: %s", data)
	}

	reloaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Searches[1].OwnerChatID != "222" {
		t.Fatalf("owner_chat_id = %q, want 222", reloaded.Searches[1].OwnerChatID)
	}
	if reloaded.Searches[0].OwnerChatID != "" {
		t.Fatalf("untouched filter must stay shared, got %q", reloaded.Searches[0].OwnerChatID)
	}
}

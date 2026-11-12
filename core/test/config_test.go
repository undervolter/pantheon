package test

import (
	"os"
	"path/filepath"
	"testing"

	"pantheon/core/config"
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

	cfg, err := config.LoadConfig(write("default.json", `{"interval_seconds":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.SendPhotos {
		t.Error("send_photos must default to true")
	}
	if cfg.MaxPhotos != 3 {
		t.Errorf("MaxPhotos = %d, want 3", cfg.MaxPhotos)
	}

	cfg, err = config.LoadConfig(write("off.json", `{"send_photos":false,"max_photos":5}`))
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

	cfg, err := config.LoadConfig(write("default.json", `{"interval_seconds":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GetANSIHyperlinks() {
		t.Error("ansi_hyperlinks must default to true")
	}

	cfg, err = config.LoadConfig(write("off.json", `{"ansi_hyperlinks":false}`))
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

	cfg, err := config.LoadConfig(write("default.json", `{"interval_seconds":5}`))
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
	cfg, err = config.LoadConfig(write("off1.json", `{"disable_backfill":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GetBackfillEnabled() {
		t.Error("disable_backfill=true must disable backfill")
	}

	// Test backfill: false
	cfg, err = config.LoadConfig(write("off2.json", `{"backfill":false}`))
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

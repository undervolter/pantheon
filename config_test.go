package main

import (
	"os"
	"path/filepath"
	"testing"
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

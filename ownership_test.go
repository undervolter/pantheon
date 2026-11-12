package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterChatIDsForOwner(t *testing.T) {
	all := []string{"111", "222"}

	if got := filterChatIDsForOwner(all, ""); len(got) != 2 {
		t.Fatalf("shared filter must go to all chats, got %v", got)
	}

	if got := filterChatIDsForOwner(all, "222"); len(got) != 1 || got[0] != "222" {
		t.Fatalf("owned filter must go only to owner, got %v", got)
	}

	if got := filterChatIDsForOwner(all, "999"); got != nil {
		t.Fatalf("owner missing from recipients must disable delivery, got %v", got)
	}
}

func TestSeenCacheIsolatedByOwner(t *testing.T) {
	c := &Checker{seenIDs: map[string]bool{}}

	c.saveSeenID(10, "111")
	if !c.isSeen(10, "111") {
		t.Fatal("offer must be seen by its owner")
	}
	if c.isSeen(10, "222") {
		t.Fatal("offer must NOT be seen by another owner")
	}

	// Общий фильтр не должен блокировать персональные алерты.
	c.saveSeenID(11, "")
	if !c.isSeen(11, "") {
		t.Fatal("offer must be seen by shared filter")
	}
	if c.isSeen(11, "111") {
		t.Fatal("shared filter must not block personal owner")
	}

	// Старые записи (просто ID) считаются просмотренными всеми.
	c.seenIDs["12"] = true
	if !c.isSeen(12, "111") || !c.isSeen(12, "222") {
		t.Fatal("legacy seen entry must block all owners")
	}
}

func TestSearchForChatAccess(t *testing.T) {
	cfg := &Config{Searches: []SearchConfig{
		{Name: "my", OwnerChatID: "111"},
		{Name: "other", OwnerChatID: "222"},
		{Name: "shared"},
	}}
	ui := NewBotUI(cfg, nil)

	if _, ok := ui.searchForChat("111", 0); !ok {
		t.Error("owner must access own filter")
	}
	if _, ok := ui.searchForChat("111", 1); ok {
		t.Error("user must not access another user's filter")
	}
	if _, ok := ui.searchForChat("111", 2); !ok {
		t.Error("shared filter must be accessible to everyone")
	}
}

func TestRenderFiltersListHidesOtherUsers(t *testing.T) {
	cfg := &Config{Searches: []SearchConfig{
		{Name: "мой телефон", OwnerChatID: "111"},
		{Name: "чужая недвижка", OwnerChatID: "222"},
		{Name: "общий фильтр"},
	}}
	ui := NewBotUI(cfg, nil)

	_, km := ui.RenderFiltersList("111")
	seen := map[string]bool{}
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			seen[b.CallbackData] = true
		}
	}

	if !seen["f_view:0"] {
		t.Error("own filter must be visible")
	}
	if seen["f_view:1"] {
		t.Error("another user's filter must be hidden")
	}
	if !seen["f_view:2"] {
		t.Error("shared filter must be visible to everyone")
	}
}

func TestSetSearchOwnerPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg := &Config{
		Searches: []SearchConfig{{Name: "phones"}, {Name: "realty"}},
		filePath: path,
	}

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

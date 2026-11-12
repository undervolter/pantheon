package test

import (
	"strings"
	"testing"
	"time"

	"pantheon/core/checker"
	"pantheon/core/config"
	"pantheon/core/olx"
)

func TestCheckerFormatAdFound(t *testing.T) {
	cfg := &config.Config{
		ANSIHyperlinks: true,
	}
	chk := checker.NewChecker(cfg, nil, nil)
	defer chk.Close()

	now := time.Date(2026, 11, 12, 5, 58, 27, 0, time.UTC)
	offer := olx.Offer{
		ID:              940639374,
		Title:           "продаю анти кваріант",
		URL:             "https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html",
		CreatedTime:     "2026-11-12T05:40:28+02:00",
		LastRefreshTime: "2026-11-12T05:43:51+02:00",
	}

	line := chk.FormatAdFound(now, offer, "Електроніка")

	// 1. Title must NOT be present in log output
	if strings.Contains(line, "продаю анти кваріант") {
		t.Errorf("ad title should not be in log line: %s", line)
	}

	// 2. Phone label must NOT be present in log output
	if strings.Contains(line, "phone:") {
		t.Errorf("phone label should not be in log line: %s", line)
	}

	// 3. Filter name must be present in brackets
	if !strings.Contains(line, "[Електроніка]") {
		t.Errorf("expected [Електроніка] in log line: %s", line)
	}

	// 4. Hyperlink must contain OSC 8 sequence with label 'link'
	if !strings.Contains(line, "\x1b]8;;https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html\x1b\\link\x1b]8;;\x1b\\") {
		t.Errorf("expected OSC 8 hyperlink in log line: %s", line)
	}

	// Now test with ANSIHyperlinks disabled
	cfgOff := &config.Config{
		ANSIHyperlinks: false,
	}
	chkOff := checker.NewChecker(cfgOff, nil, nil)
	defer chkOff.Close()

	lineOff := chkOff.FormatAdFound(now, offer, "Електроніка")
	if strings.Contains(lineOff, "\x1b]8;;") {
		t.Errorf("did not expect OSC 8 hyperlink when ansi_hyperlinks=false: %s", lineOff)
	}
	if !strings.Contains(lineOff, "https://www.olx.ua/d/uk/obyavlenie/prodayu-anti-kvarant-ID11EPgO.html") {
		t.Errorf("expected raw URL when ansi_hyperlinks=false: %s", lineOff)
	}
	if strings.Contains(lineOff, "продаю анти кваріант") {
		t.Errorf("ad title should not be in log line: %s", lineOff)
	}
	if strings.Contains(lineOff, "phone:") {
		t.Errorf("phone label should not be in log line when ansi_hyperlinks=false: %s", lineOff)
	}
}

func TestCheckerFormatStartMessage(t *testing.T) {
	cfg := &config.Config{
		IntervalSeconds: 15,
		Searches: []config.SearchConfig{
			{Name: "Електроніка", Query: ""},
			{Name: "Телефони", Query: "iphone", Disabled: true},
		},
	}
	chk := checker.NewChecker(cfg, nil, nil)
	defer chk.Close()

	msg := chk.FormatStartMessage()
	// Started pantheon v0.x.x at HH:MM:SS DD:MM:YYYY, active filters: 1, interval - 15s
	if !strings.HasPrefix(msg, "Started pantheon v") {
		t.Errorf("expected 'Started pantheon v...' prefix, got: %s", msg)
	}
	if !strings.Contains(msg, ", active filters: 1, interval - 15s") {
		t.Errorf("expected active filters and interval in start message, got: %s", msg)
	}
}

func TestCheckerBanner(t *testing.T) {
	if !strings.Contains(checker.Banner, "made by undervolter | github.com/undervolter/pantheon") {
		t.Error("banner must contain author and github repo link")
	}
	if !strings.Contains(checker.Banner, "Want some updates? t.me/pantheonUpdates") {
		t.Error("banner must contain telegram channel link")
	}
}

package checker

import (
	"fmt"
	"os"
	"strings"
	"time"

	"pantheon/core/domain"
	"pantheon/core/textutil"
	"pantheon/core/version"
)

const Banner = `                                __     __
    ____     ____ _   ____     / /_   / /_     ___     ____     ____
   / __ \   / __ ` + "`" + `/  / __ \   / __/  / __ \   / _ \   / __ \   / __ \
  / /_/ /  / /_/ /  / / / /  / /_   / / / /  /  __/  / /_/ /  / / / /
 / .___/   \__,_/  /_/ /_/   \__/  /_/ /_/   \___/   \____/  /_/ /_/
/_/

  made by undervolter | github.com/undervolter/pantheon
      Want some updates? t.me/pantheonUpdates`

// StartMessage is the first line after the banner.
func StartMessage(startedAt time.Time, activeFilters, intervalSec int) string {
	return fmt.Sprintf("Started pantheon v%s at %s, active filters: %d, interval - %ds",
		version.AppVersion, startedAt.Format("15:04:05 02.01.2006"), activeFilters, intervalSec)
}

// formatAdFound is the log line of a new listing. The title is left out on
// purpose: the log is for the operator, the card is for the reader.
func formatAdFound(o *domain.Offer, filterName string, ansiLinks bool) string {
	line := fmt.Sprintf("found new ad! id %d | created at: %s | updated at: %s | price: %s | [%s]",
		o.ID,
		secondsLabel(o.ParsedCreatedTime()),
		secondsLabel(o.ParsedRefreshTime()),
		textutil.SanitizeListingText(o.GetPrice()),
		textutil.SanitizeListingText(filterName),
	)
	if o.URL != "" {
		line += " | " + hyperlinkLabel(o.URL, "link", ansiLinks)
	}
	return line
}

func formatAdSkipped(o *domain.Offer, v Verdict, detail string) string {
	line := fmt.Sprintf("skipped ad %d, reason: %s", o.ID, v.Reason())
	if detail != "" {
		line += " (" + textutil.SanitizeListingText(detail) + ")"
	}
	return line
}

func secondsLabel(t time.Time, err error) string {
	if err != nil || t.IsZero() {
		return "unknown"
	}
	return t.Format("02.01.2006 15:04:05")
}

// hyperlinkLabel prints a URL as a clickable label in terminals that support
// OSC 8, and as the bare URL elsewhere.
func hyperlinkLabel(url, label string, enabled bool) string {
	if url == "" {
		return ""
	}
	if !enabled || !hyperlinksAllowed() {
		return url
	}
	return "\x1b]8;;" + sanitizeOSCField(url) + "\x1b\\" + sanitizeOSCField(label) + "\x1b]8;;\x1b\\"
}

// sanitizeOSCField keeps a listing URL from closing the escape sequence
// early and writing its own to the terminal.
func sanitizeOSCField(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x1b, 0x07, '\n', '\r':
			return -1
		}
		return r
	}, s)
}

func hyperlinksAllowed() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PANTHEON_ANSI_LINKS"))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return !strings.EqualFold(os.Getenv("TERM"), "dumb")
}

package logx

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func capture(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	t.Cleanup(SetOutput(&buf))
	return &buf
}

func TestSecretNeverReachesTheLog(t *testing.T) {
	const token = "123456789:AAFAKE-TOKEN-FOR-TESTS"
	AddSecret(token)
	buf := capture(t)

	// The shape net/http gives a timed-out Bot API call.
	err := &url.Error{Op: "Post", URL: "https://api.telegram.org/bot" + token + "/sendMessage", Err: errors.New("context deadline exceeded")}
	Warnf("rich send failed: %v", err)
	Errorf("plain %s", token)
	Infof(token)
	Print("banner " + token)

	if got := buf.String(); strings.Contains(got, token) {
		t.Fatalf("token leaked into the log:\n%s", got)
	}
	if !strings.Contains(buf.String(), redacted) {
		t.Fatalf("expected the placeholder in the log:\n%s", buf.String())
	}
}

func TestRedactErrKeepsTheChain(t *testing.T) {
	const token = "987654321:AASECOND-TOKEN-FOR-TESTS"
	AddSecret(token)
	sentinel := errors.New("boom")

	wrapped := RedactErr(fmt.Errorf("GET /bot%s/x: %w", token, sentinel))
	if strings.Contains(wrapped.Error(), token) {
		t.Fatalf("token left in error text: %v", wrapped)
	}
	if !errors.Is(wrapped, sentinel) {
		t.Fatal("redaction must keep the error chain")
	}

	plain := errors.New("nothing secret here")
	if RedactErr(plain) != plain {
		t.Error("an error without secrets must be returned unchanged")
	}
	if RedactErr(nil) != nil {
		t.Error("nil must stay nil")
	}
}

func TestShortValuesAreNotSecrets(t *testing.T) {
	AddSecret("abc")
	AddSecret("")
	if got := Redact("abc and nothing"); got != "abc and nothing" {
		t.Errorf("short value was treated as a secret: %q", got)
	}
}

func TestLineFormat(t *testing.T) {
	buf := capture(t)
	Infof("started %d filters", 3)
	line := buf.String()
	if !strings.HasPrefix(line, "[") || !strings.Contains(line, "] started 3 filters\n") {
		t.Errorf("unexpected line format: %q", line)
	}
}

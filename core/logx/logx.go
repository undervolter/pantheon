// Package logx is the one place the process writes its console log from.
//
// Every line goes through Redact, so a secret registered with AddSecret (the
// bot token) cannot reach the terminal or journald no matter which call site
// formatted it — a failed HTTP call carries the full request URL in its error.
package logx

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const redacted = "<redacted>"

// minSecretLen keeps a short or empty value from blanking out ordinary text.
const minSecretLen = 8

var (
	mu      sync.Mutex
	out     io.Writer = os.Stdout
	secrets []string
	clock   = time.Now
)

// SetOutput redirects the log and returns a function that restores the
// previous destination. Tests use it to capture lines.
func SetOutput(w io.Writer) (restore func()) {
	mu.Lock()
	prev := out
	out = w
	mu.Unlock()
	return func() {
		mu.Lock()
		out = prev
		mu.Unlock()
	}
}

// AddSecret registers a value that must never appear in the log.
func AddSecret(s string) {
	s = strings.TrimSpace(s)
	if len(s) < minSecretLen {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	for _, have := range secrets {
		if have == s {
			return
		}
	}
	secrets = append(secrets, s)
}

// Redact replaces every registered secret in s.
func Redact(s string) string {
	mu.Lock()
	list := secrets
	mu.Unlock()
	for _, secret := range list {
		if strings.Contains(s, secret) {
			s = strings.ReplaceAll(s, secret, redacted)
		}
	}
	return s
}

// RedactErr returns err with every registered secret removed from its text.
// The original error is kept when there is nothing to hide, so errors.Is and
// errors.As keep working in the common case.
func RedactErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	clean := Redact(msg)
	if clean == msg {
		return err
	}
	return redactedError{msg: clean, cause: err}
}

type redactedError struct {
	msg   string
	cause error
}

func (e redactedError) Error() string { return e.msg }
func (e redactedError) Unwrap() error { return e.cause }

func emit(mark, format string, args ...any) {
	line := format
	if len(args) > 0 {
		line = fmt.Sprintf(format, args...)
	}
	line = Redact(line)
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(out, "[%s] %s%s\n", clock().Format("15:04:05"), mark, line)
}

// Infof logs an ordinary event.
func Infof(format string, args ...any) { emit("", format, args...) }

// Warnf logs something the operator should look at but that did not stop work.
func Warnf(format string, args ...any) { emit("⚠️ ", format, args...) }

// Errorf logs a failed operation.
func Errorf(format string, args ...any) { emit("❌ ", format, args...) }

// Print writes text as is, without a timestamp (the start-up banner).
func Print(text string) {
	text = Redact(text)
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintln(out, text)
}

// Package store keeps the bot's state in SQLite: filters, the per-filter
// "seen" marks, and the alert queue that doubles as the delivery journal.
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go driver: no cgo, so plain cross-compilation keeps working
)

//go:embed schema.sql
var schemaSQL string

// schemaVersion is stored in PRAGMA user_version. Bump it when a release
// changes the schema in a way an older binary cannot read.
const schemaVersion = 1

// ErrNewerSchema means the database was written by a newer release.
var ErrNewerSchema = errors.New("the database was created by a newer version of pantheon")

// Store is the database handle. All access goes through one connection, which
// serialises writes; nothing here does network I/O while a transaction is open.
type Store struct {
	db   *sql.DB
	path string
}

var uriPathEscaper = strings.NewReplacer("%", "%25", "#", "%23", "?", "%3f")

// dsn builds the SQLite URI. The characters that end or reinterpret the path
// part of a URI are escaped, so a data directory called "bots#1" opens that
// directory and not a file called "bots".
func dsn(path string) string {
	q := url.Values{}
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(NORMAL)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "foreign_keys(ON)")
	if path == ":memory:" {
		return "file::memory:?" + q.Encode()
	}
	return "file:" + uriPathEscaper.Replace(filepath.ToSlash(path)) + "?" + q.Encode()
}

// Open opens (creating if needed) the database at path and applies the schema.
func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if dir := filepath.Dir(path); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, err
			}
		}
	}
	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, err
	}
	// One connection: writes are serialised, and an in-memory database (tests)
	// stays alive for as long as the handle does.
	db.SetMaxOpenConns(1)

	s := &Store{db: db, path: path}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version > schemaVersion {
		return fmt.Errorf("%w (schema %d, this build understands %d)", ErrNewerSchema, version, schemaVersion)
	}
	if _, err := s.db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if version != schemaVersion {
		if _, err := s.db.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
	}
	return nil
}

// Path returns where the database lives.
func (s *Store) Path() string { return s.path }

// Close flushes and closes the database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// withTx runs fn in a transaction and rolls it back on any error or panic.
func (s *Store) withTx(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func unix(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func fromUnix(v int64) time.Time {
	if v <= 0 {
		return time.Time{}
	}
	return time.Unix(v, 0)
}

// ---------- flags of one-time migrations ----------

func flagTx(tx *sql.Tx, key string) (bool, error) {
	var v string
	err := tx.QueryRow("SELECT value FROM meta WHERE key = ?", key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func setFlagTx(tx *sql.Tx, key, value string) error {
	_, err := tx.Exec("INSERT INTO meta(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value)
	return err
}

// Flag reports whether a one-time migration already ran.
func (s *Store) Flag(ctx context.Context, key string) (bool, error) {
	var v string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM meta WHERE key = ?", key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Prune deletes seen marks older than seenBefore and finished alerts older
// than alertsBefore. A zero time skips that part. Alerts are kept at least as
// long as seen marks: their UNIQUE(offer_id, chat_id) is the per-chat dedup.
func (s *Store) Prune(ctx context.Context, seenBefore, alertsBefore time.Time) (seen, alerts int64, err error) {
	if !seenBefore.IsZero() {
		res, err := s.db.ExecContext(ctx, "DELETE FROM seen WHERE seen_at < ?", seenBefore.Unix())
		if err != nil {
			return 0, 0, err
		}
		seen, _ = res.RowsAffected()
	}
	if !alertsBefore.IsZero() {
		if !seenBefore.IsZero() && alertsBefore.After(seenBefore) {
			alertsBefore = seenBefore
		}
		res, err := s.db.ExecContext(ctx, "DELETE FROM alerts WHERE status <> 'pending' AND created_at < ?", alertsBefore.Unix())
		if err != nil {
			return seen, 0, err
		}
		alerts, _ = res.RowsAffected()
	}
	return seen, alerts, nil
}

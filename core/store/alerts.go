package store

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Alert states.
const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)

// Scopes of seen marks imported from the 0.1.x seen_ids.txt.
const (
	legacyAny    = "*"
	legacyShared = "shared"
	legacyOwner  = "owner:"
)

const (
	flagLegacySearches = "legacy_searches_imported"
	flagLegacySeen     = "legacy_seen_imported"
)

// SeenScopes returns every scope under which a listing counts as seen for a
// filter: the filter itself plus the scopes of marks imported from 0.1.x,
// which were kept per owner.
func SeenScopes(searchID, ownerChatID string) []string {
	scopes := []string{searchID, legacyAny}
	if ownerChatID == "" {
		return append(scopes, legacyShared)
	}
	return append(scopes, legacyOwner+ownerChatID)
}

// Alert is one queued or delivered notification.
type Alert struct {
	ID         int64
	OfferID    int64
	ChatID     string
	SearchID   string
	SearchName string
	Title      string
	URL        string
	Price      string
	Status     string
	Attempts   int
	LastError  string
	Payload    string // the listing as JSON, kept until the alert is delivered
	CreatedAt  time.Time
	SentAt     time.Time
}

// Reservation asks for one alert to be queued.
type Reservation struct {
	OfferID    int64
	ChatID     string
	SearchName string
	Title      string
	URL        string
	Price      string
	Payload    string
}

// Commit is the outcome of evaluating one page of results for one filter.
type Commit struct {
	SearchID string
	// SeenIDs are all listings of the page that were evaluated, matching or
	// not. A listing skipped only because it is promoted is not among them.
	SeenIDs []int64
	// Alerts are queued together with the seen marks; empty on a silent pass.
	Alerts []Reservation
	Now    time.Time
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// SeenSet reports which of the listings were already evaluated under any of
// the scopes.
func (s *Store) SeenSet(ctx context.Context, scopes []string, offerIDs []int64) (map[int64]bool, error) {
	seen := make(map[int64]bool, len(offerIDs))
	if len(scopes) == 0 || len(offerIDs) == 0 {
		return seen, nil
	}
	// SQLite allows a few hundred bound variables per statement by default;
	// a page with backfill can exceed that, so ids go in chunks.
	const chunk = 400
	for start := 0; start < len(offerIDs); start += chunk {
		end := min(start+chunk, len(offerIDs))
		args := make([]any, 0, len(scopes)+end-start)
		for _, sc := range scopes {
			args = append(args, sc)
		}
		for _, id := range offerIDs[start:end] {
			args = append(args, id)
		}
		rows, err := s.db.QueryContext(ctx,
			"SELECT DISTINCT offer_id FROM seen WHERE scope IN ("+placeholders(len(scopes))+") AND offer_id IN ("+placeholders(end-start)+")", args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			seen[id] = true
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return seen, nil
}

// CommitPage records one evaluated page in a single transaction: the seen
// marks and the alerts they produced land together or not at all, so a crash
// cannot leave a listing marked as seen with its alert lost.
//
// The filter is re-checked inside the transaction. If it was deleted or
// paused while the page was being fetched, nothing is written and ok is false.
// The returned alerts are the ones actually queued: a listing already alerted
// to the same chat by another filter is left out.
func (s *Store) CommitPage(ctx context.Context, c Commit) (queued []Alert, ok bool, err error) {
	now := c.Now
	if now.IsZero() {
		now = time.Now()
	}
	err = s.withTx(ctx, func(tx *sql.Tx) error {
		var name string
		var disabled int
		err := tx.QueryRow("SELECT name, disabled FROM searches WHERE id = ?", c.SearchID).Scan(&name, &disabled)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && disabled != 0) {
			return nil
		}
		if err != nil {
			return err
		}
		ok = true

		if len(c.SeenIDs) > 0 {
			// The upsert refreshes seen_at, so a listing that is still on the
			// results page never ages out and comes back as new.
			stmt, err := tx.Prepare("INSERT INTO seen(scope, offer_id, seen_at) VALUES(?, ?, ?) ON CONFLICT(scope, offer_id) DO UPDATE SET seen_at = excluded.seen_at")
			if err != nil {
				return err
			}
			defer stmt.Close()
			for _, id := range c.SeenIDs {
				if _, err := stmt.Exec(c.SearchID, id, now.Unix()); err != nil {
					return err
				}
			}
		}

		for _, r := range c.Alerts {
			searchName := r.SearchName
			if searchName == "" {
				searchName = name
			}
			res, err := tx.Exec(`INSERT INTO alerts (offer_id, chat_id, search_id, search_name, title, url, price, status, payload, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?) ON CONFLICT(offer_id, chat_id) DO NOTHING`,
				r.OfferID, r.ChatID, c.SearchID, searchName, r.Title, r.URL, r.Price, r.Payload, now.Unix())
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				continue
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			queued = append(queued, Alert{
				ID: id, OfferID: r.OfferID, ChatID: r.ChatID, SearchID: c.SearchID, SearchName: searchName,
				Title: r.Title, URL: r.URL, Price: r.Price, Status: StatusPending, Payload: r.Payload, CreatedAt: now,
			})
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return queued, ok, nil
}

const alertColumns = "id, offer_id, chat_id, search_id, search_name, title, url, price, status, attempts, last_error, COALESCE(payload, ''), created_at, sent_at"

func scanAlert(row rowScanner) (Alert, error) {
	var a Alert
	var created, sent int64
	err := row.Scan(&a.ID, &a.OfferID, &a.ChatID, &a.SearchID, &a.SearchName, &a.Title, &a.URL, &a.Price,
		&a.Status, &a.Attempts, &a.LastError, &a.Payload, &created, &sent)
	a.CreatedAt, a.SentAt = fromUnix(created), fromUnix(sent)
	return a, err
}

func (s *Store) queryAlerts(ctx context.Context, query string, args ...any) ([]Alert, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// PendingChats returns the chats that have alerts waiting to be sent.
func (s *Store) PendingChats(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT chat_id FROM alerts WHERE status = 'pending' ORDER BY chat_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// PendingAlerts returns the oldest waiting alerts of a chat, in queue order.
func (s *Store) PendingAlerts(ctx context.Context, chatID string, limit int) ([]Alert, error) {
	return s.queryAlerts(ctx, "SELECT "+alertColumns+" FROM alerts WHERE status = 'pending' AND chat_id = ? ORDER BY id LIMIT ?", chatID, limit)
}

// PendingCount returns how many alerts are waiting in total.
func (s *Store) PendingCount(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts WHERE status = 'pending'").Scan(&n)
	return n, err
}

// AlertPending reports whether an alert is still waiting to be sent. The
// sender asks right before each send: between reading a batch and reaching an
// alert, its filter may have been deleted or its chat removed.
func (s *Store) AlertPending(ctx context.Context, id int64) (bool, error) {
	var status string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM alerts WHERE id = ?", id).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return status == "pending", err
}

// MarkAlertSent records a delivery. The payload is dropped: only the summary
// columns are needed from here on.
func (s *Store) MarkAlertSent(ctx context.Context, id int64, at time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE alerts SET status = 'sent', sent_at = ?, payload = NULL, attempts = attempts + 1, last_error = '' WHERE id = ?", at.Unix(), id)
	return err
}

// MarkAlertFailed gives up on an alert for good.
func (s *Store) MarkAlertFailed(ctx context.Context, id int64, reason string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE alerts SET status = 'failed', payload = NULL, attempts = attempts + 1, last_error = ? WHERE id = ?", reason, id)
	return err
}

// NoteAlertAttempt records a failed try of an alert that stays in the queue.
func (s *Store) NoteAlertAttempt(ctx context.Context, id int64, reason string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE alerts SET attempts = attempts + 1, last_error = ? WHERE id = ?", reason, id)
	return err
}

// ExpireAlerts fails the alerts that have waited since before the given time:
// a listing that could not be announced for that long is no longer news.
func (s *Store) ExpireAlerts(ctx context.Context, before time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, "UPDATE alerts SET status = 'failed', payload = NULL, last_error = 'expired' WHERE status = 'pending' AND created_at < ?", before.Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// FailPendingOf drops whatever is queued for a chat (it was removed, or
// turned out to be unreachable).
func (s *Store) FailPendingOf(ctx context.Context, chatID, reason string) (int64, error) {
	res, err := s.db.ExecContext(ctx, "UPDATE alerts SET status = 'failed', payload = NULL, last_error = ? WHERE status = 'pending' AND chat_id = ?", reason, chatID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// RecentAlerts returns the latest delivered alerts of a chat, newest first.
func (s *Store) RecentAlerts(ctx context.Context, chatID string, limit int) ([]Alert, error) {
	return s.queryAlerts(ctx, "SELECT "+alertColumns+" FROM alerts WHERE status = 'sent' AND chat_id = ? ORDER BY sent_at DESC, id DESC LIMIT ?", chatID, limit)
}

// AlertCounts returns how many alerts a chat received since the given time,
// in total and per filter.
func (s *Store) AlertCounts(ctx context.Context, chatID string, since time.Time) (total int, bySearch map[string]int, err error) {
	bySearch = map[string]int{}
	rows, err := s.db.QueryContext(ctx, "SELECT search_id, COUNT(*) FROM alerts WHERE status = 'sent' AND chat_id = ? AND sent_at >= ? GROUP BY search_id", chatID, since.Unix())
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return 0, nil, err
		}
		bySearch[id] = n
		total += n
	}
	return total, bySearch, rows.Err()
}

// LastSentAt returns when the chat last received an alert.
func (s *Store) LastSentAt(ctx context.Context, chatID string) (time.Time, error) {
	var at sql.NullInt64
	err := s.db.QueryRowContext(ctx, "SELECT MAX(sent_at) FROM alerts WHERE status = 'sent' AND chat_id = ?", chatID).Scan(&at)
	if err != nil || !at.Valid {
		return time.Time{}, err
	}
	return fromUnix(at.Int64), nil
}

// ImportLegacySeen loads a 0.1.x seen_ids.txt. Its lines are "<owner>|<id>",
// "shared|<id>" or a bare "<id>" that applied to everyone. They are kept
// under their own scopes and consulted alongside the per-filter marks until
// they age out. The file itself is left in place for a downgrade.
func (s *Store) ImportLegacySeen(ctx context.Context, path string, now time.Time) (imported int, ran bool, err error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	defer f.Close()

	type mark struct {
		scope string
		id    int64
	}
	var marks []mark
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		scope, idText := legacyAny, line
		if owner, rest, found := strings.Cut(line, "|"); found {
			idText = rest
			if owner == legacyShared {
				scope = legacyShared
			} else {
				scope = legacyOwner + owner
			}
		}
		id, perr := strconv.ParseInt(strings.TrimSpace(idText), 10, 64)
		if perr != nil || id <= 0 {
			continue
		}
		marks = append(marks, mark{scope, id})
	}
	if err := scanner.Err(); err != nil {
		return 0, false, err
	}

	err = s.withTx(ctx, func(tx *sql.Tx) error {
		done, err := flagTx(tx, flagLegacySeen)
		if err != nil || done {
			return err
		}
		ran = true
		stmt, err := tx.Prepare("INSERT INTO seen(scope, offer_id, seen_at) VALUES(?, ?, ?) ON CONFLICT(scope, offer_id) DO NOTHING")
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, m := range marks {
			if _, err := stmt.Exec(m.scope, m.id, now.Unix()); err != nil {
				return err
			}
			imported++
		}
		return setFlagTx(tx, flagLegacySeen, now.UTC().Format(time.RFC3339))
	})
	if err != nil {
		return 0, false, err
	}
	return imported, ran, nil
}

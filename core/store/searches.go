package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pantheon/core/domain"
)

// ErrNotFound means the filter does not exist (any more).
var ErrNotFound = errors.New("filter not found")

const searchColumns = `id, name, query, query_variants, exclude_words, min_price, max_price, price_mode,
	category_id, category_name, city_id, city_name, region_id, district_id, district_name,
	brand_param, brand_value, brand_name, brand_values, brand_names, owner_type, filters, ranges,
	disabled, snoozed_until, enabled_at, owner_chat_id, created_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func encodeJSON(v any, empty bool) any {
	if empty {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(data)
}

func decodeJSON(raw sql.NullString, out any) {
	if raw.Valid && raw.String != "" {
		_ = json.Unmarshal([]byte(raw.String), out)
	}
}

func scanSearch(row rowScanner) (domain.Search, error) {
	var (
		s                                          domain.Search
		variants, exclude, brandValues, brandNames sql.NullString
		filters, ranges                            sql.NullString
		disabled                                   int
	)
	err := row.Scan(&s.ID, &s.Name, &s.Query, &variants, &exclude, &s.MinPrice, &s.MaxPrice, &s.PriceMode,
		&s.Category, &s.CatName, &s.CityID, &s.CityName, &s.RegionID, &s.DistrictID, &s.DistrictName,
		&s.BrandParam, &s.BrandValue, &s.BrandName, &brandValues, &brandNames, &s.OwnerType, &filters, &ranges,
		&disabled, &s.SnoozedUntil, &s.EnabledAt, &s.OwnerChatID, &s.CreatedAt)
	if err != nil {
		return domain.Search{}, err
	}
	s.Disabled = disabled != 0
	decodeJSON(variants, &s.QueryVariants)
	decodeJSON(exclude, &s.ExcludeWords)
	decodeJSON(brandValues, &s.BrandValues)
	decodeJSON(brandNames, &s.BrandNames)
	decodeJSON(filters, &s.Filters)
	decodeJSON(ranges, &s.Ranges)
	return s, nil
}

func searchArgs(s domain.Search) []any {
	return []any{
		s.Name, s.Query,
		encodeJSON(s.QueryVariants, len(s.QueryVariants) == 0),
		encodeJSON(s.ExcludeWords, len(s.ExcludeWords) == 0),
		s.MinPrice, s.MaxPrice, s.PriceMode,
		s.Category, s.CatName, s.CityID, s.CityName, s.RegionID, s.DistrictID, s.DistrictName,
		s.BrandParam, s.BrandValue, s.BrandName,
		encodeJSON(s.BrandValues, len(s.BrandValues) == 0),
		encodeJSON(s.BrandNames, len(s.BrandNames) == 0),
		s.OwnerType,
		encodeJSON(s.Filters, len(s.Filters) == 0),
		encodeJSON(s.Ranges, len(s.Ranges) == 0),
		boolInt(s.Disabled), s.SnoozedUntil, s.EnabledAt, s.OwnerChatID, s.CreatedAt,
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// newID returns a short random filter id. It travels in callback data, so it
// must stay small; 32 bits of randomness are plenty for a few hundred filters,
// and insertSearch retries on the rare collision.
func newID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("store: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

func insertSearch(tx *sql.Tx, s *domain.Search, now time.Time) error {
	if s.CreatedAt == 0 {
		s.CreatedAt = now.Unix()
	}
	if s.EnabledAt == 0 {
		s.EnabledAt = s.CreatedAt
	}
	var position int
	if err := tx.QueryRow("SELECT COALESCE(MAX(position), -1) + 1 FROM searches").Scan(&position); err != nil {
		return err
	}
	const stmt = `INSERT INTO searches (id, position, name, query, query_variants, exclude_words, min_price, max_price, price_mode,
		category_id, category_name, city_id, city_name, region_id, district_id, district_name,
		brand_param, brand_value, brand_name, brand_values, brand_names, owner_type, filters, ranges,
		disabled, snoozed_until, enabled_at, owner_chat_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	for attempt := 0; attempt < 8; attempt++ {
		s.ID = newID()
		args := append([]any{s.ID, position}, searchArgs(*s)...)
		_, err := tx.Exec(stmt, args...)
		if err == nil {
			return nil
		}
		if !strings.Contains(strings.ToLower(err.Error()), "unique") {
			return err
		}
	}
	return errors.New("could not allocate a filter id")
}

// Searches returns every filter in creation order. Each call reads the
// database, so the result never shares maps or slices with another caller.
func (s *Store) Searches(ctx context.Context) ([]domain.Search, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+searchColumns+" FROM searches ORDER BY position, rowid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Search
	for rows.Next() {
		sc, err := scanSearch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

// Search returns one filter by id.
func (s *Store) Search(ctx context.Context, id string) (domain.Search, error) {
	sc, err := scanSearch(s.db.QueryRowContext(ctx, "SELECT "+searchColumns+" FROM searches WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Search{}, ErrNotFound
	}
	return sc, err
}

// AddSearch stores a new filter and returns it with its id.
func (s *Store) AddSearch(ctx context.Context, sc domain.Search) (domain.Search, error) {
	sc.Normalize()
	if err := sc.Validate(); err != nil {
		return domain.Search{}, err
	}
	sc.ID = ""
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		return insertSearch(tx, &sc, time.Now())
	})
	return sc, err
}

// UpdateSearch applies fn to the stored filter inside one transaction, so a
// concurrent change to the same filter cannot be lost. The id, creation time
// and position are not editable.
func (s *Store) UpdateSearch(ctx context.Context, id string, fn func(*domain.Search) error) (domain.Search, error) {
	var out domain.Search
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		sc, err := scanSearch(tx.QueryRow("SELECT "+searchColumns+" FROM searches WHERE id = ?", id))
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		created := sc.CreatedAt
		if err := fn(&sc); err != nil {
			return err
		}
		sc.ID, sc.CreatedAt = id, created
		sc.Normalize()
		const stmt = `UPDATE searches SET name=?, query=?, query_variants=?, exclude_words=?, min_price=?, max_price=?, price_mode=?,
			category_id=?, category_name=?, city_id=?, city_name=?, region_id=?, district_id=?, district_name=?,
			brand_param=?, brand_value=?, brand_name=?, brand_values=?, brand_names=?, owner_type=?, filters=?, ranges=?,
			disabled=?, snoozed_until=?, enabled_at=?, owner_chat_id=?, created_at=? WHERE id=?`
		if _, err := tx.Exec(stmt, append(searchArgs(sc), id)...); err != nil {
			return err
		}
		out = sc
		return nil
	})
	return out, err
}

// DeleteSearch removes a filter together with its seen marks. Sent alerts
// stay: they are the history and the per-chat dedup.
func (s *Store) DeleteSearch(ctx context.Context, id string) (bool, error) {
	deleted := false
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.Exec("DELETE FROM searches WHERE id = ?", id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		deleted = n > 0
		if _, err := tx.Exec("DELETE FROM seen WHERE scope = ?", id); err != nil {
			return err
		}
		// Whatever was queued for a deleted filter is no longer wanted: someone
		// who deletes a noisy filter expects it to go quiet at once. An alert
		// row names only the first filter that matched its listing, so in the
		// rare case that another filter of the same chat matched it too and
		// the alert was still waiting, that alert is dropped with this one.
		_, err = tx.Exec("UPDATE alerts SET status = 'failed', last_error = 'filter deleted', payload = NULL WHERE status = 'pending' AND search_id = ?", id)
		return err
	})
	return deleted, err
}

// DisableSearchesOf pauses every filter owned by the chat. It runs before the
// chat is removed, so its filters stop polling OLX for nobody.
func (s *Store) DisableSearchesOf(ctx context.Context, chatID string) (int, error) {
	res, err := s.db.ExecContext(ctx, "UPDATE searches SET disabled = 1 WHERE owner_chat_id = ? AND disabled = 0", chatID)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// CountSearches returns how many filters the chat owns.
func (s *Store) CountSearches(ctx context.Context, ownerChatID string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM searches WHERE owner_chat_id = ?", ownerChatID).Scan(&n)
	return n, err
}

// ImportLegacySearches moves the filters of a 0.1.x config.json into the
// database. The "done" flag and the rows are written in one transaction, and
// once the flag is set the call is a no-op: otherwise a config file the bot
// cannot rewrite would resurrect deleted filters on every start.
func (s *Store) ImportLegacySearches(ctx context.Context, list []domain.Search, now time.Time) (imported int, ran bool, err error) {
	err = s.withTx(ctx, func(tx *sql.Tx) error {
		done, err := flagTx(tx, flagLegacySearches)
		if err != nil || done {
			return err
		}
		ran = true
		for _, sc := range list {
			sc.Normalize()
			if strings.TrimSpace(sc.Name) == "" {
				sc.Name = fallbackName(sc)
			}
			// A filter in a section the bot no longer serves is kept, paused,
			// so the owner can see what happened to it.
			if sc.Hidden() {
				sc.Disabled = true
			}
			if err := insertSearch(tx, &sc, now); err != nil {
				return fmt.Errorf("import filter %q: %w", sc.Name, err)
			}
			imported++
		}
		return setFlagTx(tx, flagLegacySearches, now.UTC().Format(time.RFC3339))
	})
	if err != nil {
		return 0, false, err
	}
	return imported, ran, nil
}

func fallbackName(sc domain.Search) string {
	switch {
	case sc.Query != "":
		return sc.Query
	case sc.Brands() != "":
		return sc.Brands()
	case sc.CatName != "":
		return sc.CatName
	}
	return "OLX"
}

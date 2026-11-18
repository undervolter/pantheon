-- Schema of pantheon.db. Every statement is idempotent: the file is applied on
-- each start. Times are unix seconds.

CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Filters. Column names follow the closed edition's table where they overlap.
-- Lists and maps are JSON in TEXT columns.
CREATE TABLE IF NOT EXISTS searches (
    id             TEXT PRIMARY KEY,
    position       INTEGER NOT NULL DEFAULT 0,
    name           TEXT NOT NULL,
    query          TEXT NOT NULL DEFAULT '',
    query_variants TEXT,
    exclude_words  TEXT,
    min_price      REAL NOT NULL DEFAULT 0,
    max_price      REAL NOT NULL DEFAULT 0,
    price_mode     TEXT NOT NULL DEFAULT '',
    category_id    INTEGER NOT NULL DEFAULT 0,
    category_name  TEXT NOT NULL DEFAULT '',
    city_id        INTEGER NOT NULL DEFAULT 0,
    city_name      TEXT NOT NULL DEFAULT '',
    region_id      INTEGER NOT NULL DEFAULT 0,
    district_id    INTEGER NOT NULL DEFAULT 0,
    district_name  TEXT NOT NULL DEFAULT '',
    brand_param    TEXT NOT NULL DEFAULT '',
    brand_value    TEXT NOT NULL DEFAULT '',
    brand_name     TEXT NOT NULL DEFAULT '',
    brand_values   TEXT,
    brand_names    TEXT,
    owner_type     TEXT NOT NULL DEFAULT '',
    filters        TEXT,
    ranges         TEXT,
    disabled       INTEGER NOT NULL DEFAULT 0,
    snoozed_until  INTEGER NOT NULL DEFAULT 0,
    enabled_at     INTEGER NOT NULL DEFAULT 0,
    owner_chat_id  TEXT NOT NULL DEFAULT '',
    created_at     INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_searches_owner ON searches(owner_chat_id);

-- "This listing was already evaluated for this filter." scope is the filter
-- id; rows imported from the 0.1.x seen_ids.txt use the scopes "owner:<chat>",
-- "shared" and "*" and age out with the rest.
CREATE TABLE IF NOT EXISTS seen (
    scope    TEXT    NOT NULL,
    offer_id INTEGER NOT NULL,
    seen_at  INTEGER NOT NULL,
    PRIMARY KEY (scope, offer_id)
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS idx_seen_at ON seen(seen_at);

-- Delivery queue, per-chat dedup, and the source of the history screen and
-- the "today / 7 days" counters. UNIQUE(offer_id, chat_id) is what makes one
-- listing produce at most one alert per chat, however many filters match it.
CREATE TABLE IF NOT EXISTS alerts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    offer_id    INTEGER NOT NULL,
    chat_id     TEXT    NOT NULL,
    search_id   TEXT    NOT NULL DEFAULT '',
    search_name TEXT    NOT NULL DEFAULT '',
    title       TEXT    NOT NULL DEFAULT '',
    url         TEXT    NOT NULL DEFAULT '',
    price       TEXT    NOT NULL DEFAULT '',
    status      TEXT    NOT NULL DEFAULT 'pending',
    attempts    INTEGER NOT NULL DEFAULT 0,
    last_error  TEXT    NOT NULL DEFAULT '',
    payload     TEXT,
    created_at  INTEGER NOT NULL,
    sent_at     INTEGER NOT NULL DEFAULT 0,
    UNIQUE (offer_id, chat_id)
);
CREATE INDEX IF NOT EXISTS idx_alerts_queue ON alerts(status, chat_id, id);
CREATE INDEX IF NOT EXISTS idx_alerts_chat_sent ON alerts(chat_id, status, sent_at);

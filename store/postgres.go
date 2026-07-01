package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// postgresStore wraps a *sql.DB and provides the durable, source-of-truth
// persistence for short url -> long url mappings. Redis sits in front of this as a cache; whatever isn't found in the cache falls back here.
type postgresStore struct {
	db *sql.DB
}

const createTableSQL = `
CREATE TABLE IF NOT EXISTS url_mappings (
	short_url   TEXT PRIMARY KEY,
	long_url    TEXT NOT NULL,
	user_id     TEXT NOT NULL,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	last_read_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

// newPostgresStore opens a connection pool, verifies connectivity, and ensures the schema exists.
func newPostgresStore(dsn string) (*postgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("creating url_mappings table: %w", err)
	}

	return &postgresStore{db: db}, nil
}

// Save upserts the mapping. On conflict (same short url generated twice,
// e.g. same user shortening the same link again) we just refresh long_url.
func (p *postgresStore) Save(shortUrl, longUrl, userId string) error {
	_, err := p.db.Exec(
		`INSERT INTO url_mappings (short_url, long_url, user_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (short_url) DO UPDATE SET long_url = EXCLUDED.long_url`,
		shortUrl, longUrl, userId,
	)
	if err != nil {
		return fmt.Errorf("saving url mapping for %s: %w", shortUrl, err)
	}
	return nil
}

// Get looks up the long url for a short url. Returns ErrNotFound if no row exists.
// Also bumps last_read_at so we have real recency data should we ever want to drive eviction decisions off it.
func (p *postgresStore) Get(shortUrl string) (string, error) {
	var longUrl string
	err := p.db.QueryRow(
		`UPDATE url_mappings SET last_read_at = now()
		 WHERE short_url = $1
		 RETURNING long_url`,
		shortUrl,
	).Scan(&longUrl)

	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("fetching url mapping for %s: %w", shortUrl, err)
	}
	return longUrl, nil
}

func (p *postgresStore) Close() error {
	return p.db.Close()
}

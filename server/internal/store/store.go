// Package store is the persistence layer: a Postgres-backed implementation of
// the repository interfaces, plus connection and migration management. All SQL
// access is isolated here so the rest of the server depends on interfaces.
package store

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mmandrille_meli/arpg-dev/server/migrations"
)

// Store is the Postgres-backed repository implementation.
type Store struct {
	pool *pgxpool.Pool
}

// Connect opens a pooled connection to Postgres and verifies connectivity.
func Connect(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("store: parse config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Ping checks database connectivity (used by /readyz).
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close releases the connection pool.
func (s *Store) Close() { s.pool.Close() }

// Migrate applies all embedded migrations that have not yet been applied, in
// filename order, each in its own transaction. It is idempotent.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("store: ensure schema_migrations: %w", err)
	}

	entries, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return fmt.Errorf("store: list migrations: %w", err)
	}
	sort.Strings(entries)

	for _, name := range entries {
		version, err := versionFromName(name)
		if err != nil {
			return err
		}
		var exists bool
		if err := s.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
		).Scan(&exists); err != nil {
			return fmt.Errorf("store: check migration %d: %w", version, err)
		}
		if exists {
			continue
		}
		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("store: read migration %s: %w", name, err)
		}
		if err := s.applyMigration(ctx, version, string(body)); err != nil {
			return fmt.Errorf("store: apply migration %s: %w", name, err)
		}
	}
	return nil
}

func (s *Store) applyMigration(ctx context.Context, version int64, body string) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, body); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version)
		return err
	})
}

// versionFromName extracts the leading integer of a migration filename, e.g.
// "0001_init.sql" -> 1.
func versionFromName(name string) (int64, error) {
	base := name
	if i := strings.IndexByte(base, '_'); i >= 0 {
		base = base[:i]
	}
	v, err := strconv.ParseInt(base, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("store: bad migration filename %q: %w", name, err)
	}
	return v, nil
}

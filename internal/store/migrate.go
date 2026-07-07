// ©AngelaMos | 2026
// migrate.go

package store

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	out := make([]migration, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		verStr, _, ok := strings.Cut(e.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("migration %q: expected NNNN_name.sql", e.Name())
		}
		ver, err := strconv.Atoi(verStr)
		if err != nil {
			return nil, fmt.Errorf("migration %q: bad version prefix: %w", e.Name(), err)
		}
		body, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", e.Name(), err)
		}
		out = append(out, migration{version: ver, name: e.Name(), sql: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

func migrate(db *sql.DB) (int, error) {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	)`); err != nil {
		return 0, fmt.Errorf("create schema_migrations: %w", err)
	}

	var current int
	if err := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current); err != nil {
		return 0, fmt.Errorf("read current version: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return current, err
	}

	maxKnown := 0
	for _, m := range migrations {
		if m.version > maxKnown {
			maxKnown = m.version
		}
	}
	if current > maxKnown {
		return current, fmt.Errorf("store schema version %d is newer than this binary supports (max %d); refusing to run against a store written by a newer build", current, maxKnown)
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return current, fmt.Errorf("begin migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(m.sql); err != nil {
			_ = tx.Rollback()
			return current, fmt.Errorf("apply migration %q: %w", m.name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, m.version); err != nil {
			_ = tx.Rollback()
			return current, fmt.Errorf("record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return current, fmt.Errorf("commit migration %d: %w", m.version, err)
		}
		current = m.version
	}
	return current, nil
}

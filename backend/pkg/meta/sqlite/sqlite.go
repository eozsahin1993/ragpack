package sqlite

import (
	"context"
	"embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(name)
	s = nonAlphanumeric.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

func newTableName(displayName string) (fullID string, tableName string) {
	id := uuid.New()
	slug := slugify(displayName)
	return id.String(), slug + "_" + id.String()[:8]
}

type MetaStore struct {
	db *sqlx.DB
}

func New(path string) (*MetaStore, error) {
	// WAL lets reads proceed without waiting on a writer; busy_timeout makes a
	// second concurrent writer retry instead of failing immediately with
	// SQLITE_BUSY (SQLite still allows only one writer at a time, WAL or not).
	db, err := sqlx.Open("sqlite", path+"?_foreign_keys=on&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("sqlite: open failed: %w", err)
	}

	db.SetMaxOpenConns(10)

	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: set dialect: %w", err)
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: run migrations: %w", err)
	}

	ms := &MetaStore{db: db}
	if err := ms.upsertSystemPrompts(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: seed system prompts: %w", err)
	}

	return ms, nil
}

func (s *MetaStore) Close() error {
	return s.db.Close()
}

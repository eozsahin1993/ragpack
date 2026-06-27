package migrations

import (
	"context"
	"fmt"
	"log"

	"github.com/eozsahin1993/lancedb-go/pkg/contracts"

	"ragpack/pkg/meta"
)

type Migration struct {
	Version int
	Up      func(ctx context.Context, tbl contracts.ITable) error
}

// registry is populated automatically by each migration file's init().
// Files are processed alphabetically, so numeric prefixes guarantee order.
var registry []Migration

// Register is called by each migration file's init() to add itself to the registry.
func Register(m Migration) {
	registry = append(registry, m)
}

// CurrentVersion returns the version number of the latest registered migration.
func CurrentVersion() int {
	if len(registry) == 0 {
		return 0
	}
	return registry[len(registry)-1].Version
}

// MarkUpToDate records all registered migrations as applied for a newly created
// collection table. Call this immediately after CreateTable so the migration
// runner never tries to apply migrations to a table that was born with the
// latest schema.
func MarkUpToDate(ctx context.Context, conn contracts.IConnection, collectionID string) error {
	mTbl, err := openOrCreateMigrationsTable(ctx, conn)
	if err != nil {
		return fmt.Errorf("lancedb migrate: open migrations table: %w", err)
	}
	defer mTbl.Close()
	for _, m := range registry {
		if err := recordApplied(ctx, mTbl, collectionID, m.Version); err != nil {
			return err
		}
	}
	return nil
}

// MigrateAll applies any pending migrations to every collection table.
// Call once at server startup after connecting to LanceDB.
func MigrateAll(ctx context.Context, conn contracts.IConnection, collections []meta.Collection) error {
	if err := validateRegistry(); err != nil {
		return err
	}

	mTbl, err := openOrCreateMigrationsTable(ctx, conn)
	if err != nil {
		return fmt.Errorf("lancedb migrate: open migrations table: %w", err)
	}
	defer mTbl.Close()

	for _, col := range collections {
		tbl, err := conn.OpenTable(ctx, col.TableName)
		if err != nil {
			return fmt.Errorf("lancedb migrate: open table %s: %w", col.TableName, err)
		}
		err = apply(ctx, tbl, mTbl, col.ID)
		tbl.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func apply(ctx context.Context, collectionTbl contracts.ITable, trackingTbl contracts.ITable, collectionID string) error {
	for _, m := range registry {
		applied, err := hasApplied(ctx, trackingTbl, collectionID, m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		log.Printf("lancedb migrate: applying version %d to collection %s", m.Version, collectionID)
		if err := m.Up(ctx, collectionTbl); err != nil {
			return fmt.Errorf("lancedb migrate: %d on collection %s: %w", m.Version, collectionID, err)
		}
		if err := recordApplied(ctx, trackingTbl, collectionID, m.Version); err != nil {
			return err
		}
		log.Printf("lancedb migrate: version %d applied to collection %s", m.Version, collectionID)
	}
	return nil
}

// validateRegistry ensures migrations are registered in strict ascending order
// with no duplicates. Returns an error at startup rather than silently applying out-of-order.
func validateRegistry() error {
	for i, m := range registry {
		if m.Version <= 0 {
			return fmt.Errorf("lancedb migrations: entry %d has invalid version %d (must be > 0)", i, m.Version)
		}
		if i > 0 && m.Version != registry[i-1].Version+1 {
			return fmt.Errorf("lancedb migrations: version %d must follow %d with no gaps",
				m.Version, registry[i-1].Version)
		}
	}
	return nil
}

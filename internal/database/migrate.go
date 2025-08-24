package database

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a database migration.
type Migration struct {
	Version  string
	Filename string
	Content  string
}

// RunMigrations applies all pending database migrations.
func (d *Database) RunMigrations(ctx context.Context, migrationsFS fs.FS) error {
	migrations, err := loadMigrations(migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if err := d.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	for _, migration := range migrations {
		applied, err := d.isMigrationApplied(ctx, migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.Version, err)
		}

		if applied {
			continue
		}

		if err := d.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

func loadMigrations(migrationsFS fs.FS) ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		version := strings.TrimSuffix(filepath.Base(path), ".sql")
		migrations = append(migrations, Migration{
			Version:  version,
			Filename: path,
			Content:  string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (d *Database) ensureMigrationsTable(ctx context.Context) error {
	// First ensure the schema exists
	schemaQuery := `CREATE SCHEMA IF NOT EXISTS scry_quest;`
	_, err := d.conn.Exec(ctx, schemaQuery)
	if err != nil {
		return fmt.Errorf("failed to ensure schema: %w", err)
	}

	// Then create the migrations table
	query := `
		CREATE TABLE IF NOT EXISTS scry_quest.schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	_, err = d.conn.Exec(ctx, query)
	return err
}

func (d *Database) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM scry_quest.schema_migrations WHERE version = $1)"

	var exists bool
	err := d.conn.QueryRow(ctx, query, version).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *Database) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, migration.Content)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	insertQuery := "INSERT INTO scry_quest.schema_migrations (version) VALUES ($1)"
	_, err = tx.Exec(ctx, insertQuery, migration.Version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit(ctx)
}

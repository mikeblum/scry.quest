package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const (
	migrationsDir     = "migrations"
	migrationsDialect = "postgres"
	ansiRegex         = `(?mi)(\\x9b|\\x1b)\[[0-?]*[ -\/]*[@-~]`
)

type slogBridge struct {
	re *regexp.Regexp
}

func (s *slogBridge) Fatalf(format string, v ...interface{}) {
	slog.Error(fmt.Sprintf(s.format(format), v...))
}

func (s *slogBridge) Printf(format string, v ...interface{}) {
	slog.Info(fmt.Sprintf(s.format(format), v...))
}

func (s *slogBridge) format(format string) string {
	if s.re == nil {
		s.re = regexp.MustCompile(ansiRegex)
	}
	return strings.TrimSpace(strings.TrimSuffix(s.re.ReplaceAllString(format, ""), "\n"))
}

// RunMigrations applies all pending database migrations using Goose.
func (d *Database) RunMigrations(ctx context.Context, config Config) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	pgxConfig, err := pgx.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}
	sqlDB := stdlib.OpenDB(*pgxConfig)
	defer func() {
		_ = sqlDB.Close()
	}()

	goose.SetLogger(&slogBridge{})
	goose.SetBaseFS(embedMigrations)
	goose.SetTableName(migrationsDir)

	if err := goose.SetDialect(migrationsDialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

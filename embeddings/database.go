package embeddings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/internal/database"
)

const (
	// EnvDatabaseURL is the environment variable for database connection string
	EnvDatabaseURL     = "DATABASE_URL"
	defaultPostgresURI = "postgres://localhost/scry_quest?sslmode=disable"
)

// NewDBConn creates a new database connection and returns queries interface and connection
func NewDBConn(ctx context.Context) (*database.Queries, *pgx.Conn, func() error, error) {
	conf, err := conf.New(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	databaseURL := conf.GetPrefixedEnv(EnvDatabaseURL, defaultPostgresURI)

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cleanup := func() error {
		return conn.Close(ctx)
	}

	return database.New(conn), conn, cleanup, nil
}

// Package database provides database connection and operations for the D&D SRD application.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Database represents a database connection and query interface.
type Database struct {
	conn    *pgx.Conn
	queries *Queries
}

// Config holds the database connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewDatabase creates a new database connection with the given configuration.
func NewDatabase(ctx context.Context, config Config) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := New(conn)

	return &Database{
		conn:    conn,
		queries: queries,
	}, nil
}

// Close closes the database connection.
func (d *Database) Close(ctx context.Context) error {
	return d.conn.Close(ctx)
}

// Queries returns the query interface.
func (d *Database) Queries() *Queries {
	return d.queries
}

// Conn returns the underlying database connection.
func (d *Database) Conn() *pgx.Conn {
	return d.conn
}

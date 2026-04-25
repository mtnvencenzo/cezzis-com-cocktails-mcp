// Package db provides PostgreSQL database connection management and initialization.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// NewPool creates a new PostgreSQL connection pool using the application settings.
func NewPool(ctx context.Context, settings *config.AppSettings) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres pool config: %w", err)
	}

	poolConfig.ConnConfig.Host = settings.PostgresHost
	poolConfig.ConnConfig.Port = uint16(settings.PostgresPort)
	poolConfig.ConnConfig.Database = settings.PostgresDBName
	poolConfig.ConnConfig.User = settings.PostgresUser
	poolConfig.ConnConfig.Password = settings.PostgresPassword

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	return pool, nil
}

// EnsureDatabaseExists connects to the default 'postgres' database and creates the
// target database if it does not already exist.
func EnsureDatabaseExists(ctx context.Context, settings *config.AppSettings) error {
	dbName := settings.PostgresDBName
	if dbName == "" {
		telemetry.Logger.Warn().Msg("POSTGRES_DB not set, skipping database creation check")
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	adminConnConfig, err := pgx.ParseConfig("")
	if err != nil {
		return fmt.Errorf("failed to initialize postgres admin config: %w", err)
	}

	adminConnConfig.Host = settings.PostgresHost
	adminConnConfig.Port = uint16(settings.PostgresPort)
	adminConnConfig.Database = "postgres"
	adminConnConfig.User = settings.PostgresUser
	adminConnConfig.Password = settings.PostgresPassword

	conn, err := pgx.ConnectConfig(ctx, adminConnConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres admin database: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		telemetry.Logger.Info().Str("database", dbName).Msg("Database does not exist, creating it")
		// Database names cannot be parameterized; validate to prevent injection
		if !isValidIdentifier(dbName) {
			return fmt.Errorf("invalid database name: %s", dbName)
		}
		_, err = conn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		telemetry.Logger.Info().Str("database", dbName).Msg("Database created successfully")
	} else {
		telemetry.Logger.Info().Str("database", dbName).Msg("Database already exists")
	}

	return nil
}

// EnsureTablesExist creates the required tables if they do not already exist.
func EnsureTablesExist(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS session_tokens (
			id TEXT PRIMARY KEY,
			access_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			token_type TEXT NOT NULL,
			scope TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create session_tokens table: %w", err)
	}

	telemetry.Logger.Info().Msg("Database tables ensured")
	return nil
}

// isValidIdentifier checks that a database name contains only safe characters.
func isValidIdentifier(name string) bool {
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' && c != '-' {
			return false
		}
	}
	return len(name) > 0
}

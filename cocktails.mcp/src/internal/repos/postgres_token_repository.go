// Package repos provides data repository implementations
// for managing session tokens in PostgreSQL.
package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// SessionToken represents a user's session token
type SessionToken struct {
	ID           string    `json:"id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
}

// PostgresTokenRepository manages session tokens in PostgreSQL
type PostgresTokenRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTokenRepository creates a new PostgreSQL token repository instance
func NewPostgresTokenRepository(pool *pgxpool.Pool) *PostgresTokenRepository {
	return &PostgresTokenRepository{pool: pool}
}

// SaveToken upserts a session token into PostgreSQL
func (r *PostgresTokenRepository) SaveToken(ctx context.Context, sessionID string, sessionToken *SessionToken) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ctx, span := telemetry.Tracer.Start(ctx, "PostgreSQL.UpsertItem")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "upsert_item"),
		attribute.String("db.sql.table", "session_tokens"),
		attribute.String("db.item_id", sessionID),
	)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO session_tokens (id, access_token, refresh_token, expires_at, token_type, scope)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			token_type = EXCLUDED.token_type,
			scope = EXCLUDED.scope
	`, sessionToken.ID, sessionToken.AccessToken, sessionToken.RefreshToken,
		sessionToken.ExpiresAt, sessionToken.TokenType, sessionToken.Scope)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "PostgreSQL upsert failed")
		return fmt.Errorf("failed to save token: %w", err)
	}

	span.SetStatus(codes.Ok, "PostgreSQL upsert succeeded")
	return nil
}

// GetToken retrieves a session token from PostgreSQL
func (r *PostgresTokenRepository) GetToken(ctx context.Context, sessionID string) (*SessionToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ctx, span := telemetry.Tracer.Start(ctx, "PostgreSQL.ReadItem")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "read_item"),
		attribute.String("db.sql.table", "session_tokens"),
		attribute.String("db.item_id", sessionID),
	)

	var token SessionToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, access_token, refresh_token, expires_at, token_type, scope
		FROM session_tokens WHERE id = $1
	`, sessionID).Scan(
		&token.ID, &token.AccessToken, &token.RefreshToken,
		&token.ExpiresAt, &token.TokenType, &token.Scope,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		span.SetStatus(codes.Ok, "PostgreSQL item not found")
		telemetry.Logger.Warn().Ctx(ctx).Str("sessionId", sessionID).Msg("No token found")
		return nil, nil
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "PostgreSQL read failed")
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	span.SetStatus(codes.Ok, "PostgreSQL read succeeded")
	return &token, nil
}

// ClearTokens removes a session token from PostgreSQL
func (r *PostgresTokenRepository) ClearTokens(ctx context.Context, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ctx, span := telemetry.Tracer.Start(ctx, "PostgreSQL.DeleteItem")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "delete_item"),
		attribute.String("db.sql.table", "session_tokens"),
		attribute.String("db.item_id", sessionID),
	)

	result, err := r.pool.Exec(ctx, `DELETE FROM session_tokens WHERE id = $1`, sessionID)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "PostgreSQL delete failed")
		return fmt.Errorf("failed to delete token: %w", err)
	}

	if result.RowsAffected() == 0 {
		telemetry.Logger.Warn().Ctx(ctx).
			Str("sessionID", sessionID).
			Msg("No tokens found to clear")
	}

	span.SetStatus(codes.Ok, "PostgreSQL delete succeeded")
	return nil
}

// Package background provides background job runners for the application.
package background

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/db"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// RunInitJob runs the application initialization job after a configurable delay.
// It ensures the database and tables exist. This is intended to be called as a
// goroutine during application startup.
func RunInitJob(ctx context.Context, pool *pgxpool.Pool, settings *config.AppSettings) {
	if !settings.InitJobEnabled {
		telemetry.Logger.Info().Msg("Init job is disabled, skipping")
		return
	}

	delay := time.Duration(settings.InitDelaySeconds) * time.Second

	telemetry.Logger.Info().
		Int("delay_seconds", settings.InitDelaySeconds).
		Msg("Init job scheduled, waiting before execution")

	select {
	case <-time.After(delay):
		// delay elapsed, proceed
	case <-ctx.Done():
		telemetry.Logger.Info().Msg("Init job cancelled during delay")
		return
	}

	telemetry.Logger.Info().Msg("Running application initialization job")

	if err := db.EnsureDatabaseExists(ctx, settings); err != nil {
		telemetry.Logger.Error().Err(err).Msg("Init job: failed to ensure database exists")
		return
	}

	if err := db.EnsureTablesExist(ctx, pool); err != nil {
		telemetry.Logger.Error().Err(err).Msg("Init job: failed to ensure tables exist")
		return
	}

	telemetry.Logger.Info().Msg("Application initialization completed successfully")
}

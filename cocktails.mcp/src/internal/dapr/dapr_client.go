// Package dapr provides a lightweight Dapr sidecar client for job scheduling.
package dapr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// WaitForSidecar polls the Dapr sidecar health endpoint until it becomes ready.
// Returns true if the sidecar is ready, false if it could not be reached.
func WaitForSidecar(settings *config.AppSettings) bool {
	healthURL := fmt.Sprintf("%s/v1.0/healthz", settings.GetDaprHTTPEndpoint())

	client := &http.Client{Timeout: 5 * time.Second}

	maxAttempts := 10
	retryInterval := 3 * time.Second

	for i := range maxAttempts {
		resp, err := client.Get(healthURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				telemetry.Logger.Info().
					Int("attempt", i+1).
					Int("max_attempts", maxAttempts).
					Msg("Dapr sidecar is ready")
				return true
			}
			telemetry.Logger.Info().
				Int("status", resp.StatusCode).
				Int("attempt", i+1).
				Int("max_attempts", maxAttempts).
				Msg("Dapr sidecar not ready yet")
		} else {
			telemetry.Logger.Info().
				Int("attempt", i+1).
				Int("max_attempts", maxAttempts).
				Msg("Dapr sidecar not reachable yet")
		}

		time.Sleep(retryInterval)
	}

	return false
}

// deleteInitJob deletes any existing initialize-app job so it can be re-scheduled.
func deleteInitJob(settings *config.AppSettings) {
	jobURL := fmt.Sprintf("%s/v1.0-alpha1/jobs/initialize-app", settings.GetDaprHTTPEndpoint())

	req, err := http.NewRequest("DELETE", jobURL, nil)
	if err != nil {
		telemetry.Logger.Warn().Err(err).Msg("Failed to create delete request for init job")
		return
	}

	if settings.DaprAPIToken != "" {
		req.Header.Set("dapr-api-token", settings.DaprAPIToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		telemetry.Logger.Warn().Err(err).Msg("Failed to delete existing init job")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		telemetry.Logger.Info().Msg("Deleted existing Dapr init job")
	}
}

// ScheduleInitJob schedules the initialize-app Dapr job to run after a delay.
// This allows time for rolling updates to complete before running initialization.
// Errors are logged as warnings and do not propagate to the caller.
func ScheduleInitJob(settings *config.AppSettings) {
	// Remove any leftover job from a previous run to avoid AlreadyExists errors.
	deleteInitJob(settings)

	jobURL := fmt.Sprintf("%s/v1.0-alpha1/jobs/initialize-app", settings.GetDaprHTTPEndpoint())

	dueTime := time.Now().Add(120 * time.Second).UTC().Format(time.RFC3339)

	jobSpec := map[string]interface{}{
		"dueTime": dueTime,
		"repeats": 1,
		"data": map[string]interface{}{
			"@type": "type.googleapis.com/google.protobuf.StringValue",
			"value": "initialize",
		},
	}

	body, err := json.Marshal(jobSpec)
	if err != nil {
		telemetry.Logger.Warn().Err(err).Msg("Failed to marshal init job spec")
		return
	}

	req, err := http.NewRequest("POST", jobURL, bytes.NewReader(body))
	if err != nil {
		telemetry.Logger.Warn().Err(err).Msg("Failed to create init job request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if settings.DaprAPIToken != "" {
		req.Header.Set("dapr-api-token", settings.DaprAPIToken)
	}

	telemetry.Logger.Info().
		Str("due_time", dueTime).
		Msg("Scheduling initialize-app Dapr job")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		telemetry.Logger.Warn().Err(err).
			Msg("Could not connect to Dapr sidecar to schedule init job. App initialization can be triggered manually.")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		telemetry.Logger.Info().Msg("Successfully scheduled initialize-app job")
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		telemetry.Logger.Warn().
			Int("status", resp.StatusCode).
			Str("body", string(respBody)).
			Msg("Failed to schedule initialize-app job")
	}
}

// ScheduleInitJobBackground waits for the sidecar and schedules the init job in the background.
// This is intended to be called as a goroutine during application startup.
func ScheduleInitJobBackground(settings *config.AppSettings) {
	if !settings.DaprInitJobEnabled {
		telemetry.Logger.Info().Msg("Dapr init job is disabled, skipping")
		return
	}

	if !WaitForSidecar(settings) {
		telemetry.Logger.Warn().
			Msg("Dapr sidecar did not become ready after retries. App initialization can be triggered manually.")
		return
	}

	ScheduleInitJob(settings)
}

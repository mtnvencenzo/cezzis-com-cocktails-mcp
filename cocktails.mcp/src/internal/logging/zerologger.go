// Package logging provides a logger that uses zerolog and Application Insights.
package logging

import (
	"encoding/json"
	"fmt"
	"os"

	"cezzis.com/cezzis-mcp-server/internal/environment"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/rs/zerolog"
)

// Logger is the global logger instance.
var Logger zerolog.Logger

// InitLogger initializes the logger.
func InitLogger() (zerolog.Logger, error) {
	telemetryClient := appinsights.NewTelemetryClient(os.Getenv("APPLICATIONINSIGHTS_INSTRUMENTATIONKEY"))

	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info" // default
	}
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel // fallback
	}

	zerolog.SetGlobalLevel(level)

	// Create a console writer with pretty printing for local development
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}

	// Create an Application Insights writer
	aiLogger := &appInsightsLogger{telemetryClient: telemetryClient}

	var multiWriter zerolog.LevelWriter

	if environment.IsLocalEnv() {
		multiWriter = zerolog.MultiLevelWriter(consoleWriter, aiLogger)
	} else {
		multiWriter = zerolog.MultiLevelWriter(aiLogger)
	}

	// Create and register the hook
	appInsightsHook := &appInsightsHook{}

	Logger = zerolog.New(multiWriter).Hook(appInsightsHook).With().Timestamp().Logger()

	return Logger, nil
}

// AppInsightsHook is a hook that sends logs to Application Insights.
type appInsightsHook struct{}

type appInsightsLogger struct {
	telemetryClient appinsights.TelemetryClient
}

func (a *appInsightsLogger) Write(p []byte) (n int, err error) {
	// Parse the JSON log data from the byte slice
	var logData map[string]interface{}
	err = json.Unmarshal(p, &logData) // Unmarshal JSON into a map[string]interface{}
	if err != nil {
		// Handle the unmarshalling error (e.g., log an error message, but not to App Insights to avoid an infinite loop)
		_, _ = fmt.Fprintf(os.Stderr, "Error unmarshalling log data: %v\n", err)

		return len(p), err // Still return the length of the data processed
	}

	// Extract data from the parsed map
	message, ok := logData["message"].(string) // Get the log message
	if !ok {
		// Handle the case where the message field is not found or not a string
		message = "Log message not available" // Default message
	}

	// Extract and convert the log level
	levelStr, ok := logData["level"].(string) // Get the log level as a string
	severity := appinsights.Information       // Default severity

	if ok {
		// Convert Zerolog level string to App Insights SeverityLevel
		switch levelStr {
		case "debug":
			severity = appinsights.Verbose
		case "info":
			severity = appinsights.Information
		case "warn":
			severity = appinsights.Warning
		case "error":
			severity = appinsights.Error
		case "fatal", "panic":
			severity = appinsights.Critical
		}
	}

	// Create and send Application Insights trace telemetry
	telemetry := appinsights.NewTraceTelemetry(message, severity)

	// Extract and set the Correlation ID
	correlationID, ok := logData["correlation_id"].(string)
	if ok {
		telemetry.Properties["correlation_id"] = correlationID
		telemetry.Tags["ai.operation.id"] = correlationID // Set the operation ID using Tags
	}

	for key, value := range logData {
		if key != "message" && key != "level" && key != "time" && key != "correlation_id" {
			telemetry.Properties[key] = fmt.Sprintf("%v", value)
		}
	}

	a.telemetryClient.Track(telemetry)
	return len(p), nil
}

func (h *appInsightsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	correlationID, ok := e.GetCtx().Value("correlationID").(string)
	if ok {
		e.Str("CorrelationID", correlationID) // Add to the Zerolog event JSON output
	}
}

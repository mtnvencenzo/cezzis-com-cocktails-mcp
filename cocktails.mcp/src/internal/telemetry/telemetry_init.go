// Package telemetry provides a logger that uses zerolog and Application Insights.
package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Logger is the global logger instance.
// Tracer is the global tracer instance.
// Meter is the global meter instance.
var (
	Logger zerolog.Logger
	Tracer trace.Tracer
	Meter  metric.Meter
)

// InitTelemetry initializes the logger.
func InitTelemetry() error {
	otelClient := otelslog.NewLogger(serviceName)

	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info" // default
	}
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel // fallback
	}

	zerolog.SetGlobalLevel(level)

	// Create an OpenTelemetry writer
	otelLogger := &otelLogger{otelClient: *otelClient}

	var multiWriter zerolog.LevelWriter

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	multiWriter = zerolog.MultiLevelWriter(consoleWriter, otelLogger)

	// Create and register the hook
	otelHook := &otelHook{}

	Logger = zerolog.New(multiWriter).Hook(otelHook).With().Timestamp().Logger()
	Tracer = otel.Tracer(serviceName)
	Meter = otel.Meter(serviceName)

	return nil
}

// OtelHook is a hook that sends logs to OpenTelemetry.
type otelHook struct{}

type otelLogger struct {
	otelClient slog.Logger
}

func (a *otelLogger) Write(p []byte) (n int, err error) {
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
	severity := slog.LevelInfo                // Default severity

	if ok {
		// Convert Zerolog level string to App Insights SeverityLevel
		switch levelStr {
		case "debug":
			severity = slog.LevelDebug
		case "info":
			severity = slog.LevelInfo
		case "warn":
			severity = slog.LevelWarn
		case "error":
			severity = slog.LevelError
		case "fatal", "panic":
			severity = slog.LevelError
		}
	}

	attrs := []any{}

	for key, value := range logData {
		if key != "message" && key != "level" && key != "time" {
			attrs = append(attrs, slog.Attr{Key: key, Value: slog.AnyValue(value)})
		}
	}

	a.otelClient.Log(context.Background(), severity, message, attrs...)
	return len(p), nil
}

func (h *otelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	correlationID, ok := e.GetCtx().Value("correlationID").(string)
	if ok {
		e.Str("CorrelationID", correlationID) // Add to the Zerolog event JSON output
	}
}

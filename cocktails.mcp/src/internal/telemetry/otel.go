// Package telemetry provides a logger that uses zerolog and Application Insights.
package telemetry

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	res "go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/environment"
)

const serviceName = "cocktails.mcp"

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, version string) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	resource := res.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("cocktails-mcp"),
		semconv.ServiceNamespace("cezzis"),
		semconv.ServiceVersion(version),
		semconv.ServiceInstanceID(environment.GetHostName()),
		attribute.Key("deployment.environment").String(environment.GetEnvironmentName()),
		attribute.Key("host.host").String(environment.GetHostName()),
		attribute.Key("host.name").String(environment.GetHostName()),
	)

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	appSettings := config.GetAppSettings()

	// Set up trace provider.
	if appSettings.OTLPTraceEnabled {
		tracerProvider, err := newTracerProvider(ctx, resource)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	// Set up meter provider.
	if appSettings.OTLPMetricsEnabled {
		meterProvider, err := newMeterProvider(ctx, resource)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	// Set up logger provider.
	if appSettings.OTLPLogEnabled {
		loggerProvider, err := newLoggerProvider(ctx, resource)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
	}

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, resource *res.Resource) (*trace.TracerProvider, error) {
	appSettings := config.GetAppSettings()

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpointURL(appSettings.OTLPEndpoint),
		otlptracegrpc.WithHeaders(getHeaderMap(appSettings.OTLPHeaders)),
	}

	if appSettings.OTLPInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	client := otlptracegrpc.NewClient(opts...)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	bsp := trace.NewBatchSpanProcessor(exporter)

	traceOpts := []trace.TracerProviderOption{
		trace.WithSpanProcessor(bsp),
		trace.WithResource(resource),
	}

	if environment.IsLocalEnv() {
		traceOpts = append(traceOpts, trace.WithSampler(trace.AlwaysSample()))
	} else {
		traceOpts = append(traceOpts, trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.8))))
	}

	tracerProvider := trace.NewTracerProvider(traceOpts...)

	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context, resource *res.Resource) (*metric.MeterProvider, error) {
	appSettings := config.GetAppSettings()

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpointURL(appSettings.OTLPEndpoint),
		otlpmetricgrpc.WithHeaders(getHeaderMap(appSettings.OTLPHeaders)),
	}

	if appSettings.OTLPInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(1*time.Second))),
		metric.WithResource(resource),
	)

	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context, resource *res.Resource) (*log.LoggerProvider, error) {
	appSettings := config.GetAppSettings()

	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpointURL(appSettings.OTLPEndpoint),
		otlploggrpc.WithHeaders(getHeaderMap(appSettings.OTLPHeaders)),
	}

	if appSettings.OTLPInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	bsp := log.NewBatchProcessor(exporter)

	logProvider := log.NewLoggerProvider(
		log.WithProcessor(bsp),
		log.WithResource(resource),
	)

	return logProvider, nil
}

func getHeaderMap(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	pairs := strings.Split(headerStr, ",")

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)

		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return headers
}

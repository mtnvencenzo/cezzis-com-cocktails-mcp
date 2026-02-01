// Package aisearch provides a client for interacting with the AI Search API.
// It includes methods for retrieving cocktail details, searching for cocktails,
// and rating cocktails. The package handles authentication, request construction,
// and response parsing.
package aisearch

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// RequestEditor creates a request editor that adds OAuth bearer token
func RequestEditor() RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		// Add subscription key header
		appSettings := config.GetAppSettings()
		if appSettings.AISearchAPISubscriptionKey != "" {
			req.Header.Set("X-Key", appSettings.AISearchAPISubscriptionKey)
		}

		return nil
	}
}

// GetClient creates and returns a new AI Search API client using application settings.
// Returns an error if the client creation fails.
func GetClient() (*Client, error) {
	appSettings := config.GetAppSettings()

	if appSettings.AISearchAPIHost == "" {
		err := errors.New("AISearchAPIHost has not been configured")
		telemetry.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	httpOpts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return "[dep] aisearch-api " + r.Method + " " + r.URL.Path
		}),
	}

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport, httpOpts...),
		Timeout:   30 * time.Second,
	}

	opts := []ClientOption{
		WithHTTPClient(&httpClient),
	}

	client, err := NewClient(appSettings.AISearchAPIHost, opts...)

	if err != nil {
		telemetry.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	return client, nil
}

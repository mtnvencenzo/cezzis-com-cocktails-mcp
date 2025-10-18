// Package cocktailsapi provides a client for interacting with the Cocktails API.
// It includes methods for retrieving cocktail details, searching for cocktails,
// and rating cocktails. The package handles authentication, request construction,
// and response parsing.
package cocktailsapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// AuthenticatedRequestEditor creates a request editor that adds OAuth bearer token
func AuthenticatedRequestEditor(authManager *auth.OAuthFlowManager) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		v := ctx.Value(middleware.McpSessionIDKey)
		sid, ok := v.(string)
		if !ok || sid == "" {
			return errors.New("missing required Mcp-Session-Id header")
		}

		// Add subscription key header
		appSettings := config.GetAppSettings()
		if appSettings.CocktailsAPISubscriptionKey != "" {
			req.Header.Set("X-Key", appSettings.CocktailsAPISubscriptionKey)
		}

		// Add OAuth bearer token if authenticated
		if authManager.IsAuthenticated(ctx, sid) {
			token, err := authManager.GetAccessToken(ctx, sid)
			if err != nil {
				telemetry.Logger.Warn().Err(err).Msg("Failed to get access token")
				return err
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			telemetry.Logger.Debug().Msg("Added OAuth bearer token to request")
		}

		return nil
	}
}

// GetClient creates and returns a new cocktails API client using application settings.
// Returns an error if the client creation fails.
func GetClient() (*Client, error) {
	appSettings := config.GetAppSettings()

	if appSettings.CocktailsAPIHost == "" {
		err := errors.New("CocktailsAPIHost has not been configured")
		telemetry.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	httpOpts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return "[dep] cocktails-api " + r.Method + " " + r.URL.Path
		}),
	}

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport, httpOpts...),
		Timeout:   30 * time.Second,
	}

	opts := []ClientOption{
		WithHTTPClient(&httpClient),
	}

	client, err := NewClient(appSettings.CocktailsAPIHost, opts...)

	if err != nil {
		telemetry.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	return client, nil
}

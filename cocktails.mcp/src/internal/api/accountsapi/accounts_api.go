// Package accountsapi provides a client for interacting with the Accounts API.
// It includes methods for retrieving account details, managing user authentication,
// and handling account-related operations. The package handles authentication, request construction,
// and response parsing.
package accountsapi

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

// RequestEditor creates a request editor that adds OAuth bearer token
func RequestEditor(authManager *auth.OAuthFlowManager) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		v := ctx.Value(middleware.McpSessionIDKey)
		sid, ok := v.(string)
		if !ok || sid == "" {
			return errors.New("missing required Mcp-Session-Id header")
		}

		// Add subscription key header
		appSettings := config.GetAppSettings()
		if appSettings.AccountsAPISubscriptionKey != "" {
			req.Header.Set("X-Key", appSettings.AccountsAPISubscriptionKey)
		}

		// Add OAuth bearer token if authenticated
		if authManager.IsAuthenticated(ctx, sid) {
			token, err := authManager.GetAccessToken(ctx, sid)
			if err != nil {
				telemetry.Logger.Warn().Ctx(ctx).Err(err).Msg("Failed to get access token")
				return err
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			telemetry.Logger.Debug().Ctx(ctx).Msg("Added OAuth bearer token to request")
		}

		return nil
	}
}

// GetClient creates and returns a new accounts API client using application settings.
// Returns an error if the client creation fails.
func GetClient() (*Client, error) {
	appSettings := config.GetAppSettings()

	if appSettings.AccountsAPIHost == "" {
		err := errors.New("AccountsAPIHost has not been configured")
		telemetry.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	httpOpts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return "[dep] accounts-api " + r.Method + " " + r.URL.Path
		}),
	}

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport, httpOpts...),
		Timeout:   30 * time.Second,
	}

	opts := []ClientOption{
		WithHTTPClient(&httpClient),
	}

	client, err := NewClient(appSettings.AccountsAPIHost, opts...)

	if err != nil {
		telemetry.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	return client, nil
}

//coverage:ignore file
package testutils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/mcpserver"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints. See issue #752.
	baseURLPath = "/api-v3"
)

// Setup setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func Setup(t *testing.T) (client *cocktailsapi.Client, mux *http.ServeMux, ctx context.Context, serverURL string) {
	t.Helper()
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	ctx = context.WithValue(context.Background(), mcpserver.McpSessionIDKey, uuid.New().String())

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that. See issue #752.
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather than relative?")
		fmt.Fprintln(os.Stderr, "\tSee https://github.com/google/go-github/issues/752 for information.")
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// Create a custom transport with isolated connection pool
	transport := &http.Transport{
		// Controls connection reuse - false allows reuse, true forces new connections for each request
		DisableKeepAlives: false,
		// Maximum concurrent connections per host (active + idle)
		MaxConnsPerHost: 10,
		// Maximum idle connections maintained per host for reuse
		MaxIdleConnsPerHost: 5,
		// Maximum total idle connections across all hosts
		MaxIdleConns: 20,
		// How long an idle connection remains in the pool before being closed
		IdleConnTimeout: 20 * time.Second,
	}

	// Create HTTP client with the isolated transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	// client is the GitHub client being tested and is
	// configured to use test server.
	client, err := cocktailsapi.NewClient(config.GetAppSettings().CocktailsAPIHost)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override the HTTP client and BaseURL to use the test server.
	err = cocktailsapi.WithHTTPClient(httpClient)(client)
	if err != nil {
		t.Fatalf("Failed to set HTTP client: %v", err)
	}

	rootURL, _ := url.Parse(server.URL + baseURLPath + "/")
	err = cocktailsapi.WithBaseURL(rootURL.String())(client)
	if err != nil {
		t.Fatalf("Failed to set BaseURL: %v", err)
	}

	t.Cleanup(server.Close)

	return client, mux, ctx, server.URL
}

// TestMethod asserts that the request method is as expected.
func TestMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

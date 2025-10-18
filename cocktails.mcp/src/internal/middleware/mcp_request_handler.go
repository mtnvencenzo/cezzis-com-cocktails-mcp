package middleware

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
)

type mcpSessionKey int

const (
	// McpSessionIDKey is the context key for the MCP session ID
	McpSessionIDKey mcpSessionKey = iota
)

// McpRequestHandler is middleware that handles MCP requests.
// It supports both GET and POST methods:
//   - GET requests return a simple JSON status response for health checks.
//   - POST requests are forwarded to the provided StreamableHTTPServer for MCP processing.
//
// If the request method is neither GET nor POST, it responds with a 405 Method Not Allowed error.
//
// Parameters:
//   - next: The next HTTP handler in the chain (can be nil if not needed).
//   - streamableHTTP: The StreamableHTTPServer that processes MCP POST requests.
//
// Returns an http.Handler that implements the described behavior.
func McpRequestHandler(next http.Handler, streamableHTTP *server.StreamableHTTPServer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok", "sse":false}`))
			return
		case http.MethodPost:
			sessionID := r.Header.Get("Mcp-Session-Id")
			ctx := context.WithValue(r.Context(), McpSessionIDKey, sessionID)
			r = r.WithContext(ctx)
			streamableHTTP.ServeHTTP(w, r)
			return
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
}

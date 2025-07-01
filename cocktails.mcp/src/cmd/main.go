package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/mcp-go/server"

	"cezzis.com/cezzis-mcp-server/pkg/tools"
)

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval tools and serving requests over standard input/output or HTTP.
func main() {
	// Add a flag to choose between stdio and HTTP
	httpAddr := flag.String("http", "", "If set, serve HTTP on this address (e.g., :8080). Otherwise, use stdio.")
	flag.Parse()

	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.CocktailSearchToolHandler))
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.CocktailGetToolHandler))

	// Logging middleware for HTTP
	logMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the ResponseWriter to capture status code
			type statusRecorder struct {
				http.ResponseWriter
				status int
			}
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)
			log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.RemoteAddr, rec.status)
		})
	}

	if *httpAddr != "" {
		// HTTP mode
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Use the official streamable HTTP server for MCP
		streamableHTTP := server.NewStreamableHTTPServer(mcpServer)
		http.Handle("/mcp", logMiddleware(streamableHTTP))
		//http.Handle("/healthz", logMiddleware(http.DefaultServeMux))
		log.Printf("Serving HTTP on %s", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, nil))
	} else {
		// Stdio mode (default)
		if err := server.ServeStdio(mcpServer); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}
}

package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"cezzis.com/cezzis-mcp-server/api/cocktailsapi"
)

func main() {
	searchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cocktailsClient, err := cocktailsapi.NewClient("https://api.cezzis.com/prd/cocktails")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		martini := "mar"
		xKey := "<key-here>"
		rs, err := cocktailsClient.GetCocktailsList(context.Background(), &cocktailsapi.GetCocktailsListParams{
			FreeText: &martini,
			XKey:     &xKey,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rs.Body.Close() // Ensure the response body is closed

		bodyBytes, err := io.ReadAll(rs.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bodyBytes)
	})

	http.Handle("/", cocktailsapi.AuthMiddleware([]string{})(searchHandler))

	log.Println("Starting MCP server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}

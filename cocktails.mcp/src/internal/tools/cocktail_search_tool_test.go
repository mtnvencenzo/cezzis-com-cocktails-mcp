package tools_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"cezzis.com/cezzis-mcp-server/internal/api/aisearch"
	"cezzis.com/cezzis-mcp-server/internal/testutils"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

func Test_cocktailsearch_toolhandler_throws_on_invalid_freetext(t *testing.T) {
	// Arrange
	t.Parallel()
	testutils.LoadEnvironment("..", "..")
	_, searchClient, _, ctx, _ := testutils.Setup(t)

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "search_cocktails",
		},
		Params: mcp.CallToolParams{
			Name:      "search_cocktails",
			Arguments: map[string]interface{}{},
		},
	}

	handler := tools.NewCocktailSearchToolHandler(searchClient)

	// Act
	result, err := handler.Handle(ctx, request)
	testutils.AssertError(t, result, err, "required argument \"freeText\" not found")
}

func Test_cocktailsearch_toolhandler_returns_valid_response_for_freetext_search(t *testing.T) {
	// Arrange
	_, searchClient, mux, ctx, _ := testutils.Setup(t)

	resultRs := aisearch.CocktailsSearchRs{
		Items: []aisearch.CocktailSearchModel{
			{
				Id:               "pegu-club",
				DescriptiveTitle: "This is the pegu club",
			},
		},
	}

	jsonData, err := json.Marshal(resultRs)
	if err != nil {
		fmt.Printf("Error marshalling struct: %v\n", err)
		return
	}

	mux.HandleFunc("/api/v1/search/semantic", func(w http.ResponseWriter, r *http.Request) {
		testutils.TestMethod(t, r, "GET")
		freeText := r.URL.Query().Get("freetext")
		require.Equal(t, "Pegu Club", freeText)
		fmt.Fprint(w, string(jsonData))
	})

	testutils.LoadEnvironment("..", "..")

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "search_cocktails",
		},
		Params: mcp.CallToolParams{
			Name: "search_cocktails",
			Arguments: map[string]interface{}{
				"freeText": "Pegu Club",
			},
		},
	}

	handler := tools.NewCocktailSearchToolHandler(searchClient)

	// Act
	result, err := handler.Handle(ctx, request)
	if err != nil {
		t.Error(err)
	}

	// Assert
	// Check error
	require.False(t, result.IsError)
	require.Nil(t, err)

	// Check result
	require.NotNil(t, result)
	require.NotNil(t, result.Result)
	require.Nil(t, result.Result.Meta)
	require.Nil(t, result.Meta)

	// Check result content
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Content should be of type TextContent")

	jsonString := string(jsonData)

	require.Equal(t, jsonString, content.Text)
}

func Test_cocktailsearch_toolhandler_forwards_locale_as_accept_language(t *testing.T) {
	// Arrange
	testutils.LoadEnvironment("..", "..")
	_, searchClient, mux, ctx, _ := testutils.Setup(t)

	received := "unset"
	mux.HandleFunc("/api/v1/search/semantic", func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Get("Accept-Language")
		fmt.Fprint(w, "{}")
	})

	// Act
	result, err := tools.NewCocktailSearchToolHandler(searchClient).Handle(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "search_cocktails",
			Arguments: map[string]interface{}{"freeText": "Ginsünde", "locale": "de"},
		},
	})

	// Assert
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Equal(t, "de", received)
}

package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/testutils"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

func Test_cocktailsearch_toolhandler_throws_on_invalid_freetext(t *testing.T) {
	// Arrange
	t.Parallel()
	testutils.LoadEnvironment("..", "..")
	client, _, _ := testutils.Setup(t)

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "search_cocktails",
		},
		Params: mcp.CallToolParams{
			Name:      "search_cocktails",
			Arguments: map[string]interface{}{},
		},
	}

	handler := tools.NewCocktailSearchToolHandler(client)

	// Act
	result, err := handler.Handle(context.Background(), request)
	testutils.AssertError(t, result, err, "required argument \"freeText\" not found")
}

func Test_cocktailsearch_toolhandler_returns_valid_response_for_freetext_search(t *testing.T) {
	// Arrange
	client, mux, _ := testutils.Setup(t)

	resultRs := cocktailsapi.CocktailsListRs{
		Items: []cocktailsapi.CocktailsListModel{
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

	mux.HandleFunc("/api/v1/cocktails", func(w http.ResponseWriter, r *http.Request) {
		testutils.TestMethod(t, r, "GET")
		freeText := r.URL.Query().Get("freeText")
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

	handler := tools.NewCocktailSearchToolHandler(client)

	// Act
	result, err := handler.Handle(context.Background(), request)
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

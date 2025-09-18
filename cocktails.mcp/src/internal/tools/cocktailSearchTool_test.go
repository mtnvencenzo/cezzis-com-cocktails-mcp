package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
)

func Test_cocktailsearch_toolhandler_throws_on_invalid_freetext(t *testing.T) {
	// Arrange
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "cocktails_search",
		},
		Params: mcp.CallToolParams{
			Name:      "cocktails_search",
			Arguments: map[string]interface{}{},
		},
	}

	cocktailsAPIFactory := cocktailsapi.NewCocktailsAPIFactory()
	handler := NewCocktailSearchToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.TODO(), request)
	assertError(t, result, err, "required argument \"freeText\" not found")
}

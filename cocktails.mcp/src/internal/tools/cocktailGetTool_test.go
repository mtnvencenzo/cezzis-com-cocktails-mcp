package tools_test

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/testutils"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

func Test_cocktailget_toolhandler_returns_error_on_missing_cocktailId(t *testing.T) {
	// Arrange
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "cocktails_get",
		},
		Params: mcp.CallToolParams{
			Name:      "cocktails_get",
			Arguments: map[string]interface{}{},
		},
	}

	cocktailsAPIFactory := cocktailsapi.NewCocktailsAPIFactory()
	handler := tools.NewCocktailGetToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.Background(), request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" not found")
}

func Test_cocktailget_toolhandler_returns_error_on_invalid_cocktailId(t *testing.T) {
	// Arrange
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "cocktails_get",
		},
		Params: mcp.CallToolParams{
			Name: "cocktails_get",
			Arguments: map[string]interface{}{
				"cocktailId": "",
			},
		},
	}

	cocktailsAPIFactory := cocktailsapi.NewCocktailsAPIFactory()
	handler := tools.NewCocktailGetToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.Background(), request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" is empty")
}

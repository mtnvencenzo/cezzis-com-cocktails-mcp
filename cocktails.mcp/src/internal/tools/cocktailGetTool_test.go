package tools_test

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/test"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

func Test_cocktailget_toolhandler_throws_on_invalid_cocktailId(t *testing.T) {
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
	result, err := handler.Handle(context.TODO(), request)
	test.AssertError(t, result, err, "required argument \"cocktailId\" not found")
}

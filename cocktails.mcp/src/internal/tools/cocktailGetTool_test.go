package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func Test_CocktailGetToolHandler_throws_on_invalid_cocktailId(t *testing.T) {
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "cocktails_get",
		},
		Params: mcp.CallToolParams{
			Name:      "cocktails_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := CocktailGetToolHandler(context.TODO(), request)

	if err != nil {
		t.Errorf("Did not expect error but got one: %v", err)
	}

	if !result.IsError {
		t.Errorf("Expected error, got %v", result)
	}
}

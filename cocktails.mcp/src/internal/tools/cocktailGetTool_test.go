package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/testutils"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

func Test_cocktailget_toolhandler_returns_error_on_missing_cocktailId(t *testing.T) {
	// Arrange
	testutils.LoadEnvironment("..", "..")

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "cocktails_get",
		},
		Params: mcp.CallToolParams{
			Name:      "cocktails_get",
			Arguments: map[string]interface{}{},
		},
	}

	cocktailsAPI := testutils.NewMockCocktailsAPI()
	cocktailsAPIFactory := testutils.NewMockCocktailsAPIFactory(cocktailsAPI)
	handler := tools.NewCocktailGetToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.Background(), request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" not found")
}

func Test_cocktailget_toolhandler_returns_error_on_invalid_cocktailId(t *testing.T) {
	// Arrange
	testutils.LoadEnvironment("..", "..")

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

	cocktailsAPI := testutils.NewMockCocktailsAPI()
	cocktailsAPIFactory := testutils.NewMockCocktailsAPIFactory(cocktailsAPI)
	handler := tools.NewCocktailGetToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.Background(), request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" is empty")
}

func Test_cocktailget_toolhandler_returns_valid_response_for_cocktailId(t *testing.T) {
	// Arrange
	testutils.LoadEnvironment("..", "..")

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "cocktails_get",
		},
		Params: mcp.CallToolParams{
			Name: "cocktails_get",
			Arguments: map[string]interface{}{
				"cocktailId": "pegu-club",
			},
		},
	}

	resultRs := cocktailsapi.CocktailRs{
		Item: cocktailsapi.CocktailModel{
			Id:               "pegu-club",
			Content:          "This is the pegu club",
			DescriptiveTitle: "This is the pegu club",
		},
	}

	cocktailsAPI := testutils.NewMockCocktailsAPI()
	cocktailsAPI.SetCocktailRs(resultRs)
	cocktailsAPIFactory := testutils.NewMockCocktailsAPIFactory(cocktailsAPI)
	handler := tools.NewCocktailGetToolHandler(cocktailsAPIFactory)

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

	jsonData, _ := json.Marshal(resultRs)
	jsonString := string(jsonData)

	require.Equal(t, jsonString, content.Text)
}

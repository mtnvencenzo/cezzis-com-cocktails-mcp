package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
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
	handler := NewCocktailGetToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.TODO(), request)
	assertError(t, result, err, "required argument \"cocktailId\" not found")
}

func assertError(t *testing.T, result *mcp.CallToolResult, err error, expectedErrMsg string) {
	// Check nils
	require.Nil(t, result.Meta)
	require.NotNil(t, result)
	require.NotNil(t, result.Result)
	require.Nil(t, result.Result.Meta)

	// Check error
	require.True(t, result.IsError)
	require.NotNil(t, err)
	require.ErrorContains(t, err, expectedErrMsg)

	// Check result content for error
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Content should be of type TextContent")
	require.Equal(t, "text", textContent.Type)
	require.Equal(t, expectedErrMsg, textContent.Text)

	require.NotNil(t, textContent.Annotated)
	require.Nil(t, textContent.Annotations)
	require.Nil(t, textContent.Annotated.Annotations)
}

package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

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

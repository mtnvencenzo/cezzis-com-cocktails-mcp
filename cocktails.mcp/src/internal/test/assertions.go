// Package test provides common testing utilities for the Cezzi Cocktails MCP server.
package test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// AssertError validates error responses from MCP tool handlers.
// It checks that the error matches expected conditions and contains the expected message.
func AssertError(t *testing.T, result *mcp.CallToolResult, err error, expectedErrMsg string) {
	// Check nils
	require.Nil(t, result.Meta)
	require.NotNil(t, result)
	require.NotNil(t, result.Result)
	require.Nil(t, result.Meta)

	// Check error
	require.True(t, result.IsError)
	require.NotNil(t, err)
	require.ErrorContains(t, err, expectedErrMsg)

	// Check result content for error
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Content should be of type TextContent")
	require.Equal(t, expectedErrMsg, content.Text)
}

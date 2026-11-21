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
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/testutils"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

func Test_cocktailget_toolhandler_returns_error_on_missing_cocktailId(t *testing.T) {
	// arrange
	t.Parallel()
	testutils.LoadEnvironment("..", "..")
	client, _, _, ctx, _ := testutils.Setup(t)

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "get_cocktail",
		},
		Params: mcp.CallToolParams{
			Name:      "get_cocktail",
			Arguments: map[string]interface{}{},
		},
	}

	handler := tools.NewCocktailGetToolHandler(client)

	// Act
	result, err := handler.Handle(ctx, request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" not found")
}

func Test_cocktailget_toolhandler_returns_error_on_invalid_cocktailId(t *testing.T) {
	t.Parallel()
	testutils.LoadEnvironment("..", "..")
	client, _, _, ctx, _ := testutils.Setup(t)

	// arrange

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "get_cocktail",
		},
		Params: mcp.CallToolParams{
			Name: "get_cocktail",
			Arguments: map[string]interface{}{
				"cocktailId": "",
			},
		},
	}

	handler := tools.NewCocktailGetToolHandler(client)

	// Act
	result, err := handler.Handle(ctx, request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" is empty")
}

func Test_cocktailget_toolhandler_returns_valid_response_for_cocktailId(t *testing.T) {
	// arrange
	t.Parallel()
	testutils.LoadEnvironment("..", "..")
	client, _, mux, ctx, _ := testutils.Setup(t)

	resultRs := cocktailsapi.CocktailRs{
		Item: cocktailsapi.CocktailModel{
			Id:               "pegu-club",
			Content:          "This is the pegu club",
			DescriptiveTitle: "This is the pegu club",
		},
	}

	jsonData, err := json.Marshal(resultRs)
	if err != nil {
		fmt.Printf("Error marshalling struct: %v\n", err)
		return
	}

	mux.HandleFunc("/api/v1/cocktails/pegu-club", func(w http.ResponseWriter, r *http.Request) {
		testutils.TestMethod(t, r, "GET")
		fmt.Fprint(w, string(jsonData))
	})

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "get_cocktail",
		},
		Params: mcp.CallToolParams{
			Name: "get_cocktail",
			Arguments: map[string]interface{}{
				"cocktailId": "pegu-club",
			},
		},
	}

	handler := tools.NewCocktailGetToolHandler(client)

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

	require.Equal(t, string(jsonData), content.Text)
}

func Test_cocktailget_toolhandler_forwards_locale_as_accept_language(t *testing.T) {
	tests := []struct {
		name           string
		locale         string
		headerLanguage string
		expected       string
	}{
		{name: "locale argument", locale: "de", expected: "de"},
		{name: "locale argument wins over the request header", locale: "de", headerLanguage: "fr", expected: "de"},
		{name: "request header when there is no locale argument", headerLanguage: "fr-CA,fr;q=0.9", expected: "fr-CA,fr;q=0.9"},
		{name: "no language sends no header", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.LoadEnvironment("..", "..")
			client, _, mux, ctx, _ := testutils.Setup(t)

			if tt.headerLanguage != "" {
				ctx = context.WithValue(ctx, middleware.McpAcceptLanguageKey, tt.headerLanguage)
			}

			received := "unset"
			mux.HandleFunc("/api/v1/cocktails/pegu-club", func(w http.ResponseWriter, r *http.Request) {
				received = r.Header.Get("Accept-Language")
				fmt.Fprint(w, "{}")
			})

			arguments := map[string]interface{}{"cocktailId": "pegu-club"}
			if tt.locale != "" {
				arguments["locale"] = tt.locale
			}

			result, err := tools.NewCocktailGetToolHandler(client).Handle(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{Name: "get_cocktail", Arguments: arguments},
			})

			require.NoError(t, err)
			require.False(t, result.IsError)
			require.Equal(t, tt.expected, received)
		})
	}
}

func Test_cocktailget_toolhandler_returns_error_on_invalid_locale(t *testing.T) {
	t.Parallel()
	testutils.LoadEnvironment("..", "..")
	client, _, _, ctx, _ := testutils.Setup(t)

	result, err := tools.NewCocktailGetToolHandler(client).Handle(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_cocktail",
			Arguments: map[string]interface{}{"cocktailId": "pegu-club", "locale": "de\r\nX-Injected: 1"},
		},
	})

	require.Error(t, err)
	require.True(t, result.IsError)
	require.Contains(t, err.Error(), "not a valid language")
}

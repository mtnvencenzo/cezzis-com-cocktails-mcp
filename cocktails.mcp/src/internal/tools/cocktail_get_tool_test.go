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

func Test_cocktailget_toolhandler_returns_error_on_missing_cocktailId(t *testing.T) {
	// arrange
	t.Parallel()
	// client, mux, _ := testutils.Setup(t)

	// mux.HandleFunc("/api/v1/cocktails/%s", func(w http.ResponseWriter, r *http.Request) {
	// 	testutils.TestMethod(t, r, "GET")
	// 	fmt.Fprint(w, `{
	// 		  "sha": "s",
	// 		  "tree": [ { "type": "blob" } ],
	// 		  "truncated": true
	// 		}`)
	// })

	testutils.LoadEnvironment("..", "..")

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "get_cocktail",
		},
		Params: mcp.CallToolParams{
			Name:      "get_cocktail",
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
	// arrange
	testutils.LoadEnvironment("..", "..")

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

	cocktailsAPI := testutils.NewMockCocktailsAPI()
	cocktailsAPIFactory := testutils.NewMockCocktailsAPIFactory(cocktailsAPI)
	handler := tools.NewCocktailGetToolHandler(cocktailsAPIFactory)

	// Act
	result, err := handler.Handle(context.Background(), request)
	testutils.AssertError(t, result, err, "required argument \"cocktailId\" is empty")
}

func Test_cocktailget_toolhandler_returns_valid_response_for_cocktailId(t *testing.T) {
	// arrange
	client, mux, _ := testutils.Setup(t)

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

	testutils.LoadEnvironment("..", "..")

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

	//cocktailsAPI := testutils.NewMockCocktailsAPI()
	//cocktailsAPI.SetCocktailRs(resultRs)
	cocktailsAPIFactory := testutils.NewMockCocktailsAPIFactory(client)
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

	require.Equal(t, string(jsonData), content.Text)
}

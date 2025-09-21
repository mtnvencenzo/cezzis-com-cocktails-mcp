//coverage:ignore file

package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
)

// MockCocktailsAPI is a lightweight in-memory implementation of the
// cocktails API client used for tests. It returns simple placeholder
// responses and is intended to be wired into tests that need a
// cocktailsapi-compatible client without performing real HTTP calls.
type MockCocktailsAPI struct {
	cocktailsRs               cocktailsapi.CocktailRs
	cocktailsRsInitalized     bool
	cocktailsListRs           cocktailsapi.CocktailsListRs
	cocktailsListRsInitalized bool
}

// NewMockCocktailsAPI constructs and returns a ready-to-use
// MockCocktailsAPI value. The returned value has no internal state and
// can be copied freely.
func NewMockCocktailsAPI() MockCocktailsAPI {
	return MockCocktailsAPI{}
}

// GetCocktail implements the cocktailsapi client's GetCocktail method.
//
// In this mock it returns an empty HTTP response and a nil error. Tests
// can replace or extend this method if they need to return specific
// payloads or simulate error cases.
func (api MockCocktailsAPI) GetCocktail(ctx context.Context, id string, params *cocktailsapi.GetCocktailParams, reqEditors ...cocktailsapi.RequestEditorFn) (*http.Response, error) {
	if api.cocktailsRsInitalized {
		return createHTTPRs(api.cocktailsRs, 200, "OK")
	}

	return createHTTPRs(getGenericProblemDetails("GetCocktail"), 500, "OK")
}

// GetCocktailsList implements the cocktailsapi client's GetCocktailsList
// method.
//
// The mock returns an empty HTTP response and a nil error by default.
// Tests that require specific list data should either replace this mock
// or inspect/modify the response returned here.
func (api MockCocktailsAPI) GetCocktailsList(ctx context.Context, params *cocktailsapi.GetCocktailsListParams, reqEditors ...cocktailsapi.RequestEditorFn) (*http.Response, error) {
	if api.cocktailsListRsInitalized {
		return createHTTPRs(api.cocktailsListRs, 200, "OK")
	}

	return createHTTPRs(getGenericProblemDetails("GetCocktailsList"), 500, "OK")
}

// SetCocktailRs assigns a CocktailRs that this mock client can return
// when GetCocktail is invoked. The method returns the modified mock
// allowing fluent chaining during test setup. The provided pointer is
// stored on the copy of the mock and used by tests that inspect the
// mock's state or by a more complete mock implementation.
//
// Example:
//
//	mock := NewMockCocktailsAPI().SetCocktailRs(&cocktailsapi.CocktailRs{...})
func (api *MockCocktailsAPI) SetCocktailRs(rs cocktailsapi.CocktailRs) {
	api.cocktailsRs = rs
	api.cocktailsRsInitalized = true
}

// SetCocktailListRs assigns a CocktailsListRs that this mock client can
// return when GetCocktailsList is invoked. Like SetCocktailRs it returns
// a modified copy of the mock to support fluent construction in tests.
func (api *MockCocktailsAPI) SetCocktailListRs(rs cocktailsapi.CocktailsListRs) {
	api.cocktailsListRs = rs
	api.cocktailsListRsInitalized = true
}

func createHTTPRs(obj any, statusCode int, status string) (*http.Response, error) {
	// Create a mock response body
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	jsonString := string(jsonData)

	bodyReader := bytes.NewBufferString(jsonString)

	mockResponse := &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, status),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bodyReader), // Wrap the buffer in NopCloser to satisfy io.ReadCloser
	}

	return mockResponse, nil
}

func getGenericProblemDetails(instance string) *cocktailsapi.ProblemDetails {
	return &cocktailsapi.ProblemDetails{
		Detail:   stringPtr("Invalid mock response setup"),
		Instance: stringPtr(instance),
		Status:   int32Ptr(500),
		Title:    stringPtr("Invalid mock response setup"),
		Type:     stringPtr("Error"),
	}
}

// helper to get *string from string literal
func stringPtr(s string) *string {
	return &s
}

// helper to get *int32 from int32 literal
func int32Ptr(i int32) *int32 {
	return &i
}

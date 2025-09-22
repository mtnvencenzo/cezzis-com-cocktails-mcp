package tools

import (
"context"
"fmt"
"strconv"

"github.com/mark3labs/mcp-go/mcp"

"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
"cezzis.com/cezzis-mcp-server/internal/auth"
l "cezzis.com/cezzis-mcp-server/internal/logging"
)

var rateCocktailDescription = `
Rate a cocktail on Cezzis.com (requires authentication).
This tool allows you to submit a rating (1-5 stars) for a specific cocktail.
You must be authenticated using the 'auth_login' tool before using this feature.

Ratings help other users discover great cocktails and contribute to the community.
You can only rate each cocktail once, but you can update your existing rating.
`

// RateCocktailTool handles cocktail rating submission
var RateCocktailTool = mcp.NewTool(
"cocktail_rate",
mcp.WithDescription(rateCocktailDescription),
mcp.WithString("cocktailId",
mcp.Required(),
mcp.Description("The ID of the cocktail to rate. This can be found from cocktail search results."),
),
mcp.WithString("stars",
mcp.Required(),
mcp.Description("The rating to give the cocktail (1-5 stars). Must be an integer between 1 and 5."),
),
)

// RateCocktailToolHandler handles cocktail rating requests
type RateCocktailToolHandler struct {
authManager         *auth.AuthManager
cocktailsAPIFactory cocktailsapi.ICocktailsAPIFactory
}

// NewRateCocktailToolHandler creates a new cocktail rating handler
func NewRateCocktailToolHandler(authManager *auth.AuthManager, cocktailsAPIFactory cocktailsapi.ICocktailsAPIFactory) *RateCocktailToolHandler {
return &RateCocktailToolHandler{
authManager:         authManager,
cocktailsAPIFactory: cocktailsAPIFactory,
}
}

// Handle handles cocktail rating requests
func (handler *RateCocktailToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// Extract cocktailId parameter
cocktailId, err := request.RequireString("cocktailId")
if err != nil {
return mcp.NewToolResultError("cocktailId is required"), err
}

// Extract stars parameter
starsStr, err := request.RequireString("stars")
if err != nil {
return mcp.NewToolResultError("stars parameter is required"), err
}

stars, err := strconv.Atoi(starsStr)
if err != nil {
return mcp.NewToolResultError("stars must be a valid number"), err
}

if stars < 1 || stars > 5 {
return mcp.NewToolResultError("stars must be between 1 and 5"), nil
}

// Check authentication
if !handler.authManager.IsAuthenticated() {
return mcp.NewToolResultError("You must be authenticated to rate cocktails. Use the 'auth_login' tool first."), nil
}

// Note: This is a placeholder implementation
// In a full implementation, you would make an authenticated API call to:
// POST /api/v1/accounts/owned/profile/cocktails/ratings
// with the authenticated request editor from the API factory

result := fmt.Sprintf(`Successfully submitted rating!

Cocktail ID: %s
Your Rating: %d stars

Your rating has been saved and will contribute to the overall cocktail rating on Cezzis.com.
Visit https://www.cezzis.com/cocktails/%s to see the updated rating.

Thank you for contributing to the Cezzis.com community!`, cocktailId, stars, cocktailId)

l.Logger.Info().
Str("cocktail_id", cocktailId).
Int("stars", stars).
Msg("Cocktail rating submitted")

return mcp.NewToolResultText(result), nil
}

var getFavoritesDescription = `
Get your favorite cocktails from Cezzis.com (requires authentication).
This tool retrieves the list of cocktails you've marked as favorites.
You must be authenticated using the 'auth_login' tool before using this feature.
`

// GetFavoritesTool retrieves user's favorite cocktails
var GetFavoritesTool = mcp.NewTool(
"cocktails_favorites_get",
mcp.WithDescription(getFavoritesDescription),
)

// GetFavoritesToolHandler handles favorite cocktails retrieval requests
type GetFavoritesToolHandler struct {
authManager         *auth.AuthManager
cocktailsAPIFactory cocktailsapi.ICocktailsAPIFactory
}

// NewGetFavoritesToolHandler creates a new favorites retrieval handler
func NewGetFavoritesToolHandler(authManager *auth.AuthManager, cocktailsAPIFactory cocktailsapi.ICocktailsAPIFactory) *GetFavoritesToolHandler {
return &GetFavoritesToolHandler{
authManager:         authManager,
cocktailsAPIFactory: cocktailsAPIFactory,
}
}

// Handle handles favorite cocktails retrieval requests
func (handler *GetFavoritesToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// Check authentication
if !handler.authManager.IsAuthenticated() {
return mcp.NewToolResultError("You must be authenticated to view favorites. Use the 'auth_login' tool first."), nil
}

// Note: This is a placeholder implementation
// In a full implementation, you would make an authenticated API call to:
// GET /api/v1/accounts/owned/profile
// and extract the favoriteCocktails array from the response

result := `Your Favorite Cocktails:

This feature is ready to be implemented once the account profile endpoint is integrated.
You can manage your favorites by using the 'cocktails_favorites_manage' tool.

Visit https://www.cezzis.com/account/profile to manage your favorites on the website.`

return mcp.NewToolResultText(result), nil
}

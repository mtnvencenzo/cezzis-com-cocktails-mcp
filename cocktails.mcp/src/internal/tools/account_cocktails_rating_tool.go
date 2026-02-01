package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/accountsapi"
	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var rateCocktailDescription = `
This tool allows you to submit a rating (1-5 stars) for a specific cocktail.  The rating will be associated with your Cezzis.com account
and be reflected in the cocktail's overall rating on the website. Ratings help other users discover great cocktails and contribute to the community. 

To use this tool, provide a valid cocktail ID and your star rating as an integer between 1 and 5. A cocktail ID can be obtained from the
results of the 'cocktail_search' tool, from the cocktail details from the get_cocktail tool, or from the cocktails page of a cocktail on Cezzis.com.

If you provide an invalid rating or have already rated the cocktail, the tool will return an error.

You must be authenticated using the 'authentication_login_flow' tool prior to using this feature. Furthermore, You must have a valid and active
mcp session, the session identifier from the original initialization request must be present in the request to this tool via the Mcp-Session-Id header.
If the response returns an error about authentication, please run the 'authentication_login_flow' tool first.
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
	authManager *auth.OAuthFlowManager
	client      *accountsapi.Client
}

// NewRateCocktailToolHandler creates a new cocktail rating handler
func NewRateCocktailToolHandler(authManager *auth.OAuthFlowManager, client *accountsapi.Client) *RateCocktailToolHandler {
	return &RateCocktailToolHandler{
		authManager: authManager,
		client:      client,
	}
}

// Handle handles cocktail rating requests
func (handler *RateCocktailToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	v := ctx.Value(middleware.McpSessionIDKey)
	sessionID, ok := v.(string)
	if !ok || sessionID == "" {
		err := errors.New("missing required Mcp-Session-Id header")
		return mcp.NewToolResultError(err.Error()), err
	}

	// Extract cocktailId parameter
	cocktailID, err := request.RequireString("cocktailId")
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
	if !handler.authManager.IsAuthenticated(ctx, sessionID) {
		return mcp.NewToolResultError("You must be authenticated to rate cocktails. Use the 'authentication_login_flow' tool first."), nil
	}

	telemetry.Logger.Info().
		Ctx(ctx).
		Str("cocktail_id", cocktailID).
		Int("stars", stars).
		Msg("MCP rating cocktail: " + cocktailID)

	// default to a safe deadline if none present
	callCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		callCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	rs, callErr := handler.client.RateCocktailV1AccountsOwnedProfileCocktailsRatingsPost(
		callCtx,
		accountsapi.RateCocktailV1AccountsOwnedProfileCocktailsRatingsPostJSONRequestBody{
			CocktailId: cocktailID,
			CurrentRatings: accountsapi.AccountCocktailRatingModel{
				OneStars:    0,
				TwoStars:    0,
				ThreeStars:  0,
				FourStars:   0,
				FiveStars:   0,
				TotalStars:  0,
				Rating:      0,
				RatingCount: 0,
			},
			Stars: stars,
		}, accountsapi.RequestEditor(handler.authManager))

	if callErr != nil {
		telemetry.Logger.Err(callErr).Ctx(ctx).Msg("MCP Error rating cocktail: " + cocktailID)
		return mcp.NewToolResultError(callErr.Error()), callErr
	}

	defer func() {
		if closeErr := rs.Body.Close(); closeErr != nil {
			telemetry.Logger.Warn().Ctx(ctx).Msg(fmt.Sprintf("MCP Warning: failed to close response body: %v", closeErr))
		}
	}()

	bodyBytes, readErr := io.ReadAll(rs.Body)
	if readErr != nil {
		telemetry.Logger.Err(readErr).Ctx(ctx).Msg("MCP Error getting cocktail rs body: " + cocktailID)
		return mcp.NewToolResultError(readErr.Error()), readErr
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	if bodyString == "" {
		telemetry.Logger.Warn().Ctx(ctx).Msg("MCP Warning: empty response body when rating cocktail: " + cocktailID)
	}

	appsettings := config.GetAppSettings()

	result := fmt.Sprintf(`Successfully submitted rating!

Cocktail ID: %[1]s
Your Rating: %[2]d stars

Your rating has been saved and will contribute to the overall cocktail rating on Cezzis.com.
Visit %[3]s/cocktails/%[1]s to see the updated rating.

Thank you for contributing to the Cezzis.com community!`, cocktailID, stars, appsettings.CezzisBaseURL)

	telemetry.Logger.Info().
		Ctx(ctx).
		Str("cocktail_id", cocktailID).
		Int("stars", stars).
		Msg("Cocktail rating submitted")

	return mcp.NewToolResultText(result), nil
}

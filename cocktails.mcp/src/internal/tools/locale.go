package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/middleware"
)

const (
	localeArgument            = "locale"
	maxAcceptLanguageLength   = 128
	localeArgumentDescription = "The language to use for the response, for example 'de' or 'de-AT,de;q=0.9,en;q=0.8' (standard Accept-Language syntax). " +
		"Use the language the user is writing in. When omitted, or when the language is not supported, the default language (English) is used."
)

// localeOption adds the optional locale argument to a tool.
var localeOption = mcp.WithString(localeArgument, mcp.Description(localeArgumentDescription))

var acceptLanguagePattern = regexp.MustCompile(`^[A-Za-z0-9*;=.,_ -]+$`)

// resolveAcceptLanguage returns the Accept-Language value to send to the upstream APIs.
// The tool's locale argument wins, then the Accept-Language header of the MCP request, and when neither is present
// nil is returned so the upstream API uses its default language.
func resolveAcceptLanguage(ctx context.Context, request mcp.CallToolRequest) (*string, error) {
	if argument := strings.TrimSpace(request.GetString(localeArgument, "")); argument != "" {
		if !isValidAcceptLanguage(argument) {
			return nil, fmt.Errorf("argument %q is not a valid language, for example 'de' or 'de-AT,de;q=0.9'", localeArgument)
		}

		return &argument, nil
	}

	// The ambient header is not something the caller chose per call, so an unusable value is ignored rather than failing the tool.
	if header, ok := ctx.Value(middleware.McpAcceptLanguageKey).(string); ok {
		if header = strings.TrimSpace(header); isValidAcceptLanguage(header) {
			return &header, nil
		}
	}

	return nil, nil
}

func isValidAcceptLanguage(value string) bool {
	return value != "" && len(value) <= maxAcceptLanguageLength && acceptLanguagePattern.MatchString(value)
}

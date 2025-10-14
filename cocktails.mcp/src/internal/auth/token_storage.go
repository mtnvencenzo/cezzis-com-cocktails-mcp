package auth

import (
	"fmt"
	"time"

	"cezzis.com/cezzis-mcp-server/internal/logging"
	"cezzis.com/cezzis-mcp-server/internal/repos"
)

// TokenStorage handles secure storage and retrieval of OAuth tokens
type TokenStorage struct {
	repo *repos.CosmosAccountRepository
}

// NewTokenStorage creates a new token storage instance
func NewTokenStorage() (*TokenStorage, error) {
	repo, err := repos.NewCosmosAccountRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create token repository: %w", err)
	}

	return &TokenStorage{
		repo: repo,
	}, nil
}

// SaveToken saves tokens to encrypted storage
func (ts *TokenStorage) SaveToken(sessionID string, tokens *TokenResponse) error {
	sessionToken := &repos.SessionToken{
		ID:           sessionID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
		TokenType:    tokens.TokenType,
		Scope:        tokens.Scope,
	}

	err := ts.repo.SaveToken(sessionID, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to save tokens to repository: %w", err)
	}

	logging.Logger.Info().Msg("Tokens saved to repository")
	return nil
}

// GetToken retrieves tokens from encrypted storage
func (ts *TokenStorage) GetToken(sessionID string) (*TokenResponse, error) {

	sessionToken, err := ts.repo.GetToken(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokens from repository: %w", err)
	}

	if sessionToken == nil {
		logging.Logger.Info().Msg("No tokens found in repository")
		return nil, nil
	}

	// Check if tokens are expired
	if time.Now().After(sessionToken.ExpiresAt) {
		logging.Logger.Warn().Msg("Stored tokens are expired")
		return nil, nil
	}

	tokens := &TokenResponse{
		AccessToken:  sessionToken.AccessToken,
		RefreshToken: sessionToken.RefreshToken,
		ExpiresIn:    int(time.Until(sessionToken.ExpiresAt).Seconds()),
		TokenType:    sessionToken.TokenType,
		Scope:        sessionToken.Scope,
	}

	logging.Logger.Info().Msg("Tokens loaded from storage")
	return tokens, nil
}

// ClearTokens removes stored tokens
func (ts *TokenStorage) ClearTokens(sessionID string) error {
	if err := ts.repo.ClearTokens(sessionID); err != nil {
		return fmt.Errorf("failed to clear tokens from repository: %w", err)
	}

	logging.Logger.Info().Msg("Tokens cleared from storage")
	return nil
}

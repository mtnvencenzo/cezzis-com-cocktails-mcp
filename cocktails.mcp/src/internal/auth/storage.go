package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

// TokenStorage handles secure storage and retrieval of OAuth tokens
type TokenStorage struct {
	storageFile   string
	encryptionKey []byte
}

// StoredTokens represents tokens stored on disk
type StoredTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
}

// NewTokenStorage creates a new token storage instance
func NewTokenStorage(storageDir string) (*TokenStorage, error) {
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	storageFile := filepath.Join(storageDir, ".cezzis_tokens.enc")

	// Generate or load encryption key
	encryptionKey, err := getOrCreateEncryptionKey(storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	return &TokenStorage{
		storageFile:   storageFile,
		encryptionKey: encryptionKey,
	}, nil
}

// SaveTokens saves tokens to encrypted storage
func (ts *TokenStorage) SaveTokens(tokens *TokenResponse) error {
	storedTokens := &StoredTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
		TokenType:    tokens.TokenType,
		Scope:        tokens.Scope,
	}

	data, err := json.Marshal(storedTokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	encryptedData, err := ts.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt tokens: %w", err)
	}

	if err := os.WriteFile(ts.storageFile, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	l.Logger.Info().Msg("Tokens saved to encrypted storage")
	return nil
}

// LoadTokens loads tokens from encrypted storage
func (ts *TokenStorage) LoadTokens() (*TokenResponse, error) {
	if _, err := os.Stat(ts.storageFile); os.IsNotExist(err) {
		return nil, nil // No tokens stored
	}

	encryptedData, err := os.ReadFile(ts.storageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	data, err := ts.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tokens: %w", err)
	}

	var storedTokens StoredTokens
	if err := json.Unmarshal(data, &storedTokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	// Check if tokens are expired
	if time.Now().After(storedTokens.ExpiresAt) {
		l.Logger.Warn().Msg("Stored tokens are expired")
		return nil, nil
	}

	tokens := &TokenResponse{
		AccessToken:  storedTokens.AccessToken,
		RefreshToken: storedTokens.RefreshToken,
		ExpiresIn:    int(time.Until(storedTokens.ExpiresAt).Seconds()),
		TokenType:    storedTokens.TokenType,
		Scope:        storedTokens.Scope,
	}

	l.Logger.Info().Msg("Tokens loaded from storage")
	return tokens, nil
}

// ClearTokens removes stored tokens
func (ts *TokenStorage) ClearTokens() error {
	if err := os.Remove(ts.storageFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove tokens file: %w", err)
	}
	l.Logger.Info().Msg("Tokens cleared from storage")
	return nil
}

func (ts *TokenStorage) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(ts.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (ts *TokenStorage) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(ts.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func getOrCreateEncryptionKey(storageDir string) ([]byte, error) {
	keyFile := filepath.Join(storageDir, ".cezzis_key")

	// Try to load existing key
	if key, err := os.ReadFile(keyFile); err == nil {
		if len(key) == 32 { // AES-256
			return key, nil
		}
	}

	// Generate new key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Save key
	if err := os.WriteFile(keyFile, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encryption key: %w", err)
	}

	return key, nil
}

// Package repos provides data repository implementations
// for managing session tokens in Azure Cosmos DB.
//
// The package includes:
//   - CosmosAccountRepository: A repository for storing, retrieving, and managing
//     session tokens in an Azure Cosmos DB instance.
package repos

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/logging"
)

// SessionToken represents a user's session token
type SessionToken struct {
	ID           string    `json:"id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
}

// CosmosAccountRepository manages session tokens in Azure Cosmos DB
type CosmosAccountRepository struct {
	client        *azcosmos.Client
	dbName        string
	containerName string
}

// NewCosmosAccountRepository creates a new Cosmos DB repository instance
func NewCosmosAccountRepository() (*CosmosAccountRepository, error) {
	appSettings := config.GetAppSettings()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	if appSettings.CosmosConnectionString != "" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient := &http.Client{Transport: tr, Timeout: 30 * time.Second}

		// Create azcore.ClientOptions with the custom HTTP client
		clientOptions := azcore.ClientOptions{
			Transport: httpClient,
		}

		client, err := azcosmos.NewClientFromConnectionString(appSettings.CosmosConnectionString, &azcosmos.ClientOptions{ClientOptions: clientOptions})
		if err != nil {
			return nil, err
		}
		return &CosmosAccountRepository{
			client:        client,
			dbName:        appSettings.CosmosDatabaseName,
			containerName: appSettings.CosmosContainerName,
		}, nil
	}

	client, err := azcosmos.NewClient(appSettings.CosmosAccountEndpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	return &CosmosAccountRepository{
		client:        client,
		dbName:        appSettings.CosmosDatabaseName,
		containerName: appSettings.CosmosContainerName,
	}, nil
}

// ClearTokens removes tokens from storage
func (r *CosmosAccountRepository) ClearTokens(sessionID string) error {
	containerClient, err := r.getContainer()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rs, err := containerClient.DeleteItem(ctx, azcosmos.NewPartitionKeyString(sessionID), sessionID, nil)

	if err != nil {
		return err
	}

	if rs.RawResponse.StatusCode == 404 {
		logging.Logger.Warn().Str("sessionID", sessionID).Msg("No tokens found to clear")
		return nil
	}

	return nil
}

// SaveToken saves tokens to storage
func (r *CosmosAccountRepository) SaveToken(sessionID string, sessionToken *SessionToken) error {
	containerClient, err := r.getContainer()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bytes, err := json.Marshal(sessionToken)
	if err != nil {
		return err
	}

	rs, err := containerClient.UpsertItem(ctx, azcosmos.NewPartitionKeyString(sessionID), bytes, nil)

	if err != nil {
		return err
	}

	if rs.RawResponse.StatusCode != 200 && rs.RawResponse.StatusCode != 201 {
		return fmt.Errorf("failed to save token, status code: %d", rs.RawResponse.StatusCode)
	}

	return nil
}

// GetToken retrieves tokens from storage
func (r *CosmosAccountRepository) GetToken(sessionID string) (*SessionToken, error) {
	containerClient, err := r.getContainer()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rs, err := containerClient.ReadItem(ctx, azcosmos.NewPartitionKeyString(sessionID), sessionID, nil)

	if err != nil {
		return nil, err
	}

	if rs.RawResponse.StatusCode == 404 {
		logging.Logger.Warn().Str("sessionId", sessionID).Msg("No token found")
		return nil, nil
	}

	var sessionToken SessionToken
	if err := json.Unmarshal(rs.Value, &sessionToken); err != nil {
		return nil, err
	}

	return &sessionToken, nil
}

func (r *CosmosAccountRepository) getContainer() (*azcosmos.ContainerClient, error) {
	databaseClient, err := r.client.NewDatabase(r.dbName)
	if err != nil {
		return nil, err
	}

	containerClient, err := databaseClient.NewContainer(r.containerName)
	if err != nil {
		return nil, err
	}

	return containerClient, nil
}

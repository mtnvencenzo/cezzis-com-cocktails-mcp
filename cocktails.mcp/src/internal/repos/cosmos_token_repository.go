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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/environment"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
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

	client, err := GetCosmosClient()
	if err != nil {
		return nil, err
	}

	repo := &CosmosAccountRepository{
		client:        client,
		dbName:        appSettings.CosmosDatabaseName,
		containerName: appSettings.CosmosContainerName,
	}

	return repo, nil
}

// ClearTokens removes tokens from storage
func (r *CosmosAccountRepository) ClearTokens(ctx context.Context, sessionID string) error {
	containerClient, err := r.getContainer()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ctx, span := telemetry.Tracer.Start(ctx, "CosmosDB.DeleteItem")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "cosmosdb"),
		attribute.String("db.operation", "delete_item"),
		attribute.String("db.container", r.containerName),
		attribute.String("cosmosdb.item_id", sessionID),
	)

	rs, err := containerClient.DeleteItem(ctx, azcosmos.NewPartitionKeyString(sessionID), sessionID, nil)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(500, "Cosmos DB delete failed")
		return err
	}

	if rs.RawResponse.StatusCode == 404 {
		span.SetStatus(400, "Cosmos DB delete item not found")
		telemetry.Logger.Warn().Str("sessionID", sessionID).Msg("No tokens found to clear")
		return nil
	}

	span.SetStatus(codes.Code(rs.RawResponse.StatusCode), "Cosmos DB delete succeeded")

	return nil
}

// SaveToken saves tokens to storage
func (r *CosmosAccountRepository) SaveToken(ctx context.Context, sessionID string, sessionToken *SessionToken) error {
	containerClient, err := r.getContainer()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	bytes, err := json.Marshal(sessionToken)
	if err != nil {
		return err
	}

	ctx, span := telemetry.Tracer.Start(ctx, "CosmosDB.UpsertItem")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "cosmosdb"),
		attribute.String("db.operation", "upsert_item"),
		attribute.String("db.container", r.containerName),
		attribute.String("cosmosdb.item_id", sessionID),
	)

	rs, err := containerClient.UpsertItem(ctx, azcosmos.NewPartitionKeyString(sessionID), bytes, nil)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(500, "Cosmos DB upsert failed")
		return err
	}

	if rs.RawResponse.StatusCode != 200 && rs.RawResponse.StatusCode != 201 {
		span.SetStatus(codes.Code(rs.RawResponse.StatusCode), "Cosmos DB upsert failed")
		return fmt.Errorf("failed to save token, status code: %d", rs.RawResponse.StatusCode)
	}

	span.SetStatus(codes.Code(rs.RawResponse.StatusCode), "Cosmos DB upsert succeeded")

	return nil
}

// GetToken retrieves tokens from storage
func (r *CosmosAccountRepository) GetToken(ctx context.Context, sessionID string) (*SessionToken, error) {
	containerClient, err := r.getContainer()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ctx, span := telemetry.Tracer.Start(ctx, "CosmosDB.ReadItem")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "cosmosdb"),
		attribute.String("db.operation", "read_item"),
		attribute.String("db.container", r.containerName),
		attribute.String("cosmosdb.item_id", sessionID),
	)

	rs, err := containerClient.ReadItem(ctx, azcosmos.NewPartitionKeyString(sessionID), sessionID, nil)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(500, "Cosmos DB read failed")
		return nil, err
	}

	if rs.RawResponse.StatusCode == 404 {
		span.SetStatus(404, "Cosmos DB item not found")
		telemetry.Logger.Warn().Str("sessionId", sessionID).Msg("No token found")
		return nil, nil
	}

	span.SetStatus(codes.Code(rs.RawResponse.StatusCode), "Cosmos DB read succeeded")

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

// GetCosmosClient creates and returns a Cosmos DB client
func GetCosmosClient() (*azcosmos.Client, error) {
	appSettings := config.GetAppSettings()

	if appSettings.CosmosConnectionString != "" {
		// -------------------------------------------------------------------------------------
		// These options are for development purposes only. Since not planning
		// on using connection strings in real environments, changing the options to work with local
		// cosmos emulator.  (Note: disabling cert checks!)
		// -------------------------------------------------------------------------------------
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: environment.IsLocalEnv()}, // Skip TLS verification for local emulator (development only)
		}
		httpClient := &http.Client{Transport: tr, Timeout: 30 * time.Second}

		clientOptions := azcore.ClientOptions{
			Transport: httpClient,
		}

		client, err := azcosmos.NewClientFromConnectionString(appSettings.CosmosConnectionString, &azcosmos.ClientOptions{ClientOptions: clientOptions})
		if err != nil {
			return nil, err
		}

		return client, nil
	}

	// Use DefaultAzureCredential for authentication
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	httpOpts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return "CocktailsAPI " + operation
		}),
	}

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport, httpOpts...),
		Timeout:   30 * time.Second,
	}

	clientOptions := azcore.ClientOptions{
		Transport: httpClient,
		Retry: policy.RetryOptions{
			MaxRetries:    3,
			RetryDelay:    1 * time.Second,
			MaxRetryDelay: 3 * time.Second,
		},
	}

	client, err := azcosmos.NewClient(appSettings.CosmosAccountEndpoint, cred, &azcosmos.ClientOptions{ClientOptions: clientOptions})

	if err != nil {
		return nil, err
	}

	return client, nil
}

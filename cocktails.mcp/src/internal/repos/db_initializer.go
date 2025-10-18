package repos

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// InitializeDatabase ensures that the Cosmos DB and container exist, creating them if necessary.
func InitializeDatabase(ctx context.Context) error {
	appSettings := config.GetAppSettings()

	client, err := GetCosmosClient()
	if err != nil {
		telemetry.Logger.Err(err).Msg("Failed to get client to initialize database")
		return err
	}

	if err := createDtabaseIfNotExists(ctx, client, appSettings.CosmosDatabaseName); err != nil {
		telemetry.Logger.Err(err).Msg("Failed to create database")
		return err
	}

	dbClient, err := client.NewDatabase(appSettings.CosmosDatabaseName)
	if err != nil {
		telemetry.Logger.Err(err).Msg("Failed to get database client")
		return err
	}

	if err := createContainerIfNotExists(ctx, dbClient, appSettings.CosmosDatabaseName, appSettings.CosmosContainerName); err != nil {
		telemetry.Logger.Err(err).Msg("Failed to create container")
		return err
	}

	_, err = dbClient.NewContainer(appSettings.CosmosContainerName)
	if err != nil {
		telemetry.Logger.Err(err).Msg("Failed to get container client")
		return err
	}

	return nil
}

func createDtabaseIfNotExists(ctx context.Context, client *azcosmos.Client, dbName string) error {
	dbrs, err := client.CreateDatabase(ctx, azcosmos.DatabaseProperties{
		ID: dbName,
	}, nil)
	if err != nil {
		var respErr *azcore.ResponseError

		if errors.As(err, &respErr) {
			switch respErr.StatusCode {
			case 409:
				telemetry.Logger.Info().Msg("Database already exists")
			case 201:
				telemetry.Logger.Info().Msg("Database created")
			default:
				telemetry.Logger.Err(err).Msg("Failed to create database")
				return err
			}
		}
	} else if dbrs.RawResponse != nil && dbrs.RawResponse.StatusCode == 201 {
		telemetry.Logger.Info().Msg("Database created")
	}

	return nil
}

func createContainerIfNotExists(ctx context.Context, client *azcosmos.DatabaseClient, dbName, containerName string) error {
	ccrs, err := client.CreateContainer(ctx, azcosmos.ContainerProperties{
		ID: containerName,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{"/id"},
			Kind:  azcosmos.PartitionKeyKindHash,
		},
	}, nil)
	if err != nil {
		var respErr *azcore.ResponseError

		if errors.As(err, &respErr) {
			switch respErr.StatusCode {
			case 409:
				telemetry.Logger.Info().Msg("Container already exists")
			case 201:
				telemetry.Logger.Info().Msg("Container created")
			default:
				telemetry.Logger.Err(err).Msg("Failed to create container")
				return err
			}
		}
	} else if ccrs.RawResponse != nil && ccrs.RawResponse.StatusCode == 201 {
		telemetry.Logger.Info().Msg("Container created")
	}

	return nil
}

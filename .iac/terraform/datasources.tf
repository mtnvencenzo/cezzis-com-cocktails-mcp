data "azurerm_client_config" "current" {}

data "azurerm_application_insights" "appi" {
  name                = "appi-${var.sub}-${var.region}-${var.global_environment}-${var.global_domain}-${var.sequence}"
  resource_group_name = "rg-${var.sub}-${var.region}-${var.global_environment}-${var.global_domain}-${var.sequence}"
}

data "azurerm_resource_group" "cocktails_resource_group" {
  name = "rg-${var.sub}-${var.region}-${var.environment}-${var.domain}-${var.sequence}"
}

data "azurerm_resource_group" "global_shared_resource_group" {
  name = "rg-${var.sub}-${var.region}-${var.global_environment}-shared-${var.sequence}"
}

data "azurerm_container_registry" "shared_acr" {
  name                = "acr${var.sub}${var.region}${var.global_environment}shared${var.sequence}"
  resource_group_name = data.azurerm_resource_group.global_shared_resource_group.name
}

data "azurerm_api_management" "apim_shared" {
  name                = "apim-${var.sub}-${var.region}-${var.global_environment}-shared-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.global_shared_resource_group.name
}

data "azurerm_container_app_environment" "cae_shared" {
  resource_group_name = data.azurerm_resource_group.global_shared_resource_group.name
  name                = "cae-${var.sub}-${var.region}-${var.global_environment}-shared-${var.sequence}"
}

data "azurerm_key_vault" "cocktails_keyvault" {
  name                = "kv-${var.sub}-${var.region}-${var.environment}-${var.shortdomain}-${var.short_sequence}"
  resource_group_name = data.azurerm_resource_group.cocktails_resource_group.name
}

data "azurerm_key_vault" "global_keyvault" {
  name                = "kv-${var.sub}-${var.region}-${var.global_environment}-shared-${var.short_sequence}"
  resource_group_name = data.azurerm_resource_group.global_shared_resource_group.name
}

data "azurerm_key_vault_secret" "antiforgery_signing_secret" {
  name         = "antiforgery-signing-secret"
  key_vault_id = data.azurerm_key_vault.cocktails_keyvault.id
}

data "azurerm_container_app" "cocktails_api" {
  name                = "aca-${var.sub}-${var.region}-${var.environment}-${var.domain}api-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.cocktails_resource_group.name
}

data "azurerm_container_app" "accounts_api" {
  name                = "aca-${var.sub}-${var.region}-${var.environment}-accountsapi-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.cocktails_resource_group.name
}

data "azurerm_container_app" "aisearch_api" {
  name                = "aca-${var.sub}-${var.region}-${var.environment}-aisearchapi-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.cocktails_resource_group.name
}

data "azurerm_key_vault_secret" "otel_collector_api_key" {
  name         = "otel-collector-api-key-1"
  key_vault_id = data.azurerm_key_vault.global_keyvault.id
}

data "azurerm_key_vault_secret" "postgres_username" {
  name         = "postgres-admin-username"
  key_vault_id = data.azurerm_key_vault.global_keyvault.id
}

data "azurerm_key_vault_secret" "postgres_password" {
  name         = "postgres-admin-password"
  key_vault_id = data.azurerm_key_vault.global_keyvault.id
}

data "azurerm_postgresql_flexible_server" "postgres" {
  name                = "psqlfs-${var.sub}-eus2-${var.global_environment}-shared-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.global_shared_resource_group.name
}

data "azurerm_container_app" "otel_collector" {
  name                = "aca-${var.sub}-${var.region}-${var.global_environment}-otelcol-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.global_shared_resource_group.name
}
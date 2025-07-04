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

data "azurerm_resource_group" "global_network_resource_group" {
  name = "rg-${var.sub}-${var.region}-${var.global_environment}-network-${var.sequence}"
}

data "azurerm_virtual_network" "global_virtual_network" {
  name                = "vnet-${var.sub}-${var.region}-${var.global_environment}-network-${var.sequence}"
  resource_group_name = data.azurerm_resource_group.global_network_resource_group.name
}

data "azurerm_subnet" "container_app_environment_subnet" {
  name                 = "snet-${var.sub}-${var.region}-${var.global_environment}-${var.global_domain}containers-${var.sequence}"
  resource_group_name  = data.azurerm_resource_group.global_network_resource_group.name
  virtual_network_name = "vnet-${var.sub}-${var.region}-${var.global_environment}-network-${var.sequence}"
}

data "azurerm_resource_group" "cocktails_global_resource_group" {
  name = "rg-${var.sub}-${var.region}-${var.global_environment}-${var.domain}-${var.sequence}"
}

data "azurerm_dns_zone" "cezzis_dns_zone" {
  name                = "cezzis.com"
  resource_group_name = data.azurerm_resource_group.cocktails_global_resource_group.name
}

data "azurerm_key_vault_secret" "cocktails_api_mcp_subscription_key" {
  name         = "cocktails_api_mcp_subscription_primary_key"
  key_vault_id = data.azurerm_key_vault.cocktails_keyvault.id
}
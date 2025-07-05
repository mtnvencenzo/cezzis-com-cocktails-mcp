resource "random_password" "cocktails_mcp_localusage_subscription_keys" {
  count   = 2
  length  = 24
  special = true
  upper   = false
}

module "apim_cocktails_mcp" {
  source = "git::ssh://git@github.com/mtnvencenzo/Terraform-Modules.git//modules/apim-api"
  providers = {
    azurerm = azurerm
  }

  tags = local.tags

  environment                    = var.environment
  domain                         = var.domain
  name_discriminator             = "mcp"
  tenant_id                      = data.azurerm_client_config.current.tenant_id
  keyvault_apimhostkey_secret_id = azurerm_key_vault_secret.cocktails_mcp_apimhostkey.id

  apim_instance = {
    id                  = data.azurerm_api_management.apim_shared.id
    name                = data.azurerm_api_management.apim_shared.name
    resource_group_name = data.azurerm_api_management.apim_shared.resource_group_name
    principal_id        = data.azurerm_api_management.apim_shared.identity[0].principal_id
  }

  api = {
    version      = 1
    service_fqdn = module.aca_cocktails_mcp.ingress_fqdn
    ingress_fqdn = module.aca_cocktails_mcp.ingress_fqdn
  }

  # key_vault = {
  #   id   = data.azurerm_key_vault.cocktails_keyvault.id
  #   name = data.azurerm_key_vault.cocktails_keyvault.name
  # }

  application_insights = {
    id                  = data.azurerm_application_insights.appi.id
    instrumentation_key = data.azurerm_application_insights.appi.instrumentation_key
  }

  allowed_origins = []

  subscriptions = [
    {
      name          = "${title(var.environment)} Cezzis.Com Standard ${title(var.domain)} Mcp Subscription"
      primary_key   = random_password.cocktails_mcp_cezzis_com_subscription_keys[0].result
      secondary_key = random_password.cocktails_mcp_cezzis_com_subscription_keys[1].result
    },
    {
      name          = "${title(var.environment)} Cezzis.Com Devops ${title(var.domain)} Mcp Subscription"
      primary_key   = random_password.cocktails_mcp_devops_subscription_keys[0].result
      secondary_key = random_password.cocktails_mcp_devops_subscription_keys[1].result
    },
    {
      name          = "${title(var.environment)} Cezzis.Com LocalUsage ${title(var.domain)} Mcp Subscription"
      primary_key   = random_password.cocktails_mcp_localusage_subscription_keys[0].result
      secondary_key = random_password.cocktails_mcp_localusage_subscription_keys[1].result
    }
  ]

  operations = [
    {
      display_name        = "Get Health Status"
      method              = "GET"
      url_template        = "/healthz"
      description         = "Get health"
      success_status_code = 200
      security_type       = "anonymous"
    }
  ]
}
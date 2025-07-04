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
    #audience     = module.api_b2c_tenant.cocktails_api_app_registration_client_id
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
      display_name        = "Get Cocktail"
      method              = "GET"
      url_template        = "/cocktails/{id}"
      description         = "Gets a cocktail recipe by id"
      success_status_code = 200
      security_type       = "anonymous"

      template_parameters = [
        {
          name = "id"
          type = "string"
        }
      ]
    },
    {
      display_name        = "Get Cocktails List"
      method              = "GET"
      url_template        = "/cocktails"
      description         = "Gets a list of cocktails"
      success_status_code = 200
      security_type       = "anonymous"
    },
    {
      display_name        = "Put Cocktails List"
      method              = "PUT"
      url_template        = "/cocktails"
      description         = "Puts the list of cocktails"
      success_status_code = 204
      security_type       = "anonymous"
    },
    {
      display_name        = "Get Privacy Policy"
      method              = "GET"
      url_template        = "/legal/documents/privacy-policy"
      description         = "Gets the privacy policy"
      success_status_code = 200
      security_type       = "anonymous"
    },
    {
      display_name        = "Get Terms Of Service"
      method              = "GET"
      url_template        = "/legal/documents/terms-of-service"
      description         = "Gets the terms of service"
      success_status_code = 200
      security_type       = "anonymous"
    },
    {
      display_name        = "Get Ping"
      method              = "GET"
      url_template        = "/health/ping"
      description         = "Get health ping"
      success_status_code = 200
      security_type       = "anonymous"
    },
    {
      display_name        = "Get Version"
      method              = "GET"
      url_template        = "/health/version"
      description         = "Get version"
      success_status_code = 200
      security_type       = "anonymous"
    }
  ]
}
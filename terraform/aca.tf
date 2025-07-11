module "aca_cocktails_mcp" {
  source = "git::ssh://git@github.com/mtnvencenzo/Terraform-Modules.git//modules/container-app"
  providers = {
    azurerm = azurerm
  }

  sub                          = var.sub
  region                       = var.region
  environment                  = var.environment
  domain                       = var.domain
  name_discriminator           = "mcp"
  sequence                     = var.sequence
  tenant_id                    = data.azurerm_client_config.current.tenant_id
  resource_group_name          = data.azurerm_resource_group.cocktails_resource_group.name
  resource_group_location      = data.azurerm_resource_group.cocktails_resource_group.location
  container_app_environment_id = data.azurerm_container_app_environment.cae_shared.id
  ingress_target_port          = "8080"
  startup_probe_relative_url   = "/healthz"

  tags = local.tags

  container_registry = {
    id           = data.azurerm_container_registry.shared_acr.id
    login_server = data.azurerm_container_registry.shared_acr.login_server
  }

  key_vault = {
    id   = data.azurerm_key_vault.cocktails_keyvault.id
    name = data.azurerm_key_vault.cocktails_keyvault.name
  }

  container = {
    name       = "cocktails-mcp"
    image_name = "cocktailsmcp"
    image_tag  = var.image_tag
    cpu        = 0.25
    memory     = "0.5Gi"
  }

  secrets = [
    {
      name                  = "apim-host-key"
      key_vault_secret_name = azurerm_key_vault_secret.cocktails_mcp_apimhostkey.name
    },
    {
      name                  = "apim-cocktails-api-subscription-key"
      key_vault_secret_name = data.azurerm_key_vault_secret.cocktails_api_mcp_subscription_key.name
    }
  ]

  env_vars = [
    {
      name  = "APPLICATIONINSIGHTS_CONNECTION_STRING"
      value = data.azurerm_application_insights.appi.connection_string
    },
    {
      name  = "APPLICATIONINSIGHTS_INSTRUMENTATIONKEY"
      value = data.azurerm_application_insights.appi.instrumentation_key
    },
    {
      name  = "ApplicationInsightsAgent_EXTENSION_VERSION"
      value = "~2"
    },
    {
      name  = "COCKTAILS_API_HOST"
      value = "https://api.cezzis.com/prd/cocktails"
    },
    {
      name  = "AZUREAD_B2C_INSTANCE"
      value = "https://login.cezzis.com"
    },
    {
      name  = "AZUREAD_B2C_DOMAIN"
      value = "cezzis.onmicrosoft.com"
    },
    {
      name  = "AZUREAD_B2C_USERFLOW"
      value = "B2C_1_SignInSignUp_Policy"
    },
    {
      name  = "ENV"
      value = var.environment
    },
    {
      name  = "LOG_LEVEL"
      value = "info"
    }
  ]

  env_secret_vars = [
    {
      name        = "CocktailsMcp__ApimHostKey"
      secret_name = "apim-host-key"
    },
    {
      name        = "COCKTAILS_API_XKEY"
      secret_name = "apim-cocktails-api-subscription-key"
    }
  ]
}

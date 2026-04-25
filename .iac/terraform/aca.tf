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
  resource_group_name          = data.azurerm_resource_group.cocktails_resource_group.name
  resource_group_location      = data.azurerm_resource_group.cocktails_resource_group.location
  container_app_environment_id = data.azurerm_container_app_environment.cae_shared.id
  ingress_target_port          = "8080"
  startup_probe_relative_url   = "/mcp/v1/healthz"
  readiness_probe_relative_url = "/mcp/v1/healthz"
  liveness_probe_relative_url  = "/mcp/v1/healthz"

  tags = local.tags

  container_registry = {
    id           = data.azurerm_container_registry.shared_acr.id
    login_server = data.azurerm_container_registry.shared_acr.login_server
  }

  key_vaults = {
    cocktails = {
      id   = data.azurerm_key_vault.cocktails_keyvault.id
      name = data.azurerm_key_vault.cocktails_keyvault.name
    }
    global = {
      id   = data.azurerm_key_vault.global_keyvault.id
      name = data.azurerm_key_vault.global_keyvault.name
    }
  }


  container = {
    name       = "cocktails-mcp"
    image_name = "cocktailsmcp"
    image_tag  = var.image_tag
    cpu        = 0.25
    memory     = "0.5Gi"
  }

  env_vars = [
    {
      name  = "ACCOUNTS_API_HOST"
      value = "https://${data.azurerm_container_app.accounts_api.ingress[0].fqdn}"
    },
    {
      name  = "COCKTAILS_API_HOST"
      value = "https://${data.azurerm_container_app.cocktails_api.ingress[0].fqdn}"
    },
    {
      name  = "AISEARCH_API_HOST"
      value = "https://${data.azurerm_container_app.aisearch_api.ingress[0].fqdn}"
    },
    {
      name  = "AUTH0_ACCOUNTS_API_AUDIENCE"
      value = var.auth0_accounts_api_audience
    },
    {
      name  = "AUTH0_NATIVE_CLIENT_ID"
      value = var.auth0_native_client_id
    },
    {
      name  = "AUTH0_DOMAIN"
      value = var.auth0_domain
    },
    {
      name  = "AUTH0_SCOPES"
      value = var.auth0_scopes
    },
    {
      name  = "CEZZIS_BASE_URL"
      value = var.app_url
    },
    {
      name  = "INIT_DELAY_SECONDS"
      value = var.init_delay_seconds
    },
    {
      name  = "INIT_JOB_ENABLED"
      value = var.init_job_enabled
    },
    {
      name  = "LOG_LEVEL"
      value = var.log_level
    },
    {
      name  = "OTLP_ENDPOINT"
      value = "http://${data.azurerm_container_app.otel_collector.name}:4318"
    },
    {
      name  = "OTLP_HEADERS"
      value = "Authorization=Bearer ${data.azurerm_key_vault_secret.otel_collector_api_key.value}"
    },
    {
      name  = "OTLP_INSECURE"
      value = var.otlp_insecure
    },
    {
      name  = "OTLP_LOG_ENABLED"
      value = var.otlp_log_enabled
    },
    {
      name  = "OTLP_METRICS_ENABLED"
      value = var.otlp_metrics_enabled
    },
    {
      name  = "OTLP_TRACE_ENABLED"
      value = var.otlp_trace_enabled
    },
    {
      name  = "PORT"
      value = var.port
    },
    {
      name  = "POSTGRES_DB"
      value = var.postgres_db
    },
    {
      name  = "POSTGRES_HOST"
      value = data.azurerm_postgresql_flexible_server.postgres.fqdn
    },
    {
      name  = "POSTGRES_PORT"
      value = var.postgres_port
    },
    {
      name  = "POSTGRES_USER"
      value = var.postgres_user
    },
    {
      name  = "ENV"
      value = var.environment
    },
  ]

  secrets = [
    {
      name                  = "apim-host-key"
      key_vault_secret_name = azurerm_key_vault_secret.cocktails_mcp_apimhostkey.name
      key_vault_key         = "cocktails"
    },
    {
      name                  = "antiforgery-signing-secret"
      key_vault_secret_name = data.azurerm_key_vault_secret.antiforgery_signing_secret.name
      key_vault_key         = "cocktails"
    },
    {
      name                  = "postgres-password"
      key_vault_secret_name = data.azurerm_key_vault_secret.postgres_password.name
      key_vault_key         = "global"
    },
    {
      name                  = "otel-collector-api-key"
      key_vault_secret_name = "otel-collector-api-key-1"
      key_vault_key         = "global"
    },
  ]

  env_secret_vars = [
    {
      name        = "APIM_HOST_KEY"
      secret_name = "apim-host-key"
    },
    {
      name        = "POSTGRES_PASSWORD"
      secret_name = "postgres-password"
    }
  ]
}

module "apim_cocktails_mcp" {
  source = "git::ssh://git@github.com/mtnvencenzo/Terraform-Modules.git//modules/apim-api"
  providers = {
    azurerm = azurerm
  }

  tags                           = local.tags
  environment                    = var.environment
  domain                         = var.domain
  name_discriminator             = "mcp"
  keyvault_apimhostkey_secret_id = azurerm_key_vault_secret.cocktails_mcp_apimhostkey.id
  subscription_required          = false
  backend_url_override           = "https://${module.aca_cocktails_mcp.ingress_fqdn}"
  service_url_override           = "https://${module.aca_cocktails_mcp.ingress_fqdn}"

  apim_instance = {
    id                  = data.azurerm_api_management.apim_shared.id
    name                = data.azurerm_api_management.apim_shared.name
    resource_group_name = data.azurerm_api_management.apim_shared.resource_group_name
  }

  api = {
    version      = 1
    service_fqdn = module.aca_cocktails_mcp.ingress_fqdn
    ingress_fqdn = module.aca_cocktails_mcp.ingress_fqdn
  }

  application_insights = {
    id                  = data.azurerm_application_insights.appi.id
    instrumentation_key = data.azurerm_application_insights.appi.instrumentation_key
  }

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
      policy_xml_content  = local.apim_anonomous_operation_policy
    },
    {
      display_name        = "POST MCP"
      method              = "POST"
      url_template        = "/mcp"
      description         = "POST MCP"
      success_status_code = 200
      policy_xml_content  = local.apim_anonomous_operation_policy
    },
    {
      display_name        = "GET MCP"
      method              = "GET"
      url_template        = "/mcp"
      description         = "GET MCP"
      success_status_code = 200
      policy_xml_content  = local.apim_anonomous_operation_policy
    }
  ]
}
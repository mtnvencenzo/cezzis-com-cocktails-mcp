module "aca_cocktails_mcp_availability_test" {
  source = "git::ssh://git@github.com/mtnvencenzo/Terraform-Modules.git//modules/appinsights-availability-test"
  providers = {
    azurerm = azurerm
  }

  sub                     = var.sub
  region                  = var.region
  environment             = var.environment
  domain                  = var.domain
  name_discriminator      = "mcp"
  sequence                = var.sequence
  resource_group_name     = data.azurerm_application_insights.appi.resource_group_name
  location                = data.azurerm_application_insights.appi.location
  application_insights_id = data.azurerm_application_insights.appi.id
  description             = "Availability test for the cocktails MCP"
  http_url                = "https://${var.cezzis_site_hostname}/${var.environment}/${var.domain}/mcp/v1/healthz"

  create_alert = false

  headers = [
    {
      name  = "X-Key"
      value = random_password.cocktails_mcp_devops_subscription_keys[0].result
    }
  ]
}

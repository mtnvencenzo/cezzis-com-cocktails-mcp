module "mcp_ciam_tenant" {
  source             = "./ciam_tenant"
  environment        = var.environment
  region             = var.region
  domain             = var.domain
  tenant_id          = var.ciam_tenant_id
  tenant_domain_name = var.ciam_tenant_domain_name
  sub                = var.sub
  sequence           = var.sequence
  ciam_tenant_name   = var.ciam_tenant_name

  cocktails_api_scope_account_read  = var.cocktails_api_scope_account_read
  cocktails_api_scope_account_write = var.cocktails_api_scope_account_write
  cocktails_api_client_id           = var.cocktails_api_client_id

  tags = local.tags

  providers = {
    azuread = azuread
    azurerm = azurerm
  }
}
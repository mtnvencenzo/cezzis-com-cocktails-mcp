resource "random_password" "cocktails_mcp_cezzis_com_subscription_keys_onprem" {
  count   = 2
  length  = 24
  special = true
  upper   = false
}

resource "azurerm_key_vault_secret" "cocktails_mcp_cezzis_com_subscription_primary_key_onprem" {
  name         = "cocktails-mcp-cezzis-com-subscription-primary-key"
  value        = random_password.cocktails_mcp_cezzis_com_subscription_keys_onprem[0].result
  key_vault_id = data.azurerm_key_vault.cocktails_keyvault_onprem.id
  tags         = local.tags
}

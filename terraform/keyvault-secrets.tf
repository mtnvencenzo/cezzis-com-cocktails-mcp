# ----------------------------------------
# Backend apim host key
# ----------------------------------------

resource "random_password" "cocktails_mcp_apimhostkey" {
  length  = 32
  special = true
  upper   = true
}

resource "azurerm_key_vault_secret" "cocktails_mcp_apimhostkey" {
  name         = "cocktails-mcp-apimhostkey"
  value        = random_password.cocktails_mcp_apimhostkey.result
  key_vault_id = data.azurerm_key_vault.cocktails_keyvault.id
}



# ----------------------------------------
# Subscription Cezzis.com apim primary key
# ----------------------------------------
resource "random_password" "cocktails_mcp_cezzis_com_subscription_keys" {
  count   = 2
  length  = 24
  special = true
  upper   = false
}

resource "azurerm_key_vault_secret" "cocktails_mcp_cezzis_com_subscription_primary_key" {
  name         = "cocktails-mcp-cezzis-com-subscription-primary-key"
  value        = random_password.cocktails_mcp_cezzis_com_subscription_keys[0].result
  key_vault_id = data.azurerm_key_vault.cocktails_keyvault.id
  tags         = local.tags
}


# ----------------------------------------
# Subscription Devops apim primary key
# ----------------------------------------
resource "random_password" "cocktails_mcp_devops_subscription_keys" {
  count   = 2
  length  = 24
  special = true
  upper   = false
}

resource "azurerm_key_vault_secret" "cocktails_mcp_devops_subscription_primary_key" {
  name         = "cocktails-mcp-devops-subscription-primary-key"
  value        = random_password.cocktails_mcp_devops_subscription_keys[0].result
  key_vault_id = data.azurerm_key_vault.cocktails_keyvault.id
  tags         = local.tags
}
resource "random_uuid" "cocktails_mcp_read_own_account_scope_id" {}
resource "random_uuid" "cocktails_mcp_write_own_account_scope_id" {}

resource "azuread_application" "cocktails_mcp_app_registration" {
  display_name                   = "appr-${var.sub}-${var.region}-${var.environment}-${var.domain}mcp-${var.sequence}"
  sign_in_audience               = "AzureADMyOrg"
  fallback_public_client_enabled = true
  identifier_uris                = ["https://${var.tenant_domain_name}/${var.domain}mcp"]

  web {
    redirect_uris = [
      "http://localhost:6098/callback",                    # For local development
      "https://aca-${var.sub}-${var.region}-${var.environment}-${var.domain}mcp-${var.sequence}.azurecontainerapps.io/callback",  # For production
    ]

    implicit_grant {
      access_token_issuance_enabled = true
      id_token_issuance_enabled     = true
    }
  }

  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      # This parameter is managed by `azuread_application_identifier_uri`.
      # Details: https://github.com/hashicorp/terraform-provider-azuread/issues/428#issuecomment-1788737766
      identifier_uris,
    ]
  }

  // access to scopes available from the cocktails api app registration
  required_resource_access {
    resource_app_id = var.cocktails_api_client_id

    resource_access {
      id   = var.cocktails_api_scope_account_read # read own account
      type = "Scope"
    }

    resource_access {
      id   = var.cocktails_api_scope_account_write # write own account
      type = "Scope"
    }
  }

  required_resource_access {
    resource_app_id = data.azuread_application_published_app_ids.well_known.result.MicrosoftGraph

    resource_access {
      id   = azuread_service_principal.msgraph.app_role_ids["Application.ReadWrite.All"]
      type = "Role"
    }

    resource_access {
      id   = azuread_service_principal.msgraph.app_role_ids["User.ReadWrite.All"]
      type = "Role"
    }

    # Terraform datasource of well_known IDs doesn't contain openid and offline_access
    resource_access {
      id   = "37f7f235-527c-4136-accd-4a02d197296e" # openid
      type = "Scope"
    }

    resource_access {
      id   = "7427e0e9-2fba-42fe-b0c0-848c9e6a8182" # offline_access
      type = "Scope"
    }
  }
}

resource "azuread_service_principal" "cocktails_mcp_app_registration_serviceprincipal" {
  client_id                    = azuread_application.cocktails_mcp_app_registration.client_id
  app_role_assignment_required = false
  description                  = "Service principal for the ${azuread_application.cocktails_mcp_app_registration.display_name}"
}

resource "azuread_app_role_assignment" "cocktails_mcp_app_registration_serviceprincipal_app_rw_all_role_assignment" {
  app_role_id         = azuread_service_principal.msgraph.app_role_ids["Application.ReadWrite.All"]
  principal_object_id = azuread_service_principal.cocktails_mcp_app_registration_serviceprincipal.object_id
  resource_object_id  = azuread_service_principal.msgraph.object_id
}

resource "azuread_service_principal_password" "cocktails_mcp_app_registration_graph_api_secret" {
  service_principal_id = azuread_service_principal.cocktails_mcp_app_registration_serviceprincipal.id
  display_name         = "cocktails-mcp-graph-client-secret"
  end_date_relative    = "2yr"
}
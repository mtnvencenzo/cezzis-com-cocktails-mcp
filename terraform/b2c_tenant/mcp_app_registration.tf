resource "azuread_application" "cocktails_mcp_app_registration" {
  display_name     = "appr-${var.sub}-${var.region}-${var.environment}-${var.domain}mcp-${var.sequence}"
  sign_in_audience = "AzureADandPersonalMicrosoftAccount"
  # identifier_uris removed - not needed for client apps

  # api block is needed for requested_access_token_version
  api {
    requested_access_token_version = 2
  }

  public_client {
    redirect_uris = [
      "http://localhost:6097/callback",
    ]
  }

  lifecycle {
    prevent_destroy = false
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

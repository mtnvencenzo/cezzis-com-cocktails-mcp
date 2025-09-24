terraform {
  required_providers {

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=4.24.0"
    }

    azuread = {
      source  = "hashicorp/azuread"
      version = "=3.1.0"
    }
  }

  backend "azurerm" {}
}

provider "azurerm" {
  features {}
}

provider "azuread" {
  use_oidc      = true
  tenant_id     = var.ciam_tenant_id
  client_id     = var.ciam_terraform_app_registration_client_id
  client_secret = var.ciam_terraform_app_registration_client_secret
}
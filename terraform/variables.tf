variable "domain" {
  type        = string
  description = "Domain name. Use only lowercase letters and numbers"
  default     = "cocktails"
}

variable "shortdomain" {
  type        = string
  description = "Short domain name. Use only lowercase letters and numbers"
  default     = "cockti"
}

variable "owner" {
  type    = string
  default = "cocktails.mcp"
}

variable "product" {
  type    = string
  default = "cezzis"
}

variable "environment" {
  type        = string
  description = "Environment name, e.g. 'dev' or 'stage' or 'prod'"
}

variable "region" {
  type    = string
  default = "eus"
}

variable "sub" {
  type        = string
  description = "Subscription short identitifer to be used in resource naming"
  default     = "vec"
}

variable "sequence" {
  type        = string
  description = "The sequence number of the resource typically starting with 001"
  default     = "001"
}
variable "short_sequence" {
  type        = string
  description = "The short sequence number of the resource typically starting with 1"
  default     = "1"
}

variable "global_environment" {
  type        = string
  description = "The global environment name"
  default     = "glo"
}

variable "global_domain" {
  type        = string
  description = "The global domain name"
  default     = "shared"
}

variable "ciam_terraform_app_registration_client_id" {
  type = string
}

variable "ciam_terraform_app_registration_client_secret" {
  type = string
}

variable "image_tag" {
  type = string
}

variable "ciam_tenant_id" {
  type        = string
  description = "The ciam tenant id for this environment"
}

variable "ciam_tenant_name" {
  type        = string
  description = "The ciam tenant name (typeically the first part of the hostname)"
}

variable "ciam_tenant_domain_name" {
  type        = string
  description = "The ciam tenant domain name for this environment"
}

variable "ciam_signin_policy" {
  type        = string
  description = "The main signin policy for the cocktails webapp"
  default     = "sisu-p"
}

variable "login_subdomain" {
  type = string
}

variable "cocktails_api_scope_account_read" {
  type = string
}

variable "cocktails_api_scope_account_write" {
  type = string
}

variable "cocktails_api_client_id" {
  type = string
}
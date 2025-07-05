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

variable "b2c_tenant_id" {
  type        = string
  description = "The b2c tenant id for this environment"
}

variable "b2c_terraform_app_registration_client_id" {
  type = string
}

variable "b2c_terraform_app_registration_client_secret" {
  type = string
}

variable "image_tag" {
  type = string
}

variable "allowed_origins" {
  type    = list(string)
  default = []
}

variable "b2c_tenant_id" {
  type        = string
  description = "The b2c tenant id for this environment"
}

variable "b2c_tenant_name" {
  type        = string
  description = "The b2c tenant name (typeically the first part of the hostname)"
}

variable "b2c_signin_policy" {
  type        = string
  description = "The main signin policy for the cocktails webapp"
  default     = "B2C_1_SignInSignUp_Policy"
}
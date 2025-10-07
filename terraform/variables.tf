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

variable "image_tag" {
  type = string
}

# Auth0 Configuration Variables
variable "auth0_domain" {
  type        = string
  description = "Auth0 domain (e.g., your-domain.auth0.com)"
}

variable "auth0_audience" {
  type        = string
  description = "Auth0 API identifier/audience"
}

variable "auth0_naive_client_id" {
  type        = string
  description = "Auth0 frontend/SPA client ID for Swagger/Scalar UI"
}

variable "auth0_scopes" {
  type        = string
  description = "Auth0 scopes for the application"
  default     = "openid offline_access profile email read:owned-account write:owned-account"
}
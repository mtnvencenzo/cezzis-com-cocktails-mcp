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

variable "app_url" {
  type        = string
  description = "The public URL of the application (e.g. https://www.cezzis.com)"
}

variable "auth0_domain" {
  type = string
}

variable "auth0_accounts_api_audience" {
  type = string
}

variable "auth0_native_client_id" {
  type = string
}

variable "auth0_scopes" {
  type = string
}

variable "port" {
  type        = string
  description = "The port the application listens on"
  default     = "8080"
  validation {
    condition     = can(regex("^\\d{1,5}$", var.port)) && tonumber(var.port) > 0 && tonumber(var.port) <= 65535
    error_message = "The port must be a valid integer between 1 and 65535."
  }
}

variable "postgres_database_name" {
  type = string
}

variable "postgres_port" {
  type        = string
  description = "The port for the PostgreSQL database"
  validation {
    condition     = can(regex("^\\d{1,5}$", var.postgres_port)) && tonumber(var.postgres_port) > 0 && tonumber(var.postgres_port) <= 65535
    error_message = "The PostgreSQL port must be a valid integer between 1 and 65535."
  }
}

variable "postgres_user" {
  type = string
}

variable "allowed_origins" {
  type    = list(string)
  default = []
}

variable "init_delay_seconds" {
  type        = number
  description = "Number of seconds to delay before starting the application, allowing dependent services to become healthy"
  default     = 30
}

variable "init_job_enabled" {
  type        = bool
  description = "Whether to enable the init job that delays the application start"
  default     = true
}

variable "log_level" {
  type        = string
  description = "The log level for the application (e.g. 'debug', 'info', 'warn', 'error')"
  default     = "info"
}

variable "otlp_insecure" {
  type        = bool
  description = "Whether to allow insecure connections to the OTLP endpoint"
  default     = false
}

variable "otlp_log_enabled" {
  type        = bool
  description = "Whether to enable OTLP logging"
  default     = true
}

variable "otlp_metrics_enabled" {
  type        = bool
  description = "Whether to enable OTLP metrics"
  default     = false
}

variable "otlp_trace_enabled" {
  type        = bool
  description = "Whether to enable OTLP tracing"
  default     = true
}

variable "exposed_port" {
  type        = string
  description = "The port to expose for the container app"
  validation {
    condition     = can(regex("^\\d{1,5}$", var.exposed_port)) && tonumber(var.exposed_port) > 0 && tonumber(var.exposed_port) <= 65535
    error_message = "The exposed port must be a valid integer between 1 and 65535."
  }
}

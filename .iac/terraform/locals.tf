locals {
  tags = {
    Product     = var.product
    Environment = var.environment
    Application = var.domain
    Owner       = var.owner
  }

  apim_antiforgery_operation_policy = <<XML
      <policies>
        <inbound>
          <include-fragment fragment-id="${var.environment}-cezzis-antiforgery-jwtsignature-policy" />
          <set-backend-service backend-id="${var.environment}-${var.domain}-mcp-backend" />
          <include-fragment fragment-id="${var.environment}-${var.domain}-mcp-cors-policy" />
        </inbound>
        <backend><base /></backend>
        <outbound><base /></outbound>
        <on-error><base /></on-error>
      </policies>
    XML

  apim_anonymous_operation_policy = <<XML
      <policies>
        <inbound>
          <set-backend-service backend-id="${var.environment}-${var.domain}-mcp-backend" />
          <include-fragment fragment-id="${var.environment}-${var.domain}-mcp-cors-policy" />
        </inbound>
        <backend><base /></backend>
        <outbound><base /></outbound>
        <on-error><base /></on-error>
      </policies>
    XML

  apim_anonymous_docs_operation_policy = <<XML
      <policies>
        <inbound>
          <set-backend-service backend-id="${var.environment}-${var.domain}-mcp-docs-backend" />
        </inbound>
        <backend><base /></backend>
        <outbound><base /></outbound>
        <on-error><base /></on-error>
      </policies>
    XML
}
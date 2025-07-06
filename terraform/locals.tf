locals {
  tags = {
    Product     = var.product
    Environment = var.environment
    Application = var.domain
    Owner       = var.owner
  }

  apim_anonomous_operation_policy = <<XML
      <policies>
        <inbound>
          <set-backend-service backend-id="${var.environment}-${var.domain}-mcp-backend" />
        </inbound>
        <backend><base /></backend>
        <outbound><base /></outbound>
        <on-error><base /></on-error>
      </policies>
    XML
}
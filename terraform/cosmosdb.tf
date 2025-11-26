# module "cocktails_cosmos_containers" {
#   source = "git::ssh://git@github.com/mtnvencenzo/Terraform-Modules.git//modules/cosmos-db-sql-containers"

#   resource_group_name   = data.azurerm_resource_group.global_shared_resource_group.name
#   cosmosdb_account_name = data.azurerm_cosmosdb_account.cosmosdb_account.name
#   database_name         = data.azurerm_cosmosdb_sql_database.cosmosdb_shared_db.name

#   containers = [
#     {
#       name                  = "cezzis-${var.environment}-device-auth"
#       partition_key_paths   = ["/id"]
#       partition_key_version = 1
#       indexing_mode         = "consistent"
#       indexing_path         = "/*"
#       unique_key_paths      = []
#     }
#   ]
# }

# module "cocktails_cosmos_role_assignments" {
#   source = "git::ssh://git@github.com/mtnvencenzo/Terraform-Modules.git//modules/cosmos-db-sql-role-assignments"

#   resource_group_name   = data.azurerm_resource_group.global_shared_resource_group.name
#   cosmosdb_account_name = data.azurerm_cosmosdb_account.cosmosdb_account.name
#   database_id           = data.azurerm_cosmosdb_sql_database.cosmosdb_shared_db.id

#   database_role_assignments = [
#     {
#       name               = "54d5c255-8508-4ea2-941e-7044fe5db968" # must be a uuid
#       role_definition_id = data.azurerm_cosmosdb_sql_role_definition.cosmosdb_reader_role.id
#       principal_id       = module.aca_cocktails_mcp.managed_identity_principal_id
#     },
#     {
#       name               = "57c0645e-bc81-43cd-9607-2377e2660d2a" # must be a uuid
#       role_definition_id = data.azurerm_cosmosdb_sql_role_definition.cosmosdb_contributor_role.id
#       principal_id       = module.aca_cocktails_mcp.managed_identity_principal_id
#     }
#   ]
# }
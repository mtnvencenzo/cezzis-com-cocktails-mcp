environment      = "prd"
postgres_db      = "cezzis-cocktailsmcp-db-loc"
postgres_port    = "5432"
postgres_use_tls = true
allowed_origins = [
  "https://www.cezzis.com",
  "https://cezzis.com",
  "http://localhost:4000",
  "https://localhost:4001"
]
app_url = "https://www.cezzis.com"

# Auth0 Configuration
auth0_domain                = "login.cezzis.com"
auth0_accounts_api_audience = "https://cezzis-accounts-api"
auth0_native_client_id      = "6pNnUIHVbyakv27TtICiLyAQbXYxNjOQ"
auth0_scopes                = "openid offline_access profile email read:owned-account write:owned-account"
exposed_port                = "8080"
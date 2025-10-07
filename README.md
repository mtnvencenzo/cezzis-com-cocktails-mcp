# Cezzis.com Cocktails MCP Server

> Part of the broader Cezzis.com digital experience for discovering and sharing cocktail recipes with a broad community of cocktail enthusiasts and aficionados.

[![CI](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/actions/workflows/cezzis-mcp-cicd.yaml/badge.svg?branch=main)](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/actions/workflows/cezzis-mcp-cicd.yaml)
[![Release](https://img.shields.io/github/v/release/mtnvencenzo/cezzis-com-cocktails-mcp?include_prereleases)](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/releases)
[![License](https://img.shields.io/badge/license-Proprietary-lightgrey)](LICENSE)
![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)
[![Last commit](https://img.shields.io/github/last-commit/mtnvencenzo/cezzis-com-cocktails-mcp?branch=main)](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/commits/main)
[![Issues](https://img.shields.io/github/issues/mtnvencenzo/cezzis-com-cocktails-mcp)](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/issues)
[![Docs](https://img.shields.io/badge/docs-MCP-blue)](https://modelcontextprotocol.io)
[![Project](https://img.shields.io/badge/project-Cezzis.com%20Cocktails-181717?logo=github&logoColor=white)](https://github.com/users/mtnvencenzo/projects/2)
[![Website](https://img.shields.io/badge/website-cezzis.com-2ea44f?logo=google-chrome&logoColor=white)](https://www.cezzis.com)

An MCP (Model Context Protocol) server that gives AI agents secure, first‑class access to Cezzis.com cocktail data. It provides high‑level tools for searching cocktails, retrieving detailed recipes and metadata, authenticating users, and submitting ratings. The server runs over HTTP only and exposes a streamable MCP endpoint.

## 🧩 Cezzis.com Project Ecosystem

This server works alongside several sibling repositories:

- **cocktails-mcp** (this repo) – Model Context Protocol services that expose cocktail data to AI agents
- [**cocktails-api**](https://github.com/mtnvencenzo/cezzis-com-cocktails-api) – ASP.NET Core backend and REST API consumed by the site and integrations
- [**cocktails-web**](https://github.com/mtnvencenzo/cezzis-com-cocktails-web) – React SPA for the public experience
- [**cocktails-common**](https://github.com/mtnvencenzo/cezzis-com-cocktails-common) – Shared libraries and utilities reused across frontends, APIs, and tooling
- [**cocktails-images**](https://github.com/mtnvencenzo/cezzis-com-cocktails-images) *(private)* – Source of curated cocktail imagery and CDN assets
- [**cocktails-shared-infra**](https://github.com/mtnvencenzo/cezzis-com-cocktails-shared-infra) – Terraform compositions specific to the cocktails platform
- [**shared-infrastructure**](https://github.com/mtnvencenzo/shared-infrastructure) – Global Terraform 
modules that underpin multiple Cezzis.com workloads

## 📚 MCP Tools

The server exposes the following MCP tools:

### cocktails_search
- Purpose: Search cocktails by natural language query
- Parameters:
  - `freeText` (string, required): Search terms (name, ingredients, style)
- Returns: Array of cocktails with IDs, titles, images, and summaries

### cocktails_get
- Purpose: Retrieve full details for a specific cocktail
- Parameters:
  - `cocktailId` (string, required): ID from search results
- Returns: Full recipe with ingredients, instructions, images, ratings, and notes

### auth_login
- Purpose: Initiate login using Auth0 Device Authorization flow
- Parameters: none
- Returns: Verification URL and user code to complete in your browser

### auth_status
- Purpose: Check if you’re authenticated
- Parameters: none
- Returns: Text status

### auth_logout
- Purpose: Log out and clear stored tokens
- Parameters: none
- Returns: Text confirmation

### cocktail_rate
- Purpose: Rate a cocktail (requires authentication)
- Parameters:
  - `cocktailId` (string, required)
  - `stars` (string, required, 1–5)
- Returns: Text confirmation of submitted rating

### HTTP Endpoints
- `GET /healthz` – Health check
- `GET /version` – Version info
- `GET|POST /mcp` – Streamable MCP endpoint over HTTP


## ☁️ Cloud-Native Footprint (Azure)

Infrastructure is provisioned with Terraform (`/terraform`) and deployed into Azure using shared modules:

- **Azure Container Apps** – Hosts the MCP service (HTTP mode) with HTTPS ingress
- **Azure API Management** – Optional façade when exposing HTTP endpoints; routes and policies managed via Terraform
- **Azure Container Registry** – Stores container images published from CI/CD
- **Azure Key Vault** – Holds secrets (Cezzis API subscription keys, telemetry keys)
- **Azure Monitor / Application Insights** – Telemetry collection (logs/traces)
- **Shared Infrastructure Modules** – Sourced from the reusable Terraform modules repo for consistency

## 🛠️ Technology Stack

### Core
- Language: Go 1.25+
- Protocol: Model Context Protocol over HTTP (streamable)
- Server: Lightweight MCP server with tool registry and health/version endpoints
- Logging: zerolog (structured JSON logs)

### Integrations
- **Cezzis.com Cocktails API**: Upstream data source (requires subscription key)
- **Azure AI Search**: Powers semantic/lucene queries in the upstream API
- **Application Insights**: Optional telemetry via instrumentation key

### Authentication & Security
- API Access: `COCKTAILS_API_XKEY` subscription key injected via env/Key Vault
- Auth0 OAuth 2.1 / OIDC: End‑user authentication for personalized features (e.g., ratings)
- Secrets: Managed via environment files locally and Azure Key Vault in cloud
- Transport: HTTP/HTTPS for MCP endpoint

## 🏗️ Project Structure

```text
cocktails.mcp/
├── src/
│   ├── cmd/                    # Application entry point
│   ├── internal/
│   │   ├── api/               # Generated API client code
│   │   ├── config/            # Configuration management
│   │   ├── logging/           # Structured logging helpers
│   │   ├── middleware/        # HTTP middleware (HTTP mode)
│   │   ├── server/            # MCP server and protocol wiring
│   │   ├── testutils/         # Testing utilities
│   │   └── tools/             # MCP tool implementations
│   ├── .env                   # Environment configuration (local)
│   └── go.mod                 # Go module definition
├── dist/                      # Build outputs
└── terraform/                 # Azure resources (ACA, APIM, Key Vault, etc.)
```

## 🚀 Development Setup

1) Prerequisites
   - Go 1.25.1 or newer
   - Make (build automation)
   - Optional: Docker (container builds), Azure CLI / Terraform (infrastructure)

2) Install Dependencies
   ```bash
   make tidy
   ```

3) Environment Setup
  Create a `.env` file in `./cocktails.mcp/src/`:

  ```env
  # Required: Cezzis.com API Configuration
  COCKTAILS_API_HOST=https://api.cezzis.com/prd/cocktails
  COCKTAILS_API_XKEY=your_api_subscription_key_here

  # Auth0 (required for user-authenticated features)
  AUTH0_DOMAIN=your-tenant.us.auth0.com
  AUTH0_CLIENT_ID=your_public_client_id
  AUTH0_AUDIENCE=https://cezzis-cocktails-api
  AUTH0_SCOPES="openid offline_access profile email read:owned-account write:owned-account"

  # Optional: Application Insights (telemetry)
  APPLICATIONINSIGHTS_INSTRUMENTATIONKEY=your_app_insights_key

  # Optional: Logging
  LOG_LEVEL=info
  ENV=local
  ```

   Supported environment files: `.env`, `.env.local`, `.env.test`.

4) Run locally (HTTP)
  ```bash
  # Build binary
  make compile

  # Run HTTP server (choose a port)
  ./cocktails.mcp/dist/linux/cezzis-cocktails --http :8080
  ```

5) Testing
   ```bash
   make test
   ```
   Generates coverage artifacts (`coverage.out`, `cobertura.xml`).


## 🔐 OAuth and Authentication

This server uses Auth0 for end‑user authentication to enable personalized features (e.g., ratings).

Flow (HTTP): Device Authorization Grant
- The `auth_login` tool returns a verification URL and user code.
- Visit the URL, enter the code, and complete login.
- The server polls Auth0 and stores tokens securely once available.

Token handling:
- Access and refresh tokens are stored encrypted under `~/.cezzis/.cezzis_tokens.enc`.
- Tokens are automatically refreshed using the refresh token when near expiry.
- Logout clears stored tokens.

Required settings:
- `AUTH0_DOMAIN` – e.g., `your-tenant.us.auth0.com`
- `AUTH0_CLIENT_ID` – public SPA/native client ID configured in Auth0
- Optional: `AUTH0_AUDIENCE` if the API expects a specific audience
- Optional: `AUTH0_SCOPES` (default: `openid profile email offline_access`)

Auth tools available to MCP clients:
- `auth_login` – Initiates device code login and returns instructions.
- `auth_status` – Returns whether you’re currently authenticated.
- `auth_logout` – Clears stored tokens.

## �💻 MCP Client Setup


### Claude Desktop
Configure `~/.config/Claude/claude_desktop_config.json` for HTTP MCP:

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "url": "http://localhost:3001/mcp",
      "type": "http"
    }
  }
}
```

### Cursor
Configure `~/.cursor/mcp.json` or via Settings UI for HTTP MCP:

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "url": "http://localhost:3001/mcp",
      "type": "http"
    }
  }
}
```

### GitHub Copilot (HTTP MCP)
Configure VS Code `User/mcp.json` (Copilot MCP servers):

```json
{
  "servers": {
    "cezzis-mcp": {
      "url": "http://localhost:3001/mcp",
      "type": "http"
    }
  },
  "inputs": []
}
```
Start the server locally with `--http :8080` and Copilot Chat can call its tools over HTTP.

## 📦 Build & Deployment

- Build: `make compile` (outputs `./cocktails.mcp/dist/linux/cezzis-cocktails`)
- Container: `make docker-build` (builds image for ACA)
- Infra: Terraform under `/terraform` for ACA, APIM, Key Vault, etc.
- CI/CD: GitHub Workflows build, test, and publish artifacts/images

## 🔍 Code Quality

- `golangci-lint` for static analysis
- `gofmt` and imports tooling enforced via Make targets
- Unit tests with coverage reports

## 🔒 Security Features

- API subscription key required for upstream API access
- Secrets sourced from env files locally and Azure Key Vault in cloud
- HTTP/HTTPS transport for MCP endpoint
- Validated tool inputs and structured error handling

## 📈 Monitoring

- Optional Application Insights for logs/traces
- Health checks exposed in HTTP mode for probes

## 🤖 What is MCP?

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open standard that enables AI assistants to securely connect with external data sources and tools. Using MCP here allows agents to:

- Ask for cocktails in natural language
- Get contextual recommendations based on ingredients and styles
- Retrieve rich recipe data with measurements and techniques
- Integrate seamlessly across MCP‑compatible tools and IDEs

## 🌐 Community & Support

- 🤝 Contributing Guide – see [CONTRIBUTING.md](.github/CONTRIBUTING.md)
- 🤗 Code of Conduct – see [CODE_OF_CONDUCT.md](.github/CODE_OF_CONDUCT.md)
- 🆘 Support Guide – see [SUPPORT.md](.github/SUPPORT.md)
- 🔒 Security Policy – see [SECURITY.md](.github/SECURITY.md)

## 📄 License

This project is proprietary software. All rights reserved. See [LICENSE](LICENSE) for details.

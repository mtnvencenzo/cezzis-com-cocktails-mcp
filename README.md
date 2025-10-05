# Cezzis.com Cocktails MCP Server

> Part of the broader Cezzis.com digital experience for discovering and sharing cocktail recipes with a broad community of cocktail enthusiasts and aficionados.

[![Go](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/actions/workflows/cezzis-mcp-cicd.yaml/badge.svg)](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/actions/workflows/cezzis-mcp-cicd.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mtnvencenzo/cezzis-com-cocktails-mcp)](https://goreportcard.com/report/github.com/mtnvencenzo/cezzis-com-cocktails-mcp)
[![License](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)

An MCP (Model Context Protocol) server that gives AI agents secure, first‑class access to Cezzis.com cocktail data. It provides high‑level tools for searching cocktails and retrieving detailed recipes and metadata, and can run in stdio mode (for MCP clients) or HTTP mode (for development and debugging).

## 🧩 Cezzis.com Project Ecosystem

This server works alongside several sibling repositories:

- **cocktails-mcp** (this repo) – Model Context Protocol services that expose cocktail data to AI agents
- [**cocktails-api**](https://github.com/mtnvencenzo/cezzis-com-cocktails-api) – ASP.NET Core backend and REST API consumed by the site and integrations
- [**cocktails-web**](https://github.com/mtnvencenzo/cezzis-com-cocktails-web) – React SPA for the public experience
- [**cocktails-common**](https://github.com/mtnvencenzo/cezzis-com-cocktails-common) – Shared libraries and utilities reused across frontends, APIs, and tooling
- [**cocktails-images**](https://github.com/mtnvencenzo/cezzis-com-cocktails-images) *(private)* – Source of curated cocktail imagery and CDN assets
- [**cocktails-shared-infra**](https://github.com/mtnvencenzo/cezzis-com-cocktails-shared-infra) – Terraform compositions specific to the cocktails platform
- [**shared-infrastructure**](https://github.com/mtnvencenzo/shared-infrastructure) – Global Terraform modules that underpin multiple Cezzis.com workloads

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
- **Language**: Go 1.25+
- **Protocol**: Model Context Protocol over stdio (primary) and HTTP (dev)
- **Server**: Lightweight MCP server with tool registry and health endpoints
- **Logging**: zerolog (structured JSON logs)

### Integrations
- **Cezzis.com Cocktails API**: Upstream data source (requires subscription key)
- **Azure AI Search**: Powers semantic/lucene queries in the upstream API
- **Application Insights**: Optional telemetry via instrumentation key

### Authentication & Security
- **API Access**: `COCKTAILS_API_XKEY` subscription key injected via env/Key Vault
- **Auth0 OAuth 2.1 / OIDC**: End‑user authentication for personalized features (favorites, ratings, profile)
- **Secrets**: Managed via environment files locally and Azure Key Vault in cloud
- **Transport**: HTTPS for HTTP mode; stdio for MCP mode

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

   ```bash
   # Required: Cezzis.com API Configuration
   COCKTAILS_API_HOST=https://api.cezzis.com/prd/cocktails
   COCKTAILS_API_XKEY=your_api_subscription_key_here

  # Auth0 (required for user-authenticated features)
  AUTH0_DOMAIN=your-tenant.us.auth0.com
  AUTH0_CLIENT_ID=your_public_client_id
  # Optional audience if the API expects a specific identifier
  AUTH0_AUDIENCE=https://api.cezzis.com/prd/cocktails
  # Optional scopes (defaults: openid profile email offline_access)
  AUTH0_SCOPES=openid profile email offline_access

   # Optional: Application Insights (telemetry)
   APPLICATIONINSIGHTS_INSTRUMENTATIONKEY=your_app_insights_key

   # Optional: Logging
   LOG_LEVEL=info
   ENV=local
   ```

   Supported environment files: `.env`, `.env.local`, `.env.test`.

4) Local Development
   ```bash
   # MCP stdio mode (default)
   ./cocktails.mcp/dist/linux/cezzis-cocktails

   # HTTP mode for debugging
   ./cocktails.mcp/dist/linux/cezzis-cocktails --http :8080
   ```

5) Testing
   ```bash
   make test
   ```
   Generates coverage artifacts (`coverage.out`, `cobertura.xml`).

## 📚 MCP Tools

The server exposes two primary tools to AI clients:

### cocktail_search
- Purpose: Search cocktails by natural language query
- Parameters:
  - `query` (string) – Search terms (name, ingredients, style)
  - `limit` (optional number) – Max results
- Returns: Array of cocktails with IDs, titles, key ingredients, summaries, and images

### cocktail_get
- Purpose: Retrieve full details for a specific cocktail
- Parameters:
  - `id` (string) – Cocktail ID (from search results)
- Returns: Complete recipe with ingredients (amounts/units), instructions, images, ratings, history/notes

### HTTP Mode Endpoints (dev only)
- `GET /health` – Health check
- MCP protocol over HTTP for debugging

## � OAuth and Authentication

This MCP server uses Auth0 for end‑user authentication to enable personalized features (e.g., favorites, ratings).

Supported flows:

- Stdio/local: Authorization Code Flow with PKCE
  - The server opens your browser to Auth0’s authorize endpoint.
  - A local callback listener on http://localhost:6097/callback receives the authorization code.
  - The server exchanges the code for tokens and stores them securely (encrypted) for reuse.

- HTTP/container mode: Device Authorization Grant (Device Code)
  - The server returns a verification URL and user code.
  - You visit the URL in any browser, enter the code, and complete login.
  - The server polls Auth0 for tokens and stores them when available.

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
- `auth_login` – Initiates login. In stdio mode, triggers PKCE browser flow; in HTTP mode, returns device code instructions.
- `auth_status` – Returns whether you’re currently authenticated.

## �💻 MCP Client Setup

### Claude Desktop
Configure `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "command": "/absolute/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

### Cursor
Configure `~/.cursor/mcp.json` or via Settings UI:

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "command": "/absolute/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

### GitHub Copilot Chat
Configure VS Code Settings or `~/.config/github-copilot/mcp.json`:

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "command": "/absolute/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

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

- API subscription key required for all upstream API access
- Secrets sourced from env files locally and Azure Key Vault in cloud
- HTTPS enforcement in HTTP mode; stdio transport in MCP mode
- Minimal surface area: only two tools exposed with validated inputs

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

- 🤝 Contributing – Please open an issue or pull request in this repository to propose changes
- 🤗 Code of Conduct – Be respectful and collaborative in discussions and reviews
- 🆘 Support – Use GitHub Issues for bug reports and feature requests
- 🔒 Security – Do not disclose sensitive information in issues; contact the maintainer privately for security concerns

## 📄 License

This project is proprietary software. All rights reserved. See `LICENSE` for details.

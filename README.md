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

This repository contains the Go-based MCP server for Cezzis.com cocktails. It exposes a streamable HTTP MCP interface that lets MCP clients search cocktail data, retrieve cocktail details, authenticate against Auth0, and submit user ratings through the Cezzis platform APIs.


## Overview

The server is an integration layer, not the source of cocktail data. It registers MCP tools, forwards requests to the upstream Cocktails, AI Search, and Accounts APIs, stores authentication tokens in PostgreSQL per MCP session, and emits telemetry through OpenTelemetry.


Primary capabilities:

- Search cocktails by free text.
- Retrieve full cocktail details by cocktail ID.
- Start and manage Auth0 device-flow authentication.
- Submit authenticated cocktail ratings.
- Expose health and MCP HTTP endpoints for local and deployed environments.

## Production Environment

![Complete Diagram](./.assets/cezzis-com-mcp-interactions.drawio.svg)

## Tech Stack

- Go 1.25.1
- Model Context Protocol over streamable HTTP via `mark3labs/mcp-go`
- OpenAPI-generated API clients for upstream Cezzis services
- Auth0 device authorization flow
- PostgreSQL for MCP session token storage
- OpenTelemetry and zerolog for observability
- Kubernetes manifests for local deployed environments under `.iac/k8s`
- Terraform for production Azure infrastructure under `.iac/terraform`

## Repo Structure

```text
.
├── .iac/
│   ├── argocd/      # Argo CD manifests for cluster sync
│   ├── k8s/         # Local Kubernetes deployment manifests
│   └── terraform/   # Production Azure infrastructure
├── .vscode/         # IDE launch configuration
├── cocktails.mcp/
│   ├── http-client.env.json
│   └── src/
│       ├── cmd/         # Application entry point
│       └── internal/
│           ├── api/     # Generated API clients
│           ├── auth/    # Auth0 flow and token handling
│           ├── db/      # PostgreSQL connection and setup
│           ├── mcpserver/
│           ├── middleware/
│           ├── repos/
│           ├── telemetry/
│           └── tools/   # MCP tool definitions and handlers
├── Dockerfile
├── makefile
└── mcp.http
```

## HTTP Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/mcp/v1/health/ping` | Basic health check returning `{"status": "ok"}` |
| GET | `/mcp/v1/health/readiness` | Readiness probe returning `{"status": "ready"}` |
| GET | `/mcp/v1/health/liveness` | Liveness probe returning `{"status": "alive"}` |
| GET | `/mcp/v1/health/version` | Build version response |
| GET | `/mcp/v1/mcp` | MCP probe endpoint returning `{"status":"ok", "sse":false}` |
| POST | `/mcp/v1/mcp` | Streamable HTTP MCP endpoint used by MCP clients |

Tool execution requests rely on the `Mcp-Session-Id` header so the server can associate requests with an MCP session.

## MCP Tools

| Tool | Description |
| --- | --- |
| `search_cocktails` | Searches cocktail data using the upstream AI Search API |
| `get_cocktail` | Returns detailed cocktail data for a specific cocktail ID |
| `convert_to_plaintext` | Converts markdown or HTML-rich content into plain text |
| `authentication_login_flow` | Starts the Auth0 device login flow |
| `auth_status` | Returns the authentication state for the current MCP session |
| `authentication_logout_flow` | Clears tokens for the current MCP session |
| `cocktail_rate` | Submits a cocktail rating for an authenticated user |

## Quick Start

### Prerequisites

- Go 1.25.1+
- PostgreSQL
- Valid values for the upstream API hosts and subscription keys
- Auth0 settings if you want to use authenticated tools

### 1. Configure environment

Create a `.env` file in `cocktails.mcp/src` with the values your environment needs:

```env
PORT=7999
ENV=loc

COCKTAILS_API_HOST=https://your-host/cocktails
COCKTAILS_API_XKEY=replace-me

ACCOUNTS_API_HOST=https://your-host/accounts
ACCOUNTS_API_XKEY=replace-me

AISEARCH_API_HOST=https://your-host/search
AISEARCH_API_XKEY=replace-me

AUTH0_DOMAIN=your-tenant.us.auth0.com
AUTH0_NATIVE_CLIENT_ID=replace-me
AUTH0_ACCOUNTS_API_AUDIENCE=https://api.cezzis.com/accounts
AUTH0_SCOPES=openid offline_access profile email read:owned-account write:owned-account

CEZZIS_BASE_URL=https://www.cezzis.com

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=cezzis_cocktails_mcp
POSTGRES_USER=postgres
POSTGRES_PASSWORD=replace-me
POSTGRES_USE_TLS=false
```

### 2. Build and run

```bash
make compile
./cocktails.mcp/dist/linux/cezzis-cocktails
```

The server listens on the configured `PORT`. By default that is `7999`.

### 3. Run from VS Code

For IDE-based debugging, use the launch configuration in `.vscode/launch.json`. It runs the Go application from `cocktails.mcp/src/cmd` with local environment loading enabled.

### 4. Optional local Kubernetes deployment

The manifests in `.iac/k8s` are for a local deployed environment. They define the Deployment, Service, Ingress, ConfigMap, and ExternalSecret wiring used when running the app in a local cluster.

To sync that setup through Argo CD:

```shell
# app
kubectl apply -f https://raw.githubusercontent.com/mtnvencenzo/cezzis-com-cocktails-mcp/refs/heads/main/.iac/argocd/cezzis-com-cocktails-mcp.yaml

# image updater
kubectl apply -f https://raw.githubusercontent.com/mtnvencenzo/cezzis-com-cocktails-mcp/refs/heads/main/.iac/argocd/image-updater.yaml
```

## Production Notes

Production infrastructure for this application is defined in `.iac/terraform`. In production, the MCP app is hosted in Azure Container Apps, sits behind Azure API Management, and is reached externally through Azure Front Door.

## License

This project is proprietary software. All rights reserved. See [LICENSE](LICENSE) for details.

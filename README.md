# Cezzis Cocktails MCP Server 🍸

[![Go](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/actions/workflows/cezzis-mcp-cicd.yaml/badge.svg)](https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/actions/workflows/cezzis-mcp-cicd.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mtnvencenzo/cezzis-com-cocktails-mcp)](https://goreportcard.com/report/github.com/mtnvencenzo/cezzis-com-cocktails-mcp)
[![License](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)

A Model Context Protocol (MCP) server that provides seamless access to the Cezzis.com cocktails API, enabling AI agents and applications to search and retrieve detailed cocktail recipes and information.

## Table of Contents

- [Features](#features)
- [What is MCP?](#what-is-mcp-)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [MCP Client Setup](#mcp-client-setup)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

## Features ✨

- **🔍 Cocktail Search Tool**: Search cocktails by name, ingredient, or recipe details with intelligent semantic search
- **📖 Cocktail Retrieval Tool**: Get detailed information about specific cocktails including full recipes, instructions, and ratings  
- **🤖 MCP Integration**: Seamless integration with AI agents and LLM applications
- **🔒 Secure API Access**: Built-in authentication and API key management
- **🚀 High Performance**: Powered by Azure AI Search with Lucene indexing
- **📋 Comprehensive Logging**: Built-in structured logging with zerolog
- **🐳 Container Support**: Docker support for easy deployment
- **🧪 Well Tested**: Comprehensive test coverage with Go testing framework

### MCP Tools Available

#### 1. Cocktail Search Tool
- **Purpose:** Search for cocktails using flexible natural language queries
- **Capabilities:**
  - Search by cocktail name (e.g., "Old Fashioned", "Margarita")
  - Search by ingredients (e.g., "gin and tonic", "contains lime") 
  - Search by style or category (e.g., "whiskey cocktails", "tiki drinks")
- **Returns:** List of matching cocktails with summaries, images, and key ingredients

#### 2. Cocktail Get Tool  
- **Purpose:** Retrieve detailed information about a specific cocktail by ID
- **Capabilities:**
  - Full recipe details with precise measurements
  - Step-by-step preparation instructions
  - High-quality cocktail images
  - User ratings and reviews
  - Historical context and variations
- **Returns:** Complete cocktail information including ingredients, directions, and metadata

## What is MCP? 🤖

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open standard that enables AI assistants to securely connect with external data sources and tools. It provides a universal way for AI models to access and interact with various APIs, databases, and services.

### Why MCP for Cocktails? 🍹

This server bridges AI assistants with the Cezzis.com cocktails database through MCP, enabling:

- **Natural Language Queries**: Ask for cocktails in plain English ("What's a good whiskey cocktail for winter?")
- **Contextual Recommendations**: Get personalized suggestions based on available ingredients
- **Rich Recipe Data**: Access detailed recipes, techniques, and cocktail history
- **Seamless Integration**: Works with any MCP-compatible AI assistant or application

## Getting Started 🚀

### Prerequisites

- **Go**: Version 1.25.1 or later ([Install Go](https://golang.org/doc/install))
- **Make**: For build automation (usually pre-installed on Linux)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp.git
   cd cezzis-com-cocktails-mcp
   ```

2. **Build the application**
   ```bash
   make compile
   ```
   This creates the executable at `./cocktails.mcp/dist/linux/cezzis-cocktails`

3. **Configure environment variables** (see [Configuration](#configuration))

4. **Set up your MCP client** (see [MCP Client Setup](#mcp-client-setup))

## Configuration ⚙️

The server requires several environment variables for API access and authentication. Create a `.env` file in `./cocktails.mcp/src/`:

```bash
# Required: Cezzis.com API Configuration
COCKTAILS_API_HOST=https://api.cezzis.com/prd/cocktails
COCKTAILS_API_XKEY=your_api_subscription_key_here

# Required: Azure AD B2C Configuration (for authentication)
AZUREAD_B2C_INSTANCE=https://your_tenant.b2clogin.com
AZUREAD_B2C_DOMAIN=your_tenant.onmicrosoft.com  
AZUREAD_B2C_USERFLOW=B2C_1_SignInSignUp_Policy

# Optional: Application Insights (for telemetry)
APPLICATIONINSIGHTS_INSTRUMENTATIONKEY=your_app_insights_key

# Optional: Logging
LOG_LEVEL=info
ENV=local
```

### Environment Files

The server supports multiple environment files:
- `.env` - Base configuration
- `.env.local` - Local development overrides (recommended for development)
- `.env.test` - Test environment configuration

### Getting API Access

To obtain a `COCKTAILS_API_XKEY`:
1. Visit [Cezzis.com Developer Portal](https://api.cezzis.com) 
2. Sign up for an account
3. Subscribe to the Cocktails API
4. Copy your subscription key

## MCP Client Setup 💻

### Claude Desktop

1. Build and configure the server (see steps above)
2. Add to your Claude Desktop configuration file at `~/.config/Claude/claude_desktop_config.json`:

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

1. Build and configure the server (see steps above) 
2. Add to your Cursor MCP configuration:

**Via Settings UI:**
- Open Cursor → `Ctrl + Shift + P` → "View: Open MCP Settings"

**Via Configuration File at `~/.cursor/mcp.json`:**

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "command": "/absolute/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

### GitHub Copilot

1. Build and configure the server (see steps above)
2. Configure GitHub Copilot MCP integration:

**Via VS Code Settings:**
1. Open VS Code Settings (`Ctrl + ,`)
2. Search for "copilot mcp"
3. Add the server configuration:

```json
{
  "github.copilot.chat.mcp.servers": {
    "cezzis-cocktails": {
      "command": "/absolute/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

**Via Configuration File at `~/.config/github-copilot/mcp.json`:**

```json
{
  "mcpServers": {
    "cezzis-cocktails": {
      "command": "/absolute/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

3. Restart VS Code to apply changes

## Development 🛠️

### Project Structure

```
cocktails.mcp/
├── src/
│   ├── cmd/                    # Application entry point
│   ├── internal/
│   │   ├── api/               # Generated API client code
│   │   ├── config/            # Configuration management  
│   │   ├── logging/           # Structured logging
│   │   ├── middleware/        # HTTP middleware
│   │   ├── server/            # MCP server implementation
│   │   ├── testutils/         # Testing utilities
│   │   └── tools/             # MCP tool implementations
│   ├── .env                   # Environment configuration
│   └── go.mod                 # Go module definition
└── dist/                      # Build outputs
```

### Available Make Targets

```bash
# Development
make tidy          # Update Go modules
make lint          # Run linters and fix issues
make fmt           # Format Go code  
make test          # Run tests with coverage
make imports       # Fix imports

# Building  
make compile       # Build for Linux
make clean         # Clean build artifacts
make docker-build  # Build Docker image

# Running
./cocktails.mcp/dist/linux/cezzis-cocktails --http :8080  # HTTP mode
./cocktails.mcp/dist/linux/cezzis-cocktails               # Stdio mode (MCP)
```

### Running Tests

```bash
make test
```

This runs the full test suite with:
- Unit tests for all packages
- Coverage report (generates `coverage.out` and `cobertura.xml`)
- HTML coverage report

### Code Generation

The project uses [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate API client code from OpenAPI specifications. Generated files are marked with `DO NOT EDIT` comments.

## API Documentation 📚

The server exposes cocktail search and retrieval capabilities through two main MCP tools:

### Cocktail Search
- **Tool**: `cocktail_search`
- **Purpose**: Search cocktails by query
- **Parameters**: 
  - `query` (string): Natural language search query
  - `limit` (optional int): Maximum results to return

### Cocktail Get  
- **Tool**: `cocktail_get`
- **Purpose**: Get detailed cocktail information
- **Parameters**:
  - `id` (string): Cocktail ID from search results

### HTTP Mode (Development)

When run with `--http` flag, the server also exposes HTTP endpoints for testing:

- `GET /health` - Health check endpoint
- MCP protocol over HTTP for debugging

## Contributing 🤝

We welcome contributions! Please see our contributing guidelines:

### Prerequisites

1. **Go 1.25.1+** with modules enabled
2. **Development tools** (installed via Makefile):
   ```bash
   # Install required tools
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/incu6us/goimports-reviser/v3@latest
   go install github.com/quantumcycle/go-ignore-cov@latest
   go install github.com/t-yuki/gocover-cobertura@latest
   ```

### Development Workflow

1. **Fork and clone** the repository
2. **Create a feature branch**: `git checkout -b feature/my-feature`
3. **Make changes** following the coding standards
4. **Run tests**: `make test`
5. **Run linting**: `make lint`  
6. **Commit changes** with [conventional commits](https://conventionalcommits.org/)
7. **Push and create** a pull request

### Coding Standards

- Follow standard Go conventions and idioms
- Use `golangci-lint` for code quality (config in `.golangci.yaml`)
- Maintain test coverage above 80%
- Document public APIs with Go doc comments
- Use structured logging with zerolog

### Commit Message Format

We use [GitVersion](https://gitversion.net/) for semantic versioning:

- `feat: add new cocktail search filter` (+semver: minor)
- `fix: resolve API timeout issue` (+semver: patch)  
- `feat!: change search API response format` (+semver: major)
- `docs: update README examples` (+semver: none)

## License 📄

This project is licensed under a proprietary license. All rights reserved. Please contact the repository owner for more information about usage and distribution rights.

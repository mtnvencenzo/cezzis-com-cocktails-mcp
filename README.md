# cezzis-com-cocktails-mcp 🍸

## Cocktails MCP Tools & API Integration 🍸🔍

This project exposes two powerful MCP tools that allow both developers and non-developers to seamlessly integrate cocktail search and data into their applications, chatbots, or workflows. These tools act as a bridge between the Model Context Protocol (MCP) and the cezzis.com cocktails API, unlocking advanced search and retrieval capabilities for cocktail recipes and information.

### 1. Cocktail Search Tool
- **Purpose:** Enables searching for cocktails by name, ingredient, or recipe details.
- **How it works:**
  - Accepts flexible search queries (e.g., "gin and tonic", "contains lime", "classic whiskey drinks").
  - Forwards these queries to the cezzis.com cocktails API `/search` endpoint.
  - Returns a list of matching cocktails, each with summary info, images, and key ingredients.
- **Use cases:**
  - Powering conversational agents that recommend drinks.
  - Integrating cocktail search into mobile/web apps.
  - Enabling voice assistants to answer cocktail-related questions.

### 2. Cocktail Get Tool
- **Purpose:** Retrieves detailed information about a specific cocktail by its unique ID.
- **How it works:**
  - Accepts a cocktail ID (from a search result or known recipe).
  - Calls the cezzis.com cocktails API `/get` endpoint.
  - Returns full recipe details, preparation instructions, images, ratings, and historical context.
- **Use cases:**
  - Displaying full cocktail recipes in apps or websites.
  - Enabling step-by-step drink preparation guides.
  - Fetching cocktail metadata for analytics or recommendations.

### Advanced Search Backed by Azure AI Search & Lucene
The cezzis.com cocktails API endpoints are powered by a robust backend leveraging **Azure AI Search** and **Lucene-based indexes**. This architecture provides:
- **Intelligent, semantic search:** Understands natural language queries and ingredient combinations.
- **High relevance:** Results are ranked using AI-driven scoring and classic information retrieval techniques.
- **Scalability:** Supports thousands of cocktail recipes and complex queries with low latency.
- **Rich filtering:** Search by ingredient, style, flavor profile, or even historical/geographic context.

Whether you're a developer building a new app, or a non-developer looking to add cocktail discovery to your platform, these MCP tools and the cezzis.com API make it easy to deliver the best cocktail search experience available.

## What is MCP? 🤖

Model Context Protocol (MCP) is an open protocol designed to enable large language models (LLMs) and intelligent agents to interact with external APIs and tools in a standardized, secure, and extensible way. MCP provides a structured context for models to understand available actions, endpoints, and data schemas, making it easier for LLMs to perform complex tasks by leveraging external services.

### How does MCP help search for cocktails? 🍹

In this repository, MCP is used to bridge LLMs and agents with the cocktails API from cezzis.com. By exposing the cocktails search endpoint through MCP, LLMs can:
- Discover and understand the available search capabilities and parameters.
- Programmatically query the cocktails database for recipes, ingredients, and suggestions.
- Integrate cocktail search functionality into conversational or agent-based workflows, enabling richer and more interactive user experiences.

This approach allows developers to build intelligent assistants or chatbots that can answer cocktail-related queries, recommend drinks, or help users explore new recipes, all powered by the cezzis.com cocktails API and its internall AI Search capabilities.

## Features ✨
- MCP integration for LLMs and agents
- Secure and extensible API access
- Search cocktails by name, ingredient, or recipe
- Easy local development and deployment

## Getting Started 🚀

### Prerequisites
- Go (latest version recommended)
- make

### Building the app for local usage 🛠️

Navigate to the project root directory and run:

```bash
make compile
```

This will build the application and place the executable at `./cocktails.mcp/dist/linux/cezzis-cocktails`

## Setting up with Claude Desktop 💻

Make sure an entry exists in this file for the deploy path of the local executable:
> ~/.config/Claude/claude_desktop_config.json (Linux/macOS) or C:\Users\<user>\AppData\Roaming\Claude\claude_desktop_config.json (Windows)

```json
{
  "mcpServers": {
    "mcp-cocktails-go": {
        "command": "/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
      }
    }
}
```

## Setting up with Cursor 💻

Make sure an entry exists in this file for the deploy path of the local executable:
> ~/.cursor/mcp.json (Linux/macOS) or C:\Users\<user>\.cursor\mcp.json (Windows)

Or open the mcp settings within cursor via `Ctrl Shift P` > `View: Open Mcp Settings`

```json
{
  "mcpServers": {
    "mcp-cocktails-go": {
        "command": "/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
      }
    }
}
```

## Setting up with GitHub Copilot 💻

GitHub Copilot supports MCP servers through its configuration. Add an entry to your GitHub Copilot MCP configuration:

**For VS Code/GitHub Copilot Chat:**
1. Open VS Code Settings (Ctrl/Cmd + ,)
2. Search for "copilot mcp"
3. Add the MCP server configuration:

```json
{
  "github.copilot.chat.mcp.servers": {
    "mcp-cocktails-go": {
      "command": "/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

**Alternative configuration file approach:**
Create or edit the GitHub Copilot MCP configuration file at:
> ~/.config/github-copilot/mcp.json (Linux/macOS) or %APPDATA%\github-copilot\mcp.json (Windows)

```json
{
  "mcpServers": {
    "mcp-cocktails-go": {
      "command": "/path/to/your/project/cocktails.mcp/dist/linux/cezzis-cocktails"
    }
  }
}
```

After configuration, restart VS Code and you should be able to use cocktail search and lookup tools in GitHub Copilot Chat.

## Installs
`go install -v github.com/go-delve/delve/cmd/dlv@latest`

## Contributing 🤝

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](../../issues).

## License 📄

This project is licensed under a proprietary license. All rights reserved. Please contact the repository owner for more information about usage and distribution rights.

# cezzis-com-cocktails-mcp 🍸

Model Context Protocol (MCP) application to allow LLMs and agents to connect and use the cocktails API search endpoint from cezzis.com.

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

Navigate to the `/src` directory and run:

```bash
make compile
```

## Setting up with Claude Desktop 💻

Make sure an entry exists in this file for the deploy path of the local exe:
> C:\Users\<user>\AppData\Roaming\Claude\claude_desktop_config.json

```json
{
  "mcpServers": {
	"mcp-cocktails-go": {
      "command": "D:\\Github\\cezzis-com-cocktails-mcp\\cocktails.mcp\\dist\\cezzis-cocktails.exe",
      "args": ["--stdio"]
    }
  }
}
```

## Setting up with Cursor 💻

Make sure an entry exists in this file for the deploy path of the local exe:
> C:\Users\rvecc\.cursor\mcp.json

Or open the mcp settings within cursor via `Ctrl Shift P` > `View: Open Mcp Settings`

```json
{
  "mcpServers": {
	"mcp-cocktails-go": {
      "command": "D:\\Github\\cezzis-com-cocktails-mcp\\cocktails.mcp\\dist\\cezzis-cocktails.exe",
      "args": ["--stdio"]
    }
  }
}
```


## Contributing 🤝

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](../../issues).

## License 📄

This project is licensed under a proprietary license. All rights reserved. Please contact the repository owner for more information about usage and distribution rights.

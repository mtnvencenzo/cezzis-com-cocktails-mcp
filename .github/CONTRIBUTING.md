# 🍸 Contributing to Cezzis Cocktails MCP Server

Thank you for your interest in contributing to the Cezzis Cocktails MCP Server! We welcome contributions that help improve the MCP tooling, developer experience, and integration with the broader Cezzis.com ecosystem.

## 📋 Table of Contents

- [Getting Started](#-getting-started)
- [Development Setup](#-development-setup)
- [Contributing Process](#-contributing-process)
- [Code Standards](#-code-standards)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Getting Help](#-getting-help)

## 🚀 Getting Started

### 🧰 Prerequisites

Before you begin, ensure you have the following installed:
- Go 1.25+
- Make
- Docker (optional, for containerized development)
- Terraform (optional, for IaC under `terraform/`)
- Git

### 🗂️ Project Structure

```text
cocktails.mcp/
├── src/
│   ├── cmd/                    # Entry point
│   ├── internal/               # Packages (auth, tools, server, logging, config)
│   └── go.mod                  # Go module definition
├── terraform/                  # Infrastructure as Code (Azure)
└── .github/                    # GitHub workflows and templates
```

## 💻 Development Setup

1. **Fork and Clone the Repository**
   ```bash
   git clone https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp.git
   cd cezzis-com-cocktails-mcp
   ```

2. **Restore Dependencies**
   ```bash
   make tidy
   ```

3. **Run locally**
   ```bash
   # Build & test
   make compile
   make test

   # Run MCP server
   ./cocktails.mcp/dist/linux/cezzis-cocktails           # stdio (MCP) mode
   ./cocktails.mcp/dist/linux/cezzis-cocktails --http :8080  # HTTP mode
   ```

4. **Docker (Optional)**
   ```bash
   make docker-build
   ```

## 🔄 Contributing Process

### 1. 📝 Before You Start

- **Check for existing issues** to avoid duplicate work
- **Create or comment on an issue** to discuss your proposed changes
- **Wait for approval** from maintainers before starting work (required for this repository)

### 2. 🛠️ Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following our [code standards](#-code-standards)

3. **Test your changes**
   ```bash
   make test
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat(mcp): add new tool or behavior for ..."
   ```
   
   Use [conventional commit format](https://www.conventionalcommits.org/):
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `style:` for formatting changes
   - `refactor:` for code refactoring
   - `test:` for adding tests
   - `chore:` for maintenance tasks

### 3. 📬 Submitting Changes

1. **Push your branch**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request**
   - Use our [PR template](pull_request_template.md)
   - Fill out all sections completely
   - Link related issues using `Closes #123` or `Fixes #456`
   - Request review from maintainers

## 📏 Code Standards

### 🧩 MCP (Go)

- Go idioms and effective Go
- Keep public APIs documented with Go doc comments
- Structured logging with zerolog
- Avoid global state; prefer dependency injection via constructors
- Clear separation of concerns across internal packages

### 🧪 Code Quality

```bash
# Build
make compile

# Run tests
make test

# Lint and formatting
make lint
make fmt
```

### 🌱 Infrastructure (Terraform)

- **Terraform**: Use Terraform best practices
- **Variables**: Define all variables in `variables.tf`
- **Documentation**: Document all resources and modules
- **State**: Never commit `.tfstate` files

## 🧪 Testing

### 🧪 Unit Tests
```bash
make test
```


### 📏 Test Requirements

- **Unit Tests**: All new features must include unit tests
- **E2E Tests**: Critical user flows should have E2E test coverage
- **Coverage**: Maintain minimum 80% code coverage
- **Test Naming**: Use descriptive test names that explain the behavior

## 🆘 Getting Help

### 📡 Communication Channels

- **Issues**: Use GitHub issues for bugs and feature requests
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Email**: Contact maintainers directly for sensitive issues

### 📄 Issue Templates

Use our issue chooser:
- https://github.com/mtnvencenzo/cezzis-com-cocktails-mcp/issues/new/choose

### ❓ Common Questions

**Q: How do I run the application locally?**
A: Follow the [Development Setup](#-development-setup) section above. Use the compiled binary in `cocktails.mcp/dist/linux/cezzis-cocktails`.

**Q: How do I run tests?**
A: Use `make test`.

**Q: Can I contribute without approval?**
A: No, all contributors must be approved by maintainers before making changes.

**Q: How do I report a security vulnerability?**
A: Please email the maintainers directly rather than creating a public issue.

## 📜 License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (see [LICENSE](../LICENSE)).

---

**Happy Contributing! 🍸**

For any questions about this contributing guide, please open an issue or contact the maintainers.

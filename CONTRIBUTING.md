# Contributing to Construct

Thank you for your interest in contributing to Construct! This document provides guidelines and information to help you contribute effectively to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [Community](#community)

## Getting Started

### Prerequisites

Before you begin contributing, ensure you have:

- **Go 1.24+** installed
- **Git** for version control
- A **GitHub account**
- Familiarity with Go and basic command-line tools
- An Anthropic API key for testing

### First Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/construct.git
   cd construct
   ```

3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/furisto/construct.git
   ```

4. **Install development dependencies** (if any):
   ```bash
   # The project uses Go modules, so dependencies are managed automatically
   go mod download
   ```

## Development Setup

### Building from Source

Construct uses a Go workspace with multiple modules. To build the CLI:

```bash
# Build the CLI
cd frontend/cli
go build -o construct

# Optionally install it locally for testing
sudo mv construct /usr/local/bin/construct
```

### Setting Up the Development Environment

1. **Install the daemon**:
   ```bash
   construct daemon install
   ```

2. **Configure a test provider**:
   ```bash
   export ANTHROPIC_API_KEY="your-test-api-key"
   construct modelprovider create test-provider --type anthropic
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./backend/agent/...
```

### Linting

The project uses `.golangci.yml` for linting configuration:

```bash
# Install golangci-lint if you haven't already
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the linter
golangci-lint run

# Run with auto-fix where possible
golangci-lint run --fix
```

## Project Structure

Construct follows a modular architecture with clear separation of concerns:

```
construct/
├── api/                    # Protocol Buffer definitions (ConnectRPC)
├── backend/                # Core backend logic
│   ├── agent/             # Agent runtime and execution
│   ├── analytics/         # Metrics and analytics
│   ├── api/               # API implementation (ConnectRPC handlers)
│   ├── event/             # Event system
│   ├── memory/            # Conversation memory and storage
│   ├── model/             # LLM Provider integration
│   ├── prompt/            # Agent prompts
│   ├── secret/            # Secret/credential management
│   └── tool/              # Tool implementations (file ops, commands, etc.)
├── frontend/
│   └──cli/                # Command-line interface
├── shared/                # Shared utilities and types
├── docs/                  # Documentation
└── dev/                   # Development tools and scripts
```

### Key Components

- **Backend**: Contains the daemon logic, agent runtime, model providers, and tool execution
- **Frontend/CLI**: User-facing command-line interface
- **API**: ConnectRPC protocol definitions for client-daemon communication
- **Shared**: Common types and utilities used across modules

## Development Workflow

### Branch Strategy

We follow a standard Git workflow:

1. **main** branch: Stable, production-ready code
2. **Feature branches**: Named `feature/description` or `fix/issue-number`
3. **Pull requests**: All changes must go through PR review

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes**

3. **Test your changes** thoroughly:
   ```bash
   go test ./...
   golangci-lint run
   ```

4. **Commit your changes** with clear, descriptive messages:
   ```bash
   git add .
   git commit -m "Add feature: description of what you did"
   ```

5. **Keep your branch up to date**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

6. **Push to your fork**:
   ```bash
   git push origin feature/my-new-feature
   ```

### Commit Message Guidelines

Write clear, concise commit messages that explain **why** the change was made, not just what changed:

**Good**:
- `Add timeout handling to prevent hung tool executions`
- `Fix race condition in agent message processing`
- `Refactor model provider interface for better extensibility`

**Avoid**:
- `Update code`
- `Fix bug`
- `Changes`

## Coding Standards

### Go Style Guidelines

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

**Key principles**:

1. **Formatting**: Use `gofmt` or `goimports` (enforced by CI)
2. **Naming**: Use clear, descriptive names following Go conventions
   - Interfaces: `Reader`, `Writer`, `AgentExecutor`
   - Variables: camelCase for local, PascalCase for exported
   - Constants: PascalCase or ALL_CAPS for exported constants
3. **Error Handling**: Always handle errors explicitly, never ignore them
   ```go
   // Good
   result, err := someFunction()
   if err != nil {
       return fmt.Errorf("failed to execute: %w", err)
   }
   
   // Bad
   result, _ := someFunction()
   ```

4. **Testing**:
   - Write table-driven tests where appropriate
   - Test both success and error cases
   - Use descriptive test names: `TestAgentExecute_WithTimeout_ReturnsError`

### Code Review Checklist

Before submitting a PR, ensure:

- [ ] Code follows Go style guidelines
- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] New features have tests
- [ ] Documentation is updated (if applicable)
- [ ] Commit messages are clear and descriptive
- [ ] No sensitive information (API keys, credentials) is committed

## Testing

### Writing Tests

```go
func TestAgentCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   *AgentInput
        want    *Agent
        wantErr bool
    }{
        {
            name: "valid agent creation",
            input: &AgentInput{
                Name:   "test-agent",
                Model:  "test-model",
                Prompt: "You are a test assistant",
            },
            want: &Agent{
                Name:   "test-agent",
                Model:  "test-model",
                Prompt: "You are a test assistant",
            },
            wantErr: false,
        },
        {
            name: "missing required field",
            input: &AgentInput{
                Name: "test-agent",
                // Model is missing
            },
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CreateAgent(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateAgent() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateAgent() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Ensure your PR**:
   - Has a clear, descriptive title
   - References any related issues
   - Includes a description of what changed and why
   - Passes all CI checks
   - Has been rebased on the latest `main` branch

2. **PR Description Template**:
   ```markdown
   ## Description
   Brief description of what this PR does.

   ## Motivation
   Why is this change needed? What problem does it solve?

   ## Changes
   - List of specific changes made
   - Can be bullet points

   ## Testing
   How was this tested? What tests were added?

   ## Checklist
   - [ ] Tests pass locally
   - [ ] Linting passes
   - [ ] Documentation updated (if needed)
   - [ ] Breaking changes documented (if any)
   ```

3. **Review Process**:
   - A maintainer will review your PR
   - Address any feedback or requested changes
   - Once approved, a maintainer will merge your PR

### What to Expect

- **Initial Review**: Usually within 2-3 business days
- **Feedback**: We aim to provide constructive, actionable feedback
- **Iteration**: Be prepared to make changes based on review comments
- **Merge**: Once approved and CI passes, your PR will be merged

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. **Clear title** describing the issue
2. **Steps to reproduce** the bug
3. **Expected behavior** vs actual behavior
4. **Environment details**:
   - OS and version
   - Go version
   - Construct version
   - Model provider used
5. **Relevant logs** or error messages
6. **Screenshots** (if applicable)

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).

### Feature Requests

When requesting features:

1. **Describe the feature** clearly
2. **Explain the use case** and problem it solves
3. **Provide examples** of how it would be used
4. **Suggest implementation** (optional, but helpful)

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).

### Security Issues

**Do not** open public issues for security vulnerabilities. Instead:

- Email security concerns to the maintainers
- Provide detailed information about the vulnerability
- Allow time for a fix before public disclosure

## Community

### Getting Help

- **Documentation**: Check [docs/](./docs/) for guides and references
- **GitHub Discussions**: Ask questions and share ideas
- **GitHub Issues**: Report bugs and request features

### Ways to Contribute

You don't have to write code to contribute! Here are other ways to help:

- **Documentation**: Improve guides, add examples, fix typos
- **Bug Reports**: Report issues you encounter
- **Testing**: Test new features and provide feedback
- **Feature Ideas**: Suggest improvements and new capabilities
- **Code Review**: Review pull requests from other contributors
- **Community Support**: Help answer questions in discussions

### Recognition

We value all contributions! Contributors will be:

- Listed in release notes for significant contributions
- Credited in commit messages when appropriate
- Acknowledged in the project README (for major contributions)

## Development Tips

### Useful Commands

```bash
# Run daemon in foreground for debugging
construct daemon run --listen-unix /tmp/construct.sock

# View daemon logs (Linux)
journalctl --user -u construct -f

# View daemon logs (macOS)
tail -f ~/Library/Logs/construct.log

# Clean build artifacts
go clean -cache
rm -rf ~/.construct/test-*

# Update dependencies
go get -u ./...
go mod tidy
```

### Debugging

1. **Enable verbose logging**:
   ```bash
   construct -v <command>
   ```

2. **Run daemon in foreground**:
   ```bash
   construct daemon run --listen-unix
   ```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./backend/agent

# Memory profiling
go test -memprofile=mem.prof -bench=. ./backend/agent

# Analyze profiles
go tool pprof cpu.prof
```

## License

By contributing to Construct, you agree that your contributions will be licensed under the project's license.

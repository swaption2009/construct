<p align="center">
  <img src="logo.jpeg" alt="Construct Logo" width="600"
</p>

An API-first, multi-agent coding assistant designed for superior tool calling performance.

## Overview

Construct is a next-generation coding assistant that breaks away from traditional black-box AI assistants. Built with an API-first approach, it offers unparalleled customization and extensibility while supporting multiple collaborative agents that can work together on complex tasks.

## Key Features

### API-first Architecture
Everything in Construct can be configured via API, making it highly customizable and integrable with existing workflows and tools. This is in stark contrast to traditional coding assistants that operate as black boxes with limited configuration options.

### Multi-Agentic System
Construct supports multiple agents by default that can work together on a task. The system handles agent handoffs and delegations automatically, allowing for specialized agents to tackle different aspects of a problem.

### CodeAct Tool Calling
Construct uses CodeAct tool calling with JavaScript for superior tool call performance. This approach provides more reliable and efficient tool execution compared to traditional methods.


### Additional Features

- **Multiple Model Providers**: Support for various AI models including Anthropic, OpenAI, DeepSeek, and more
- **Language SDKs**: SDKs available for multiple programming languages
- **Model Context Protocol**: Enhanced context management for improved model performance
- **Parallel Tool Use**: Execute multiple tools simultaneously for faster operations
- **Checkpoints**: Save and restore the state of your work at any point

## Architecture

Construct is built with a modular architecture that separates concerns between:

- **Backend**: Handles agent runtime, model providers, and tool execution
- **API Layer**: Provides a consistent interface for all operations
- **Frontend CLI**: Offers an intuitive terminal interface for interacting with the system

The multi-agent system allows for specialized agents to collaborate on tasks, with the runtime managing message passing and coordination between agents.

## Getting Started

### Prerequisites

- Go 1.19 or later (for building from source)
- A supported operating system (macOS, Linux)

### Installation

```bash
# Clone the repository
git clone https://github.com/furisto/construct
cd construct

# Build the CLI
cd frontend/cli
go build -o construct

# Install the daemon
./construct daemon install
```

### Quick Start

1. **Install and start the daemon**:
   ```bash
   construct daemon install
   ```

2. **Create a model provider** (required before creating agents):
   ```bash
   # For OpenAI
   construct modelprovider create "openai" --type openai --api-key "your-api-key"
   
   # For Anthropic
   construct modelprovider create "anthropic" --type anthropic --api-key "your-api-key"
   ```

3. **Create models**:
   ```bash
   construct model create "gpt-4" --provider "openai" --context-window 8192
   construct model create "claude-3-5-sonnet" --provider "anthropic" --context-window 200000
   ```

4. **Create your first agent**:
   ```bash
   construct agent create "coder" --prompt "You are a helpful coding assistant" --model "claude-3-5-sonnet"
   ```

5. **Start a conversation**:
   ```bash
   construct new --agent coder
   # or ask a quick question
   construct ask "Help me write a hello world function in Python"
   ```

## Usage Examples

### Interactive Conversations

```bash
# Start a new interactive session
construct new --agent coder

# Resume a previous conversation
construct resume --last

# Work in a specific directory
construct new --agent coder --workspace /path/to/project
```

### Quick Questions with Context

```bash
# Ask a simple question
construct ask "What is the time complexity of quicksort?"

# Include files for context
construct ask "Review this code for bugs" --file main.go --file utils.go

# Use piped input
cat error.log | construct ask "What's causing this error?"

# Complex analysis with multiple turns
construct ask "Analyze this architecture and suggest improvements" \
  --agent architect --max-turns 10 --file architecture.md
```

### Agent Management

```bash
# List all agents
construct agent list

# Create specialized agents
construct agent create "debugger" \
  --prompt "You are an expert at debugging code and finding issues" \
  --model "gpt-4" \
  --description "Debugging specialist"

construct agent create "reviewer" \
  --prompt-file ./prompts/code-reviewer.txt \
  --model "claude-3-5-sonnet"

# Edit agent configuration
construct agent edit coder

# Get agent details
construct agent get coder --output json
```

### Task and Message Management

```bash
# Create a new task
construct task create --agent coder --workspace /project

# List recent tasks
construct task list --agent coder

# View task details
construct task get <task-id>

# List messages in a conversation
construct message list --task <task-id>

# View specific message
construct message get <message-id> --output yaml
```

### Configuration

```bash
# Set default agent for new conversations
construct config set cmd.new.agent "coder"

# Set default output format
construct config set output.format "json"

# Configure max turns for ask command
construct config set cmd.ask.max-turns 10

# View current configuration
construct config get cmd.new.agent
```

## Documentation

For more detailed documentation, please refer to:

- [CLI Reference](docs/cli_reference.md) - Complete reference for all CLI commands
- [API Reference](https://docs.construct.sh/api) (Coming soon)
- [User Guide](https://docs.construct.sh/guide) (Coming soon)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[License information will be provided here]

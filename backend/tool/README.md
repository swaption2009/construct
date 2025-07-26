# Tool System Architecture

The tool system provides a layered architecture for implementing and executing tools within the Construct AI agent framework. This document describes the architecture, components, and how to work with the system.

## Overview

The tool system is designed with a clear separation of concerns:

- **Core Logic**: Pure business logic implementations in domain-specific packages
- **Integration Layer**: JavaScript runtime integration for agent execution
- **Shared Infrastructure**: Common utilities, error handling, and testing framework
- **Native Interface**: Alternative tool interface for non-JavaScript contexts

## Architecture

```
backend/tool/
├── base/           # Shared infrastructure
├── filesystem/     # File system operations
├── search/         # Text search operations  
├── system/         # System command execution
├── communication/  # Agent communication tools
├── codeact/        # JavaScript runtime integration
└── native/         # Native tool interface
```

### Layer Responsibilities

#### Core Packages (`filesystem/`, `search/`, `system/`, `communication/`)

These packages contain pure business logic implementations:

- **Input/Output Types**: Structured data types for tool inputs and results
- **Core Functions**: Business logic that can be tested independently
- **Error Handling**: Uses `base.ToolError` for consistent error reporting
- **No Runtime Dependencies**: Independent of JavaScript runtime or session context

Example:
```go
// filesystem/create.go
func CreateFile(ctx context.Context, fs afero.Fs, input *CreateFileInput) (*CreateFileResult, error) {
    // Pure business logic
}
```

#### Integration Layer (`codeact/`)

The CodeAct package provides JavaScript runtime integration:

- **JavaScript Bindings**: Converts between `sobek.Value` and Go types
- **Session Management**: Accesses task context, agent ID, and database
- **Error Propagation**: Uses `session.Throw()` for JavaScript error handling
- **Tool Descriptions**: Rich documentation for agent consumption

Example:
```go
// codeact/create_file.go
func NewCreateFileTool() Tool {
    return NewOnDemandTool("create_file", description, inputHandler, execHandler)
}
```

#### Shared Infrastructure (`base/`)

Common utilities used across all packages:

- **Error System**: Standardized error codes and user-friendly suggestions
- **Testing Framework**: Reusable test setup for database and filesystem scenarios
- **Interfaces**: Base contracts (though these are minimal by design)

#### Native Interface (`native/`)

Alternative tool interface for non-JavaScript contexts:

- **Schema Generation**: JSON Schema generation from Go types
- **Direct Execution**: Bypasses JavaScript runtime for performance-critical scenarios
- **Tool Registration**: Alternative registration mechanism

## Tool Categories

### Filesystem Tools
- **create_file**: Create new files with content
- **edit_file**: Modify existing files with diff-based editing
- **read_file**: Read file contents with line range support
- **list_files**: Directory listing with filtering and metadata
- **find_file**: Search for files by name patterns

### Search Tools
- **grep**: Text search using ripgrep or fallback grep

### System Tools  
- **execute_command**: Run system commands with output capture

### Communication Tools
- **handoff**: Transfer tasks between agents
- **submit_report**: Submit structured reports
- **ask_user**: Request user input
- **print**: Output messages to user

## Error Handling

The system uses a standardized error approach:

```go
// Core packages use base.ToolError
return base.NewCustomError("File not found", []string{
    "Check the file path",
    "Use list_files to verify the location",
})

// CodeAct layer propagates errors
if err != nil {
    session.Throw(err)
}
```

Error codes provide consistent categorization and user-friendly suggestions for common issues.

## Testing

All core logic includes comprehensive tests using the shared testing framework:

```go
setup := &base.ToolTestSetup[*CreateFileInput, *CreateFileResult]{
    Call: func(ctx context.Context, services *base.ToolTestServices, input *CreateFileInput) (*CreateFileResult, error) {
        return CreateFile(ctx, services.FS, input)
    },
    // Database and filesystem verification
}
```

The testing framework provides:
- **Database Setup**: In-memory SQLite with schema migration
- **Filesystem Mocking**: Using `afero.Fs` for isolated testing
- **Scenario Testing**: Structured test cases with setup/verification
- **Debug Support**: Schema and data inspection utilities

## Adding New Tools

### 1. Create Core Implementation

```go
// backend/tool/[category]/[tool].go
package category

type ToolInput struct {
    // Input fields
}

type ToolResult struct {
    // Output fields  
}

func Tool(ctx context.Context, deps Dependencies, input *ToolInput) (*ToolResult, error) {
    // Business logic implementation
    if err != nil {
        return nil, base.NewCustomError("Error message", []string{"suggestion"})
    }
    return &ToolResult{}, nil
}
```

### 2. Create CodeAct Integration

```go
// backend/tool/codeact/[tool].go
package codeact

func NewToolTool() Tool {
    return NewOnDemandTool(
        "tool_name",
        toolDescription,
        inputHandler,
        execHandler,
    )
}

func inputHandler(session *Session, args []sobek.Value) (any, error) {
    // Convert JavaScript values to Go input struct
}

func execHandler(session *Session) func(call sobek.FunctionCall) sobek.Value {
    return func(call sobek.FunctionCall) sobek.Value {
        // Call core implementation and handle errors
        err := category.Tool(session.Context, deps, input)
        if err != nil {
            session.Throw(err)
        }
        return session.VM.ToValue(result)
    }
}
```

### 3. Add Tests

```go
// backend/tool/[category]/[tool]_test.go
func TestTool(t *testing.T) {
    setup := &base.ToolTestSetup[*ToolInput, *ToolResult]{
        Call: func(ctx context.Context, services *base.ToolTestServices, input *ToolInput) (*ToolResult, error) {
            return Tool(ctx, services.FS, input)
        },
    }
    
    setup.RunToolTests(t, []base.ToolTestScenario[*ToolInput, *ToolResult]{
        // Test scenarios
    })
}
```

### 4. Register Tool

```go
// frontend/cli/cmd/daemon_run.go
registry.RegisterTool(codeact.NewToolTool())
```

## Design Principles

### Separation of Concerns
- Core logic is independent of runtime environment
- Integration layer handles runtime-specific concerns
- Shared infrastructure promotes consistency

### Testability
- Core functions can be tested without JavaScript runtime
- Comprehensive test coverage with realistic scenarios
- Shared testing framework reduces boilerplate

### Error Handling
- Consistent error types across all tools
- User-friendly error messages with actionable suggestions
- Proper error propagation through layers

### Performance
- Native interface available for performance-critical scenarios
- Minimal overhead in core implementations
- Efficient filesystem and database operations

## Migration Notes

This architecture was established through systematic migration from a flat package structure. The migration preserved:

- All existing functionality and behavior
- Complete test coverage and scenarios  
- Tool descriptions and documentation
- Error handling and user experience

The layered approach improves maintainability, testability, and allows for alternative execution contexts while maintaining backward compatibility.

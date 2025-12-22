# Claude Agent SDK Go

A Go implementation of the Claude Agent SDK for building interactive AI agents with Claude.

## Overview

The Claude Agent SDK Go provides a clean, idiomatic Go interface for building applications that interact with Claude through the Claude CLI. It enables:

- Interactive multi-turn conversations with streaming support
- Custom tool integration via MCP (Model Context Protocol) servers
- Fine-grained control over permissions, hooks, and agent behavior
- Session management for stateful conversations
- Comprehensive error handling and timeout management

This SDK mirrors the Python implementation while following Go best practices and idioms.

## Prerequisites

The Claude CLI must be installed and available in your PATH:

```bash
npm install -g @anthropic-ai/claude-code
```

Go 1.25 or later is required.

## Installation

```bash
go get github.com/victorarias/claude-agent-sdk-go
```

## Quick Start

### Simple One-Shot Query

The simplest way to use the SDK - send a prompt and get a complete response:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/victorarias/claude-agent-sdk-go/sdk"
    "github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    messages, err := sdk.RunQuery(ctx, "What is 2+2?")
    if err != nil {
        panic(err)
    }

    for _, msg := range messages {
        if m, ok := msg.(*types.AssistantMessage); ok {
            fmt.Println(m.Text())
        }
    }
}
```

### Interactive Streaming Conversation

For multi-turn conversations with real-time output:

```go
package main

import (
    "context"
    "fmt"

    "github.com/victorarias/claude-agent-sdk-go/sdk"
    "github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
    ctx := context.Background()

    // Create client with options
    client := sdk.NewClient(
        types.WithModel("claude-sonnet-4-5"),
        types.WithMaxTurns(10),
        types.WithSystemPrompt("You are a helpful assistant."),
    )

    // Connect in streaming mode
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    defer client.Close()

    // Send a query
    if err := client.SendQuery("Tell me about Go interfaces"); err != nil {
        panic(err)
    }

    // Receive response messages
    for {
        msg, err := client.ReceiveMessage()
        if err != nil {
            panic(err)
        }

        switch m := msg.(type) {
        case *types.AssistantMessage:
            fmt.Print(m.Text())
        case *types.ResultMessage:
            fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
            return
        }
    }
}
```

## Features

- **Simple API**: One-shot queries with `RunQuery()` or streaming with `Client`
- **Streaming Support**: Real-time message streaming for interactive experiences
- **Session Management**: Resume and continue conversations across multiple queries
- **Custom Tools**: Host MCP servers to extend Claude with custom capabilities
- **Hooks System**: Pre/post tool-use hooks for logging, validation, and control
- **Permission Control**: Fine-grained control over tool execution permissions
- **Error Handling**: Comprehensive error types for robust error handling
- **Timeout Management**: Context-based timeouts for all operations
- **Model Configuration**: Support for different Claude models and configurations
- **Budget Control**: Set token and cost limits for queries

## Configuration Options

The SDK supports extensive configuration through functional options:

```go
client := sdk.NewClient(
    types.WithModel("claude-opus-4-5"),           // Choose Claude model
    types.WithMaxTurns(20),                       // Limit conversation turns
    types.WithSystemPrompt("Custom instructions"), // Set system prompt
    types.WithMaxTokens(4096),                    // Limit response tokens
    types.WithPermissionMode(types.PermissionAlways), // Tool permission mode
    types.WithTimeout(10*time.Minute),            // Set timeout
    types.WithBudget(1.0),                        // Set cost budget in USD
)
```

## Examples

Comprehensive examples are available in the [`examples/`](./examples) directory:

| Example | Description |
|---------|-------------|
| [simple](./examples/simple/) | One-shot query with minimal setup |
| [streaming](./examples/streaming/) | Interactive multi-turn conversation |
| [hooks](./examples/hooks/) | Pre/post tool use hooks for logging and control |
| [mcp-server](./examples/mcp-server/) | SDK-hosted MCP server with custom tools |
| [permissions](./examples/permissions/) | Custom tool permission handling |
| [session](./examples/session/) | Session resume and conversation continuation |
| [error-handling](./examples/error-handling/) | Error types and recovery patterns |

Run any example:

```bash
go run ./examples/simple/ "What is 2+2?"
go run ./examples/streaming/
go run ./examples/hooks/
```

See the [examples README](./examples/README.md) for more details.

## Error Handling

The SDK provides specific error types for common scenarios:

```go
import (
    "errors"
    "github.com/victorarias/claude-agent-sdk-go/types"
)

messages, err := sdk.RunQuery(ctx, prompt)
if err != nil {
    switch {
    case errors.Is(err, types.ErrCLINotFound):
        fmt.Println("Install Claude CLI: npm i -g @anthropic-ai/claude-code")
    case errors.Is(err, types.ErrCLIVersion):
        fmt.Println("Update Claude CLI: npm update -g @anthropic-ai/claude-code")
    case errors.Is(err, context.DeadlineExceeded):
        fmt.Println("Request timed out")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── sdk/            # High-level client API
├── types/          # Core types, messages, and options
├── internal/       # Internal implementation (parser, subprocess, etc.)
├── examples/       # Example applications
└── docs/           # Implementation plans and documentation
```

### Implementation Plans

Detailed implementation plans are available in [`docs/plans/`](./docs/plans):
- [01-types-interfaces.md](./docs/plans/01-types-interfaces.md) - Core types and interfaces
- [02-transport-layer.md](./docs/plans/02-transport-layer.md) - Subprocess and CLI communication
- [03-query-protocol.md](./docs/plans/03-query-protocol.md) - Message parsing and control protocol
- [04-client-api.md](./docs/plans/04-client-api.md) - Client API design
- [05-integration.md](./docs/plans/05-integration.md) - Examples and integration tests
- [06-mcp-servers.md](./docs/plans/06-mcp-servers.md) - MCP server support

## Contributing

Contributions are welcome! Please ensure all tests pass before submitting pull requests.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Reference

This implementation mirrors the official Python SDK. The Python reference is maintained in the `reference/` directory (gitignored) for comparison during development.

# Claude Agent SDK Go - Implementation Plans

> **For Claude:** Use superpowers:executing-plans to implement each plan task-by-task with TDD.

## Overview

Go port of Anthropic's Claude Agent SDK (Python). Spawns the `claude` CLI as a subprocess and handles bidirectional JSON streaming for control protocol.

**Feature Parity Target:** 100% compatibility with [claude-agent-sdk-python](https://github.com/anthropics/claude-agent-sdk-python)

## Plan Sequence

Execute plans in order. Each plan depends on the previous ones.

```
Plan 01: Types & Interfaces
         ↓
Plan 02: Transport Layer (subprocess)
         ↓
Plan 03: Query/Control Protocol
         ↓
Plan 04: Client API
         ↓
Plan 05: Integration & Examples
         ↓
Plan 06: MCP Server Support
```

## Plans Overview

| # | Plan | Status | Tasks | Est. LOC |
|---|------|--------|-------|----------|
| 1 | [Types & Interfaces](./01-types-interfaces.md) | NOT STARTED | 10 | ~800 |
| 2 | [Transport Layer](./02-transport-layer.md) | NOT STARTED | 11 | ~600 |
| 3 | [Query Protocol](./03-query-protocol.md) | NOT STARTED | 10 | ~600 |
| 4 | [Client API](./04-client-api.md) | NOT STARTED | 8 | ~400 |
| 5 | [Integration](./05-integration.md) | NOT STARTED | 10 | ~400 |
| 6 | [MCP Servers](./06-mcp-servers.md) | NOT STARTED | 7 | ~800 |

**Total: 56 tasks, ~3,600 LOC**

## Complete Feature List

### Core Features

| Feature | Plan | Description |
|---------|------|-------------|
| One-shot queries | 04 | `Query(ctx, prompt)` returns all messages |
| Streaming queries | 04 | `QueryStream(ctx, prompt)` returns channels |
| Interactive sessions | 04 | `Connect`, `SendQuery`, `ReceiveMessage` |
| Session resume | 04 | `WithResume(sessionID)` continues conversation |
| Context managers | 04 | `WithClient(Run)` for resource cleanup |

### Message Types

| Type | Plan | Description |
|------|------|-------------|
| UserMessage | 01 | User input with text content |
| AssistantMessage | 01 | Claude's response with content blocks |
| SystemMessage | 01 | System events (init, notifications) |
| ResultMessage | 01 | Final result with cost, session info |
| StreamEvent | 01 | Partial updates during streaming |

### Content Blocks

| Block | Plan | Description |
|-------|------|-------------|
| TextBlock | 01 | Plain text content |
| ThinkingBlock | 01 | Extended thinking content |
| ToolUseBlock | 01 | Tool invocation with ID, name, input |
| ToolResultBlock | 01 | Tool execution result |
| ServerToolUseBlock | 01 | MCP server tool invocation |

### Transport Layer

| Feature | Plan | Description |
|---------|------|-------------|
| CLI discovery | 02 | PATH, bundled CLI, custom path |
| Version checking | 02 | Semantic version validation (≥2.0.0) |
| Write serialization | 02 | Mutex prevents concurrent write races |
| TOCTOU prevention | 02 | State checks inside locks |
| Speculative JSON | 02 | Handles partial JSON across lines |
| Graceful shutdown | 02 | 5s timeout then force kill |
| Windows support | 02 | Command length validation (8191 chars) |

### Hook System

| Hook | Plan | Description |
|------|------|-------------|
| PreToolUse | 01, 04 | Called before tool execution |
| PostToolUse | 01, 04 | Called after tool execution |
| UserPromptSubmit | 01 | Called when user submits prompt |
| Stop | 01, 04 | Called when Claude stops |
| SubagentStop | 01 | Called when subagent stops |
| PreCompact | 01 | Called before context compaction |

### Permission System

| Feature | Plan | Description |
|---------|------|-------------|
| PermissionBypass | 01 | Skip all permission checks |
| PermissionAllow | 01 | Allow specific tool |
| PermissionAllowAlways | 01 | Allow tool for session |
| PermissionDeny | 01 | Deny specific invocation |
| PermissionDenyAll | 01 | Deny all future requests |
| canUseTool callback | 04 | Custom permission handling |

### MCP Server Support

| Feature | Plan | Description |
|---------|------|-------------|
| MCPServerBuilder | 06 | Fluent API for server creation |
| MCPToolHandler | 06 | Custom tool implementation |
| WithMCPServer | 06 | Register with client |
| MCPServerManager | 06 | Lifecycle management |
| Concurrent calls | 06 | Thread-safe tool invocation |

### Configuration Options

| Option | Plan | Description |
|--------|------|-------------|
| WithModel | 01 | Set Claude model |
| WithMaxTurns | 01 | Limit conversation turns |
| WithSystemPrompt | 01 | Set system message |
| WithCwd | 01 | Set working directory |
| WithEnv | 01 | Set environment variables |
| WithTools | 01 | Limit allowed tools |
| WithPermissionMode | 01 | Set permission behavior |
| WithCLIPath | 01 | Custom CLI binary path |
| WithTimeout | 01 | Set operation timeout |
| WithResume | 04 | Resume previous session |
| WithContinue | 04 | Continue from last message |

### Error Types

| Error | Plan | Description |
|-------|------|-------------|
| ErrCLINotFound | 01 | CLI binary not found |
| ErrVersionMismatch | 01 | CLI version too old |
| ErrConnectionFailed | 01 | Failed to connect |
| ErrProcessExited | 01 | CLI exited unexpectedly |
| ErrParseFailed | 01 | JSON parse error |
| ProcessError | 01 | Exit code and stderr |
| VersionError | 01 | Version comparison details |

### File Operations

| Feature | Plan | Description |
|---------|------|-------------|
| File checkpointing | 01 | Save file state before changes |
| RewindFiles | 03 | Restore files to checkpoint |
| FileState tracking | 01 | Track modified files |

### Sandbox Configuration

| Feature | Plan | Description |
|---------|------|-------------|
| SandboxConfig | 01 | Container isolation settings |
| WithSandbox | 01 | Enable sandboxed execution |

## Module Structure

```
claude-agent-sdk-go/
├── go.mod
├── go.sum
├── sdk.go                 # Package exports, version constants
├── doc.go                 # Package documentation
├── types.go               # All type definitions
├── options.go             # ClaudeAgentOptions with functional pattern
├── errors.go              # Error types and sentinel errors
├── transport.go           # Transport interface
├── subprocess.go          # SubprocessTransport implementation
├── query.go               # Query/control protocol
├── client.go              # Client API
├── parser.go              # Message parser
├── mcp_types.go           # MCP protocol types
├── mcp_builder.go         # MCP server builder
├── mcp_handler.go         # MCP message handler
├── mcp_transport.go       # MCP stdio transport
├── mcp_lifecycle.go       # MCP server lifecycle
├── integration_test.go    # Integration tests
└── examples/
    ├── README.md
    ├── simple/main.go
    ├── streaming/main.go
    ├── hooks/main.go
    ├── mcp-server/main.go
    ├── permissions/main.go
    ├── session/main.go
    └── error-handling/main.go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
    ctx := context.Background()

    // Simple one-shot query
    messages, err := sdk.Query(ctx, "What is 2+2?")
    if err != nil {
        log.Fatal(err)
    }

    for _, msg := range messages {
        if m, ok := msg.(*sdk.AssistantMessage); ok {
            fmt.Println(m.Text())
        }
    }
}
```

## Advanced Usage

### Streaming with Hooks

```go
client := sdk.NewClient(
    sdk.WithModel("claude-sonnet-4-5"),
    sdk.WithPreToolUseHook(func(name string, input map[string]any) *sdk.HookResult {
        fmt.Printf("Using tool: %s\n", name)
        return &sdk.HookResult{Decision: sdk.HookDecisionAllow}
    }),
)

if err := client.Connect(ctx, ""); err != nil {
    log.Fatal(err)
}
defer client.Close()

client.SendQuery("List files in current directory")
for msg := range client.Messages() {
    // Handle messages...
}
```

### Custom MCP Tools

```go
server := sdk.NewMCPServerBuilder("custom-tools").
    WithTool("weather", "Get weather for city", schema, func(input map[string]any) (any, error) {
        city := input["city"].(string)
        return fmt.Sprintf("Weather in %s: Sunny, 72°F", city), nil
    }).
    Build()

client := sdk.NewClient(sdk.WithMCPServer(server))
```

## Verification Commands

```bash
# Run all tests
go test ./... -v

# Run with race detector
go test ./... -race

# Run specific plan tests
go test -run TestTypes -v
go test -run TestTransport -v
go test -run TestQuery -v
go test -run TestClient -v
go test -run TestMCP -v

# Run integration tests (requires Claude CLI)
CLAUDE_TEST_INTEGRATION=1 go test -tags=integration -v

# Generate documentation
go doc ./...
```

## Dependencies

- Go 1.21+
- Claude CLI 2.0.0+ (`npm install -g @anthropic-ai/claude-code`)

## Implementation Notes

### Critical Concurrency Features

1. **Write Mutex (`writeMu`)**: All writes to CLI stdin are serialized to prevent interleaved JSON from concurrent MCP tool calls.

2. **TOCTOU Prevention**: State checks (e.g., `isClosed`) happen inside locks to prevent time-of-check-time-of-use races.

3. **Speculative JSON Parsing**: Uses `jsonAccumulator` to handle JSON that spans multiple lines or arrives in partial chunks.

### Python SDK Reference

The implementation mirrors the Python SDK's architecture:

- `SubprocessCLITransport` → `SubprocessTransport`
- `ClaudeCodeClient` → `Client`
- `ClaudeAgentOptions` → `Options` with functional pattern
- Message types maintain 1:1 correspondence

See `.reference/claude-agent-sdk-python/` for the Python implementation reference.

## Next Steps After Implementation

1. ✅ Create git repository
2. ✅ Push to GitHub
3. Add CI/CD with GitHub Actions
4. Publish to pkg.go.dev
5. Add benchmarks and performance tests
6. Consider adding gRPC transport option

**Plans complete and saved.** Ready for execution!

# Claude Agent SDK Go Examples

This directory contains examples demonstrating common SDK usage patterns.

## Examples

| Example | Description |
|---------|-------------|
| [simple](./simple/) | One-shot query with minimal setup |
| [streaming](./streaming/) | Interactive multi-turn conversation |
| [hooks](./hooks/) | Pre/post tool use hooks for logging and control |
| [mcp-server](./mcp-server/) | SDK-hosted MCP server with custom tools |
| [runtime-controls](./runtime-controls/) | Runtime model/permission/session/MCP control APIs |
| [unstable-session](./unstable-session/) | Unstable v2 create/resume/prompt session APIs |
| [permissions](./permissions/) | Custom tool permission handling |
| [session](./session/) | Session resume and conversation continuation |
| [error-handling](./error-handling/) | Error types and recovery patterns |

## Running Examples

All examples require the Claude CLI to be installed:

```bash
npm install -g @anthropic-ai/claude-code
```

Run any example:

```bash
go run ./examples/simple/ "What is 2+2?"
go run ./examples/streaming/
go run ./examples/hooks/
go run ./examples/runtime-controls/
go run ./examples/unstable-session/
```

## Common Patterns

### Context with Timeout
Always use context timeouts for production:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

### Defer Close
Always defer Close to ensure cleanup:
```go
client := sdk.NewClient()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Error Checking
Use errors.Is for specific error handling:
```go
if errors.Is(err, sdk.ErrCLINotFound) {
    fmt.Println("Please install: npm i -g @anthropic-ai/claude-code")
}
```

# Claude Agent SDK Go - Implementation Plans

> **For Claude:** Use superpowers:executing-plans to implement each plan task-by-task with TDD.

## Overview

Go port of Anthropic's Claude Agent SDK. Spawns the `claude` CLI as a subprocess and handles bidirectional JSON streaming for control protocol.

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
```

## Plans Overview

| # | Plan | Status | Tasks | Est. LOC |
|---|------|--------|-------|----------|
| 1 | [Types & Interfaces](./01-types-interfaces.md) | NOT STARTED | 8 | ~600 |
| 2 | [Transport Layer](./02-transport-layer.md) | NOT STARTED | 10 | ~500 |
| 3 | [Query Protocol](./03-query-protocol.md) | NOT STARTED | 9 | ~500 |
| 4 | [Client API](./04-client-api.md) | NOT STARTED | 7 | ~300 |
| 5 | [Integration](./05-integration.md) | NOT STARTED | 5 | ~200 |

**Total: 39 tasks, ~2,100 LOC**

## Module Structure

```
claude-agent-sdk-go/
├── go.mod
├── go.sum
├── sdk.go                 # Package exports
├── types.go               # All type definitions
├── options.go             # ClaudeAgentOptions
├── transport.go           # Transport interface
├── subprocess.go          # SubprocessTransport implementation
├── query.go               # Query/control protocol
├── client.go              # ClaudeSDKClient
├── parser.go              # Message parser
├── errors.go              # Error types
└── examples/
    ├── simple/main.go
    └── streaming/main.go
```

## Quick Start

```bash
# After all plans complete
go get github.com/victorarias/claude-agent-sdk-go

# Simple query
client := sdk.NewClient(nil)
defer client.Close()

for msg := range client.Query(ctx, "Hello Claude") {
    fmt.Println(msg)
}
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
```

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-12-22

Initial release of the Claude Agent SDK Go, a Go implementation of the Claude Agent SDK.

### Added

- **Core Client API**: `sdk.Query()` for one-shot queries and `sdk.NewClient()` for interactive sessions with bidirectional streaming
- **Message Types**: Complete type system for system messages, user messages, assistant messages, stream events, and control messages
- **Session Management**: Create, resume, and fork sessions with full session lifecycle control
- **Streaming Support**: Real-time streaming of assistant responses with partial message support
- **Custom Tools**: MCP (Model Context Protocol) server integration for extending Claude with custom tools
- **Hooks System**: Pre/post tool use hooks, user prompt hooks, and lifecycle hooks for fine-grained control
- **Permission Management**: Approval callbacks for tool usage and other sensitive operations
- **Agent Support**: Programmatic subagent definition and filesystem-based agent loading
- **Comprehensive Options**:
  - Model selection and configuration
  - System prompt customization (preset or custom)
  - Token budget controls (max tokens, thinking tokens, budget USD)
  - Settings sources (user, project, local)
  - Working directory and additional directories
  - Plugin loading (local plugins)
  - Tool configuration (preset or custom tool list)
  - Retry and timeout settings
- **Error Handling**: Structured error types for API errors, CLI errors, validation errors, and timeouts
- **Context Integration**: Full support for Go context.Context for cancellation and timeouts
- **Examples**: 16 comprehensive examples covering all major features:
  - Simple one-shot queries
  - Interactive streaming sessions
  - Custom tools via MCP servers
  - Hook implementations
  - Permission handling
  - Agent and subagent usage
  - Session management
  - Error handling patterns
  - Budget controls
  - Interrupt handling
  - System prompts
  - Tool configuration
  - Settings sources
  - Prompt variations
  - Advanced tool usage
- **Testing**: Full test coverage including unit tests and integration tests
- **Documentation**: Complete README with API examples, CONTRIBUTING guide, and implementation plans

### Internal

- **Subprocess Management**: Robust Claude CLI subprocess lifecycle with proper cleanup
- **Message Parser**: JSON-RPC message parsing with type inference
- **MCP Protocol**: Complete MCP protocol implementation for tool definitions and invocations
- **Control Protocol**: Approval flow, hook handling, and session management
- **Transport Layer**: Stdio-based communication with the Claude CLI process

[unreleased]: https://github.com/victorarias/claude-agent-sdk-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/victorarias/claude-agent-sdk-go/releases/tag/v0.1.0

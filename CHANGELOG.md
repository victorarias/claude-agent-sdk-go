# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1](https://github.com/victorarias/claude-agent-sdk-go/compare/v0.1.0...v0.1.1) (2025-12-25)


### Bug Fixes

* pass SDK MCP servers to CLI for tool discovery ([cc72bb1](https://github.com/victorarias/claude-agent-sdk-go/commit/cc72bb1702fc698392d26d936f85ade36fd75d70))
* set last-release-sha to v0.1.0 commit ([eb75822](https://github.com/victorarias/claude-agent-sdk-go/commit/eb75822ae2e8c970b39aac37f0e86884d1abe18c))


### Documentation

* add experimental warning to README ([4c25d07](https://github.com/victorarias/claude-agent-sdk-go/commit/4c25d07b5bd9f5db1bc4a999a346f3658c471d76))


### Miscellaneous

* remove reference SDK folder ([1413810](https://github.com/victorarias/claude-agent-sdk-go/commit/141381024d202092df926fe3b319207311f4ea5e))


### Continuous Integration

* add workflow_dispatch trigger to release-please ([0cd6278](https://github.com/victorarias/claude-agent-sdk-go/commit/0cd627851cc177a1338012a4a97f33778802331a))
* use explicit config-file and manifest-file for release-please ([a23fae5](https://github.com/victorarias/claude-agent-sdk-go/commit/a23fae56083a7417b0b88debd732de7cc30839ad))

## [Unreleased]

### Bug Fixes

- parse `result`, `structured_output`, and `usage` fields from SDK `result` messages so structured output consumers can reliably read successful query payloads
- parse `rate_limit_event` messages as first-class SDK messages so query streams no longer fail with `unknown message type: rate_limit_event`

## [0.1.0] - 2025-12-22

First official release. Not yet tested in production.

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
- **Documentation**: Complete README with API examples and CONTRIBUTING guide

### Internal

- **Subprocess Management**: Robust Claude CLI subprocess lifecycle with proper cleanup
- **Message Parser**: JSON-RPC message parsing with type inference
- **MCP Protocol**: Complete MCP protocol implementation for tool definitions and invocations
- **Control Protocol**: Approval flow, hook handling, and session management
- **Transport Layer**: Stdio-based communication with the Claude CLI process

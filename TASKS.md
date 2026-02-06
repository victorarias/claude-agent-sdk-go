# Python SDK Parity Tasks (from 2025-11-01 onward)

This file tracks missing features identified when comparing this Go port against `anthropics/claude-code-sdk-python` up to v0.1.31.

## Tasks

- [x] 1. Add MCP tool annotations support for SDK-hosted tools
  - Add annotation fields to Go MCP tool types
  - Include annotations in `tools/list` responses for SDK MCP servers
  - Add/adjust tests for annotation serialization and bridge behavior

- [x] 2. Send `agents` via initialize control request (stdin), not `--agents` CLI args
  - Stop passing `--agents` in subprocess command construction
  - Extend `Query.Initialize(...)` path to include agent payload
  - Wire client options agents into initialize request
  - Remove obsolete command-length `--agents` fallback logic/tests

- [x] 3. Add public MCP status API
  - Add `Query.GetMCPStatus()` control request (`subtype: mcp_status`)
  - Add `Client.GetMCPStatus()` wrapper method
  - Add tests for request/response handling

- [x] 4. Add `tool_use_result` support on `UserMessage`
  - Add typed field to message model
  - Parse it from raw user messages
  - Add tests for parsing and type behavior

- [x] 5. Add new hook events and typed hook inputs/outputs
  - Add events: `PostToolUseFailure`, `Notification`, `SubagentStart`, `PermissionRequest`
  - Add missing input fields: `tool_use_id` (Pre/PostToolUse), subagent stop metadata fields
  - Add hook-specific output support fields
  - Add tests for JSON/control parsing and callback typing

- [x] 6. Extend hook helper API surface
  - PreToolUse helper supports optional `additionalContext`
  - PostToolUse helper supports optional `updatedMCPToolOutput`
  - Add helpers for new events where useful
  - Add tests for helper output payloads

- [x] 7. Honor `CLAUDE_CODE_STREAM_CLOSE_TIMEOUT` for initialize timeout behavior
  - Read env var (milliseconds) and apply `max(env/1000, 60)` semantics for initialize timeout
  - Keep existing stream-close timeout behavior intact
  - Add tests for env override parsing and clamping

- [x] 8. Fail fast pending control requests on transport/read failure
  - Ensure pending control waits are unblocked with errors when transport closes/errors
  - Avoid waiting full request timeout in broken-transport scenarios
  - Add regression tests for pending request fast-fail

## Verification

- [x] Run targeted unit tests for each implemented task
- [x] Run full test suite: `go test ./...`

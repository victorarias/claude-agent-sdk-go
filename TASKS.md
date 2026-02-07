# TypeScript SDK Parity Roadmap

Source of truth for parity: `@anthropic-ai/claude-agent-sdk@0.2.34` (`sdk.d.ts` + `sdk.mjs`).

Status legend:
- [ ] pending
- [~] in progress
- [x] complete

## Implementation Order (Foundational -> Advanced)

- [x] 1. Core options + CLI flag parity foundation
  - Add missing permission modes: `delegate`, `dontAsk`
  - Add missing `Options` fields used directly by CLI launch:
    - `agent`, `session_id`, `resume_session_at`
    - `debug`, `debug_file`
    - `strict_mcp_config`
    - `allow_dangerously_skip_permissions`
    - `persist_session` (tri-state behavior; only pass `--no-session-persistence` when explicitly false)
  - Extend subprocess command building for corresponding flags:
    - `--agent`, `--session-id`, `--resume-session-at`
    - `--debug` / `--debug-file`
    - `--strict-mcp-config`
    - `--allow-dangerously-skip-permissions`
    - `--no-session-persistence`
  - Add option helpers + tests

- [x] 2. Agent definition parity + initialize payload parity
  - Extend `AgentDefinition` with TS fields (`disallowedTools`, `mcpServers`, `criticalSystemReminder_EXPERIMENTAL`, `skills`, `maxTurns`)
  - Ensure initialize request serializes all supported fields
  - Add serialization + initialize payload tests

- [x] 3. Control protocol type parity (request/response surface)
  - Add control request types/subtypes:
    - `set_max_thinking_tokens`
    - `mcp_set_servers`, `mcp_reconnect`, `mcp_toggle`
    - `rewind_files` with optional `dry_run`
  - Add initialization response typed model (`commands`, `models`, `account`, output style data)
  - Add parser tests for new control types

- [x] 4. Query/Client control API parity
  - Implement query methods:
    - `SetMaxThinkingTokens(*int / clear)` semantics
    - `InitializationResult`, `SupportedCommands`, `SupportedModels`, `AccountInfo`
    - `ReconnectMCPServer`, `ToggleMCPServer`, `SetMCPServers`
    - `RewindFiles` options/result structure parity (incl. dry-run)
  - Add `Client` wrappers with connection checks
  - Add unit tests for each control method

- [x] 5. Dynamic MCP server management parity
  - Implement runtime replacement/toggling/reconnect for process-managed MCP servers
  - Keep SDK-hosted MCP server behavior aligned with TS expectations
  - Add integration-style tests around MCP control flow

- [x] 6. Hook event/type parity expansion
  - Add missing events: `SessionStart`, `SessionEnd`, `Setup`, `TeammateIdle`, `TaskCompleted`
  - Add typed hook input/output structs + helper builders
  - Ensure parser + hook callback routing handles all events

- [x] 7. Message type parity expansion
  - Add missing SDK message types/subtypes and parser support:
    - auth/task notification/files persisted/hook progress/hook response/tool summary variants
  - Add robust unknown-message handling tests

- [x] 8. Permission callback context parity
  - Add fields from TS control payload (`decision_reason`, `tool_use_id`, `agent_id`, description)
  - Thread through `CanUseTool` context and tests

- [x] 9. Behavioral validation parity
  - Enforce TS-equivalent option validation where applicable:
    - `canUseTool` vs `permissionPromptToolName` exclusivity (already present, keep)
    - `fallbackModel != model`
    - bypass-permissions safety checks
    - continue/resume/session-id interaction constraints
  - Add validation-focused tests

- [x] 10. Docs + examples parity sweep
  - Update README/options docs for new fields and control APIs
  - Add/refresh examples for session controls + MCP management

## Current Execution

- Current step: Completed roadmap (items 1-10)
- Next step: open a new parity sweep against the newest TypeScript SDK release

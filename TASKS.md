# TypeScript SDK Full Parity Sweep (0.2.34)

Source of truth: `@anthropic-ai/claude-agent-sdk@0.2.34` (`sdk.d.ts` and runtime behavior where applicable).

Status legend:
- [ ] pending
- [~] in progress
- [x] complete

## Execution Order

- [x] 1. Message model parity foundation (assistant/user/system/result metadata)
  - Add missing message fields used by TS payloads:
    - user/assistant/system `uuid`, `session_id`
    - user `isSynthetic`, `isReplay`
    - result `uuid`, `stop_reason`, `usage`, `modelUsage`, `permission_denials`, `errors`
  - Add supporting typed structs for model usage + permission denials
  - Add/refresh unit tests for JSON marshalling compatibility

- [x] 2. Parser parity for enriched message payloads
  - Parse all new fields from item 1 for assistant/user/result/system messages
  - Ensure existing behavior remains backward compatible
  - Add parser tests for each new field family

- [x] 3. Missing system subtype models/parsing
  - Add typed system message models + parser support:
    - `status`
    - `compact_boundary`
  - Add parser tests for each subtype

- [x] 4. Query control parity: setModel clear/default semantics
  - Align `set_model` control behavior with TS optional model semantics
  - Keep backwards compatibility for existing `SetModel(string)` usage
  - Add tests for set/clear model flows

- [x] 5. Option/helper parity expansion
  - Add missing option helpers for existing fields:
    - fallback model
    - permission prompt tool name
    - additional directories
    - extra args
    - user
  - Add tests

- [x] 6. Sandbox schema parity expansion
  - Add missing sandbox network fields:
    - `allowedDomains`
    - `allowManagedDomainsOnly`
  - Add sandbox `ripgrep` config support
  - Add tests

- [x] 7. MCP status typing parity expansion
  - Extend MCP status typed model:
    - `serverInfo`
    - `tools` metadata
    - richer config typing surface (backward-compatible)
  - Add tests around typed parsing from control responses

- [x] 8. Public constants/type ergonomics parity
  - Add exported hook event list constant equivalent to TS `HOOK_EVENTS`
  - Add exported exit reason constants/equivalent typing surface
  - Add tests

- [x] 9. Unstable session API parity (Go-adapted)
  - Add `UnstableV2CreateSession`
  - Add `UnstableV2ResumeSession`
  - Add `UnstableV2Prompt`
  - Add lightweight session abstraction with send/stream/close/session-id access
  - Add tests

- [x] 10. Docs/examples parity sweep for new APIs
  - Update README and examples index for new capabilities
  - Add minimal example usage for unstable session APIs

- [x] 11. Process execution option parity
  - Add options for:
    - `pathToClaudeCodeExecutable` (alias to CLI path resolution)
    - `executable`
    - `executableArgs`
  - Update subprocess launch path to support runtime wrapper execution
  - Add option + subprocess command tests

- [x] 12. CLI preset/empty-tools argument parity
  - Align `tools` command mapping with TS/Python transport:
    - `tools=[]` => `--tools ""`
    - preset tools => `--tools default`
  - Align `system_prompt` preset mapping:
    - preset w/o append should not emit `--system-prompt`
    - preset with append should emit `--append-system-prompt`
  - Add subprocess command tests for all above cases

- [x] 13. Quality hardening + release test harness
  - Fix runtime-wrapper/native execution correctness:
    - native entrypoints ignore `executable` and `executableArgs`
    - wrapper mode supports extensionless Node/Bun/Deno shebang scripts only
  - Add parser robustness:
    - preserve structured `tool_result.content` by JSON-encoding non-string payloads
    - surface parse errors when assistant/user messages contain only invalid blocks
    - support top-level `stream_event` fallback fields (`event_type`, `index`, `delta`)
  - Add stronger end-to-end subprocess harness:
    - fake CLI script test that exercises `Client.Connect` + initialize controls
    - verify `SetModel` and `ClearModel` payload semantics through real subprocess transport
  - Add fuzz harness and CI gate for parser robustness:
    - `internal/parser/fuzz_test.go` panic-safety fuzzers
    - `make fuzz-parser` bounded fuzz target
    - `.github/workflows/fuzz.yml` scheduled + PR/push bounded fuzz runs
  - Validate with full suite + lint

## Current Execution

- Current step: Final validation + summarize
- Next step: Open follow-up parity items for typed non-string tool_result content API (if desired)

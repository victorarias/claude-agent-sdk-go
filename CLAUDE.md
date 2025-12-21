# Claude Agent SDK Go

Go implementation of the Claude Agent SDK, mirroring the Python SDK at `reference/`.

## Plans

Implementation plans are in `docs/plans/`:
- `01-types-interfaces.md` - Core types, errors, messages, options
- `02-transport-layer.md` - Subprocess management, CLI communication
- `03-query-protocol.md` - Message parsing, control protocol, hooks
- `04-client-api.md` - Client API, session management
- `05-integration.md` - Examples and integration tests
- `06-mcp-servers.md` - MCP server support

## Task Tracking (Beads)

Use `bd` (beads) for tracking work, not TodoWrite. Run `bd prime` to learn commands.

### Quick Reference

```bash
bd ready                    # Show unblocked work
bd create --title="..." --type=task --priority=2  # Create (priority 0-4)
bd update <id> --status=in_progress               # Claim work
bd close <id>               # Complete
bd dep add <child> <parent> # child depends on parent
bd sync                     # Sync with git (run at session end)
```

### Workflow

1. **Epics:** Create for features requiring 3+ tasks
2. **Plan references:** Link via `--description "docs/plans/foo.md#section"`
3. **Dependencies:** Use `bd dep add` when order matters
4. **Discovery:** Add tasks as you find bugs or new work

## Reference

Python SDK is cloned to `reference/` (gitignored) for comparison during implementation.

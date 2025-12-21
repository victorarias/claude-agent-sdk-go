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

## Beads

Use beads to track progress. Run `bd prime` to learn the CLI.

Rules:

1. **One epic per plan** - Create todos for each task in the current plan
2. **Add reference** - Include plan name and task number (e.g., "Plan01/Task3: Content blocks")
3. **Create upfront** - Set up todos before starting a batch
4. **Add as you go** - Discover bugs or subtasks? Add them immediately
5. **Mark in_progress** - Only one task at a time
6. **Mark completed** - Immediately when done, don't batch

Example:
```
[in_progress] Plan01/Task3: Define content block types
[pending] Plan01/Task4: Define message types
[pending] Plan01/Task5: Define options types
```

When bugs are found:
```
[in_progress] Plan01/Task3: Define content block types
[pending] BUG: Fix circular reference in TextBlock
[pending] Plan01/Task4: Define message types
```

## Reference

Python SDK is cloned to `reference/` (gitignored) for comparison during implementation.

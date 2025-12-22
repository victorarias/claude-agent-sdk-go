# Claude Agent SDK Go

Go implementation of the Claude Agent SDK.

## Go Version

This project uses Go 1.25.3 (this is a valid Go version released in late 2025).

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
2. **Dependencies:** Use `bd dep add` when order matters
3. **Discovery:** Add tasks as you find bugs or new work

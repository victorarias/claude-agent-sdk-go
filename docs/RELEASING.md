# Releasing

This project uses [Release Please](https://github.com/googleapis/release-please) for automated releases with [Conventional Commits](https://www.conventionalcommits.org/).

## How It Works

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Push commits   │────▶│  Release Please  │────▶│  Release PR     │
│  to main        │     │  analyzes them   │     │  created/updated│
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                          │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  GitHub Release │◀────│  Tag created     │◀────│  Merge PR       │
│  published      │     │  (v0.2.0)        │     │                 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

1. **You push commits** to `main` using conventional commit format
2. **Release Please** automatically creates/updates a "Release PR"
3. **When you're ready** to release, merge the Release PR
4. **A tag is created** and GitHub Release is published automatically

## Conventional Commits

Format: `<type>(<scope>): <description>`

### Types and Version Bumps

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor (0.X.0) |
| `fix` | Bug fix | Patch (0.0.X) |
| `docs` | Documentation only | Patch |
| `style` | Formatting, no code change | Patch |
| `refactor` | Code change, no new feature or fix | Patch |
| `perf` | Performance improvement | Patch |
| `test` | Adding tests | Patch |
| `chore` | Maintenance tasks | Patch |
| `ci` | CI/CD changes | Patch |
| `build` | Build system changes | Patch |

### Breaking Changes (Major Version)

Add `!` after type or include `BREAKING CHANGE:` in footer:

```bash
# Method 1: Add ! after type
feat!: change Client.Query signature

# Method 2: Footer
feat: redesign hook system

BREAKING CHANGE: HookCallback signature changed from (input) to (input, context)
```

### Examples

```bash
# Patch release (0.0.X)
git commit -m "fix: handle nil pointer in parser"
git commit -m "docs: update README examples"
git commit -m "test: add coverage for edge cases"

# Minor release (0.X.0)
git commit -m "feat: add streaming support"
git commit -m "feat(hooks): add pre-compact hook type"

# Major release (X.0.0)
git commit -m "feat!: rename Client to Agent"
git commit -m "fix!: change error types to implement unwrap"
```

### Scope (Optional)

Scope indicates which part of the codebase changed:

```bash
feat(sdk): add new query option
fix(parser): handle malformed JSON
docs(examples): add streaming example
test(hooks): improve coverage
```

Common scopes for this project:
- `sdk` - Main SDK package
- `types` - Types package
- `parser` - Internal parser
- `subprocess` - Subprocess management
- `mcp` - MCP server support
- `hooks` - Hook system
- `examples` - Example code

## Release Process

### Regular Release

1. **Work normally** - commit with conventional commits
2. **Check the Release PR** - Release Please creates/updates it automatically
3. **Review the changelog** - Ensure it looks correct
4. **Merge when ready** - This triggers the release

### Manual Release (Emergency)

If you need to release without waiting for Release Please:

```bash
# 1. Update version in .release-please-manifest.json
# 2. Update CHANGELOG.md manually
# 3. Commit changes
git commit -m "chore: release v0.2.0"

# 4. Create and push tag
git tag v0.2.0
git push origin v0.2.0
```

## Configuration Files

| File | Purpose |
|------|---------|
| `.github/workflows/release-please.yml` | GitHub Actions workflow |
| `release-please-config.json` | Release Please configuration |
| `.release-please-manifest.json` | Current version tracking |

## FAQ

### Why didn't my commit appear in the changelog?

- Commits must follow conventional commit format exactly
- The type must be one of: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `build`
- Merge commits and commits without a type are ignored

### How do I release multiple changes at once?

Just keep merging to main. Release Please accumulates all changes in the Release PR until you merge it.

### Can I edit the changelog before release?

Yes! Edit the Release PR directly. Release Please will preserve your edits.

### What if I need to skip a release?

Just don't merge the Release PR. It will keep accumulating changes.

### How do I see what will be in the next release?

Check the open Release PR - it shows all pending changes and the proposed version.

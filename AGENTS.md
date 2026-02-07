# Agent Instructions

## Session Completion

When ending a work session, complete these steps:

1. Run quality gates (tests, lint, build) for changed code.
2. Create follow-up issues for anything intentionally deferred.
3. Push all commits to the remote branch.
4. Confirm a clean working tree and branch sync:

```bash
git pull --rebase
git push
git status
```

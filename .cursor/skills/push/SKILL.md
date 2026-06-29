---
name: push
description: Sync myLife vault changes to GitHub via jj describe and jj git push. Use when the user says pushして, GitHubに同期.
---

# push

## Prerequisites

- jj-managed repo (`.jj/` present)
- Remote: `git@github.com:solle458/Competitive-Programming-eXecutor.git`

## Workflow

1. `jj status` and `goreleaser release --snapshot --clean`
2. If no changes, report and exit
3. Show diff summary to user
4. `jj describe -r @ -m "<user-approved summary>"`
5. Optional: `jj git fetch`
6. `jj git push --bookmark main`
7. Report success or conflict (never force push)
8. Reasoning new version(format v{major}.{minor}.{patch})
8. `git tag {new-version}`
9. `git push origin {new-version}`

## Rules

- **Never** `--force` push
- Do not commit `.cursor/mcp.json`, secrets, or `.env`
- Offer after any Skill that wrote to the vault

## Commit message examples

- `tasks: english comment`
- `cursor: myLife MVP Skills`

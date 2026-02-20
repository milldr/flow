# PRD-003: flow prompt

**Depends on:** [002-mvp.md](./002-mvp.md)

## Summary

`flow prompt` takes a natural language description of what the user is working on and uses AI to populate the workspace state file with repos, branches, and paths.

## Command

```bash
flow prompt <name>
```

The workspace must already exist (created via `flow init`).

## Flow

1. User runs `flow prompt <name>`
2. Command reads user input (interactive text prompt or stdin)
3. Input is sent to an AI model along with available context
4. AI returns structured output: description, repos, branches, paths
5. State file at `~/.flow/workspaces/<name>/state.yaml` is updated
6. Summary of changes is printed

## Input

The user provides a free-form text description. Examples:

```
Add IPv6 support to VPC module across atmos and docs repos
```

```
Fix the authentication bug in api-gateway. The issue is in the token validation
middleware. Also need to update the shared auth library and the integration tests
in the e2e repo.
```

```
Set up a new feature branch for the billing migration. Repos: billing-service,
payment-gateway, and the terraform infra repo. Branch off main for all of them.
```

Input can come from:
- Interactive prompt (default)
- Stdin (`flow prompt calm-delta < description.txt`)
- Inline argument (`flow prompt calm-delta "Add IPv6 support..."`)

## Context

To make useful inferences, the AI needs context beyond just the prompt. Available context sources:

| Source | How | What it provides |
|--------|-----|-----------------|
| GitHub orgs/repos | `gh` CLI | List of repos the user has access to |
| Git config | `git config` | User's name, email, default branch conventions |
| Existing workspaces | `~/.flow/workspaces/` | Previously used repos, branches, URL patterns |
| Current state file | `state.yaml` | Any manually added repos (additive updates) |

The AI should use context to:
- Resolve short repo names ("atmos") to full URLs ("github.com/cloudposse/atmos")
- Infer branch naming conventions from existing workspaces
- Suggest default branches when not specified
- Avoid duplicating repos already in the state file

## Output

### State file update

The AI populates/updates these fields:
- `metadata.description` — cleaned-up description from the prompt
- `metadata.name` — may rename from the generated name to something meaningful
- `spec.repos[].url` — full repo URL
- `spec.repos[].branch` — branch name (existing or suggested new branch)
- `spec.repos[].path` — local path for the worktree

### Terminal output

```
📦 Updating workspace: calm-delta

  Description: Add IPv6 support to VPC module
  Name: vpc-ipv6

  Repos:
    + github.com/cloudposse/atmos @ feature/ipv6 → ./atmos
    + github.com/cloudposse/docs @ feature/ipv6 → ./docs

✅ State updated

  flow render vpc-ipv6    # Create worktrees
```

### Confirmation

Before writing, show the user what will change and ask for confirmation. The user should be able to:
- Accept all changes
- Reject and re-prompt
- Edit the state file directly instead

## AI Provider

Use the Anthropic API (Claude) for parsing. Configuration:

| Setting | Source | Default |
|---------|--------|---------|
| API key | `$ANTHROPIC_API_KEY` | Required |
| Model | `$FLOW_MODEL` | `claude-sonnet-4-20250514` |

If no API key is set, print a helpful error:

```
Error: ANTHROPIC_API_KEY is not set.

flow prompt requires an Anthropic API key to use AI parsing.
Get one at: https://console.anthropic.com/
```

## Prompt Design

The AI prompt should:
1. Include the user's description
2. Include available context (repos, orgs, existing state)
3. Request structured JSON output matching the state file schema
4. Be specific about field formats (full URLs, valid branch names, relative paths)

The response should be parsed as structured JSON, validated, and applied to the state file.

## Edge Cases

- **No repos mentioned:** Ask the user to be more specific, or create state with just the description
- **Ambiguous repo names:** Present options and let the user choose
- **AI unavailable:** Fail with a clear error, suggest `flow state` as fallback
- **Existing repos in state:** Additive — don't remove repos already in the state file
- **Rate limits:** Retry with backoff, fail gracefully

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| AI provider | Anthropic (Claude) | Best structured output, tool use support |
| Input format | Free-form text | Lowest friction |
| Output format | Structured JSON from AI | Reliable parsing |
| Context gathering | `gh` CLI + local state | No extra auth setup if gh is configured |
| Confirmation | Required before write | Prevent unexpected state changes |

## Acceptance Criteria

- [ ] `flow prompt <name>` reads user input and calls AI
- [ ] AI populates state file with description, repos, branches, paths
- [ ] User sees a summary and confirms before state is written
- [ ] Works with repos from multiple GitHub orgs
- [ ] Falls back gracefully when `gh` CLI is not available
- [ ] Helpful error when `ANTHROPIC_API_KEY` is not set
- [ ] Additive — does not remove existing repos from state
- [ ] Stdin input works (`echo "..." | flow prompt <name>`)

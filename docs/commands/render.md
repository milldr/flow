# flow render

Materialize the workspace. Clones repos (or fetches if cached), creates worktrees for each repo/branch pair.

![flow render](gif/render.gif)

## Usage

```
flow render <workspace>
```

## Examples

```bash
flow render vpc-ipv6       # by name
flow render calm-delta     # by ID
```

## Behavior

- Idempotent — safe to run repeatedly
- Existing worktrees are skipped
- Cached bare repos are fetched for updates
- New branches are created from the default branch if they don't exist in the remote

## Output

```
✓ Workspace ready

  flow exec demo -- <command>
  flow exec demo -- cursor .
  flow exec demo -- claude
```

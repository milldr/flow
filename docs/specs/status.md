# Status

The status spec defines how Flow resolves the status of each repo in a workspace. Statuses are evaluated top-to-bottom — the first passing check wins. A workspace's overall status is the least-advanced repo (highest index in the spec).

## Location

```
~/.flow/status.yaml                              # Global default
~/.flow/workspaces/<workspace-id>/status.yaml     # Workspace override (fully replaces global)
```

## Schema

```yaml
apiVersion: flow/v1
kind: Status
spec:
  statuses:
    - name: closed
      description: All PRs merged or closed
      check: >-
        gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state merged --json number
        | jq -e 'length > 0' > /dev/null 2>&1
        && gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json number
        | jq -e 'length == 0' > /dev/null 2>&1
    - name: in-review
      description: Non-draft PR open
      check: >-
        gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json isDraft
        | jq -e 'map(select(.isDraft == false)) | length > 0' > /dev/null 2>&1
    - name: in-progress
      description: Local diffs or draft PR
      check: >-
        git -C "$FLOW_REPO_PATH" diff --name-only "origin/$FLOW_REPO_BRANCH" 2>/dev/null | grep -q .
        || gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json isDraft
        | jq -e 'map(select(.isDraft)) | length > 0' > /dev/null 2>&1
    - name: open
      description: Workspace created, no changes yet
      default: true
```

## Fields

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | Must be `flow/v1` |
| `kind` | Yes | Must be `Status` |
| `spec.statuses[]` | Yes | Must contain at least one entry |
| `spec.statuses[].name` | Yes | Unique status name |
| `spec.statuses[].description` | No | Human-readable description |
| `spec.statuses[].check` | Yes* | Shell command evaluated via `sh -c` (*not required for the default entry) |
| `spec.statuses[].default` | No | Exactly one entry must have `default: true` |

## Environment Variables

Each check command receives the following environment variables:

| Variable | Description |
|----------|-------------|
| `FLOW_REPO_URL` | Repo URL from the state file |
| `FLOW_REPO_BRANCH` | Branch name from the state file |
| `FLOW_REPO_PATH` | Directory path in the workspace |

## Resolution

1. Checks run top-to-bottom per repo. The first check that exits `0` determines that repo's status.
2. If no check passes, the default status is used.
3. A workspace's overall status is the least-advanced repo — the one with the highest index in the statuses list.
4. If a workspace has its own `status.yaml`, it fully replaces the global spec (no merging).

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
statuses:
  - name: closed
    description: PR merged
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state merged --json number | jq -e 'length > 0' > /dev/null 2>&1
  - name: in-review
    description: PR open
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length > 0' > /dev/null 2>&1
  - name: in-progress
    description: Active development
    default: true
```

## Fields

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | Must be `flow/v1` |
| `kind` | Yes | Must be `Status` |
| `statuses` | Yes | Must contain at least one entry |
| `statuses[].name` | Yes | Unique status name |
| `statuses[].description` | No | Human-readable description |
| `statuses[].check` | Yes* | Shell command evaluated via `sh -c` (*not required for the default entry) |
| `statuses[].default` | No | Exactly one entry must have `default: true` |

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

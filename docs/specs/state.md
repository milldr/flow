# State

Each workspace is defined by a `state.yaml` file. Run `flow edit state <workspace>` to open it in your editor.

## Location

```
~/.flow/workspaces/<workspace-id>/state.yaml
```

## Schema

```yaml
apiVersion: flow/v1
kind: State
metadata:
  name: vpc-ipv6
  description: IPv6 support across VPC services
  created: "2026-02-18T12:00:00Z"
spec:
  repos:
    - url: github.com/acme/vpc-service
      branch: feature/ipv6
    - url: github.com/acme/subnet-manager
      branch: feature/ipv6
      path: subnet-manager    # optional, defaults to repo name
```

## Fields

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | Must be `flow/v1` |
| `kind` | Yes | Must be `State` |
| `metadata.name` | No | Human-friendly workspace name |
| `metadata.description` | No | Optional description |
| `metadata.created` | Yes | RFC 3339 timestamp (set automatically on init) |
| `spec.repos` | Yes | Must contain at least one repo |
| `spec.repos[].url` | Yes | Git remote URL |
| `spec.repos[].branch` | Yes | Branch to check out |
| `spec.repos[].path` | No | Directory name in the workspace (defaults to repo name) |

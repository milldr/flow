# flow state

Open the workspace state file in your editor (`$EDITOR` or `vim`).

![flow state](gif/state.gif)

## Usage

```
flow state <workspace>
```

## Examples

```bash
flow state vpc-ipv6        # by name
flow state calm-delta      # by ID
```

## State file format

```yaml
apiVersion: flow/v1
kind: State
metadata:
  name: vpc-ipv6
  description: IPv6 support across VPC services
  created: "2026-02-18T12:00:00Z"
spec:
  repos:
    - url: git@github.com:acme/vpc-service.git
      branch: feature/ipv6
      path: vpc-service
    - url: git@github.com:acme/subnet-manager.git
      branch: feature/ipv6
```

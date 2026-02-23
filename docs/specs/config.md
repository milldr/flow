# Config

The global config file stores Flow-wide settings.

## Location

```
~/.flow/config.yaml
```

This file is created automatically on first run.

## Schema

```yaml
apiVersion: flow/v1
kind: Config
spec:
  agents:
    - name: claude
      exec: claude
      default: true
    - name: cursor
      exec: cursor .
```

## Fields

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | Must be `flow/v1` |
| `kind` | Yes | Must be `Config` |
| `spec.agents[]` | No | List of agent tools. The agent marked `default: true` is shown in `flow render` output. When omitted, a generic `<command>` placeholder is shown. |
| `spec.agents[].name` | Yes | Identifier for the agent |
| `spec.agents[].exec` | Yes | Command to run via `flow exec` |
| `spec.agents[].default` | No | When `true`, this agent's `exec` is used in render output |

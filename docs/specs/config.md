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
```

## Fields

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | Must be `flow/v1` |
| `kind` | Yes | Must be `Config` |

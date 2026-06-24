# Copilot Slim — Commands

## Build

```bash
make build          # → bin/slim
go run ./cmd/slim setup --all
```

## CLI

All commands accept `--target copilot` (default). Cursor and Claude return not-implemented in Phase 1.

```bash
slim setup [--all] [--module NAME]... [--dry-run]
slim off [--all] [--module NAME]...
slim status
slim doctor [--json]
slim config list
slim config set FEATURE on|off
slim logs
slim paths
```

## Logs and reports

| Path | Contents |
|------|----------|
| `~/.copilot/slim/manifest.json` | Installed modules |
| `~/.copilot/slim/config.yaml` | Feature on/off |
| `~/.copilot/slim/logs/hook.log` | Hook events |
| `~/.copilot/slim/sessions/*.json` | Session stop summaries |
| `~/.copilot/slim/bin/slim` | Installed binary (hooks call this) |
| `~/.copilot/hooks/copilot-slim.json` | Hook registration |

## Feature counts

`setup`, `status`, and `doctor` read `templates/feature-registry.yaml` and print:

```
Modules: N/M installed | Features: N/M active (X OFF, Y not shipped)
```

Add features to the registry to update counts automatically.

## Testing

See [scripts/README.md](../scripts/README.md) for build and test scripts.

```bash
make test-all
```

## Troubleshooting

1. `slim doctor` — find OFF or misconfigured features  
2. `slim config set bash_compress off` — disable one hook feature  
3. `slim off --module hooks` — remove hooks, keep skills  
4. `slim off --all` — full opt-out  

See [PLAN.md](PLAN.md) for full troubleshooting workflow.

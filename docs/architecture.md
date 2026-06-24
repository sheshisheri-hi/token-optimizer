# Architecture

Copilot Slim is a Go CLI with IDE-agnostic core and per-IDE runtimes.

```
cmd/slim          → CLI entry
internal/core/    → manifest, config, registry, measurement, jsonmerge
internal/hook/    → hook bridge + compress (shared across IDEs)
internal/runtime/
  copilot/        → ~/.copilot/*
  cursor/         → ~/.cursor/*
  claude/         → stub
```

## Data flow

1. `slim setup --target cursor` installs binary to `~/.cursor/slim/bin/slim`
2. Cursor calls hooks in `~/.cursor/hooks.json`
3. Hook bridge reads `config.yaml`, may return `updated_input` / `updatedInput`
4. Events logged to `logs/hook.log`; compressions to `events.jsonl`

## Extension

Add `internal/runtime/<ide>/` implementing `runtime.Runtime`. Shared hook logic stays in `internal/hook/`.

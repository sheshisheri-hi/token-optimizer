# Copilot Slim

Org-owned toolkit to reduce **AI coding agent token usage** in **Cursor** and **GitHub Copilot (VS Code)**.

Hooks intercept verbose tool output before it hits the model. Skills and settings cut always-on context bloat.

| Doc | Purpose |
|-----|---------|
| [docs/PLAN.md](docs/PLAN.md) | Full design and roadmap |
| [docs/commands.md](docs/commands.md) | CLI reference |
| [docs/architecture.md](docs/architecture.md) | How it fits together |
| [docs/troubleshooting.md](docs/troubleshooting.md) | When something breaks |
| [docs/onboarding.md](docs/onboarding.md) | 4-week adoption guide |
| [scripts/README.md](scripts/README.md) | Automated test scripts |

---

## What it does

| Layer | What | Token impact |
|-------|------|----------------|
| **Hooks** | Rewrite safe shell commands (`git log`, `rg`, `pytest`…) to shorter output | **Highest** — trims dynamic per-step bloat |
| **Skills** | 4 global skills: output-control, token-economy, mcp-hygiene, workflow-habits | High — smaller always-on guidance |
| **Settings** | Merge MCP hygiene + IDE defaults into your config | Medium |
| **Measurement** | Log compressions and sessions locally (`events.jsonl`) | ROI tracking — no cloud telemetry |

### Hook features (on by default)

| Feature | Behavior |
|---------|----------|
| `bash_compress` | `git log` → `git log --oneline -30`, quieter test runners, etc. |
| `search_compress` | Cap `rg` / `grep` / `find` with `head -80` |
| `session_logging` | Write session summary on `stop` |
| `session_continuity` | Checkpoint in `sessions/latest.json` for next `session-start` |
| `lean_output_nudges` | Observe-only logging (full injection blocked upstream) |
| `loop_detection` | Off by default — logs repeated commands |

### What it saves (examples)

| Before | After (hook rewrite) |
|--------|----------------------|
| `git log` (hundreds of lines) | `git log --oneline -30` |
| `rg pattern .` (unbounded) | `rg pattern . \| head -80` |
| `pytest tests/` (verbose) | `pytest tests/ -q --tb=no` |

Report rollup:

```bash
./bin/slim --target cursor report --days 7
# Last 7 days: N events | N compressions | ~N bytes saved | N sessions logged
```

---

## Quick start

### Cursor (recommended if you use Cursor daily)

```bash
git clone https://github.com/sheshisheri-hi/token-optimizer.git
cd token-optimizer
make build
./bin/slim --target cursor setup --all
./bin/slim --target cursor doctor --probe
```

**Restart Cursor** (`Cmd+Shift+P` → Reload Window).

### GitHub Copilot (VS Code)

```bash
make build
./bin/slim setup --all          # default target: copilot
./bin/slim doctor --probe
```

**Restart VS Code** after setup.

### Preview paths (no writes)

```bash
./scripts/show-paths.sh
./bin/slim --target cursor setup --all --dry-run
```

---

## What gets installed

### Cursor (`--target cursor`)

| Path | Purpose |
|------|---------|
| `~/.cursor/slim/bin/slim` | Installed CLI (hooks call this) |
| `~/.cursor/slim/manifest.json` | Install ledger |
| `~/.cursor/slim/config.yaml` | Feature toggles |
| `~/.cursor/slim/logs/hook.log` | Hook activity |
| `~/.cursor/slim/sessions/` | Session summaries + `latest.json` |
| `~/.cursor/slim/events.jsonl` | Compression/session events (measurement) |
| `~/.cursor/slim/capabilities.json` | Per-feature probe status |
| `~/.cursor/hooks.json` | **Merged** — slim hooks added, existing hooks kept |
| `~/.cursor/skills/*/SKILL.md` | 4 compressed skills |
| `~/.cursor/settings.json` | MCP hygiene fragment |
| `~/Library/Application Support/Cursor/User/settings.json` | Cursor IDE defaults (macOS) |

### Copilot (`--target copilot`, default)

| Path | Purpose |
|------|---------|
| `~/.copilot/slim/` | Data home (same layout as Cursor) |
| `~/.copilot/hooks/copilot-slim.json` | Dedicated hook file (does not touch other hooks) |
| `~/.copilot/skills/` | Global skills |
| `~/.copilot/settings.json` | Copilot settings merge |
| `~/Library/Application Support/Code/User/settings.json` | VS Code settings merge (macOS) |

Both targets are independent — you can run both.

---

## Commands

| Command | Purpose |
|---------|---------|
| `./bin/slim setup --all` | Install Copilot modules (idempotent) |
| `./bin/slim --target cursor setup --all` | Install Cursor modules |
| `./bin/slim doctor` / `doctor --probe` | Health check; `--probe` refreshes capabilities |
| `./bin/slim status` | Module and feature counts |
| `./bin/slim logs` | Log paths + last 20 hook lines |
| `./bin/slim report --days 7` | Compression/session rollup |
| `./bin/slim config list` | Feature toggles |
| `./bin/slim config set bash_compress off` | Disable one feature |
| `./bin/slim off --module hooks` | Opt out of hooks only |
| `./bin/slim paths` | Path cheat sheet |

Use `--target cursor` on any command for Cursor data home.

---

## How to test (manual — your machine)

`make test-all` validates the **code** in isolated temp homes. To verify **your** install after setup:

### 1. Agent test in Cursor

1. Open **Agent** chat (not Ask-only).
2. Ask it to run a shell command, e.g. *"Run `git status` in this repo."*
3. End the turn.

### 2. Check hooks fired

```bash
./bin/slim --target cursor logs
```

**Expect** lines like:

```
target=cursor event=pre-tool-use
target=cursor event=post-tool-use
target=cursor event=stop
```

### 3. Test compression directly

```bash
echo '{"tool_name":"Shell","tool_input":{"command":"git log"}}' \
  | ./bin/slim internal hook pre-tool-use --target cursor
```

**Expect:**

```json
{"permission":"allow","updated_input":{"command":"git log --oneline -30"}}
```

Then confirm it was recorded:

```bash
grep "tool rewrite" ~/.cursor/slim/logs/hook.log
./bin/slim --target cursor report --days 1
```

**Expect:** `1 compressions` (or more) and `~512 bytes saved` per rewrite.

### 4. Session checkpoint

```bash
cat ~/.cursor/slim/sessions/latest.json
```

**Expect:** `endedAt`, `hint` for session continuity.

### 5. Doctor

```bash
./bin/slim --target cursor doctor
```

**Expect:** hooks, skills, settings, vscode, measurement = **OK**. No `hook.log not found` warning after step 2.

### Pass/fail checklist

| Check | Pass? |
|-------|-------|
| `hook.log` has `pre-tool-use` / `stop` | |
| Manual `git log` returns `oneline -30` | |
| `report` shows compressions ≥ 1 | |
| `sessions/latest.json` exists | |
| `doctor` — hooks module OK | |

---

## Automated tests (developers)

```bash
make build
make test-all          # build + unit + smoke + integration + hooks + cursor
```

| Script | What it checks |
|--------|----------------|
| `test-smoke` | Dry-run, doctor, config (reads your home, no install) |
| `test-integration` | Full Copilot setup in **temp HOME** |
| `test-cursor` | Full Cursor setup in **temp HOME** |
| `test-hooks` | Hook events + log files in **temp HOME** |

See [scripts/README.md](scripts/README.md) for expected output per script.

---

## Doctor output (what it means)

```
Modules: 5/6 installed | Features: 4/6 active (2 OFF, 0 not shipped)
```

| Status | Meaning |
|--------|---------|
| **OK** | Module installed and feature active |
| **OFF** | Not installed or disabled in `config.yaml` |
| **DEGRADED** | Active but limited (e.g. lean_output observe-only) |

Common warnings:

| Warning | Meaning |
|---------|---------|
| `hook.log not found yet` | Normal before first agent session — run step 2 above |
| `hooks-rtk OFF` | Optional RTK proxy — install [rtk](https://github.com/rtk-ai/rtk) and `slim setup --module hooks-rtk` |

---

## Per-feature opt-out

```bash
# Disable one hook feature (instant, no reinstall)
./bin/slim --target cursor config set bash_compress off

# Disable all hooks, keep skills
./bin/slim --target cursor off --module hooks

# Env override (one shot)
ORG_SLIM_BASH_COMPRESS=0 cursor   # if slim invoked from hook env
```

---

## Repo templates (per-project)

Copy into your project — **not** installed globally:

```
templates/repo/
├── .github/copilot-instructions.md
├── .github/instructions/api.instructions.md   # applyTo example
├── agents/token-saver.agent.md
└── .copilotignore
```

---

## Architecture

- **Go CLI** (`bin/slim`) with IDE-agnostic `internal/core` and per-IDE `internal/runtime/*`
- **Copilot** and **Cursor** fully supported; **Claude** stub (`--target claude` → not available yet)
- Shared compression logic in `internal/hook/compress/`

---

## Problem statement

Background and research: [problem-statements/token-optimizer.md](problem-statements/token-optimizer.md)

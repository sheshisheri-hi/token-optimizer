# Scripts — build and test Copilot Slim

Runnable helpers for local development and CI. All scripts assume you are at the repo root (or call them via `./scripts/...`).

## Quick reference

| Script | What it does | Touches `~/.copilot`? |
|--------|----------------|------------------------|
| [`build.sh`](build.sh) | `make build` → `bin/slim` | No |
| [`test-unit.sh`](test-unit.sh) | `go test ./...` | No |
| [`test-smoke.sh`](test-smoke.sh) | CLI dry-run, status, doctor, paths | **Read only** (no install) |
| [`test-integration.sh`](test-integration.sh) | Full `setup` / `off` cycle | **No** — uses temp `HOME` |
| [`test-hooks.sh`](test-hooks.sh) | Hook events + log/session files | **No** — uses temp `HOME` |
| [`show-paths.sh`](show-paths.sh) | Dry-run install paths for **your** `~/.copilot` | **No** — preview only |
| [`test-all.sh`](test-all.sh) | Runs all of the above in order | No (except smoke reads real home) |

**Want to see paths without running tests?**

```bash
./scripts/show-paths.sh    # dry-run on your real HOME — lists every target path
```

```bash
# Typical local workflow
./scripts/build.sh
./scripts/test-smoke.sh          # fast, safe
./scripts/test-integration.sh    # full install in isolated HOME
./scripts/test-all.sh            # everything
```

Makefile shortcuts:

```bash
make build
make test          # same as test-unit.sh
make test-smoke
make test-all
```

---

## show-paths.sh

**Command:** `./scripts/show-paths.sh`

**Safe preview** of every path slim touches on **your real machine** — uses `setup --all --dry-run` only.

**Expect output including:**

- `would copy binary → ~/.copilot/slim/bin/slim`
- `would write hook → ~/.copilot/hooks/copilot-slim.json`
- `would write ~/.copilot/skills/.../SKILL.md` (×4)
- `settings: deferred` / `vscode: deferred` (not written yet)
- Summary table of binary, manifest, config, hooks, skills, logs

---

## build.sh

**Command:** `./scripts/build.sh`

**Expect:**

- Exit code `0`
- Binary at `bin/slim`
- Prints slim help header

**If it fails:** install Go 1.21+ and run `go mod tidy` from repo root.

---

## test-unit.sh

**Command:** `./scripts/test-unit.sh`

**Expect:**

- Exit code `0`
- Packages without `_test.go` show `[no test files]` — normal until unit tests are added

**What it validates:** Go packages compile and any `_test.go` files pass.

---

## test-smoke.sh

**Command:** `./scripts/test-smoke.sh`

**Safe to run anytime.** Does not run `setup --all` without `--dry-run`. Only reads your real `~/.copilot/slim` for `status` / `doctor` / `config list`.

| Check | Command exercised | Expected |
|-------|-------------------|----------|
| Help | `slim --help` | Contains "token optimization", "setup" |
| Dry-run | `slim setup --all --dry-run` | Full path list printed in cyan (`would copy`, `would write`); settings/vscode show `deferred` |
| Hooks module | `slim setup --module hooks --dry-run` | `would write hook` |
| Status | `slim status` | `Modules: N/M installed` |
| Doctor | `slim doctor` | `MODULES`, feature list, tally footer |
| JSON | `slim doctor --json` | Valid JSON with `"Modules"` |
| Paths | `slim paths` | Command cheat sheet |
| Config | `slim config list` | `bash_compress` and other toggles |
| Claude stub | `slim --target claude setup --all` | Exit non-zero, `not available yet` |
| Cursor dry-run | `slim --target cursor setup --all --dry-run` | `would merge hooks`, skill paths |

---

## test-integration.sh

**Command:** `./scripts/test-integration.sh`

Uses a **temporary `HOME`** (`mktemp`) so your real Copilot profile is never modified. Cleaned up on exit.

| Step | What happens | Expected |
|------|----------------|----------|
| `setup --all` | Installs to `$TEST_HOME/.copilot/` | Cyan lines: `install …/slim/bin/slim`, `write …/hooks/copilot-slim.json`, skill paths |
| Path summary | `print_copilot_paths` | Full table of binary, manifest, config, hooks, skills, logs |
| Hook JSON | Parses registration file | References `internal hook` commands |
| `status` | After install | At least 1 module installed |
| `doctor` | After install | `hooks` module `OK` |
| `config set` | Disable bash_compress | `config.yaml` contains `bash_compress: false` |
| Hook events | Pipe JSON to `internal hook` | `hook.log` created, `sessions/*.json` after `stop` |
| `logs` | Print paths | Shows `hook.log` path |
| `off --module hooks` | Partial uninstall | Completes without error |
| Re-install | `setup --module hooks` | Hook JSON present again |

---

## test-hooks.sh

**Command:** `./scripts/test-hooks.sh`

Focused hook bridge test in isolated `HOME` (hooks module only).

| Event | Input | Expected |
|-------|--------|----------|
| `session-start` | `{}` | Exit `0`, line in `hook.log` |
| `pre-tool-use` | `{}` or bash payload | Exit `0`, line in `hook.log` |
| `post-tool-use` | `{}` | Exit `0` |
| `stop` | `{}` | Exit `0`, `sessions/*.json` created |

**Note:** `pre-tool-use` with `git log` may produce **empty stdout** today (observe-only mode). When bash compression ships, expect JSON with `updatedInput` on stdout.

---

## test-all.sh

**Command:** `./scripts/test-all.sh`

Runs: `build` → `test-unit` → `test-smoke` → `test-integration` → `test-hooks`

Stops on first failure. Use in CI or before opening a PR.

---

## Manual tests (real Copilot / VS Code)

These are **not** automated — run after `./bin/slim setup --all` on your machine:

1. **Restart VS Code** after setup.
2. Open Copilot agent chat and run a bash tool (e.g. `git log`).
3. **Expect:** `~/.copilot/slim/logs/hook.log` gains new lines with `event=pre-tool-use`.
4. End the agent session.
5. **Expect:** new file under `~/.copilot/slim/sessions/`.
6. Run `./bin/slim doctor` — hooks/skills modules should show `OK` when installed.

---

## Troubleshooting scripts

| Problem | Fix |
|---------|-----|
| `Binary not found: bin/slim` | Run `./scripts/build.sh` |
| Integration test permission errors | Ensure `mktemp` works; don't run as root |
| Smoke `doctor` shows 0/6 modules | Normal if you never ran `setup --all` on real home |
| `test-all` fails at integration | Run `./scripts/test-integration.sh` alone to see which assert failed |

---

## Adding new tests

1. Add Go tests under `internal/.../*_test.go` — picked up by `test-unit.sh`.
2. Add CLI checks to `test-smoke.sh` (read-only) or `test-integration.sh` (isolated HOME).
3. Document the new check in this README under the relevant script section.

Shared assertions live in [`lib/common.sh`](lib/common.sh): `pass`, `fail`, `assert_output_contains`, `assert_file`, `assert_fails_with_pattern`, `print_summary`.

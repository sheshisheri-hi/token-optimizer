#!/usr/bin/env bash
# Shared helpers for Copilot Slim test scripts.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
SLIM="${SLIM_BIN:-$ROOT_DIR/bin/slim}"

PASS=0
FAIL=0
SKIP=0

bold()  { printf '\033[1m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
red()   { printf '\033[31m%s\033[0m\n' "$*"; }
yellow(){ printf '\033[33m%s\033[0m\n' "$*"; }
cyan()  { printf '\033[36m  → %s\033[0m\n' "$*"; }
dim()   { printf '\033[2m  %s\033[0m\n' "$*"; }

# info_path LABEL PATH — always print resolved path (exists or not)
info_path() {
  local label="$1"
  local path="$2"
  if [[ -e "$path" ]]; then
    cyan "$label: $path"
  else
    yellow "  → $label: $path (not present)"
  fi
}

# show_cmd_output TITLE CMD... — run command and print full stdout (for path visibility)
show_cmd_output() {
  local title="$1"
  shift
  bold "$title"
  echo "   $*"
  local out
  set +e
  out="$("$@" 2>&1)"
  local code=$?
  set -e
  echo "$out" | while IFS= read -r line; do
    if [[ "$line" =~ ^(install|write|would\ copy|would\ write|archived|settings:|vscode:) ]]; then
      cyan "$line"
    else
      dim "$line"
    fi
  done
  return "$code"
}

pass() {
  PASS=$((PASS + 1))
  green "  PASS: $*"
}

fail() {
  FAIL=$((FAIL + 1))
  red "  FAIL: $*"
}

skip() {
  SKIP=$((SKIP + 1))
  yellow "  SKIP: $*"
}

section() {
  echo
  bold "== $* =="
}

require_slim() {
  if [[ ! -x "$SLIM" ]]; then
    red "Binary not found: $SLIM"
    echo "Run: ./scripts/build.sh   or   make build"
    exit 1
  fi
}

# run_step NAME CMD...
# Runs a command; on failure records FAIL and returns 1 (does not exit shell).
run_step() {
  local name="$1"
  shift
  echo ">> $name"
  echo "   $*"
  if "$@"; then
    pass "$name"
    return 0
  else
    fail "$name (exit $?)"
    return 1
  fi
}

# assert_exit CODE CMD...
assert_exit() {
  local expected="$1"
  shift
  set +e
  "$@" >/dev/null 2>&1
  local actual=$?
  set -e
  if [[ "$actual" -eq "$expected" ]]; then
    pass "exit $expected: $*"
    return 0
  fi
  fail "expected exit $expected, got $actual: $*"
  return 1
}

# assert_fails_with_pattern PATTERN CMD...
assert_fails_with_pattern() {
  local pattern="$1"
  shift
  local out
  set +e
  out="$("$@" 2>&1)"
  local code=$?
  set -e
  if [[ "$code" -eq 0 ]]; then
    fail "expected failure, got exit 0: $*"
    return 1
  fi
  if echo "$out" | grep -qE "$pattern"; then
    pass "failed with /$pattern/: $*"
    return 0
  fi
  fail "failed (exit $code) but output missing /$pattern/"
  echo "--- output ---"
  echo "$out"
  echo "--------------"
  return 1
}

assert_output_contains() {
  local pattern="$1"
  shift
  local out
  set +e
  out="$("$@" 2>&1)"
  local code=$?
  set -e
  if [[ "$code" -ne 0 ]]; then
    fail "command failed (exit $code): $*"
    return 1
  fi
  if echo "$out" | grep -qE "$pattern"; then
    pass "output matches /$pattern/: $*"
    return 0
  fi
  fail "output missing /$pattern/"
  echo "--- output ---"
  echo "$out"
  echo "--------------"
  return 1
}

# assert_file PATH [description]
assert_file() {
  local path="$1"
  local desc="${2:-$path}"
  if [[ -f "$path" ]]; then
    pass "file exists: $desc"
    cyan "$path"
    return 0
  fi
  fail "file missing: $desc ($path)"
  return 1
}

# assert_dir PATH [description]
assert_dir() {
  local path="$1"
  local desc="${2:-$path}"
  if [[ -d "$path" ]]; then
    pass "dir exists: $desc"
    cyan "$path"
    return 0
  fi
  fail "dir missing: $desc ($path)"
  return 1
}

# print_copilot_paths HOME_DIR — summary of where slim installs (test or real home)
print_copilot_paths() {
  local home="${1:-$HOME}"
  local copilot="$home/.copilot"
  local slim="$copilot/slim"
  section "Install paths (under HOME=$home)"
  echo "Binary & data:"
  info_path "slim CLI (installed copy)" "$slim/bin/slim"
  info_path "manifest" "$slim/manifest.json"
  info_path "config (feature toggles)" "$slim/config.yaml"
  info_path "hook activity log" "$slim/logs/hook.log"
  info_path "session summaries" "$slim/sessions/"
  echo
  echo "Copilot integration:"
  info_path "hook registration" "$copilot/hooks/copilot-slim.json"
  echo
  echo "Skills (global):"
  for skill in output-control token-economy mcp-hygiene workflow-habits; do
    info_path "$skill" "$copilot/skills/$skill/SKILL.md"
  done
  echo
  echo "Settings (Phase 1 — not auto-merged yet):"
  dim "~/.copilot/settings.json — deferred (settings module stub)"
  dim "VS Code User/settings.json — deferred (vscode module stub)"
}

print_summary() {
  echo
  bold "Summary"
  echo "  Passed: $PASS"
  echo "  Failed: $FAIL"
  echo "  Skipped: $SKIP"
  if [[ "$FAIL" -gt 0 ]]; then
    red "Some checks failed."
    exit 1
  fi
  green "All checks passed."
}

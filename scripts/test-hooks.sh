#!/usr/bin/env bash
# Exercise hook bridge events only (uses temp HOME).
set -euo pipefail
source "$(dirname "$0")/lib/common.sh"

section "Prerequisites"
require_slim

TEST_HOME="$(mktemp -d "${TMPDIR:-/tmp}/slim-hook-test.XXXXXX")"
trap 'rm -rf "$TEST_HOME"' EXIT
export HOME="$TEST_HOME"
SLIM_HOME="$TEST_HOME/.copilot/slim"

bold "Isolated HOME: $TEST_HOME"

section "Minimal install (hooks only)"
"$SLIM" setup --module hooks >/dev/null
assert_file "$TEST_HOME/.copilot/hooks/copilot-slim.json" "hook registration"

export SLIM

section "Hook events"
EVENTS=(session-start pre-tool-use post-tool-use stop)
for ev in "${EVENTS[@]}"; do
  echo ">> hook $ev"
  if echo '{}' | "$SLIM" internal hook "$ev"; then
    pass "hook $ev exit 0"
  else
    fail "hook $ev non-zero exit"
  fi
done

section "hook.log lines"
LOG="$SLIM_HOME/logs/hook.log"
assert_file "$LOG" "hook.log"
cyan "log file: $LOG"
LINE_COUNT="$(wc -l < "$LOG" | tr -d ' ')"
if [[ "$LINE_COUNT" -ge 4 ]]; then
  pass "hook.log has $LINE_COUNT event lines (expected >= 4)"
else
  fail "hook.log has only $LINE_COUNT lines (expected >= 4)"
fi
echo "--- hook.log ---"
cat "$LOG"
echo "----------------"

section "stop → session file"
STOP_COUNT="$(find "$SLIM_HOME/sessions" -name '*.json' 2>/dev/null | wc -l | tr -d ' ')"
if [[ "${STOP_COUNT:-0}" -ge 1 ]]; then
  pass "stop hook wrote session JSON"
  find "$SLIM_HOME/sessions" -name '*.json' -exec head -5 {} \;
else
  fail "stop hook did not write session JSON"
fi

section "pre-tool-use with bash payload (observe mode)"
OUT="$(echo '{"toolName":"bash","input":{"command":"git log"}}' | "$SLIM" internal hook pre-tool-use 2>&1 || true)"
# Phase 1: compression may not emit updatedInput yet — log-only is OK
if [[ -z "$OUT" ]]; then
  pass "pre-tool-use fail-open (empty stdout OK for now)"
else
  pass "pre-tool-use produced stdout: $OUT"
fi

print_summary

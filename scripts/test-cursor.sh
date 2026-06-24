#!/usr/bin/env bash
# Cursor target tests (isolated HOME).
set -euo pipefail
source "$(dirname "$0")/lib/common.sh"

section "Prerequisites"
require_slim
export SLIM_TARGET="--target cursor"

TEST_HOME="$(mktemp -d "${TMPDIR:-/tmp}/slim-cursor-test.XXXXXX")"
trap 'rm -rf "$TEST_HOME"' EXIT
export HOME="$TEST_HOME"

bold "Isolated HOME: $TEST_HOME"

section "setup --all (cursor)"
show_cmd_output "cursor setup" "$SLIM" --target cursor setup --all
assert_file "$TEST_HOME/.cursor/hooks.json" "hooks.json"
assert_file "$TEST_HOME/.cursor/slim/manifest.json" "manifest"

section "compression"
OUT="$(echo '{"tool_name":"Shell","tool_input":{"command":"git log"}}' | "$SLIM" --target cursor internal hook pre-tool-use)"
if echo "$OUT" | grep -q 'oneline'; then
  pass "cursor bash compression stdout"
else
  fail "missing compression in: $OUT"
fi

section "doctor"
assert_output_contains "MODULES" "$SLIM" --target cursor doctor

print_summary

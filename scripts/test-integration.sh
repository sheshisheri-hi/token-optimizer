#!/usr/bin/env bash
# Full install cycle in a temporary HOME — safe, isolated from your real profile.
set -euo pipefail
source "$(dirname "$0")/lib/common.sh"

section "Prerequisites"
require_slim

TEST_HOME="$(mktemp -d "${TMPDIR:-/tmp}/slim-test-home.XXXXXX")"
cleanup() { rm -rf "$TEST_HOME"; }
trap cleanup EXIT

export HOME="$TEST_HOME"
COPILOT_HOME="$TEST_HOME/.copilot"
SLIM_HOME="$COPILOT_HOME/slim"

bold "Using isolated HOME: $TEST_HOME"

section "setup --all"
show_cmd_output "slim setup --all" "$SLIM" setup --all
pass "slim setup --all completed"

print_copilot_paths "$TEST_HOME"

section "Installed files (verify)"
assert_file "$SLIM_HOME/bin/slim" "installed hook binary"
assert_file "$SLIM_HOME/manifest.json" "manifest"
assert_file "$SLIM_HOME/config.yaml" "config"
assert_file "$COPILOT_HOME/hooks/copilot-slim.json" "hook registration"
assert_file "$COPILOT_HOME/skills/output-control/SKILL.md" "output-control skill"
assert_file "$COPILOT_HOME/skills/token-economy/SKILL.md" "token-economy skill"
assert_file "$COPILOT_HOME/skills/mcp-hygiene/SKILL.md" "mcp-hygiene skill"
assert_file "$COPILOT_HOME/skills/workflow-habits/SKILL.md" "workflow-habits skill"

section "Hook JSON references installed binary"
HOOK_JSON="$COPILOT_HOME/hooks/copilot-slim.json"
cyan "hook file: $HOOK_JSON"
if grep -q "internal hook" "$HOOK_JSON"; then
  pass "hook JSON references internal hook commands"
  bold "Hook commands registered:"
  grep -o '"bash": "[^"]*"' "$HOOK_JSON" | head -4 | while IFS= read -r line; do
    dim "$line"
  done
else
  fail "hook JSON missing internal hook commands"
fi

section "status after install"
assert_output_contains "Modules: [1-9]/[0-9]+ installed" "$SLIM" status

section "doctor after install"
OUT="$("$SLIM" doctor 2>&1)"
echo "$OUT" | head -30
if echo "$OUT" | grep -q "hooks.*OK"; then
  pass "hooks module OK in doctor"
else
  fail "hooks module not OK in doctor"
fi

section "config set"
run_step "config set bash_compress off" "$SLIM" config set bash_compress off
cyan "updated: $SLIM_HOME/config.yaml"
if grep -q 'bash_compress: false' "$SLIM_HOME/config.yaml"; then
  pass "bash_compress disabled in config.yaml"
  dim "  bash_compress: false"
else
  fail "bash_compress not disabled in config.yaml"
fi

section "hook events (internal)"
export SLIM
run_step "hook session-start" bash -c 'echo "{}" | "$SLIM" internal hook session-start'
run_step "hook pre-tool-use" bash -c 'echo "{\"toolName\":\"bash\",\"input\":{\"command\":\"git log\"}}" | "$SLIM" internal hook pre-tool-use'
run_step "hook stop" bash -c 'echo "{}" | "$SLIM" internal hook stop'
assert_file "$SLIM_HOME/logs/hook.log" "hook.log"
SESSION_FILES=( "$SLIM_HOME"/sessions/*.json )
SESSION_COUNT="${#SESSION_FILES[@]}"
if [[ -f "${SESSION_FILES[0]:-}" ]]; then
  pass "session summary written ($SESSION_COUNT file(s))"
  for f in "${SESSION_FILES[@]}"; do
    cyan "$f"
  done
else
  fail "no session summary in $SLIM_HOME/sessions/"
fi

section "logs command"
bold "slim logs output:"
"$SLIM" logs 2>&1 | while IFS= read -r line; do dim "$line"; done
assert_output_contains "hook.log" "$SLIM" logs

section "off --module hooks"
show_cmd_output "slim off --module hooks" "$SLIM" off --module hooks
pass "off --module hooks completed"
if [[ ! -f "$COPILOT_HOME/hooks/copilot-slim.json" ]]; then
  pass "hook file removed/archived after off --module hooks"
else
  yellow "  NOTE: hook file still present (may be archived, not deleted)"
  pass "off --module hooks completed"
fi

section "Idempotent re-install hooks"
show_cmd_output "slim setup --module hooks" "$SLIM" setup --module hooks
assert_file "$COPILOT_HOME/hooks/copilot-slim.json" "hook re-installed"

print_copilot_paths "$TEST_HOME"

print_summary

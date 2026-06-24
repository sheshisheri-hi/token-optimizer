#!/usr/bin/env bash
# Read-only CLI smoke tests — does not install to your real ~/.copilot.
set -euo pipefail
source "$(dirname "$0")/lib/common.sh"

section "Prerequisites"
require_slim

section "Help and version"
assert_output_contains "Copilot token usage" "$SLIM" --help
assert_output_contains "setup" "$SLIM" --help

section "Dry-run setup (no disk writes)"
show_cmd_output "slim setup --all --dry-run" "$SLIM" setup --all --dry-run
assert_output_contains "would copy binary" "$SLIM" setup --all --dry-run
assert_output_contains "would write hook" "$SLIM" setup --all --dry-run
assert_output_contains "output-control" "$SLIM" setup --all --dry-run
assert_output_contains "dry-run: manifest and config not written" "$SLIM" setup --all --dry-run

section "Dry-run paths on this machine (real ~/.copilot)"
bold "These are the paths slim would use on your profile (dry-run only):"
show_cmd_output "preview" "$SLIM" setup --all --dry-run

section "Single module dry-run"
show_cmd_output "hooks only" "$SLIM" setup --module hooks --dry-run

section "Status and doctor (reads ~/.copilot/slim if present)"
assert_output_contains "Modules:" "$SLIM" status
assert_output_contains "Modules:" "$SLIM" doctor
assert_output_contains "MODULES" "$SLIM" doctor

section "Doctor JSON"
assert_output_contains '"Modules"' "$SLIM" doctor --json

section "Paths cheat sheet"
assert_output_contains "slim setup" "$SLIM" paths

section "Config list (defaults when not installed)"
assert_output_contains "bash_compress" "$SLIM" config list

section "Cursor dry-run"
show_cmd_output "cursor preview" "$SLIM" --target cursor setup --all --dry-run
assert_output_contains "would merge hooks" "$SLIM" --target cursor setup --all --dry-run

section "Unimplemented targets"
assert_fails_with_pattern "not available yet" "$SLIM" --target claude setup --all

print_summary

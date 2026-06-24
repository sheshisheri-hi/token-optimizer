#!/usr/bin/env bash
# Show where slim installs files — dry-run on your real home, no writes.
set -euo pipefail
source "$(dirname "$0")/lib/common.sh"

require_slim

bold "Copilot Slim — install path preview"
dim "No files are written (dry-run). HOME=$HOME"
echo

show_cmd_output "slim setup --all --dry-run" "$SLIM" setup --all --dry-run

echo
print_copilot_paths "$HOME"

echo
bold "VS Code / Copilot settings"
yellow "  settings and vscode modules are Phase 1 stubs — not merged yet."
dim "  When implemented:"
dim "    ~/.copilot/settings.json          (MCP hygiene)"
dim "    ~/Library/Application Support/Code/User/settings.json   (macOS)"
dim "    ~/.config/Code/User/settings.json (Linux)"

echo
bold "Cursor (your daily driver)"
show_cmd_output "slim --target cursor setup --all --dry-run" "$SLIM" --target cursor setup --all --dry-run
dim "  ~/.cursor/hooks.json   (merged, not overwritten)"
dim "  ~/.cursor/slim/        (data home)"
dim "  ~/.cursor/skills/      (global skills)"

echo
bold "After real install, also run (from repo root):"
dim "  ./bin/slim logs     # log paths + recent hook lines"
dim "  ./bin/slim paths    # command cheat sheet"
dim "  ./bin/slim doctor   # per-module health"
echo
dim "Or add to PATH:  export PATH=\"$ROOT_DIR/bin:\$PATH\"  then run: slim logs"

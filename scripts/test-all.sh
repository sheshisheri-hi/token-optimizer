#!/usr/bin/env bash
# Run every test script in order. Stops on first script failure.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SCRIPTS=(
  build.sh
  test-unit.sh
  test-smoke.sh
  test-integration.sh
  test-hooks.sh
  test-cursor.sh
)

FAILED=0
for s in "${SCRIPTS[@]}"; do
  echo
  printf '\033[1m########################################\033[0m\n'
  printf '\033[1m# %s\033[0m\n' "$s"
  printf '\033[1m########################################\033[0m\n'
  if bash "$ROOT/scripts/$s"; then
    echo "OK: $s"
  else
    echo "FAILED: $s"
    FAILED=$((FAILED + 1))
    break
  fi
done

echo
if [[ "$FAILED" -gt 0 ]]; then
  echo "test-all: FAILED at $s"
  exit 1
fi
echo "test-all: all scripts passed."

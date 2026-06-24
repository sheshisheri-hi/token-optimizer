#!/usr/bin/env bash
# Run Go unit tests (internal packages).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "Running: go test ./..."
if go test ./... -count=1; then
  echo "Unit tests passed."
else
  echo "Unit tests failed."
  exit 1
fi

# When no _test.go files exist yet, go test still exits 0 with:
#   ?   github.com/.../cmd/slim   [no test files]
echo
echo "Note: packages without _test.go files report [no test files] — that is OK for now."

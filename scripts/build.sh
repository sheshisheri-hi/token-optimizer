#!/usr/bin/env bash
# Build bin/slim from repo root.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
make build
echo "Built: $ROOT/bin/slim"
"$ROOT/bin/slim" --help >/dev/null 2>&1 || true
exec "$ROOT/bin/slim" --help 2>&1 | head -3

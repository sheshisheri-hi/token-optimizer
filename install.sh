#!/usr/bin/env bash
# Deprecated bootstrap — builds or runs slim setup.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
if [[ -x "$ROOT/bin/slim" ]]; then
  exec "$ROOT/bin/slim" setup "$@"
fi
echo "install.sh: building slim then running setup..."
make -C "$ROOT" build
exec "$ROOT/bin/slim" setup --all "$@"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT="${OUTPUT:-$ROOT_DIR/bin/navigation}"
GOCACHE="${GOCACHE:-$ROOT_DIR/.gocache}"

mkdir -p "$(dirname "$OUTPUT")"
mkdir -p "$GOCACHE"

cd "$ROOT_DIR"
if [ ! -d "$ROOT_DIR/frontend/node_modules" ]; then
  (cd "$ROOT_DIR/frontend" && npm ci)
fi
(cd "$ROOT_DIR/frontend" && npm run build)
GOCACHE="$GOCACHE" CGO_ENABLED="${CGO_ENABLED:-1}" go build -trimpath -ldflags="-s -w" -o "$OUTPUT" .

echo "Built $OUTPUT"

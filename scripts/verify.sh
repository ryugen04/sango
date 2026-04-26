#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

export GOCACHE="$ROOT_DIR/.cache/go-build"
mkdir -p "$GOCACHE"

go test ./...
go run . audit inventory --root . --format text

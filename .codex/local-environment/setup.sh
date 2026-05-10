#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

worktree_name="${SANGO_CODEX_WORKTREE:-codex/local}"

if go run . worktree list 2>/dev/null | grep -q "^[* ] ${worktree_name}[[:space:]]"; then
  echo "[sango] Codex worktree already exists: ${worktree_name}"
  exit 0
fi

go run . worktree create "${worktree_name}"

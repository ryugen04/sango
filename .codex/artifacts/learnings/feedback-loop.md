# Feedback Loop Ledger

## Permanent Changes

- `internal/log/collector.go`: Collector を JSONL 契約へ戻し、`go test ./...` の failure を解消した
- `.codex/hooks/*`: Codex-only / ai-dlc 前提の local hook を追加した
- `.codex/rules/*`: `go test`, `go run`, `make verify` を軸にした project rules を追加した
- `.github/workflows/*`: CI と tag-based release を repo 内へ追加した
- `Makefile`, `scripts/verify.sh`: 次セッションでも同じ検証入口を使えるようにした

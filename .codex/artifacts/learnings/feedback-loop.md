# Feedback Loop Ledger

## Permanent Changes

- `internal/log/collector.go`: Collector を JSONL 契約へ戻し、`go test ./...` の failure を解消した
- `.codex/hooks/*`: Codex-only / ai-dlc 前提の local hook を追加した
- `.codex/rules/*`: `go test`, `go run`, `make verify` を軸にした project rules を追加した
- `.github/workflows/*`: CI と tag-based release を repo 内へ追加した
- `Makefile`, `scripts/verify.sh`: 次セッションでも同じ検証入口を使えるようにした
- `.codex/docs/migration-status-roadmap.md`: 移行の完了/未完了を汎用形式で追跡できるようにした
- `.codex/docs/asset-portability-and-troubleshoot.md`: 再利用資産の分類と troubleshoot 運用モデルを標準化した
- `.codex/docs/templates/troubleshoot-card.yaml`: 障害知見を YAML カードとして管理する雛形を追加した

## Recovery Learnings

- active plan に必須見出しが欠けた場合、PreToolUse hook が Git mutating 操作だけでなく plan 修復と `.codex/artifacts/learnings/**` への記録まで止め、自己回復不能になり得る
- 改善案は、active plan 不備時でも `.codex/plans/active/**` と `.codex/artifacts/learnings/**` への修復操作は許可し、`commit` / `push` / `PR create` などの Git mutating 操作だけを継続して止めること

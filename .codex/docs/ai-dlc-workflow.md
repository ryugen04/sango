# ai-dlc Workflow for sango

## Objective

- Codex 単独で次のセッションに再入しやすい運用を repo 内に残す
- 実装だけでなく、plan / execution / learnings を同じ repo で追跡する
- GitHub 上の CI / release と local verify を一致させる

## Working Agreement

- 実装前に active plan を作成する
- 実装中は対象テストを優先し、最後に `go test ./...` を実行する
- セッション終了前に execution artifact を更新する
- 問題の再発防止は `.codex/artifacts/learnings/` に残す
- Claude 系コマンドは使わない

## Standard Commands

- `make verify`
- `go test ./...`
- `go run . audit inventory --root . --format text`

## Artifacts

- plan: `.codex/plans/active/`
- execution: `.codex/artifacts/execution/`
- learnings: `.codex/artifacts/learnings/`

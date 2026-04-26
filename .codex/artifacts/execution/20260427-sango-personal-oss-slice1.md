# sango personal OSS slice 1 execution

対象 plan:
- path: `.codex/plans/active/20260426-sango-cli-tui-agent-roadmap.md`
- status: in_progress

## Summary

- `sango` の価値提案を個人 OSS 向けに再定義した
- README / docs / `init` を新しい方向性へ寄せた
- `status` `doctor` `runbook` `troubleshoot` `worktree status` に JSON 出力を追加した
- `version` コマンドを追加し、release workflow でタグ埋め込みするようにした
- 全体計画の handoff 文書として `.codex/docs/personal-oss-roadmap.md` を追加した

## Commands

- `go test ./...`
- `make verify`
- `go run . --help`
- `go run . version`
- `go run . status --json --config testdata/valid.yaml`
- `go run . doctor --json --config testdata/valid.yaml`
- `go run . runbook search api --json --config testdata/valid.yaml`
- `go run . troubleshoot --json --config testdata/valid.yaml`
- `go run . worktree status --json --config testdata/valid.yaml`

## Validation

- `go test ./...` は成功
- `make verify` は成功
- `status` / `doctor` / `runbook` / `troubleshoot` / `worktree status` の JSON 出力を確認
- root help に `version` コマンドが露出していることを確認

## Remaining Tasks

- `status` / `logs` / `doctor` の text 出力整合性を詰める
- README / docs のコマンド説明を実装ベースで棚卸しする
- 「最初の10分」で価値が伝わる導入例を追加する
- dashboard の責務を最小化し、必要なら回帰検知を追加する

# sango Codex Workflow

`sango` は Codex 主体、`ai-dlc` 前提で運用する。
このディレクトリは project 固有の plan / artifacts / hooks / rules の source of truth とし、共通思想は dotfiles の `codex-common` を参照する。

## Structure

- `config.toml`: project-scoped Codex config
- `hooks.json`, `hooks/`: project local hooks
- `rules/`: Codex 実行ルール
- `plans/templates/`: ai-dlc 用テンプレート
- `plans/active/`: 実行中 plan
- `plans/archive/`: 完了済み plan
- `artifacts/execution/`: 実装・検証結果
- `artifacts/learnings/`: 恒久化した learnings / feedback loop

## ai-dlc Flow

1. `plan`: `.codex/plans/active/` に active plan を作る
2. `implement`: plan に沿って実装する
3. `validate`: テストと CLI 確認を行う
4. `execution`: `.codex/artifacts/execution/` に結果を残す
5. `learnings`: `.codex/artifacts/learnings/` に再発防止を残す

## Codex-only Policy

- この repo では Claude wrapper を前提にしない
- review artifact は任意だが、execution artifact は必須
- `claude -p` や `claude-plan-review` などの Claude 実行は project hook/rule で禁止する
- 実装は active plan を source of truth とする

## Test Policy

- 変更中は対象テストを優先する
- 最終確認では `go test ./...` を通す
- ローカル検証は `make verify` を正本とする

## Release Policy

- CI は push / pull_request で `make verify` を実行する
- release は `v*` tag で GitHub Release を作成する
- バイナリは `sango` として publish する

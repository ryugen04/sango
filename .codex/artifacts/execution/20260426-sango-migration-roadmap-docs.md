# migration roadmap docs execution

対象 plan:
- path: `.codex/plans/active/20260426-sango-migration-roadmap-docs.md`
- status: approved

## Summary

- 移行の完了/未完了を管理するため、`migration-status-roadmap.md` を追加した
- 旧 `.claude` / YAML 資産の再利用方針と troubleshoot 管理モデルを `asset-portability-and-troubleshoot.md` に整理した
- troubleshoot カードの共通雛形として `templates/troubleshoot-card.yaml` を追加した
- `.codex/README.md` に追加ドキュメントへの導線を追記した

## Commands

- `GOCACHE=$(pwd)/.cache/go-build make verify`
- `GOCACHE=$(pwd)/.cache/go-build go test ./cmd ./internal/audit`
- `rg -n -i '/Users/|github.com/' .codex/docs .codex/README.md .codex/plans/active/20260426-sango-migration-roadmap-docs.md`

## Validation

- `make verify` は sandbox の port bind 制限で `internal/port`, `internal/process` が失敗した
- `go test ./cmd ./internal/audit` は成功した
- 新規/更新ドキュメントに禁止識別子のヒットはなかった

## Remaining

- ポート bind が可能な実環境で `make verify` を再確認する

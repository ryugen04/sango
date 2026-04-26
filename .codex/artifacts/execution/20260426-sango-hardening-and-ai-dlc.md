# sango hardening execution

対象 plan:
- path: `.codex/plans/active/20260426-sango-hardening-and-ai-dlc.md`
- status: approved
- review artifact: なし（Codex-only workflow）

## Summary

- `internal/log/collector.go` を修正し、stdout/stderr を JSONL として収集する契約を回復した
- `.codex` に config/hooks/rules/templates/docs を追加し、Codex-only / ai-dlc 運用を repo 内へ定着させた
- `Makefile`, `scripts/verify.sh`, `.github/workflows/ci.yml`, `.github/workflows/release.yml` を追加し、ローカル検証と GitHub 導線を揃えた

## Commands

- `GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build make verify`
- `GOCACHE=$(pwd)/.cache/go-build go run . audit inventory --root . --format text`
- `rg -n 'sango-tools|github.com/sango-tools/sango' .`

## Validation

- `go test ./...` が成功した
- `make verify` が成功した
- `audit inventory` は warning 継続で exit 0 を確認した
- `sango-tools` 残存は 0 件だった

## Gate Check

- active plan 前提を満たしていたか: yes
- execution artifact を作成したか: yes
- feedback open 解消: 該当なし

## Risks

- release workflow は GitHub 上の tag push 実行を未確認
- hook/rules は Codex CLI 実行環境での実運用を今後 1 セッション確認したい

## Remaining Tasks

- tag ベース release の実地確認
- 必要なら README に配布バイナリのインストール例を追加

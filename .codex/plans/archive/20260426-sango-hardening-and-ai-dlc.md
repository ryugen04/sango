---
id: 20260426-sango-hardening-and-ai-dlc
title: sango hardening（full test 復旧 + Codex ai-dlc 整備 + CI/release）
status: approved
owner: codex
created_at: 2026-04-26T11:10:00+09:00
updated_at: 2026-04-26T11:10:00+09:00
related_issue:
active_branch: main
task_size: large
split_strategy: log修正 -> Codex workflow整備 -> CI/release -> 検証
artifacts:
  plan:
    - .codex/plans/active/20260426-sango-hardening-and-ai-dlc.md
  execution:
    - .codex/artifacts/execution/20260426-sango-hardening-and-ai-dlc.md
  learnings:
    - .codex/artifacts/learnings/feedback-loop.md
    - .codex/artifacts/learnings/feedback-loop.jsonl
---

# sango hardening（full test 復旧 + Codex ai-dlc 整備 + CI/release）

## Context

- `~/dev/projects/sango` は `sango-tree` から移植済みだが、`go test ./...` は `internal/log` で失敗している。
- 今後の別リモートセッションでも継続しやすいように、repo 配下 `.codex` を Codex-only / ai-dlc 前提で整備したい。
- GitHub 上の `github.com/ryugen04/sango` を前提に、CI と release 導線を repo 内へ用意する。

## 完了基準

- [x] `go test ./...` が成功する
- [x] `.codex` に config/hooks/rules/templates/docs が揃い、Claude 依存なしで運用方針が明文化されている
- [x] `make verify` と GitHub Actions の CI/release 導線が追加されている
- [x] execution / learnings が更新されている
- [x] ローカル commit が作成されている

## Phase Checklist

### Phase 1: Investigation
- [x] `internal/log` 失敗原因を確定する（Owner: Codex）
- [x] `codex-common` の ai-dlc 構成から流用対象を決める

### Phase 2: Implementation
- [x] `internal/log` を修正して JSONL 契約を回復する
- [x] `.codex` の config/hooks/rules/templates/docs を Codex-only で整備する
- [x] CI/release/verify 導線を追加する

### Phase 3: Verification
- [x] `go test ./...` を実行して成功する
- [x] `make verify` と `go run . audit inventory --root . --format text` を確認する
- [x] `sango-tools` / Claude 依存の不要残存がないことを確認する

### Phase 4: Close
- [x] execution artifact を更新する
- [x] learnings / feedback-loop を更新する
- [x] ローカル commit を作成する

## Agent Assignment

| Phase | Owner | Support | Output |
| --- | --- | --- | --- |
| Investigation | Codex | - | failure analysis, workflow design |
| Implementation | Codex | - | code, docs, workflows |
| Verification | Codex | - | test results |
| Close | Codex | - | execution, learnings, commit |

## Review Loop

1. active plan を更新して作業範囲を固定する
2. 実装と検証を行う
3. execution / learnings を更新する
4. commit を作成して次セッションに引き継ぐ

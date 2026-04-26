---
id: 20260426-sango-migration-roadmap-docs
title: 移行状況・再利用資産・troubleshoot運用の文書化
status: approved
owner: codex
created_at: 2026-04-26T12:30:00+09:00
updated_at: 2026-04-26T12:52:00+09:00
related_issue:
active_branch: codex/bootstrap-hardening
task_size: medium
split_strategy: active plan更新 -> 汎用ドキュメント追加 -> 検証 -> commit
artifacts:
  plan:
    - .codex/plans/active/20260426-sango-migration-roadmap-docs.md
  execution:
    - .codex/artifacts/execution/20260426-sango-migration-roadmap-docs.md
  learnings:
    - .codex/artifacts/learnings/feedback-loop.md
---

# 移行状況・再利用資産・troubleshoot運用の文書化

## Context

- 移行作業の進捗と未完了項目を、次セッションでも読める形で残したい。
- 旧運用の `.claude` / YAML 資産のうち、汎用化して他プロジェクトへ使えるものを整理したい。
- リポジトリには project 固有情報を残さず、再利用できる方針と運用モデルを明文化したい。

## 完了基準

- [x] 移行の「完了済み / 未完了」を文書化したドキュメントがある
- [x] `.claude` / YAML 資産の「テンプレート化対象 / ライブラリ化対象 / 非移植対象」が明記されている
- [x] troubleshoot の表現形式と運用ライフサイクルが定義されている
- [x] execution artifact を追加し、関連検証コマンドを実行する
- [x] ローカル commit を作成する

## Phase Checklist

### Phase 1: Investigation
- [x] 既存 `.codex` / README / 実装から移行済み項目を抽出する
- [x] 文書化対象の未整備領域を確定する

### Phase 2: Implementation
- [x] migration status / roadmap ドキュメントを追加する
- [x] 再利用資産と troubleshoot モデルのドキュメントを追加する
- [x] `.codex/README.md` に参照導線を追加する

### Phase 3: Verification
- [x] ドキュメント整合性を確認する
- [x] 最低限の回帰確認コマンドを実行する

### Phase 4: Close
- [x] execution artifact を追加する
- [x] commit を作成する

## Agent Assignment

| Phase | Owner | Support | Output |
| --- | --- | --- | --- |
| Investigation | Codex | - | migration notes |
| Implementation | Codex | - | docs update |
| Verification | Codex | - | command results |
| Close | Codex | - | execution + commit |

## Review Loop

1. active plan を作成して作業範囲を固定する
2. ドキュメントを更新する
3. 検証コマンドを実行する
4. execution artifact を残して commit する

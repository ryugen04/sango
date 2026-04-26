---
id: YYYYMMDD-topic-slug
title: タイトル
status: draft
owner: codex
created_at: 2026-04-26T00:00:00+09:00
updated_at: 2026-04-26T00:00:00+09:00
related_issue:
active_branch:
task_size: small # small | medium | large
split_strategy:
artifacts:
  plan:
  execution:
  learnings:
---

# Title

## Context

- 背景
- 現状
- 参照ドキュメント

## 完了基準

- [ ] ユーザーに見える完了条件
- [ ] 技術的な完了条件
- [ ] 必須テストと確認

## Phase Checklist

### Phase 1: Investigation

- [ ] 必要な調査対象を列挙する（Owner: Codex）
- [ ] 影響範囲を確定する

### Phase 2: Implementation

- [ ] 実装する
- [ ] 必要なテストを追加・更新する
- [ ] 関連ドキュメントを更新する

### Phase 3: Verification

- [ ] 対象テストを実行する
- [ ] CLI / CI / release 導線を確認する

### Phase 4: Close

- [ ] execution artifact を残す
- [ ] learnings を反映する
- [ ] 必要なら commit / handoff を記録する

## Agent Assignment

| Phase | Owner | Support | Output |
| --- | --- | --- | --- |
| Investigation | Codex | - | context notes |
| Implementation | Codex | - | code changes |
| Verification | Codex | - | test results |
| Close | Codex | - | execution, handoff |

## Review Loop

1. plan を作成する
2. 実装と検証を行う
3. execution / learnings を残して close する

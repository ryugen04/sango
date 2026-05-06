---
id: 202605071130-worktree-interactive-cli
title: worktree 引数省略時のインタラクティブ CLI 導線
status: in_progress
owner: codex
created_at: 2026-05-07T11:30:00+09:00
updated_at: 2026-05-07T12:25:00+09:00
parent_plan: .codex/plans/active/20260426-sango-cli-tui-agent-roadmap.md
workflow:
  origin_mode: resume_existing_workspace
  execution_intent: autonomous_until_git_boundary
  safety_domain: git_finish
target_root: /home/glaucus03/dev/projects/sango
allowed_paths:
  - cmd/**
  - README.md
  - docs/worktree.md
  - .codex/plans/active/**
  - .codex/artifacts/learnings/**
forbidden_paths:
  - .codex/config.toml
  - .codex/hooks/**
  - .agents/**
  - secrets/**
subagents: []
outputs:
  - 引数なしで対話入力できる worktree CLI 実装
  - `sango up` の対象 repo / service / `default-ports` を選べる対話導線
  - 最小限のドキュメント更新
  - 関連テスト
  - hook blocker を自己回復する `.codex` 更新
  - commit / push / PR

test_plan:
  - go test ./cmd/... ./internal/worktree/...
  - go test ./...
  - 必要なら対象コマンドの help 表示を目視確認
approval_gates:
  - `2026-05-07` のユーザー指示で `commit` / `push` / `PR create` は実行許可済み
  - destructive な cleanup はしない
rollback:
  - 追加した対話ヘルパーと各コマンド変更を戻せば元に戻る
  - `.codex` の回復差分だけを戻せば hook blocker 前の状態へ戻せる
status_notes:
  - `cmd/worktree_create.go` にはサービス選択だけ既存の対話処理がある
  - `create` / `switch` / `remove` は必須引数で入口が閉じている
  - `rebase` / `sync` / `verify` は active fallback のみで一覧選択がない
  - 前セッションの hook blocker により plan 修復と learning 記録が未完了
  - active plan 修復、learning 記録、feedback loop close、`go test ./...`、commit までは完了
---

# Original Request

`sango worktree create` などで、引数なしの場合に CLI 上で選択や入力を行って作成・操作できるようにしたい。repo 選択や default 利用有無も対話で決めたい。

# Context

- 前セッションは active plan の必須見出し不足により PreToolUse hook が false positive を起こし、Git mutating 操作だけでなく plan 修復と `.codex/artifacts/learnings/**` 更新まで停止した
- 今回は既存の CLI 実装変更を保持したまま、まず active plan を自己回復し、その後 learning 記録更新、検証、commit、push、PR 作成まで進める
- 既存の source 変更は `cmd/**`, `README.md`, `docs/worktree.md` にあり、`.codex/**` は回復と feedback loop 閉鎖のために最小限だけ更新する

# Scope

- `worktree create` で branch 未指定時に名前を入力できる
- 既存 worktree を対象にするコマンドで、対象未指定時に候補から選択できる
- `sango up` で対象 repo / service / default ports 利用有無を対話選択できる
- 既存の active fallback は壊さず、非対話環境では明確なエラーを返す
- hook blocker の自己回復に必要な plan / learning / feedback loop 更新
- 変更確認後の commit / push / PR 作成

# 完了基準

- active plan に必須見出し `## Context`, `## 完了基準`, `## Phase Checklist`, `## Agent Assignment`, `## Review Loop` が揃っている
- blocker 改善事項が `.codex/artifacts/learnings/**` に記録され、`fb-20260506-002` が close されている
- 既存実装変更と `.codex` 更新を含めて検証が完了している
- commit を作成し、現在の branch を push し、PR URL を取得して報告できる

# Phases

## Phase 1: 現状整理

- [x] 対象コマンドの引数仕様と既存対話処理を確認
- [x] 親 plan との関係を整理
- [x] hook blocker の原因を特定

## Phase 2: 実装

- [x] 共通の対話ヘルパーを追加
- [x] `worktree create` の branch 入力を対話対応
- [x] `switch` `remove` `rebase` `sync` `verify` の対象選択を対話対応
- [x] `sango up` の repo / service / `default-ports` 選択を対話対応
- [x] help / docs 表記を引数任意に合わせて更新
- [x] active plan と learning 記録を自己回復

## Phase 3: 検証と Git finish

- [x] `.codex` 更新を確認
- [x] `go test ./...` を実行
- [ ] push / PR を完了

# Checkpoints

- [x] 非対話環境では曖昧な待機に入らない
- [x] アクティブ worktree がある場合の既存 UX と矛盾しない
- [x] 表記上も「引数省略可」が分かる
- [ ] hook blocker を再発させない記録が残っている

# Validation

- `gofmt -w cmd/interactive.go cmd/interactive_test.go cmd/worktree_create.go cmd/worktree_switch.go cmd/worktree_remove.go cmd/worktree_rebase.go cmd/worktree_sync.go cmd/worktree_verify.go cmd/include_refresh.go cmd/up.go`
- `go test ./cmd ./internal/worktree ./internal/service`
- `go test ./...`

# Residual Notes

- `sango up` の対話導線は今回 `repo` / `service` / `profile` / `default target` の選択まで。`--worktree` 自体の対話選択は未追加。
- `down` / `restart` には同系統の対話導線をまだ広げていない。

# Phase Checklist

- [x] active plan の必須見出し不足を補完する
- [x] hook blocker の learning を `.codex/artifacts/learnings/**` に記録する
- [x] `fb-20260506-002` を close する `.codex` 更新を行う
- [x] 変更差分を確認し、必要な検証を再実行する
- [x] commit を作成する
- [ ] branch を push する
- [ ] PR を作成する

# Agent Assignment

| Phase | Owner | Support | Output |
| --- | --- | --- | --- |
| hook blocker 自己回復 | Codex | - | fixed active plan と learning 更新 |
| 実装差分の検証 | Codex | - | test results と residual risks |
| Git finish | Codex | - | commit, push, PR |

# Review Loop

1. active plan を source of truth として先に修復する
2. `.codex` 更新で feedback loop と blocker learning を閉じる
3. source 差分と `.codex` 差分をまとめて再検証する
4. Git mutating 操作は commit, push, PR の順に進め、各段階で失敗時はその場で原因を記録する

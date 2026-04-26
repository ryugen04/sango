---
id: 20260426-sango-bootstrap-and-migration
title: sango Codex 初期セットアップとコード移植
status: approved
owner: codex
---

# sango Codex 初期セットアップとコード移植

## Context

- 正規リポジトリは `~/dev/projects/sango`
- 現在 `~/dev/projects/sango` には旧試行のコードが残っている
- `~/dev/projects/sango-tree` の実装を移植し、Codex 用の基本運用ファイルを追加する

## 完了基準

- [x] `AGENTS.md` と `.codex/` 基本構成が追加されている
- [x] `sango-tree` のコードが `~/dev/projects/sango` に移植されている
- [x] module path が `github.com/ryugen04/sango` である
- [x] 差分に直接関係するテストが成功する
- [x] ローカル commit が作成されている

## Phase Checklist

### Phase 1
- [x] Codex 初期セットアップを追加する
- [x] active plan を作成する

### Phase 2
- [x] `sango-tree` からコードを移植する
- [x] 旧試行の不要ファイルを置き換える

### Phase 3
- [x] テストと CLI 動作を確認する
- [x] ローカル commit を作成する

## Agent Assignment

| Phase | Owner | Support | Output |
| --- | --- | --- | --- |
| Setup | Codex | - | AGENTS, .codex |
| Migration | Codex | - | migrated code |
| Verify | Codex | - | test results, commit |

## Review Loop

1. 初期セットアップを作成する
2. コードを移植する
3. テストと確認を行う
4. ローカル commit を作成する

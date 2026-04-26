# Personal OSS Roadmap

この文書は、`sango` を「個人開発向けの実用的な CLI/TUI」として磨くための handoff 用正本です。  
日々の実装は active plan を source of truth とし、この文書は背景・方針・次の優先順を引き継ぐために使います。

## Product Definition

`sango` は、複数 repo / 複数 worktree を使う個人開発者が、開発環境を迷わず起動・切替・確認できるようにする CLI/TUI です。

### Priorities

1. 毎日使う操作を短くする
2. 挙動と出力契約を安定させる
3. CLI を正本にし、TUI は補助に留める
4. 機能追加より概念削減を優先する

### Non-goals

- 大規模チーム向けの権限管理
- 常設 Web UI
- 重いプラグイン機構
- daemon / server mode 前提の設計
- 早期の公開ライブラリ化

## What Was Completed In Slice 1

- README 冒頭を個人 OSS 向けの価値提案に更新した
- 主コマンドを `日常利用` / `診断` / `補助` に再分類した
- `version` コマンドを追加した
- `status` `doctor` `runbook` `troubleshoot` `worktree status` に `--json` を追加した
- `sango init` テンプレートを現行の思想に寄せて更新した
- AI / script 利用向けの出力契約を README / docs に明記した

## Source Of Truth

- 実装計画: `.codex/plans/active/20260426-sango-cli-tui-agent-roadmap.md`
- 実行記録: `.codex/artifacts/execution/20260427-sango-personal-oss-slice1.md`

## Next Recommended Order

### 1. Surface Consistency

- `status` `logs` `doctor` の用語と表示粒度を揃える
- help 文と README のコマンド表記を棚卸しする
- `runbook` / `troubleshoot` の JSON / text の形を必要最小限まで寄せる

### 2. First 10 Minutes Experience

- README に最短導入例を追加する
- `init` テンプレートと設定例を README / docs と同期する
- 「最小の 2 サービス構成」で試せる例を用意する

### 3. TUI Role Cleanup

- dashboard の責務を「一覧・軽操作・確認」に固定する
- CLI に寄せるべき操作を明確化する
- 最低限の回帰検知方針を決める

## Decision Rules For Future Work

- 新機能は「毎日使う操作が短くなるか」で評価する
- JSON 出力を増やす場合は、既存の stdout/stderr/exit code 契約を壊さない
- 重い抽象化や拡張機構は、実運用で必要性が証明されるまで入れない

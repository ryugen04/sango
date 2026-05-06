---
id: 20260426-sango-cli-tui-agent-roadmap
title: sango 個人OSS再定義ロードマップ
status: in_progress
owner: codex
created_at: 2026-04-26T18:55:00+09:00
updated_at: 2026-04-27T09:20:00+09:00
related_issue:
active_branch: main
task_size: medium
split_strategy: phased roadmap
artifacts:
  plan: .codex/plans/active/20260426-sango-cli-tui-agent-roadmap.md
  execution: .codex/artifacts/execution/20260427-sango-personal-oss-slice1.md
  child_plans:
    - .codex/plans/active/202605071130-worktree-interactive-cli.md
  learnings:
---

# sango 個人OSS再定義ロードマップ

## Context

- このリポジトリは別PCから移行された途中段階で、現状をいったん白紙に近い視点で見直したい
- 方向性は「個人OSSとして実用的で、触ってみたくなるほど魅力的な CLI/TUI ツール」を優先する
- 小機能を増やすより、ユースケースの芯を定めてシンプルに磨き込む
- Codex / Claude からの利用しやすさは重要だが、機能過多より一貫した使い心地を優先する

## 現状評価

### 強み

- `up/down/status/logs/doctor/worktree/audit/runbook/troubleshoot` まで機能面の骨格がある
- `internal/*` にロジックが分離され、ユニットテストが広く整備されている
- `audit inventory` は JSON 出力があり、AI/自動化連携の入口になりうる
- `dashboard` によりTUIの基礎も存在する
- CI / release workflow が最低限整っている

### ボトルネック

- 外部公開用の Go API がなく、再利用しやすい package が無い
- `status` `doctor` `runbook` `worktree status` など主要コマンドが人間向け表示中心
- TUI は存在するがテストがなく、機能も監視・操作の最小セットに留まる
- ドキュメントと実装にずれがあるため、導入時の信頼性を損ねやすい
- exit code / stderr / JSON schema の契約がコマンド横断で揃っていない
- 何を最も解決するプロダクトなのかが、現状の機能群からは少し散って見える

## Product Goal

### 目標像

1. 複数リポジトリと複数 worktree を使う個人開発者が、開発環境を迷わず起動・切替・確認できる
2. CLI だけで日常利用が完結し、TUI は「状況把握と軽操作を気持ちよくする補助」に留める
3. AI エージェントからは、少数の安定した JSON コマンドで状態取得と操作ができる
4. 作者ひとりでも保守できる程度に概念数が少なく、挙動が予測しやすい

### 非目標

- 小さな便利機能を無制限に追加しない
- フルGUI、過度なプラグイン機構、重い常駐アーキテクチャを先に導入しない
- 汎用PaaSや大規模チーム向けプラットフォームを目指さない

## Design Principles

1. 一番よく使う操作を最短にする
2. 設定より挙動を単純にする
3. 出力契約を安定させ、AI と人間の両方に優しくする
4. 1機能追加より1概念削減を優先して評価する
5. TUI は補助線、CLI を正本にする

## 完了基準

- [ ] `sango` のコアユースケースが 1 文で説明できる
- [ ] 毎日使う主操作が少数のコマンドに整理されている
- [ ] 主要コマンドに `--json` と安定した終了コード方針がある
- [ ] README / docs / `sango init` テンプレートが同じ思想と設定モデルを示す
- [ ] TUI の役割が明確化され、CLI と競合しない
- [ ] 作者ひとりで維持しにくい機能候補が棚上げまたは削除されている

## Phase Checklist

### Phase 0: Analysis and Planning

- [x] リポジトリ構成・主要コマンド・ドキュメント・既存 plan を確認する
- [x] CLI / TUI / Agent 利用観点で現状の強み・不足を整理する
- [x] 分析結果を active plan として記録する
- [x] `make verify` を実行して現状の検証状態を確認する
- [x] 個人OSSとしての再整理が必要だと判断した

### Phase 1: Reframe the Product

- [x] `sango` の対象ユーザーとコアユースケースを 1 つに絞る
- [x] 「やること」と「やらないこと」を README 冒頭で伝えられる形にする
- [x] コマンド群を「日常利用」「診断」「補助」の3種に再分類する
- [x] 個人OSSとして重すぎる候補機能を棚上げリストに移す

### Phase 2: Stabilize the Contract

- [x] `status` `doctor` `runbook` `troubleshoot` `worktree status` に JSON 出力を追加する
- [x] 終了コードと stderr 出力ポリシーを整理する
- [x] JSON schema とサンプル出力を README / docs に記載する

### Phase 3: Simplify the Surface

- [x] 主コマンドの導線を絞り、入口として推すコマンドを少数にする
- [ ] `status` / `logs` / `doctor` の出力整合性を揃える
- [x] `version` コマンドを追加する
- [ ] コマンド名・help文・README を同じ用語へ揃える

### Phase 4: Fix Documentation Drift

- [ ] README と docs のコマンド説明を実装に合わせて棚卸しする
- [x] `docs/configuration.md` の `worktree.include` 記述を実装へ合わせる
- [x] `sango init` のテンプレートを現行機能に見合う内容へ更新する
- [x] AI 向け利用ガイドを追加する
- [ ] 「最初の10分」で価値が伝わる導入例を整備する

### Phase 5: Clarify CLI and TUI Roles

- [ ] dashboard の役割を「一覧・軽操作・確認」に限定して整理する
- [ ] dashboard の操作導線を見直し、最低限の主要操作を明確化する
- [ ] TUIの回帰検知方針を追加する

### Phase 6: Optional Future Work

- [ ] 公開 package 化は、明確な再利用需要が見えた後に再評価する
- [ ] 常駐 daemon / server mode は必要性が実証されるまで導入しない
- [ ] runbook / troubleshoot の拡張は、実利用で効く最小限だけ残す

### Phase 7: Handoff and Delivery

- [x] 全体計画を他者へ引き継げる handoff doc を `.codex/docs/` に追加する
- [x] execution artifact に今回の実装・検証・残タスクを記録する
- [ ] 現在の変更を単一コミットにまとめる
- [ ] PR を作成し、全体計画の引き継ぎ先を本文に明記する

## Future Feature Policy

### 積極的に検討してよいもの

- 毎日使う操作を短くする改善
- 状態確認を速く・確実にする改善
- AI から安全に呼びやすくする JSON 出力整備
- 設定量を減らし、初回成功率を上げる改善

### 慎重に扱うもの

- プラグイン機構
- 常駐プロセス前提の設計
- ネットワーク越し制御や multi-user 機能
- 機能追加のための機能追加になりやすい抽象化

## Potential Future Features

1. `sango ps` のような、より短い日常監視コマンド
2. `sango open` のような、主要サービスURLやログへの最短導線
3. テンプレート付き `init` の改善と、典型パターンのスターター提供
4. worktree 切替時の安全ガード強化
5. AI向けの `status --json` `doctor --json` を中心とした安定API化

## Explicitly Deprioritized

1. 大規模チーム向け権限管理
2. Web UI の常設
3. 複雑なプラグイン marketplace
4. 早期の公開ライブラリ化
5. 監視基盤やクラウド制御の内製化

## Priority Order

1. プロダクトの芯を絞る
2. 日常利用コマンドを磨く
3. JSON-first な CLI 契約の整備
4. ドキュメントと導入体験の同期
5. TUI の役割整理と最小改善

## Immediate Next Slice

1. README 冒頭で定義するべきコア価値を 1 文に固める
2. 日常利用で本当に推すコマンドセットを選別する
3. その上で `status` と `doctor` から `--json` を実装する
4. README / docs / init template を同じ思想へ揃える

## Active Child Plans

- `.codex/plans/active/202605071130-worktree-interactive-cli.md`: `worktree create` などの引数省略時インタラクティブ導線追加

## Agent Assignment

| Phase | Owner | Support | Output |
| --- | --- | --- | --- |
| Reframe the Product | Codex | - | product definition |
| Stabilize the Contract | Codex | - | JSON CLI design |
| Simplify the Surface | Codex | - | CLI polish |
| Fix Documentation Drift | Codex | - | synced docs |
| Clarify CLI and TUI Roles | Codex | - | UX changes |

## Investigation Notes

- `go test ./...` は現時点で成功
- `dashboard` は存在するが `internal/tui` にテストが無い
- `audit inventory` はすでに JSON 化されており、今後の標準I/O設計の基準に使える
- `docs/configuration.md` の `worktree.include.common` は実装の `root` と不一致
- 公開 package が無いため、現状は「ライブラリ」というより「CLIアプリ」に近い
- 魅力的な個人OSSにするには、機能拡張より「使う理由がすぐ分かること」が先
- `status` `doctor` `runbook` `troubleshoot` `worktree status` は JSON 出力に揃えた
- 診断系コマンドの fail/warn は、終了コードではなく出力内容で判定する方針に寄せた

## Review Loop

1. 俯瞰分析を更新する
2. active plan を source of truth として維持する
3. 最小の実装単位ごとに検証する
4. CLI/TUI/Agent の3視点で回帰を確認する

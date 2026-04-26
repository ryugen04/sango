# Asset Portability and Troubleshoot Model

この文書は、旧運用資産を他プロジェクトへ安全に再利用するための分類基準と、troubleshoot の管理モデルを定義します。

## 1. 資産分類ポリシー

### `template`（再利用テンプレート化）

- ルール文書、チェックリスト、plan 雛形
- 汎用プロンプト、レビュー観点、handoff 形式
- CI 設定のうち環境依存値を持たない部分

### `library`（実装として共通化）

- 設定読み込み・バリデーション
- インベントリ収集、分類、レポート出力
- include 展開、整合性検証、ガード処理

### `non-portable`（リポジトリ外に隔離）

- 個人名、組織名、ホスト名、絶対パス
- トークン、シークレット、内部 URL
- 特定環境でのみ意味を持つ運用メモ

## 2. 旧 `.claude` / YAML 資産の扱い

| 資産種別 | 推奨アクション | 置き場 |
| --- | --- | --- |
| アシスタント向け運用ルール | `template` 化 | `.codex/rules/` |
| 計画・実行・振り返りの雛形 | `template` 化 | `.codex/plans/templates/` |
| 実行フロー定義（YAML） | 汎用部を `template` 化 | `.github/workflows/` |
| 設定スキーマと検証 | `library` 化 | Go パッケージ |
| 障害検知チェック | `library` 化 + 設定駆動化 | Go パッケージ + `sango.yaml` |
| 固有環境向けメモ | `non-portable` 扱い | リポジトリ外 |

## 3. ライブラリ化の判断基準

- 複数プロジェクトで同一入出力を要求する
- テストで仕様を固定できる
- 依存先が限定され、境界が明確
- 固有環境情報を引数で注入できる

## 4. Troubleshoot 管理モデル

### レイヤ分離

- `detection`: 実行可能なチェック（自動判定）
- `diagnosis`: 症状と原因候補の整理（人が判断）
- `resolution`: 手順化された復旧と再発防止（runbook）

### データ管理

- カードは YAML で定義し、履歴を Git で管理する
- 推奨配置: `.codex/artifacts/troubleshooting/cards/`
- テンプレート: `.codex/docs/templates/troubleshoot-card.yaml`

### 運用ライフサイクル

1. `detect`: 症状を検知し、カードを新規作成
2. `triage`: 影響範囲と優先度を確定
3. `mitigate`: 暫定復旧手順を実施
4. `fix`: 恒久対策を実装
5. `verify`: 再発検知の自動チェックを追加
6. `close`: `last_verified` を更新してクローズ

## 5. ガバナンス

- `audit inventory` を定期実行し、運用資産の増減を可視化する
- PR 時に「project 固有情報を含まない」レビュー項目を必須化する
- troubleshoot カードは定期的に `last_verified` を更新し、陳腐化を防ぐ

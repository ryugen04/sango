# sango

複数 repo / 複数 worktree の開発環境を、迷わず起動・切替・確認するための個人開発向け CLI/TUI。

**sango** = サンゴ（珊瑚）。海中に根を張るサンゴの木（coral tree）のように、複数のサービスを一つの幹から分岐させて育てる。

## 何をするツールか

- `sango.yaml` ひとつで、複数サービスの起動・停止・状態確認をまとめる
- Git worktree を前提に、ブランチごとの開発環境を切り替えやすくする
- CLI を正本にして、TUI は状況把握と軽操作に絞る

## やらないこと

- 大規模チーム向けの権限管理や常設Web UI
- 重いプラグイン機構や daemon/server 前提の設計
- 小さな便利機能を無制限に増やすこと

## 特徴

- **日常運用を短くする**: `init` `up` `status` `logs` `down` を中心に毎日使う操作を短くする
- **Worktree前提で扱える**: ブランチごとの独立環境とポートオフセットをまとめて扱える
- **状態確認に強い**: 構造化ログ、doctor、troubleshoot、runbook を同じCLIに揃える
- **AI からも扱いやすい**: `logs --json` に加えて `status --json` `doctor --json` を提供する

## インストール

```bash
go install github.com/ryugen04/sango@latest
```

## クイックスタート

```bash
# 1. 設定ファイルを生成
sango init

# 2. sango.yaml を編集してサービスを定義
# 3. サービスを起動
sango up

# 4. 状態を確認
sango status

# 4-1. AI / スクリプト向けに JSON で確認
sango status --json

# 5. 停止
sango down
```

## 主コマンド

### 日常利用

| コマンド | 説明 |
|---------|------|
| `sango init` | `sango.yaml` テンプレートを生成 |
| `sango up [services...] [--profile name]` | サービスを起動。未指定時は対話で対象や `default-ports` を選択可能 |
| `sango down [--all]` | サービスを停止 |
| `sango restart [services...] [--profile name]` | サービスを再起動 |
| `sango status [--json]` | 現在の worktree を基準に状態を確認 |
| `sango logs [services...] [-f] [--json]` | ログを表示・フォロー |
| `sango version` | バージョンを表示 |

### 診断

| コマンド | 説明 |
|---------|------|
| `sango doctor [--fix] [--json]` | 開発環境の状態を確認 |
| `sango troubleshoot [service] [--fix] [--json]` | サービス単位の診断を実行 |
| `sango runbook search <keyword> [--json]` | 既知トラブルの手順を検索 |
| `sango audit inventory [--root dir] [--format text|json|both] [--include-runtime]` | 運用資産を棚卸し |
| `sango worktree verify [branch]` | include 状態を検証 |

### 補助

| コマンド | 説明 |
|---------|------|
| `sango clone [--shallow]` | リポジトリをbare clone＆初期worktree作成 |
| `sango worktree create [branch]` | ワークツリーを作成。未指定時は対話で新規名または既存リモートブランチを選択 |
| `sango worktree list` | ワークツリー一覧 |
| `sango worktree status [--json]` | 全ワークツリーの状態を表示 |
| `sango worktree switch [branch]` | アクティブワークツリーを切替。未指定時は選択 |
| `sango worktree remove [branch]` | ワークツリーを削除。未指定時は選択 |
| `sango runbook list [--service name]` | Runbook一覧を表示 |
| `sango dashboard` | TUI で一覧・軽操作を行う |

## Audit

`sango audit inventory` は、運用資産を `config` / `workflow` / `runtime` / `other` に分類して report-only で出力する。

```bash
# text + JSON（既定）
sango audit inventory

# runtime を含めて JSON のみ出力
sango audit inventory --format json --include-runtime
```

- 既定対象: `.claude`, `.codex`, `.sango`, `.hwt`
- 既定除外: `.sango/logs`, `.sango/pids`, `.sango/bare`, `.sango/work`, `.sango/template-cache`, `.sango/locks`
- JSON 出力先既定: `.sango/audit/inventory.json`

## Development

```bash
make verify
go test ./...
go run . audit inventory --root . --format text
go run . status --json --config testdata/valid.yaml
go run . doctor --json --config testdata/valid.yaml
```
- CI は `.github/workflows/ci.yml`、release は `.github/workflows/release.yml` を参照する
- release は `v*` tag push を入口にする

### Linux sandbox warning

Codex 実行時に `bubblewrap` / `user namespace` 警告が出る場合は `sango doctor` で診断できる。

```bash
go run . doctor --config testdata/valid.yaml
```

`Linux sandbox:` で始まる項目が fail/warn の場合、表示される Fix に従って `sysctl` 設定を調整する。

## AI / スクリプト利用

- `sango status --json`: 現在の worktree、サービス一覧、全worktree概要を取得
- `sango doctor --json`: 診断結果と pass/fail/warn の集計を取得
- `sango troubleshoot --json`: サービス単位の診断結果を取得
- `sango runbook search <keyword> --json`: 既知トラブル候補を取得
- `sango worktree status --json`: 全 worktree の横断状態を取得
- `sango logs --json`: JSONL 形式のログをそのまま取得

終了コードは、コマンド実行そのものが失敗した場合のみ非0になる。診断結果の fail/warn は各 JSON / テキスト出力の内容で判定する。

## 設定ファイル

`sango.yaml` の基本構造:

```yaml
name: my-project
version: "1.0"

services:
  api:
    type: process
    port: 3000
    command: npm start
    working_dir: ./api
    depends_on: [postgres]
    healthcheck:
      url: http://localhost:3000/health
      interval: 5s
      retries: 3

  postgres:
    type: docker
    image: postgres:16
    port: 5432
    shared: true
    volumes:
      - pgdata:/var/lib/postgresql/data

ports:
  strategy: fixed
  base_offset: 100
  range: [3000, 9999]

profiles:
  backend:
    services: [api, postgres]

worktree:
  base_dir: worktrees
  default_branch: main
  auto_setup: true
  create:
    default_services: [api]
  include:
    root: []

doctor:
  checks:
    - name: Node.js
      command: node --version
      expect: "v"
      fix: "brew install node"
```

詳細は [docs/configuration.md](docs/configuration.md) を参照。

## ドキュメント

- [設定リファレンス](docs/configuration.md) - `sango.yaml` の全スキーマ
- [Worktreeガイド](docs/worktree.md) - Worktreeベースの並行開発ワークフロー

## ライセンス

MIT

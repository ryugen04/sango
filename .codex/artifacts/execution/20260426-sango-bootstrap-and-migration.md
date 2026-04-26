# Execution

- Codex 初期セットアップとして `AGENTS.md` と `.codex/` 構成を追加
- `~/dev/projects/sango-tree` から `~/dev/projects/sango` へ rsync でコードを移植
- `module github.com/ryugen04/sango` を確認
- `GOCACHE=$(pwd)/.cache/go-build go test ./cmd ./internal/audit` を実行して成功
- `GOCACHE=$(pwd)/.cache/go-build go run . audit inventory --root . --format text` を実行して warning 継続・exit 0 を確認

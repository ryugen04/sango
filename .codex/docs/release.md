# Release Notes

## Local Verification

1. `make verify`
2. `go test ./...`
3. `go run . audit inventory --root . --format text`

## Tag-based Release

- `v*` tag push で GitHub Actions の release workflow が動く
- 各 target 向けに `sango` バイナリを build し、GitHub Release に添付する

## Current Targets

- darwin-amd64
- darwin-arm64
- linux-amd64
- linux-arm64
- windows-amd64

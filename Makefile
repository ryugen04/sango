.PHONY: verify test build audit

verify:
	./scripts/verify.sh

test:
	GOCACHE=$(PWD)/.cache/go-build go test ./...

build:
	GOCACHE=$(PWD)/.cache/go-build go build ./...

audit:
	GOCACHE=$(PWD)/.cache/go-build go run . audit inventory --root . --format text

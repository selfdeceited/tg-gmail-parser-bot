.PHONY: build lint check test test-db

build:
	go build ./...

lint:
	golangci-lint run ./...

## test runs all unit tests with the race detector.
## Requires all module dependencies to be downloaded (go mod download).
test:
	go test -race ./...

## test-db runs only internal/db tests — works without the full module cache.
test-db:
	go test -race ./internal/db/...

## check runs lint then build — use this before committing.
check: lint build

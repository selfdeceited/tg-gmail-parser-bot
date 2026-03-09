.PHONY: build lint check

build:
	go build ./...

lint:
	golangci-lint run ./...

## check runs lint then build — use this before committing.
check: lint build

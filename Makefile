.PHONY: build test

build:
	go build ./cmd/nala

test:
	go test ./...

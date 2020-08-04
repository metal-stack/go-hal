.ONESHELL:
CGO_ENABLED := $(or ${CGO_ENABLED},0)
GO := go
GO111MODULE := on

.DEFAULT_GOAL := test

.PHONY: gofmt
gofmt:
	GO111MODULE=off $(GO) fmt ./...

.PHONY: test
test: gofmt
	CGO_ENABLED=1 $(GO) test ./... -coverprofile=coverage.out -covermode=atomic && go tool cover -func=coverage.out

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: cli
cli:
	$(GO) build -o bin/cli ./cli

.PHONY: golint
golint:
	golangci-lint run

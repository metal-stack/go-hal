.ONESHELL:
CGO_ENABLED := $(or ${CGO_ENABLED},0)
GO := go
GO111MODULE := on

.PHONY: test
test: fmt
	CGO_ENABLED=1 $(GO) test ./... -coverprofile=coverage.out -covermode=atomic && go tool cover -func=coverage.out

.PHONY: fmt
fmt:
	GO111MODULE=off $(GO) fmt ./...

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: cli
cli:
	$(GO) build -o bin/cli ./cli
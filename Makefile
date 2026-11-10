GO ?= go

.PHONY: all test cover fuzz lint vet audit tidy help

all: vet lint test

## test: Run tests with the race detector
test:
	$(GO) test -race -shuffle=on ./...

## cover: Run tests and open an HTML coverage report
cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## fuzz: Fuzz the error parser for 30s
fuzz:
	$(GO) test -run='^$$' -fuzz=FuzzParseError -fuzztime=30s .

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## audit: Scan for known vulnerabilities
audit:
	govulncheck ./...

## tidy: Tidy go.mod
tidy:
	$(GO) mod tidy

## help: Show this help
help:
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

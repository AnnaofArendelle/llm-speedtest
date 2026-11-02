SHELL := /bin/bash
MODULE := github.com/AnnaofArendelle/llm-speedtest
BIN := llm-speedtest
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)
GOFLAGS := -trimpath
export CGO_ENABLED := 0
# Go packages of this module (excludes stray Go files inside web/node_modules).
PKGS := $(shell go list ./... | grep -v /web/node_modules/)

.PHONY: all build web web-install test race cover lint vuln fmt tidy snapshot skill screenshots run serve clean help check fixtures e2e web-test

all: web build ## Build frontend and binary

build: ## Build the binary into ./bin
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BIN) ./cmd/$(BIN)

web-install: ## Install frontend dependencies
	cd web && npm ci

web: ## Build the frontend into web/dist
	cd web && npm run build
	@touch web/dist/.gitkeep

test: ## Run Go tests
	go test $(PKGS)

race: ## Run Go tests with the race detector
	go test -race -count=1 $(PKGS)

cover: ## Run tests with coverage and enforce the threshold
	./scripts/coverage.sh

lint: ## Run golangci-lint and go vet
	go vet $(PKGS)
	golangci-lint run ./...

vuln: ## Run govulncheck
	govulncheck $(PKGS)

fmt: ## Format Go and web sources
	gofmt -s -w $$(git ls-files '*.go')
	cd web && npm run format

tidy: ## Tidy go.mod
	go mod tidy

snapshot: ## Build all release artifacts locally without publishing
	./scripts/package-skill.sh
	goreleaser release --snapshot --clean --skip=publish

skill: ## Package the AI skill into build/llm-speedtest-skill.zip
	./scripts/package-skill.sh

screenshots: ## Regenerate README screenshots from demo data
	cd web && npm run screenshots

run: build ## Run the server with demo data
	./bin/$(BIN) serve --demo

serve: build ## Run the server
	./bin/$(BIN) serve

clean: ## Remove build output
	rm -rf bin dist build coverage.out coverage.html

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

check: ## Run every CI check locally
	./scripts/ci-check.sh

web-test: ## Run the web unit tests with coverage
	cd web && npm run test:coverage

fixtures: ## Regenerate web/src/test/fixtures from the Go API (after changing response shapes)
	go test ./internal/server -run TestAPIContract -update-fixtures

e2e: build ## Run the Playwright end-to-end suite against a local demo server
	./scripts/e2e.sh

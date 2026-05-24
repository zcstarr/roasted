.PHONY: help build install test generate lint check-schemas check tidy
.DEFAULT_GOAL := help

GOLANGCI_LINT_VERSION ?= v2.6.1

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

build: ## Compile all packages
	go build ./...

install: ## Install the roasted CLI into $(shell go env GOBIN)
	go install ./cmd/roasted

test: ## Run tests with race detector and coverage
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

tidy: ## Tidy go.mod/go.sum
	go mod tidy

generate: ## Regenerate JSON schemas from Go types
	go generate ./pkg

lint: ## Run golangci-lint (install: https://golangci-lint.run/welcome/install/)
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)"; \
		exit 1; \
	}
	golangci-lint run ./...

check-schemas: generate ## Fail if generated schemas are out of date
	@git diff --exit-code schemas/ || { \
		echo "schemas/ is out of date. Run 'make generate' and commit the result."; \
		exit 1; \
	}

check: lint test check-schemas ## Run all CI checks locally

.PHONY: all check lint test test-interop test-race tidy vet build help

GO ?= go

all: check ## Run the full check suite

check: tidy vet lint test-race ## tidy + vet + lint + race tests (CI gate)

build: ## Compile all packages
	$(GO) build ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golangci-lint (if installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed; skipping (https://golangci-lint.run)"; \
	fi

test: ## Run tests
	$(GO) test ./...

test-interop: ## Install the TS SDK fixture and run Go<->JS interop tests
	npm --prefix testdata/interop-js install
	$(GO) test -count=1 -tags interop ./...

test-race: ## Run tests with the race detector
	$(GO) test -race ./...

tidy: ## Tidy go.mod/go.sum
	$(GO) mod tidy

# Bump the BSR-generated SDK (pulls protocolbuffers/go transitively).
update-sdk: ## Update the buf.build generated connect SDK to latest
	$(GO) get buf.build/gen/go/photon-hq/imessage/connectrpc/go@latest
	$(GO) mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

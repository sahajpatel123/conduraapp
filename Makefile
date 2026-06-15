# Synaptic Makefile
# ------------------------------------------------------------------
# This Makefile drives the developer workflow for the Synaptic daemon.
# Run `make help` to see all targets.

# ------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------

BINARY_NAME := synapticd
CLI_NAME := synaptic
TUI_NAME := synaptic-tui
VERSION          := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
COMMIT           := $(shell git rev-parse HEAD 2>/dev/null || echo "none")
BUILD_DATE       := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS          := -s -w \
                    -X 'github.com/sahajpatel123/synapticapp/internal/version.Version=$(VERSION)' \
                    -X 'github.com/sahajpatel123/synapticapp/internal/version.Commit=$(COMMIT)' \
                    -X 'github.com/sahajpatel123/synapticapp/internal/version.BuildDate=$(BUILD_DATE)'

PKG              := ./...
COVERAGE_FILE    := coverage.out
COVERAGE_HTML    := coverage.html

# Tools
GO               := go
GOLANGCI_LINT    := golangci-lint
GOIMPORTS        := goimports
GOFUMPT          := gofumpt

# ------------------------------------------------------------------
# Default
# ------------------------------------------------------------------

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Synaptic — Makefile targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------
# Build
# ------------------------------------------------------------------

.PHONY: build
build: ## Build synapticd, synaptic, and synaptic-tui into ./bin
	@mkdir -p bin
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/synapticd
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(CLI_NAME)    ./cmd/synaptic
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(TUI_NAME)    ./cmd/synaptic-tui
	@echo "Built: bin/$(BINARY_NAME), bin/$(CLI_NAME), bin/$(TUI_NAME)"

.PHONY: build-daemon
build-daemon: ## Build only synapticd
	@mkdir -p bin
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/synapticd

.PHONY: build-cli
build-cli: ## Build only the CLI
	@mkdir -p bin
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(CLI_NAME) ./cmd/synaptic

.PHONY: build-tui
build-tui: ## Build only the TUI
	@mkdir -p bin
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(TUI_NAME) ./cmd/synaptic-tui

.PHONY: build-all
build-all: ## Cross-compile for all supported platforms
	@mkdir -p dist
	@for os in darwin linux windows; do \
	  for arch in amd64 arm64; do \
	    ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
	    echo "Building $$os/$$arch..."; \
	    GOOS=$$os GOARCH=$$arch $(GO) build -ldflags "$(LDFLAGS)" \
	      -o dist/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/synapticd; \
	    GOOS=$$os GOARCH=$$arch $(GO) build -ldflags "$(LDFLAGS)" \
	      -o dist/$(CLI_NAME)-$$os-$$arch$$ext ./cmd/synaptic; \
	    GOOS=$$os GOARCH=$$arch $(GO) build -ldflags "$(LDFLAGS)" \
	      -o dist/$(TUI_NAME)-$$os-$$arch$$ext ./cmd/synaptic-tui; \
	  done; \
	done
	@echo "All builds complete in dist/"

.PHONY: install
install: build ## Build and install to GOPATH/bin
	$(GO) install -ldflags "$(LDFLAGS)" ./cmd/synapticd
	$(GO) install -ldflags "$(LDFLAGS)" ./cmd/synaptic
	$(GO) install -ldflags "$(LDFLAGS)" ./cmd/synaptic-tui

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ dist/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

.PHONY: release-snapshot
release-snapshot: ## Local GoReleaser snapshot (no publish)
	goreleaser release --snapshot --clean

.PHONY: build-gui
build-gui: ## Build Wails desktop app for current OS/arch
	chmod +x scripts/build-gui.sh
	./scripts/build-gui.sh

.PHONY: gen-manifest
gen-manifest: ## Generate unsigned update manifest from dist/checksums.txt
	@test -f dist/checksums.txt || (echo "run release-snapshot first" && exit 1)
	go run ./cmd/gen-update-manifest generate --unsigned \
	  --version $(VERSION) \
	  --checksums dist/checksums.txt \
	  --base-url "https://github.com/sahajpatel123/synapticapp/releases/download/$(VERSION)" \
	  --out dist/update-manifest.json

# ------------------------------------------------------------------
# Test
# ------------------------------------------------------------------

.PHONY: test
test: ## Run all unit tests with race detection
	$(GO) test -race -count=1 -timeout=120s $(PKG)

.PHONY: test-short
test-short: ## Run only short tests (skip integration)
	$(GO) test -race -count=1 -short -timeout=60s $(PKG)

.PHONY: test-integration
test-integration: ## Run integration tests
	$(GO) test -race -count=1 -timeout=300s -tags=integration ./test/integration/...

.PHONY: coverage
coverage: ## Run tests with coverage report
	$(GO) test -race -count=1 -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(PKG)
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total

.PHONY: bench
bench: ## Run benchmarks
	$(GO) test -bench=. -benchmem -run=^$$ $(PKG)

# ------------------------------------------------------------------
# Lint & Format
# ------------------------------------------------------------------

.PHONY: lint
lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run --timeout=5m ./...

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	$(GOLANGCI_LINT) run --fix --timeout=5m ./...

.PHONY: fmt
fmt: ## Format code (gofmt + goimports + gofumpt)
	$(GO) fmt $(PKG)
	@which $(GOIMPORTS) >/dev/null 2>&1 && $(GOIMPORTS) -w -local "github.com/sahajpatel123/synapticapp" . || echo "goimports not installed; skipping"
	@which $(GOFUMPT) >/dev/null 2>&1 && $(GOFUMPT) -w . || echo "gofumpt not installed; skipping"

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKG)

# ------------------------------------------------------------------
# Verify (CI-equivalent)
# ------------------------------------------------------------------

.PHONY: verify
verify: vet fmt lint test ## Run all checks (vet, fmt, lint, test)

# ------------------------------------------------------------------
# Dev
# ------------------------------------------------------------------

.PHONY: run-daemon
run-daemon: build-daemon ## Run synapticd with default config
	./bin/$(BINARY_NAME)

.PHONY: run-cli
run-cli: build-cli ## Run the CLI (pass args via CLI_ARGS=...)
	./bin/$(CLI_NAME) $(CLI_ARGS)

.PHONY: deps
deps: ## Download and tidy dependencies
	$(GO) mod download
	$(GO) mod tidy

.PHONY: deps-upgrade
deps-upgrade: ## Upgrade all dependencies to their latest minor/patch versions
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: tools
tools: ## Install dev tools (goimports, gofumpt)
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install mvdan.cc/gofumpt@latest

# ------------------------------------------------------------------
# Daemon convenience
# ------------------------------------------------------------------

.PHONY: daemon-init
daemon-init: build-cli ## Initialize ~/.synaptic/ (first-run setup)
	./bin/$(CLI_NAME) init

.PHONY: daemon-start
daemon-start: build-daemon ## Start the daemon in the background
	./bin/$(BINARY_NAME) --daemon

.PHONY: daemon-stop
daemon-stop: ## Stop the running daemon
	./bin/$(CLI_NAME) stop

.PHONY: daemon-status
daemon-status: build-cli ## Show daemon status
	./bin/$(CLI_NAME) status

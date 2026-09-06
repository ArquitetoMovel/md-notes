# ==============================================================================
# Makefile for md-notes (mdn)
# Cross-platform build automation for macOS, Linux, and Windows
# ==============================================================================

BINARY_NAME := mdn
CMD_DIR     := ./cmd/mdn
BIN_DIR     := bin
DIST_DIR    := dist

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.BuildDate=$(BUILD_DATE)'

GO_BUILD := CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)"

# Target platforms
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: all help build build-all universal-darwin package test test-coverage fmt vet tidy clean

all: build

help: ## Show available make targets
	@echo "md-notes build management system"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for current host platform
	@echo "==> Building $(BINARY_NAME) for host..."
	@mkdir -p $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "==> Built: $(BIN_DIR)/$(BINARY_NAME)"

build-all: ## Cross-compile binaries for macOS, Linux, and Windows (amd64 and arm64)
	@echo "==> Cross-compiling for all platforms (version: $(VERSION))..."
	@rm -rf $(BIN_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		OUTPUT_DIR=$(BIN_DIR)/$${OS}_$${ARCH}; \
		EXT=""; \
		if [ "$$OS" = "windows" ]; then EXT=".exe"; fi; \
		echo "    -> Building $${OS}/$${ARCH}..."; \
		mkdir -p $$OUTPUT_DIR; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build -trimpath -ldflags="$(LDFLAGS)" -o $$OUTPUT_DIR/$(BINARY_NAME)$$EXT $(CMD_DIR) || exit 1; \
	done
	@echo "==> All binaries successfully built in $(BIN_DIR)/"

universal-darwin: ## Build macOS Universal Binary (combines amd64 and arm64 via lipo)
	@echo "==> Building macOS Universal Binary..."
	@mkdir -p $(BIN_DIR)/darwin_amd64 $(BIN_DIR)/darwin_arm64 $(BIN_DIR)/darwin_universal
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/darwin_amd64/$(BINARY_NAME) $(CMD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/darwin_arm64/$(BINARY_NAME) $(CMD_DIR)
	lipo -create -output $(BIN_DIR)/darwin_universal/$(BINARY_NAME) \
		$(BIN_DIR)/darwin_amd64/$(BINARY_NAME) \
		$(BIN_DIR)/darwin_arm64/$(BINARY_NAME)
	@echo "==> macOS Universal Binary ready at $(BIN_DIR)/darwin_universal/$(BINARY_NAME)"

package: build-all universal-darwin ## Package cross-platform binaries into tar.gz/zip and generate SHA256 checksums
	@echo "==> Packaging release archives..."
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		ARCHIVE_NAME="$(BINARY_NAME)-$(VERSION)-$${OS}-$${ARCH}"; \
		if [ "$$OS" = "windows" ]; then \
			echo "    -> Zipping $${ARCHIVE_NAME}.zip..."; \
			(cd $(BIN_DIR)/$${OS}_$${ARCH} && zip -q -9 ../../$(DIST_DIR)/$${ARCHIVE_NAME}.zip $(BINARY_NAME).exe); \
		else \
			echo "    -> Compressing $${ARCHIVE_NAME}.tar.gz..."; \
			tar -czf $(DIST_DIR)/$${ARCHIVE_NAME}.tar.gz -C $(BIN_DIR)/$${OS}_$${ARCH} $(BINARY_NAME); \
		fi; \
	done
	@if [ -f "$(BIN_DIR)/darwin_universal/$(BINARY_NAME)" ]; then \
		echo "    -> Compressing $(BINARY_NAME)-$(VERSION)-darwin-universal.tar.gz..."; \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-universal.tar.gz -C $(BIN_DIR)/darwin_universal $(BINARY_NAME); \
	fi
	@echo "==> Generating SHA256 checksums..."
	@(cd $(DIST_DIR) && \
		if command -v sha256sum >/dev/null 2>&1; then \
			sha256sum *.tar.gz *.zip > checksums.txt; \
		else \
			shasum -a 256 *.tar.gz *.zip > checksums.txt; \
		fi)
	@echo "==> Release artifacts ready in $(DIST_DIR)/:"
	@ls -lh $(DIST_DIR)

test: ## Run unit tests
	@echo "==> Running tests..."
	go test -v ./...

test-coverage: ## Run unit tests with coverage report
	@echo "==> Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage report generated at coverage.html"

fmt: ## Format Go source code
	@echo "==> Formatting code..."
	go fmt ./...

vet: ## Run go vet linter
	@echo "==> Running go vet..."
	go vet ./...

tidy: ## Tidy Go module dependencies
	@echo "==> Tidying dependencies..."
	go mod tidy

clean: ## Clean build artifacts and temporary files
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out coverage.html

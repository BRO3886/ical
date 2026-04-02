BINARY_NAME=ical
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build install test clean completions release formula help

all: build

build: ## Build the binary (includes EventKit via cgo)
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/ical/

install: build ## Install the binary to $GOPATH/bin
	go install $(LDFLAGS) ./cmd/ical/

test: ## Run tests
	go test ./... -v

release: ## Build release tarballs for GitHub upload (arm64 + amd64)
	@mkdir -p bin
	@for arch in arm64 amd64; do \
		echo "Building ical-darwin-$$arch..."; \
		CGO_ENABLED=1 GOARCH=$$arch go build $(LDFLAGS) -o bin/ical ./cmd/ical/; \
		chmod +x bin/ical; \
		tar -czf bin/ical-darwin-$$arch.tar.gz -C bin ical; \
		rm bin/ical; \
	done
	@echo ""
	@echo "SHA256 checksums (copy into Formula/ical.rb or run 'make formula'):"
	@shasum -a 256 bin/ical-darwin-arm64.tar.gz bin/ical-darwin-amd64.tar.gz
	@echo ""
	@echo "Upload bin/ical-darwin-{arm64,amd64}.tar.gz to GitHub Releases"

formula: ## Generate Homebrew formula (run after 'make release')
	@if [ ! -f bin/ical-darwin-arm64.tar.gz ] || [ ! -f bin/ical-darwin-amd64.tar.gz ]; then \
		echo "Run 'make release' first to build the tarballs"; exit 1; \
	fi
	@ARM64_SHA=$$(shasum -a 256 bin/ical-darwin-arm64.tar.gz | awk '{print $$1}'); \
	AMD64_SHA=$$(shasum -a 256 bin/ical-darwin-amd64.tar.gz | awk '{print $$1}'); \
	VER=$$(echo "$(VERSION)" | sed 's/^v//'); \
	echo 'class Ical < Formula'; \
	echo '  desc "CLI for macOS Calendar — fast native EventKit bindings"'; \
	echo '  homepage "https://github.com/BRO3886/ical"'; \
	echo "  version \"$$VER\""; \
	echo '  license "MIT"'; \
	echo ''; \
	echo '  on_arm do'; \
	echo "    url \"https://github.com/BRO3886/ical/releases/download/$(VERSION)/ical-darwin-arm64.tar.gz\""; \
	echo "    sha256 \"$$ARM64_SHA\""; \
	echo '  end'; \
	echo ''; \
	echo '  on_intel do'; \
	echo "    url \"https://github.com/BRO3886/ical/releases/download/$(VERSION)/ical-darwin-amd64.tar.gz\""; \
	echo "    sha256 \"$$AMD64_SHA\""; \
	echo '  end'; \
	echo ''; \
	echo '  def install'; \
	echo '    bin.install "ical"'; \
	echo '  end'; \
	echo ''; \
	echo '  test do'; \
	echo '    system "#{bin}/ical", "--version"'; \
	echo '  end'; \
	echo 'end'

clean: ## Remove built binaries
	rm -rf bin/

completions: build ## Generate shell completion scripts
	mkdir -p completions
	./bin/$(BINARY_NAME) completion bash > completions/ical.bash
	./bin/$(BINARY_NAME) completion zsh > completions/_ical
	./bin/$(BINARY_NAME) completion fish > completions/ical.fish

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

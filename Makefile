BINARY_NAME=ical
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build install test clean completions release help

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
	@echo "Upload bin/ical-darwin-{arm64,amd64}.tar.gz to GitHub Releases"

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

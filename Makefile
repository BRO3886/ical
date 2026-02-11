VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build clean test install completions

build:
	go build $(LDFLAGS) -o bin/cal ./cmd/cal

clean:
	rm -rf bin/

test:
	go test ./... -v

install:
	go install $(LDFLAGS) ./cmd/cal

completions: build
	mkdir -p completions
	./bin/cal completion bash > completions/cal.bash
	./bin/cal completion zsh > completions/_cal
	./bin/cal completion fish > completions/cal.fish

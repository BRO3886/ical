VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build clean test install completions

build:
	go build $(LDFLAGS) -o bin/ical ./cmd/ical

clean:
	rm -rf bin/

test:
	go test ./... -v

install:
	go install $(LDFLAGS) ./cmd/ical

completions: build
	mkdir -p completions
	./bin/ical completion bash > completions/ical.bash
	./bin/ical completion zsh > completions/_ical
	./bin/ical completion fish > completions/ical.fish

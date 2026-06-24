.PHONY: build test test-smoke test-integration test-hooks test-all clean run

BINARY := bin/slim
MAIN   := ./cmd/slim
VERSION ?= dev

build:
	@mkdir -p bin
	go build -ldflags "-X github.com/sheshisheri-hi/token-optimizer/internal/version.Version=$(VERSION)" -o $(BINARY) $(MAIN)

test:
	go test ./... -count=1

test-smoke:
	@bash scripts/test-smoke.sh

test-integration:
	@bash scripts/test-integration.sh

test-hooks:
	@bash scripts/test-hooks.sh

test-cursor:
	@bash scripts/test-cursor.sh

show-paths:
	@bash scripts/show-paths.sh

test-all:
	@bash scripts/test-all.sh

clean:
	rm -rf bin dist

run: build
	./$(BINARY)

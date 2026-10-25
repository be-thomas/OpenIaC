GO ?= go
BIN := bin/openiac
PLAYGROUND := playground

.PHONY: all build test vet fmt tidy wasm playground clean hooks conformance-inputs

all: build

build:
	$(GO) build -o $(BIN) ./cmd/openiac

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

# Install repository-local git hooks.
hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/commit-msg scripts/check-commit-message.sh

# Refresh the conformance input corpus from the configured public sources.
conformance-inputs:
	$(GO) run ./cmd/openiac-conformance collect -v

# Build the browser playground. Copies wasm_exec.js from the local toolchain,
# which lives in a different directory depending on the Go version.
wasm:
	@mkdir -p $(PLAYGROUND)
	GOOS=js GOARCH=wasm $(GO) build -o $(PLAYGROUND)/main.wasm ./wasm
	@if [ -f "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(PLAYGROUND)/wasm_exec.js; \
	else \
		cp "$$($(GO) env GOROOT)/misc/wasm/wasm_exec.js" $(PLAYGROUND)/wasm_exec.js; \
	fi

playground: wasm

clean:
	rm -rf bin out $(PLAYGROUND)/main.wasm $(PLAYGROUND)/wasm_exec.js

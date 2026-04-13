.PHONY: dev run build test

AIR ?= air
RUN_ARGS ?= --listen :2222 --log-level debug

dev:
	@command -v $(AIR) >/dev/null 2>&1 || { \
		echo "air is required for hot reload."; \
		echo "Install it with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	}
	$(AIR) -c air.toml

run:
	go run ./cmd/e4 serve $(RUN_ARGS)

build:
	go build ./...

test:
	go test ./...

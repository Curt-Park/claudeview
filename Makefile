BINARY_NAME := claudeview
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS     := -ldflags "-X main.Version=$(VERSION) -s -w"

.PHONY: build install test bdd lint clean demo cross-build

## build: build the binary for the current platform
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

## install: install to ~/.local/bin
install: build
	mkdir -p ~/.local/bin
	cp bin/$(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@echo "Installed to ~/.local/bin/$(BINARY_NAME)"

## test: run all tests
test:
	go test -race -count=1 ./...

## bdd: run BDD integration tests
bdd:
	go test -race -count=1 ./internal/ui/bdd/...

## fmt: run formatting
fmt:
	go fmt ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## demo: run with synthetic demo data
demo: build
	./bin/$(BINARY_NAME) --demo

## clean: remove build artifacts
clean:
	rm -rf bin/

## cross-build: build for all supported platforms
cross-build:
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64  .
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64  .

## help: show this help
help:
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

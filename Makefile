.DEFAULT_GOAL := help
.PHONY: all fmt lint staticcheck test coverage deps build build-all install clean help setup generate

BINARY_NAME := deepviz
INSTALL_PATH := $(HOME)/.local/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

## Execute main tasks collectively
all: setup fmt lint staticcheck test

## Install required tools with mise
setup:
	@if ! command -v mise > /dev/null 2>&1; then \
		echo "❌ mise is not installed"; \
		exit 1; \
	fi
	@mise install

## Format code
fmt: setup
	gofmt -s -w .

## Run static analysis (go vet)
lint: setup
	go vet ./...

## Run static analysis (staticcheck)
staticcheck: setup
	go run honnef.co/go/tools/cmd/staticcheck ./...

## Run tests
test: setup
	go test -json ./... | go tool tparse -all

## Generate coverage report
coverage: setup
	@mkdir -p build
	go test -coverprofile=build/coverage.out ./...
	go tool cover -html=build/coverage.out -o build/coverage.html
	@echo "Coverage report: build/coverage.html"

## Build binary (current platform)
build: setup
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/deepviz

## Multi-platform build (all platforms)
build-all: setup
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/deepviz
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/deepviz
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/deepviz
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/deepviz
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/deepviz
	@echo "✅ Multi-platform build completed in dist/"

## Install binary (~/.local/bin/)
install: build
	mkdir -p $(INSTALL_PATH)
	cp bin/$(BINARY_NAME) $(INSTALL_PATH)/

## Organize and verify dependencies
deps: setup
	go mod tidy
	go mod verify

## Generate code from OpenAPI schema
generate: setup
	@mkdir -p internal/genai/interactions
	@echo "Generating Interactions API client..."
	@go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen-interactions.yaml https://ai.google.dev/api/interactions.openapi.json
	@echo "✅ Code generation completed"

## Remove build artifacts
clean:
	rm -rf bin build dist internal/genai

## Display help
help:
	@make2help $(MAKEFILE_LIST)

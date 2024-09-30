# Makefile for Go Project

# Project variables
PROJECT_NAME := gollm
BINARY_NAME := $(PROJECT_NAME)
BUILD_DIR := build
GO_FILES := $(shell find . -name '*.go' -not -path "./vendor/*")

# Tools
GOFMT := gofmt
GOLINT := golint
GOTEST := go test
GOBUILD := go build
GOCLEAN := go clean

# Default target
all: format lint test build

# Format code using gofmt
.PHONY: format
format:
	@echo "==> Formatting code..."
	@$(GOFMT) -s -w $(GO_FILES)

# Lint code using golint
.PHONY: lint
lint:
	@echo "==> Linting code..."
	@if ! [ -x "$$(command -v $(GOLINT))" ]; then \
		echo "Installing golint..."; \
		go install golang.org/x/lint/golint@latest; \
	fi
	@$(GOLINT) ./...

# Run tests
.PHONY: test
test:
	@echo "==> Running tests..."
	@$(GOTEST) -v ./...

# Build the project
.PHONY: build
build:
	@echo "==> Building binary..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)

# Clean build files
.PHONY: clean
clean:
	@echo "==> Cleaning build files..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Install project dependencies (if necessary)
.PHONY: deps
deps:
	@echo "==> Installing dependencies..."
	@go mod tidy

# Run all targets (default behavior)
.PHONY: all
all: format lint test build

# Help message
.PHONY: help
help:
	@echo "Available make commands:"
	@echo "  make format      - Format all Go source files"
	@echo "  make lint        - Run linters on the source code"
	@echo "  make test        - Run unit tests"
	@echo "  make build       - Build the project binary"
	@echo "  make clean       - Remove build files and clean the project"
	@echo "  make deps        - Install project dependencies"
	@echo "  make all         - Format, lint, test, and build the project"

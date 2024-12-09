.PHONY: build test clean install release snapshot

# Build variables
BINARY_NAME=take-cli
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/take

# Run tests
test:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.txt

# Install the binary
install:
	$(GOINSTALL) $(LDFLAGS) ./cmd/take

# Create a snapshot release
snapshot:
	goreleaser release --snapshot --rm-dist

# Create a release
release:
	goreleaser release --rm-dist

# Install development dependencies
dev-deps:
	$(GOGET) -u golang.org/x/lint/golint
	$(GOGET) -u github.com/goreleaser/goreleaser

# Run linter
lint:
	golint ./...

# Format code
fmt:
	go fmt ./...

# Run all checks
check: fmt lint test

# Generate shell completions
completions:
	mkdir -p completions
	cp completions/take.{bash,zsh,fish} completions/

# Install shell integration
install-shell:
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/install.ps1
else
	bash scripts/install.sh
endif
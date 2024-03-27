# Makefile for building the Go CLI application

# The output binary name
BINARY_NAME=mendix-cli

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOPKG=$(GOBASE)/cmd/$(BINARY_NAME)

# Go commands
GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOGET=go get

# Build targets
all: clean deps test build-macos build-windows build-macos-arm64

# Build for macOS
build-macos:
	@echo "Building for macOS amd64..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-darwin-amd64 $(GOPKG)

build-macos-arm64:
	@echo "Building for macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-darwin-arm64 $(GOPKG)

# Build for Windows
build-windows:
	@echo "Building for Windows amd64..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-windows-amd64.exe $(GOPKG)

# Clean up binaries
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(GOBIN)/$(BINARY_NAME)*

# Run tests
test:
	@echo "Running tests"
	@$(GOTEST) -v ./...

# Fetch dependencies
deps:
	@echo "Fetching dependencies"
	@go mod tidy

.PHONY: all build-macos build-windows clean test deps

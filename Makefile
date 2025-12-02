# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=scanner
BINARY_UNIX=$(BINARY_NAME)_unix

# Build for current platform
build:
	$(GOBUILD) -tags 'netgo osusergo' -a -installsuffix cgo -o $(BINARY_NAME) -v ./cmd/scanner

# Build for Raspberry Pi (ARM)
build-pi:
	@echo "Building for Raspberry Pi (ARM64)..."
	@echo "Note: CGO is disabled to avoid cross-compilation issues"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOARM=7 $(GOBUILD) -tags 'purego nowasm noasm' -o $(BINARY_NAME)-arm64 ./cmd/scanner

# Install dependencies
deps:
	$(GOMOD) download

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-arm64

# Run the application
run:
	$(GOBUILD) -tags 'netgo osusergo' -a -installsuffix cgo -o $(BINARY_NAME) -v ./cmd/scanner
	./$(BINARY_NAME)

# Install the binary
install:
	$(GOBUILD) -tags 'netgo osusergo' -a -installsuffix cgo -o $(BINARY_NAME) -v ./cmd/scanner
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build for current platform"
	@echo "  build-pi     - Build for Raspberry Pi (ARM64) with CGO disabled"
	@echo "  deps         - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run the application"
	@echo "  install      - Install the binary to /usr/local/bin"
	@echo "  help         - Show this help message"

.PHONY: build build-pi deps test clean run install help
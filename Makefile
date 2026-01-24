.PHONY: build-gui build-cli install-cli test clean help

CLI_BINARY := bin/spotiflac-cli
CLI_SOURCE := ./cmd/cli

.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo ""
	@echo "  make build-gui        - Build the GUI application with Wails"
	@echo "  make build-cli        - Build the SpotiFLAC CLI"
	@echo "  make build-all        - Build both GUI and CLI"
	@echo "  make install-cli      - Install the CLI globally"
	@echo "  make test             - Run tests"
	@echo "  make clean            - Clean build files"
	@echo "  make dev-cli          - Quick build and test the CLI"
	@echo ""

build-gui:
	@echo "Building GUI application..."
	wails build

build-cli:
	@echo "Building CLI..."
	@mkdir -p bin
	go build -o $(CLI_BINARY) $(CLI_SOURCE)
	@echo "CLI built: $(CLI_BINARY)"

build-all: build-gui build-cli
	@echo "All binaries built"

install-cli:
	@echo "Installing CLI..."
	go install $(CLI_SOURCE)
	@echo "CLI installed globally (spotiflac-cli)"

test:
	@echo "Running tests..."
	go test ./... -v

clean:
	@echo "Cleaning..."
	rm -rf build/ bin/
	go clean
	@echo "Cleanup complete"

dev-cli: build-cli
	@echo ""
	@echo "Usage:"
	@echo "  ./$(CLI_BINARY) album <spotify-url>"
	@echo ""
	@echo "Example:"
	@echo "  ./$(CLI_BINARY) album https://open.spotify.com/album/..."
	@echo ""

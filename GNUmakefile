.PHONY: all build test install clean security

# Name of your final executable binary
BINARY_NAME=goas
# Path to the main entry point
MAIN_PATH=./cmd/goas

all: test build

build:
	@echo "Building the binary..."
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PATH)

test:
	@echo "Running unit tests..."
	go test -v -race ./...

coverage:
	@echo "Running tests with coverage analysis..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

install:
	@echo "Installing binary to your local Go bin path..."
	go install $(MAIN_PATH)

security:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out
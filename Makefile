# Makefile to build the scan and api code
# Go variables
GO = go
GOPATH = $(shell $(GO) env GOPATH)
GOFILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Project variables
BINARY_SCAN = scan
BINARY_API = api

# Build scan
.PHONY: build-scan
build-scan:
	$(GO) build -o $(BINARY_SCAN) ./cmd/scan

# Build api
.PHONY: build-api
build-api:
	$(GO) build -o $(BINARY_API) ./cmd/api

# Build all
.PHONY: build
build: build-scan build-api
	@echo "Build complete for both scan and api."

# Clean the build
.PHONY: clean
clean:
	rm -f $(BINARY_SCAN) $(BINARY_API)
	@echo "Clean complete."

# Run tests
.PHONY: test
test:
	$(GO) test ./...
	@echo "Tests run complete."
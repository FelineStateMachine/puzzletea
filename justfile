version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
ldflags := "-X main.version=" + version

# Build the binary (default recipe).
build:
    go build -ldflags '{{ldflags}}' -o puzzletea

# Build and run.
run: build
    ./puzzletea

# Run all tests.
test:
    go test ./...

# Run linter.
lint:
    golangci-lint run ./...

# Format all Go files with gofumpt.
fmt:
    gofumpt -w .

# Tidy module dependencies.
tidy:
    go mod tidy

# Install the binary to $GOPATH/bin.
install:
    go install -ldflags '{{ldflags}}'

# Remove build artifacts.
clean:
    rm -f puzzletea
    rm -rf dist

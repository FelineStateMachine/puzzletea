version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
ldflags := "-X github.com/FelineStateMachine/puzzletea/cmd.Version=" + version

# Build the binary (default recipe).
build:
    go build -ldflags '{{ldflags}}' -o puzzletea

# Build and run.
run: build
    ./puzzletea

# Run all tests.
test:
    go test ./...

# Run tests in short mode (skips slow generator tests).
test-short:
    go test -short ./...

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

# Generate VHS GIFs (requires vhs: https://github.com/charmbracelet/vhs).
vhs: build
    vhs vhs/menu.tape
    vhs vhs/nonogram.tape
    vhs vhs/nurikabe.tape
    vhs vhs/sudoku.tape
    vhs vhs/shikaku.tape
    vhs vhs/takuzu.tape
    vhs vhs/wordsearch.tape
    vhs vhs/hashiwokakero.tape
    vhs vhs/help.tape
    vhs vhs/hitori.tape
    vhs vhs/lightsout.tape
    vhs vhs/stats.tape

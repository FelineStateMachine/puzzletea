# AGENTS.md - PuzzleTea Development Guide

## Build, Test, and Quality Commands

### Build & Run
```bash
just              # build with version from git tags
just run          # build and run
just install      # install to $GOPATH/bin
just clean        # remove binary and dist/
```
Without `just`: `go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o puzzletea`

### Testing
```bash
just test                                    # go test ./...
just test-short                              # go test -short ./... (skips slow generator tests)
go test ./nonogram/                          # single package
go test ./sudoku/ -run TestGenerateGrid      # specific test
go test ./resolve/ -run TestCategory -v      # specific test, verbose
```

### Linting & Formatting
```bash
just lint         # golangci-lint run ./...
just fmt          # gofumpt -w .
just tidy         # go mod tidy
```

**Always run `just fmt` and `just lint` before committing.**

---

## Project Structure
```
puzzletea/
├── main.go, cmd.go, model.go   # Entry point, CLI (cobra), root TUI model
├── update.go, view.go          # Root model update loop, view rendering
├── spawn.go, debug.go          # Game spawn/save lifecycle, debug overlay
├── game/       # Plugin interfaces (Gamer, Mode, Spawner), cursor, keys, style
├── store/      # SQLite persistence (GameRecord, CRUD)
├── ui/         # Shared UI: menu list, table, styles, MenuItem
├── daily/      # Daily puzzle seeding, RNG, mode selection
├── resolve/    # CLI argument resolution (category/mode name matching)
├── namegen/    # Adjective-noun name generator
├── hashiwokakero/, hitori/, lightsout/, nonogram/
├── sudoku/, takuzu/, wordsearch/       # Puzzle game packages
├── plan/       # Design/planning documents
└── vhs/        # GIF preview tapes
```

Each puzzle package follows a consistent file structure:
- **Capitalized**: `Model.go`, `Gamemode.go`, `Export.go`
- **Lowercase**: `grid.go`, `keys.go`, `style.go`, `generator.go`, `<game>_test.go`

---

## Code Style Guidelines

### Formatting
- Use `gofumpt` (stricter than gofmt) -- run `just fmt`
- No comments required unless explaining non-obvious logic
- Keep lines under ~100 characters

### Imports
Two groups separated by a blank line: stdlib, then everything else (internal + external sorted together). When there are many internal imports, a third group separating internal from external is acceptable.
```go
import (
    "errors"
    "fmt"

    "github.com/FelineStateMachine/puzzletea/game"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)
```

### Naming Conventions
- **Types**: PascalCase (`Model`, `NonogramMode`, `Entry`)
- **Unexported types**: camelCase or lowercase (`grid`, `state`, `menuItem`)
- **Variables/Fields**: camelCase (`rowHints`, `currentHints`)
- **Constants**: camelCase (`mainMenuView`, `gameSelectView`)
- **Interfaces**: PascalCase (`Gamer`, `Spawner`, `Mode`)

### Type Declarations
Prefer grouped type aliases: `type ( grid [][]rune; state string )`

### Interface Compliance
Use compile-time checks: `var _ game.Gamer = Model{}`, `var _ game.Spawner = NonogramMode{}`

### Error Handling
- Return descriptive errors: `errors.New("puzzle width does not support row tomography definition")`
- Check errors immediately; use `fmt.Errorf` with `%w` only when wrapping adds context

### Styling
Use `lipgloss.AdaptiveColor` with ANSI 256 palette numbers. Avoid hex colors.

---

## Plugin Architecture

### Gamer Interface (game/gamer.go)
Every puzzle `Model` must implement:
```go
type Gamer interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Gamer, tea.Cmd)
    View() string
    GetDebugInfo() string
    GetFullHelp() [][]key.Binding
    GetSave() ([]byte, error)
    IsSolved() bool
    Reset() Gamer
    SetTitle(string) Gamer
}
```

### Mode/Spawner Interfaces
```go
type Spawner interface { Spawn() (Gamer, error) }
type SeededSpawner interface {
    Spawner
    SpawnSeeded(rng *rand.Rand) (Gamer, error)
}
```

### Puzzle Package Exports
Every puzzle package exports `Modes`, `DailyModes` (`[]list.Item`), `HelpContent` (`string`), `NewMode(...)`, `New(...)`, `ImportModel([]byte) (*Model, error)`, and `DefaultKeyMap`.

Each package registers itself via `init()`:
```go
func init() {
    game.Register("Nonogram", func(data []byte) (game.Gamer, error) {
        return ImportModel(data)
    })
}
```

---

## Testing Conventions

### Section Comments with Priority
```go
// --- generateTomography (P0) ---     // P0 = critical
// --- Grid serialization (P1) ---     // P1 = important
// --- generateRandomState (P2) ---    // P2 = generators/slow
```

### Table-Driven Tests with Subtests
```go
func TestGenerateTomography(t *testing.T) {
    tests := []struct {
        name     string
        grid     grid
        wantRows TomographyDefinition
    }{
        {name: "all filled row", grid: grid{...}, wantRows: ...},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic using t.Errorf / t.Fatalf only (no assertion libraries)
        })
    }
}
```

### Save/Load Round-Trip Pattern
```go
data, err := m.GetSave()
if err != nil { t.Fatal(err) }
loaded, err := ImportModel(data)
if err != nil { t.Fatal(err) }
// verify state preserved
```

### Slow Test Gating
```go
if testing.Short() {
	t.Skip("skipping slow generator test in short mode")
}
```

---

## Global Keybindings
`Ctrl+N` Main menu | `Ctrl+C` Quit (saves abandoned) | `Ctrl+E` Debug overlay | `Ctrl+H` Full help | `Ctrl+R` Reset puzzle | `Enter` Select | `Escape` Back

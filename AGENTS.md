# AGENTS.md - PuzzleTea Development Guide

## Build, Test, and Quality Commands

### Build & Run
```bash
just              # build with version from git tags
just run          # build and run
just install      # install to $GOPATH/bin
just clean        # remove binary and dist/
```
Without `just`: `go build -ldflags "-X github.com/FelineStateMachine/puzzletea/cmd.Version=$(git describe --tags --always --dirty)" -o puzzletea`

### CLI Seed Flags
```bash
puzzletea new --set-seed myseed              # deterministic game/mode/puzzle selection
puzzletea new nonogram epic --with-seed s1   # deterministic puzzle in selected game/mode
```
- `--set-seed` cannot be combined with positional game/mode arguments.
- `--with-seed` is used with explicit game/mode arguments for mode-local reproducibility.

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

Enabled linters (`.golangci.yml`): errcheck, gofumpt (extra-rules), gosimple, govet, ineffassign, misspell (US locale), staticcheck, unused.

---

## Project Structure
```
puzzletea/
├── main.go             # Entry point: wires cmd package
├── app/                # Root TUI model (Elm architecture)
│   ├── model.go, update.go, view.go, keys.go, spawn.go, debug.go
├── cmd/                # CLI commands (Cobra)
│   ├── root.go, new.go, continue.go, list.go
├── config/             # Persistent JSON config (~/.puzzletea/config.json)
├── theme/              # Color theming (WCAG-compliant palettes, contrast utils)
├── stats/              # XP/level math, streaks, card rendering
├── game/               # Plugin interfaces, cursor, keys, style, border helpers
├── store/              # SQLite persistence (~/.puzzletea/history.db)
├── ui/                 # Shared UI: menu list, main menu, table, panel, styles
├── daily/              # Daily puzzle seeding, RNG, mode selection
├── resolve/            # CLI argument resolution (category/mode name matching)
├── namegen/            # Adjective-noun name generator
├── hashiwokakero/, hitori/, lightsout/, nonogram/
├── shikaku/, sudoku/, takuzu/, wordsearch/  # Puzzle game packages
└── vhs/                # VHS tape files for GIF previews
```

Each puzzle package follows a consistent file structure:
- **Capitalized**: `Model.go`, `Gamemode.go`, `Export.go`
- **Lowercase**: `grid.go`, `keys.go`, `style.go`, `generator.go`, `mouse.go`, `<game>_test.go`
- **Docs**: `help.md` (embedded via `//go:embed`), `README.md`

---

## Code Style Guidelines

### Formatting
- Use `gofumpt` (stricter than gofmt, extra-rules enabled) -- run `just fmt`
- No comments required unless explaining non-obvious logic
- Keep lines under ~100 characters
- US English spelling enforced by misspell linter

### Imports
Two groups separated by a blank line: stdlib, then everything else (internal + external sorted together). When there are many internal imports, a third group separating internal from external is acceptable.
```go
import (
    "errors"
    "fmt"

    "github.com/FelineStateMachine/puzzletea/game"
    "charm.land/bubbles/v2/key"
    tea "charm.land/bubbletea/v2"
    "charm.land/lipgloss/v2"
)
```
Note: always alias bubbletea as `tea`.

### Naming Conventions
- **Types**: PascalCase (`Model`, `NonogramMode`, `Entry`)
- **Unexported types**: camelCase or lowercase (`grid`, `state`, `menuItem`)
- **Variables/Fields**: camelCase (`rowHints`, `currentHints`)
- **Constants**: camelCase (`mainMenuView`, `gameSelectView`)
- **Interfaces**: PascalCase (`Gamer`, `Spawner`, `Mode`)

### Type Declarations
Prefer grouped type blocks: `type ( grid [][]rune; state string )`

### Interface Compliance
Use compile-time checks in grouped var blocks:
```go
var (
    _ game.Mode          = NonogramMode{}
    _ game.Spawner       = NonogramMode{}
    _ game.SeededSpawner = NonogramMode{}
)
```

### Error Handling
- Return descriptive errors: `errors.New("puzzle width does not support row tomography definition")`
- Check errors immediately; use `fmt.Errorf` with `%w` only when wrapping adds context
- No assertion libraries in tests -- use `t.Errorf` / `t.Fatalf` only

### Styling
Use the `theme` package for colors (`theme.Current().Accent`, etc.) and `game/style.go` shared accessors (`CursorFG()`, `CursorBG()`, `ConflictFG()`). Use `compat.AdaptiveColor` from `charm.land/lipgloss/v2/compat` for adaptive light/dark colors.

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

Every mode type embeds `game.BaseMode` via `game.NewBaseMode(title, description)`.

### Puzzle Package Exports
Every puzzle package exports: `Modes`, `DailyModes` (`[]list.Item`), `HelpContent` (`string`, from `//go:embed help.md`), `NewMode(...)`, `New(...)`, `ImportModel([]byte) (*Model, error)`, and `DefaultKeyMap`.

Each package registers itself via `init()` in `Gamemode.go`:
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
// --- TitleBarView (P3) ---           // P3 = low-priority UI
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

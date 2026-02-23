# AGENTS.md - PuzzleTea Development Guide

## Commands

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

Always run `just fmt` and `just lint` before committing.

Enabled linters (`.golangci.yml`): errcheck, gofumpt (extra-rules), gosimple, govet, ineffassign, misspell (US locale), staticcheck, unused.

---

## Architecture

PuzzleTea is a terminal puzzle game collection built with Go using the **Bubble Tea TUI framework** (Elm architecture: Model-Update-View).

### Technology Stack
- **TUI**: Bubble Tea v2 (`charm.land/bubbletea/v2`, always aliased as `tea`) + Bubbles + Lip Gloss
- **CLI**: Cobra
- **PDF generation**: go-pdf/fpdf (half-letter size: 139.7mm × 215.9mm)
- **Persistence**: SQLite (`~/.puzzletea/history.db`)

### Control Flow
```
main() → cmd.RootCmd (Cobra)
  ├─ Default: Launch TUI (app.InitialModel → Elm loop)
  ├─ --new / --continue / --set-seed: direct game launch
  └─ Subcommands: new, continue, list, export-pdf
```

### Key Packages

| Package | Role |
|---------|------|
| `app/` | Root TUI model; 9 puzzle categories wired at startup |
| `cmd/` | Cobra CLI commands including `export-pdf` |
| `game/` | Plugin interfaces (`Gamer`, `Mode`, `Spawner`, `PrintAdapter`), registry |
| `pdfexport/` | PDF pipeline: JSONL parsing → per-game rendering → cover art |
| `store/` | SQLite persistence |
| `theme/` | 365 WCAG-compliant color themes |
| `stats/` | XP/level/streak system |
| `config/` | Persistent JSON config (`~/.puzzletea/config.json`) |
| `resolve/` | CLI argument matching for game/mode names |
| `daily/` | Deterministic daily puzzle seeding |
| `ui/` | Shared TUI components (menus, tables, panels) |

### Puzzle Packages
Eight printable games: `nonogram`, `sudoku`, `nurikabe`, `shikaku`, `wordsearch`, `hashiwokakero`, `hitori`, `takuzu`. One game without PDF export: `lightsout`.

Each puzzle package exposes: `Modes`, `DailyModes`, `HelpContent`, `NewMode(...)`, `New(...)`, `ImportModel([]byte)`, `DefaultKeyMap`, and registers itself via `init()` in `Gamemode.go`.

### Plugin Registration
```go
// In Gamemode.go init():
game.Register("Nonogram", func(data []byte) (game.Gamer, error) {
    return ImportModel(data)
})
// Optional PDF export registration:
game.RegisterPrintAdapter(adapter)
```

### PDF Export Pipeline
```
export-pdf command
  → ParseJSONLFiles()                  # parse schema puzzletea.export.v1
  → adapter.BuildPDFPayload()          # game-specific save → typed struct
  → pdfexport.OrderPuzzlesForPrint()   # difficulty-based ordering
  → pdfexport.WritePDF()               # cover + title pages + puzzle bodies + back
```

### File Layout per Puzzle Package
- **Capitalized**: `Model.go`, `Gamemode.go`, `Export.go`
- **Lowercase**: `grid.go`, `keys.go`, `style.go`, `generator.go`, `mouse.go`, `<game>_test.go`
- **Docs**: `help.md` (embedded via `//go:embed`), `README.md`

---

## Code Style

### Formatting
- Use `gofumpt` (stricter than gofmt, extra-rules enabled) — run `just fmt`
- Keep lines under ~100 characters
- US English spelling enforced by misspell linter

### Imports
Two groups: stdlib, then everything else (internal + external sorted together). Three groups acceptable when there are many internal imports.
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

### Naming
- **Types**: PascalCase (`Model`, `NonogramMode`, `Entry`)
- **Unexported types**: camelCase or lowercase (`grid`, `state`, `menuItem`)
- **Variables/Fields**: camelCase (`rowHints`, `currentHints`)
- **Constants**: camelCase (`mainMenuView`, `gameSelectView`)
- **Interfaces**: PascalCase (`Gamer`, `Spawner`, `Mode`)

### Type Declarations
Prefer grouped type blocks: `type ( grid [][]rune; state string )`

### Interface Compliance (compile-time checks)
```go
var (
    _ game.Mode          = NonogramMode{}
    _ game.Spawner       = NonogramMode{}
    _ game.SeededSpawner = NonogramMode{}
)
```

### Styling
Use `theme.Current().Accent` etc. from the `theme` package, and `game/style.go` shared accessors (`CursorFG()`, `CursorBG()`, `ConflictFG()`). Use `compat.AdaptiveColor` from `charm.land/lipgloss/v2/compat` for adaptive colors.

### Error Handling
Return descriptive errors; use `fmt.Errorf` with `%w` only when wrapping adds context. No assertion libraries in tests — use `t.Errorf`/`t.Fatalf` only.

---

## Testing Conventions

### Section Comments with Priority
```go
// --- generateTomography (P0) ---   // P0 = critical
// --- Grid serialization (P1) ---   // P1 = important
// --- generateRandomState (P2) ---  // P2 = generators/slow
// --- TitleBarView (P3) ---         // P3 = low-priority UI
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
            // test logic using t.Errorf / t.Fatalf only
        })
    }
}
```

### Slow Test Gating
```go
if testing.Short() {
    t.Skip("skipping slow generator test in short mode")
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

---

## Gamer Interface
Every puzzle `Model` must implement (from `game/gamer.go`):
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

Every mode type embeds `game.BaseMode` via `game.NewBaseMode(title, description)`.

## Global Keybindings
`Ctrl+N` Main menu | `Ctrl+C` Quit | `Ctrl+E` Debug overlay | `Ctrl+H` Full help | `Ctrl+R` Reset puzzle | `Enter` Select | `Escape` Back

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
just test         # go test ./...
just test-short   # go test -short ./... (skips slow generator tests)
go test ./nonogram/              # single package
go test ./sudoku/ -run TestGenerateGrid  # specific test
```

### Linting & Formatting
```bash
just lint         # golangci-lint run ./...
just fmt          # gofumpt -w .
just tidy         # go mod tidy
```
**Always run `just fmt` and `just lint` before committing.**

---

## Code Style Guidelines

### Formatting
- Use `gofumpt` - run `just fmt`
- No comments required unless explaining non-obvious logic
- Keep lines under ~100 characters
- Group imports: stdlib, puzzletea internal, Bubble Tea ecosystem

### File Naming
- **Capitalized**: `Gamemode.go`, `Model.go`, `Export.go`
- **Lowercase**: `grid.go`, `keys.go`, `style.go`, `generator.go`, `<game>_test.go`

### Naming Conventions
- **Types**: PascalCase (`Model`, `NonogramMode`, `grid`)
- **Variables/Fields**: camelCase (`rowHints`, `currentHints`)
- **Constants**: camelCase for constants
- **Interfaces**: PascalCase (`Gamer`, `Spawner`, `Mode`)

### Type Declarations
Prefer type aliases for clarity:
```go
type (
    grid  [][]rune
    state string
)
```

### Interface Compliance
Use compile-time checks:
```go
var _ game.Gamer = Model{}
var _ game.Mode = NonogramMode{}
var _ game.Spawner = NonogramMode{}
```

### Error Handling
- Return descriptive errors: `return Model{}, errors.New("puzzle width does not support row tomography definition")`
- Check errors immediately; avoid wrapping unless adding context
- Use `fmt.Errorf` with `%w` when wrapping

### Imports (grouped, blank line between)
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

---

## TUI Patterns

### Model Structure
```go
type Model struct {
    width, height int
    rowHints      TomographyDefinition
    colHints      TomographyDefinition
    cursor        game.Cursor
    grid          grid
    keys          KeyMap
    modeTitle     string
    showFullHelp  bool
    currentHints  Hints
    solved        bool
}
```

### Update Method Pattern
```go
func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
    switch msg := msg.(type) {
    case game.HelpToggleMsg:
        m.showFullHelp = msg.Show
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.FillTile):
            m.updateTile(filledTile)
        }
        m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
    }
    m.updateKeyBindings()
    return m, nil
}
```

### Required Gamer Methods
`Init() tea.Cmd`, `Update(tea.Msg)`, `View() string`, `GetDebugInfo() string`, `GetFullHelp() string`, `GetSave() []byte`, `IsSolved() bool`, `SetTitle(string)`

### Styling
Use `lipgloss.AdaptiveColor` with ANSI 256 palette numbers:
```go
lipgloss.AdaptiveColor{Light: "255", Dark: "255"}
```
Avoid hex colors (`#rrggbb`).

---

## Testing Conventions

Section comments with priority:
```go
// --- generateTomography (P0) ---
```

Table-driven tests with subtests:
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
            // test logic
        })
    }
}
```

Save/load round-trip:
```go
m, _ := New(mode, hints)
data := m.GetSave()
loaded, err := ImportModel(data)
if err != nil { t.Fatal(err) }
// verify state preserved
```

Priority: (P0) = critical, (P1) = important, (P2) = generators

---

## Project Structure
```
puzzletea/
├── main.go, cmd.go, model.go, resolve.go   # Core app
├── game/           # Plugin interface, cursor, keys, style
├── store/          # SQLite persistence
├── namegen/        # Adjective-noun name generator
├── nonogram/, sudoku/, wordsearch/, hashiwokakero/, lightsout/, takuzu/  # Puzzle games
└── vhs/           # GIF preview tapes
```

---

## Global Keybindings
- `Ctrl+N`: Main menu | `Ctrl+C`: Quit (saves abandoned) | `Ctrl+E`: Debug overlay | `Ctrl+H`: Full help | `Enter`: Select | `Escape`: Back

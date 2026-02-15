# Hitori Rendering Implementation Plan

## Overview

This document outlines the implementation of the Hitori puzzle rendering using Bubble Tea and Lipgloss. The rendering builds on patterns established in the existing Nonogram implementation (`nonogram/style.go`, `nonogram/Model.go`).

## Hitori Puzzle Rules

1. **No duplicate numbers** in any row/column among white (unshaded) cells
2. **No adjacent black cells** - black cells cannot share an edge (diagonal allowed)
3. **All white cells orthogonally connected** - single contiguous region

## Implementation Components

### 1. Files to Create

| File | Purpose |
|------|---------|
| `hitori/Model.go` | Main game model, implements `game.Gamer` interface |
| `hitori/style.go` | Lipgloss styles and rendering functions |
| `hitori/grid.go` | Grid data structure and manipulation |
| `hitori/keys.go` | Key bindings |
| `hitori/Gamemode.go` | Menu entry and spawner |
| `hitori/solver.go` | Validation/solving logic |

### 2. Model Structure

```go
type Model struct {
    width, height int
    grid          grid        // current puzzle state (numbers + shading)
    solutionMask  [][]bool    // true = cell should be white in solution
    cursor        game.Cursor
    keys          KeyMap
    modeTitle     string
    showFullHelp  bool
    solved        bool
}
```

### 3. Grid Cell States

Unlike Nonogram, Hitori has three states per cell:
- **Number cell**: displayed with number (rune '1'-'9')
- **Marked (X)**: player marked as black
- **Cleared**: player marked as white (empty)

```go
const (
    filledTile = 'X'   // marked black
    emptyTile = ' '    // cleared/white
    // Numbers 1-9 stored directly as rune
)
```

### 4. Key Bindings (from keys.go)

- Arrow keys / WASD: Move cursor
- `X` or `Z`: Mark cell as black (shaded)
- `Space` or `Backspace`: Clear cell (mark as white)
- `Enter`: Toggle cell state (number → X → space → number)
- `Ctrl+R`: Reset puzzle
- `Ctrl+H`: Toggle help

### 5. Rendering with Lipgloss

Following `nonogram/style.go` patterns:

```go
// Cell styles
var (
    baseStyle = lipgloss.NewStyle()
    
    numberStyle = baseStyle.
        Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "250"}).
        Background(lipgloss.AdaptiveColor{Light: "255", Dark: "235"})
    
    markedStyle = baseStyle.
        Foreground(lipgloss.AdaptiveColor{Light: "88", Dark: "196"}).
        Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"})
    
    cursorStyle = baseStyle.
        Bold(true).
        Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
        Background(lipgloss.AdaptiveColor{Light: "130", Dark: "214"})
    
    solvedStyle = baseStyle.
        Background(lipgloss.AdaptiveColor{Light: "151", Dark: "22"})
)
```

### 6. Grid Rendering

The grid renders similarly to Nonogram:
- Each cell is `cellWidth = 3` characters wide
- Separator lines every 5 cells (`spacerEvery = 5`)
- Use `lipgloss.JoinHorizontal` / `lipgloss.JoinVertical` to assemble
- Crosshair highlight on cursor row/column

```go
func gridView(g grid, c game.Cursor, solved bool) string {
    // Iterate rows, render each cell with appropriate style
    // Apply cursor highlighting, crosshair for cursor row/col
    // Join with vertical separators
}
```

### 7. Validation (Solver)

The `IsSolved()` check must verify:
1. No duplicate numbers in any row among non-marked cells
2. No adjacent marked cells
3. All non-marked cells are orthogonally connected

### 8. Game.Gamer Interface Methods

```go
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd)
func (m Model) View() string
func (m Model) GetDebugInfo() string
func (m Model) GetFullHelp() [][]key.Binding
func (m Model) GetSave() ([]byte, error)
func (m Model) IsSolved() bool
func (m Model) Reset() game.Gamer
func (m Model) SetTitle(string) game.Gamer
```

### 9. Integration

Register in main menu via `cmd.go`:
```go
game.Registry["hitori"] = hitori.ImportModel
```

---

## References

- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Lipgloss: https://github.com/charmbracelet/lipgloss
- Nonogram implementation: `nonogram/style.go`, `nonogram/Model.go`

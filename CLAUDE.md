# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PuzzleTea is a terminal-based puzzle game framework built with the Bubble Tea TUI library. It provides a plugin architecture for different puzzle types with a shared menu system and debug interface.

## Commands

### Building and Running
```bash
# Run the application
go run .

# Build the binary
go build -o puzzletea

# Update dependencies
go mod tidy
```

### Testing
```bash
# Run all tests (no tests exist yet)
go test ./...

# Run tests for a specific package
go test ./nonogram
```

## Architecture

### Core Application Structure

The application follows a **plugin-based architecture** where puzzle games implement the `game.Gamer` interface:

- **Main entry point** (`main.go`): Initializes the Bubble Tea program
- **Root model** (`model.go`): Manages application state with three views:
  - `gameSelectView`: Category selection (Nonogram, Word Search, Sudoku, Hashiwokakero)
  - `modeSelectView`: Difficulty/mode selection within a category
  - `gameView`: Active game instance

Navigation flow: select a game category → select a mode/difficulty → play the game. `Escape` returns from mode select to game select, `Ctrl+N` returns to game select from anywhere.

### Plugin Interface (`game/gamer.go`)

All puzzle games must implement:

1. **`game.Gamer`** - The active game instance (Init, Update, View, GetDebugInfo, GetFullHelp, GetSave)
2. **`game.Spawner`** - Creates a new game instance: `Spawn() (Gamer, error)`
3. **`game.BaseMode`** - Embed this in your mode struct to get `Title()`, `Description()`, and `FilterValue()` for free
4. **`game.Category`** - Groups related modes under a heading in the menu; satisfies `Mode` but not `Spawner`

### Adding a New Puzzle Type

1. Create a new package under the project root (e.g., `mypuzzle/`)
2. Define a mode struct that embeds `game.BaseMode` and implements `game.Spawner`
3. Use compile-time interface checks: `var _ game.Mode = MyMode{}` and `var _ game.Spawner = MyMode{}`
4. Export a `Modes` variable (`[]list.Item`) listing available difficulties
5. Implement `game.Gamer` on a `Model` struct
6. Register the category in `model.go`'s `GameCategories` slice

Each package follows a consistent file structure:
- `Gamemode.go`: Mode struct, `NewMode()`, `Spawn()`, `Modes` var
- `Model.go`: `Model` struct implementing `game.Gamer`
- `Export.go`: Exported `Save` struct, `GetSave()` method (on Model), `ImportModel([]byte) (*Model, error)` function
- `keys.go`: `KeyMap` struct with game-specific keybindings, `var DefaultKeyMap = KeyMap{...}`
- `style.go`: lipgloss styling definitions
- `generator.go`: Puzzle generation logic (where applicable)

### Current Implementations

**Nonogram** (`nonogram/`): Grid-based puzzle with row/column hints. Three tile states: filled (`.`), marked (`-`), empty (` `). Victory when current grid tomography matches hint tomography.

**Word Search** (`wordsearch/`): Find hidden words in a letter grid. Select start and end positions to highlight words. Supports 8 directions (Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft) with configurable subsets per difficulty.

**Sudoku** (`sudoku/`): Classic 9x9 sudoku with configurable number of provided cells. Uses `cell` type with provided/user-entered distinction.

**Hashiwokakero** (`hashiwokakero/`): Connect islands with bridges (1 or 2). Two input modes: navigation mode (move cursor between islands) and bridge mode (select an island, then cycle bridges in a direction). Victory requires all islands satisfied and connected. Uses `Island`, `Bridge`, `Puzzle` types.

### Global Keybindings

Handled by root model:
- `Ctrl+N`: Return to game select from any view
- `Ctrl+C`: Quit application
- `Ctrl+E`: Toggle debug overlay
- `Enter`: Select item in menus
- `Escape`: Return from mode select to game select

Individual games handle their own gameplay keybindings via `KeyMap` types in `keys.go`.

In `model.go`, key events are handled in a single `switch msg.Type` block. Each key type (e.g. `tea.KeyEnter`) has **one `case`** with an if-chain that checks `m.state` to determine per-view behavior. Never add a duplicate `case` for the same key type — add new state checks within the existing case.

### Continue View

The continue view (`continueView`) displays saved games in a `table.Model`. Game records are cached in `m.continueGames` (populated by `initContinueTable()` each time the user navigates to the view). Use this slice directly by index to look up a selected game — no need to re-query the store. Pressing Enter on a row calls the matching package's `ImportModel()` to restore the game.

### Debug System

The debug overlay (`Ctrl+E`) displays game-specific markdown rendered via `glamour`. Each game provides debug info via `GetDebugInfo()` — typically cursor position, dimensions, solved state, and game-specific details.

### Grid System Pattern

Puzzle implementations use a similar grid pattern:
- Internal grid type: 2D slice (`[][]rune` for nonogram/wordsearch, `[][]cell` for sudoku)
- `state` type: String serialization for save/load
- Conversion functions: `newGrid(state)` and `grid.String()`

Hashiwokakero uses a different approach with `Island` and `Bridge` structs rather than a 2D grid.

### Color Convention

All colors should use `lipgloss.AdaptiveColor{Light: "X", Dark: "Y"}` with ANSI 256 palette numbers to support both light and dark terminals. Avoid raw hex colors (`#rrggbb`) as they require true-color support and don't adapt to terminal background.

## Module Information

- **Module path**: `github.com/FelineStateMachine/puzzletea`
- **Go version**: 1.24.5
- **Key dependencies**:
  - `github.com/charmbracelet/bubbletea`: TUI framework
  - `github.com/charmbracelet/bubbles`: TUI components (list)
  - `github.com/charmbracelet/lipgloss`: Terminal styling
  - `github.com/charmbracelet/glamour`: Markdown rendering (debug view)

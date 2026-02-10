# CLAUDE.md

## Project Overview

PuzzleTea is a terminal-based puzzle game framework built with the Bubble Tea TUI library. It provides a plugin architecture for different puzzle types with a shared menu system, save/load persistence, and debug interface.

## Commands

### Building and Running
```bash
just              # build with version from git tags
just run          # build and run
just install      # install to $GOPATH/bin
just clean        # remove binary and dist/
```

Or without just:
```bash
go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o puzzletea
```

### CLI Usage
```bash
puzzletea                          # launch TUI menu (default)
puzzletea new <game> [mode]        # start a new game directly
puzzletea new nonogram medium      # example: new nonogram, medium mode
puzzletea new lights-out           # hyphenated names, default (first) mode
puzzletea new hashi                # aliases: hashi, bridges, lights, ws, words
puzzletea continue <name>          # resume a saved game by its unique name
puzzletea list                     # print saved games table to stdout
puzzletea list --all               # include abandoned games
puzzletea --new nonogram:medium    # flag alias (colon-separated game:mode)
puzzletea --continue <name>        # flag alias
```

### Testing and Quality
```bash
just test         # go test ./...
just lint         # golangci-lint run ./...
just fmt          # gofumpt -w .
just tidy         # go mod tidy
```

### VHS GIF Previews
```bash
just vhs          # generate all game preview GIFs (requires vhs + ffmpeg + ttyd)
vhs vhs/menu.tape # generate a single GIF
```

## Architecture

### Core Application Structure

The application follows a **plugin-based architecture** where puzzle games implement the `game.Gamer` interface.

- **`main.go`**: Entry point; calls `rootCmd.Execute()` (Cobra).
- **`cmd.go`**: Cobra command definitions (`rootCmd`, `newCmd`, `continueCmd`, `listCmd`) and root-level `--new`/`--continue` flag aliases. Contains `launchNewGame()` and `continueGame()` helpers that open the store, resolve names, spawn/import the game, and launch the TUI in `gameView` state.
- **`resolve.go`**: Case-insensitive, hyphen/underscore-tolerant game and mode name resolution. `resolveCategory(name)` matches against `GameCategories` with alias support (e.g. `hashi` → Hashiwokakero, `ws` → Word Search). `resolveMode(cat, name)` matches mode titles; defaults to first mode when name is empty.
- **`model.go`**: Root model managing application state across five views:
  - `mainMenuView`: Top-level menu (Daily Puzzle, Generate, Continue)
  - `gameSelectView`: Category selection
  - `modeSelectView`: Difficulty/mode selection within a category
  - `gameView`: Active game instance
  - `continueView`: Saved game browser (table)

Navigation: main menu → select category → select mode → play. `Escape` goes back one level, `Ctrl+N` returns to main menu from anywhere.

In `model.go`, key events are handled in a single `switch msg.Type` block. Each key type (e.g. `tea.KeyEnter`) has **one `case`** with an if-chain that checks `m.state`. Never add a duplicate `case` for the same key type.

### Plugin Interface (`game/`)

All puzzle games must implement:

1. **`game.Gamer`** — Active game instance: `Init`, `Update`, `View`, `GetDebugInfo`, `GetFullHelp`, `GetSave`, `IsSolved`, `SetTitle`
2. **`game.Spawner`** — Creates a new game: `Spawn() (Gamer, error)`
3. **`game.BaseMode`** — Embed in mode structs for free `Title()`, `Description()`, `FilterValue()`
4. **`game.Category`** — Groups modes under a heading; satisfies `Mode` but not `Spawner`

Shared utilities in `game/`:
- **`cursor.go`**: `Cursor` type with `Move()` for grid-based navigation
- **`keys.go`**: `CursorKeyMap` and `DefaultCursorKeyMap` for shared arrow/WASD/vim bindings
- **`style.go`**: `TitleBarView()`, `DebugHeader()`, `DebugTable()` for consistent rendering

Save/load uses `game.Registry` — each package registers its import function via `init()` calling `game.Register(name, fn)`.

### Adding a New Puzzle Type

1. Create a new package under the project root (e.g., `mypuzzle/`)
2. Define a mode struct embedding `game.BaseMode` that implements `game.Spawner`
3. Add compile-time checks: `var _ game.Mode = MyMode{}` and `var _ game.Spawner = MyMode{}`
4. Export a `Modes` variable (`[]list.Item`) listing available difficulties
5. Implement `game.Gamer` on a `Model` struct
6. Register the import function in an `init()`: `game.Register("My Puzzle", func(data []byte) (game.Gamer, error) { ... })`
7. Add the category to `GameCategories` in `model.go`

Each package follows a consistent file structure:
- `Gamemode.go`: Mode struct, `NewMode()`, `Spawn()`, `Modes` var, `init()` with `game.Register()`
- `Model.go`: `Model` struct implementing `game.Gamer`
- `Export.go`: `Save` struct, `GetSave()`, `ImportModel([]byte)` function
- `keys.go`: `KeyMap` struct with game-specific keybindings
- `style.go`: lipgloss styling definitions
- `generator.go`: Puzzle generation logic (where applicable)
- `grid.go`: Grid type, serialization, and helpers (for grid-based games)

### Current Implementations

- **Nonogram** (`nonogram/`): Fill cells to match row/column hints. Three tile states: filled, marked, empty. Victory when grid tomography matches hint tomography.
- **Word Search** (`wordsearch/`): Find hidden words in a letter grid. Supports 8 directions with configurable subsets per difficulty.
- **Sudoku** (`sudoku/`): Classic 9x9 with configurable provided cells. Uses `cell` type with provided/user-entered distinction.
- **Hashiwokakero** (`hashiwokakero/`): Connect islands with bridges (1 or 2). Navigation mode and bridge mode. Victory requires all islands satisfied and connected.
- **Lights Out** (`lightsout/`): Toggle lights in a cross pattern to turn all off. Grid sizes from 3x3 to 9x9.

### Supporting Packages

- **`store/`**: SQLite-based persistence (via `modernc.org/sqlite`). Stores game records with status tracking (new, in_progress, completed, abandoned). DB at `~/.puzzletea/history.db`.
- **`namegen/`**: Generates unique adjective-noun names for saved games.

### Global Keybindings

- `Ctrl+N`: Return to main menu
- `Ctrl+C`: Quit (saves current game as abandoned)
- `Ctrl+E`: Toggle debug overlay
- `Ctrl+H`: Toggle full help display
- `Enter`: Select in menus
- `Escape`: Go back one level

### Continue View

Displays saved games in a `table.Model`. Records cached in `m.continueGames` (populated by `initContinueTable()`). Use this slice by index — no need to re-query the store. Enter calls the matching package's `ImportModel()`.

### Debug System

Debug overlay (`Ctrl+E`) renders game-specific markdown via `glamour`. Games provide info via `GetDebugInfo()` using `game.DebugHeader()` and `game.DebugTable()` helpers.

### Grid System Pattern

Grid-based games (nonogram, wordsearch, sudoku) use:
- Internal grid type: 2D slice (`[][]rune`, `[9][9]cell`, etc.)
- `state` type: String serialization for save/load
- `newGrid(state)` and `grid.String()` conversion functions

Hashiwokakero uses `Island` and `Bridge` structs instead of a 2D grid. Lights Out uses `[][]bool` directly.

### Versioning

Version is injected at build time via `-ldflags "-X main.version=..."`. The `version` variable lives in `cmd.go` and defaults to `"dev"`. The justfile derives it from `git describe --tags`. GoReleaser (`.goreleaser.yml`) handles cross-platform release builds and injects the tag version automatically.

### Formatting and Linting

All Go code must pass `gofumpt` (strict superset of `gofmt`) and `golangci-lint` with the project's `.golangci.yml` config. Enabled linters: gofumpt, govet, errcheck, staticcheck, unused, gosimple, ineffassign, misspell. Run `just fmt` and `just lint` before committing.

### Color Convention

All colors must use `lipgloss.AdaptiveColor{Light: "X", Dark: "Y"}` with ANSI 256 palette numbers. Avoid hex colors (`#rrggbb`) — they require true-color support and don't adapt to terminal background.

### VHS Tape Files (`vhs/`)

Each game has a `.tape` file in `vhs/` that generates a GIF preview using [charmbracelet/vhs](https://github.com/charmbracelet/vhs). All tapes share a common preamble:

```tape
Set FontSize 13
Set Width 854
Set Height 480
Set Padding 40
Set Theme "Catppuccin Mocha"

Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Env CI ""
```

The three `Env` lines are required for CI color support. `termenv` (used by lipgloss/bubbletea) checks `CI` env var and returns no-color when set — GitHub Actions sets `CI=true`. `COLORTERM=truecolor` sets the color profile, and `CLICOLOR_FORCE=1` is a fallback.

Each tape uses a `Hide` block to build the binary and clear the screen before `Show`ing the gameplay commands. Use `Ctrl+L` (not `clear`) to clear the screen in the hidden block.

GIFs are not tracked in git — they are generated by CI and committed via PR.

### CI/CD Workflows

- **`ci.yml`**: Runs tests and linting on PRs and pushes to main.
- **`release.yml`**: GoReleaser builds on tagged releases.
- **`vhs.yml`**: Generates VHS GIF previews. Triggers on release, `workflow_dispatch`, or pushes to main that change `vhs/*.tape`. Creates a PR (`chore/regenerate-vhs-gifs` branch) with the updated GIFs rather than pushing directly to main (required by branch protection rules).

## Module Information

- **Module path**: `github.com/FelineStateMachine/puzzletea`
- **Go version**: 1.24.5
- **Key dependencies**:
  - `github.com/charmbracelet/bubbletea`: TUI framework
  - `github.com/charmbracelet/bubbles`: TUI components (list, table)
  - `github.com/charmbracelet/lipgloss`: Terminal styling
  - `github.com/charmbracelet/glamour`: Markdown rendering (debug view)
  - `github.com/spf13/cobra`: CLI command/flag management
  - `modernc.org/sqlite`: Pure-Go SQLite driver (game persistence)

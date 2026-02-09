# PuzzleTea

A terminal-based puzzle game collection built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Five puzzle types, multiple difficulty modes, save/load persistence, and a plugin architecture for adding new games.

## Games

| Game | Description | Modes |
|------|-------------|-------|
| **Nonogram** | Fill cells to match row and column hints | Easy (5x5), Medium (10x10), Hard (15x15), Extra (5x10) |
| **Sudoku** | Classic 9x9 grid | Beginner, Easy, Medium, Hard, Expert, Diabolical |
| **Word Search** | Find hidden words in a letter grid | Easy, Medium, Hard (3-8 directions) |
| **Hashiwokakero** | Connect islands with bridges | 12 modes across 7x7 to 13x13 grids |
| **Lights Out** | Toggle lights to turn all off | Easy (3x3) to Extreme (9x9) |

## Install

### From release binaries

Download the latest binary for your platform from the [Releases](https://github.com/FelineStateMachine/puzzletea/releases) page.

### From source

Requires Go 1.24+.

```bash
go install github.com/FelineStateMachine/puzzletea@latest
```

Or clone and build:

```bash
git clone https://github.com/FelineStateMachine/puzzletea.git
cd puzzletea
just        # or: go build -o puzzletea
```

## Usage

Launch the interactive menu:

```
puzzletea
```

Or jump straight into a game:

```bash
puzzletea new nonogram medium
puzzletea new sudoku hard
puzzletea new lights-out
puzzletea new hashi easy
```

Manage saved games:

```bash
puzzletea list                     # show saved games
puzzletea continue amber-falcon    # resume by name
```

Flag aliases also work:

```bash
puzzletea --new nonogram:medium
puzzletea --continue amber-falcon
```

## Controls

### Global

| Key | Action |
|-----|--------|
| `Enter` | Select |
| `Escape` | Go back |
| `Ctrl+N` | Return to main menu |
| `Ctrl+C` | Quit |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+H` | Toggle full help |

### Navigation

Arrow keys, WASD, and Vim bindings (hjkl) are supported for grid movement across all games.

## Game Persistence

Games are automatically saved to `~/.puzzletea/history.db` (SQLite). Navigating away saves progress; quitting with `Ctrl+C` marks the game as abandoned. Completed games are preserved and can be revisited.

## Building

[just](https://github.com/casey/just) is used as the command runner:

```bash
just              # build
just run          # build and run
just test         # run tests
just lint         # run golangci-lint
just fmt          # format with gofumpt
just install      # install to $GOPATH/bin
just clean        # remove build artifacts
```

## Adding a New Puzzle

PuzzleTea uses a plugin architecture. To add a new puzzle type:

1. Create a package (e.g., `mypuzzle/`)
2. Implement the `game.Gamer` interface on a Model struct
3. Define modes embedding `game.BaseMode` that implement `game.Spawner`
4. Register an import function in `init()` for save/load support
5. Add the category to `GameCategories` in `model.go`

See any existing game package (e.g., `nonogram/`) for the full pattern.

## License

See [LICENSE](LICENSE) for details.

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Pure-Go SQLite

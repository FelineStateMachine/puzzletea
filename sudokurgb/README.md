# Sudoku RGB

Sudoku RGB is a 9x9 multiset-constraint puzzle inspired by Ripeto.

## How to Play

Fill every cell with one of three values:

- `1` = `▲`
- `2` = `■`
- `3` = `●`

The goal is for every row, every column, and every 3x3 box to contain the
same multiset:

`{1,1,1,2,2,2,3,3,3}`

Unlike classic Sudoku, repeated values are allowed within a house up to their
quota of three. A conflict appears only when a house contains **more than**
three copies of a value.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| Mouse left-click | Focus a cell |
| `1` / `2` / `3` | Place `▲` / `■` / `●` |
| `Backspace` | Clear cell |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Escape` | Return to main menu |

## Modes

| Mode | Clues | Description |
|------|-------|-------------|
| Beginner | 60 | Gentle intro to RGB quota logic |
| Easy | 54 | Early rows and boxes resolve quickly |
| Medium | 48 | Mixed row, column, and box pressure |
| Hard | 42 | More ambiguous houses and cross-checking |
| Expert | 36 | Sparse givens require deeper scanning |
| Diabolical | 30 | Tightest clue budget in the launch set |

## Quick Start

```bash
puzzletea new "sudoku rgb" beginner
puzzletea new ripeto medium
puzzletea new "rgb sudoku" diabolical
```

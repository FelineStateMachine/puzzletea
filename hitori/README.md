# Hitori

Shade cells to eliminate duplicate numbers in every row and column.

![Hitori gameplay](../vhs/hitori.gif)

## How to Play

The grid starts filled with numbers. Shade some cells black so that three
rules are satisfied:

1. **No duplicates**: No number appears more than once in any row or column among the unshaded (white) cells.
2. **No adjacent shading**: No two shaded cells share an edge (diagonal is fine).
3. **White connectivity**: All unshaded cells must form a single orthogonally connected region.

The puzzle is solved when all three rules hold simultaneously.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `x` | Shade cell (toggle) |
| `z` | Circle cell as confirmed white (toggle) |
| `Backspace` | Clear cell to unmarked |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Description |
|------|------|-------------|
| Mini | 5x5 | Gentle introduction |
| Easy | 6x6 | Straightforward logic |
| Medium | 8x8 | Moderate challenge |
| Tricky | 9x9 | Requires careful deduction |
| Hard | 10x10 | Advanced logic chains |
| Expert | 12x12 | Maximum challenge |

## Quick Start

```bash
puzzletea new hitori mini
puzzletea new hitori medium
puzzletea new hitori expert
```

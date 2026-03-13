# Fillomino

Grow orthogonally connected regions until each region contains exactly the number written in its cells.

![Fillomino gameplay](../vhs/fillomino.gif)

## How to Play

Each number belongs to a region whose area must match that number.
Grow and merge regions until every filled cell belongs to a valid orthogonally
connected group.

1. Every region must be orthogonally connected.
2. A region's area must equal the number written in its cells.
3. Regions with the same number may not touch orthogonally.
4. Given cells are fixed and cannot be changed.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| Mouse left-click | Focus a cell |
| `1`-`9` | Place number |
| `Backspace` / `Delete` | Clear cell |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Escape` | Return to main menu |

## Modes

| Mode | Size | Notes |
|------|------|-------|
| Mini 5x5 | 5x5 | Fast intro board |
| Easy 6x6 | 6x6 | Gentle deduction |
| Medium 8x8 | 8x8 | Balanced board |
| Hard 10x10 | 10x10 | Longer chains |
| Expert 12x12 | 12x12 | Broad region planning |

## Quick Start

```bash
puzzletea new fillomino
puzzletea new fillomino "Medium 8x8"
puzzletea new fillomino "Hard 10x10" --with-seed sample-seed
```

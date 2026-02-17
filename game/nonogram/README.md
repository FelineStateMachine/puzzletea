# Nonogram

Fill cells to match row and column hints, revealing a hidden pattern.

![Nonogram gameplay](../vhs/nonogram.gif)

## How to Play

Navigate the grid and fill or mark cells based on the numeric hints along
each row and column. The hints tell you how many consecutive filled cells
appear in that line, in order. Hints turn green when a row or column is
correctly satisfied.

- **Filled** cells count toward hint satisfaction.
- **Marked** cells are visual reminders that a cell should stay empty.
- The puzzle is solved when every row and column matches its hints.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `z` | Fill cell |
| `x` | Mark cell |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Density | Description |
|------|------|---------|-------------|
| Easy 5x5 | 5x5 | ~35% | Simple hints |
| Medium 5x5 | 5x5 | ~50% | Balanced challenge |
| Hard 5x5 | 5x5 | ~65% | Dense hints |
| Easy 10x10 | 10x10 | ~35% | Simple hints |
| Medium 10x10 | 10x10 | ~50% | Balanced challenge |
| Hard 10x10 | 10x10 | ~65% | Dense hints |
| Easy 15x15 | 15x15 | ~35% | Simple hints |
| Medium 15x15 | 15x15 | ~50% | Balanced challenge |
| Hard 15x15 | 15x15 | ~65% | Dense hints |
| Easy 20x20 | 20x20 | ~35% | Simple hints |
| Medium 20x20 | 20x20 | ~50% | Balanced challenge |
| Hard 20x20 | 20x20 | ~65% | Dense hints |

## Quick Start

```bash
puzzletea new nonogram easy-5x5
puzzletea new nonogram medium-10x10
puzzletea new nonogram hard-20x20
```

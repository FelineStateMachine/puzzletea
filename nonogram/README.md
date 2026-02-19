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
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Fill | Description |
|------|------|------|-------------|
| Mini | 5x5 | ~65% | Quick puzzle, straightforward hints |
| Pocket | 5x5 | ~50% | Compact but balanced |
| Teaser | 5x5 | ~35% | Small but tricky |
| Standard | 10x10 | ~67% | Classic size, dense hints |
| Classic | 10x10 | ~52% | The typical nonogram experience |
| Tricky | 10x10 | ~37% | Sparse hints require reasoning |
| Large | 15x15 | ~69% | Bigger grid, constraining hints |
| Grand | 15x15 | ~54% | A substantial challenge |
| Epic | 20x20 | ~71% | A epic undertaking |
| Massive | 20x20 | ~56% | Truly massive puzzle |

## Quick Start

```bash
puzzletea new nonogram mini
puzzletea new nonogram classic
puzzletea new nonogram massive
```

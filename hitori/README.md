# Hitori

Shade cells so each row and column contains no duplicate numbers.

![Hitori gameplay](../vhs/hitori.gif)

## How to Play

Shade cells to remove them from consideration. The remaining un-shaded cells
must satisfy three rules:

- **No duplicates**: Each row and column must contain unique numbers (shaded
  cells don't count).
- **No adjacent shaded**: Shaded cells cannot share an edge.
- **All connected**: All un-shaded cells must form a single connected group.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `z` | Shade cell |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Clues | Description |
|------|------|-------|--------------|
| Tiny | 5x5 | 55% | Quick introduction |
| Easy | 6x6 | 50% | Simple patterns |
| Medium | 8x8 | 45% | Standard difficulty |
| Hard | 10x10 | 40% | Longer deductions |
| Expert | 12x12 | 35% | Complex logic |

## Quick Start

```bash
puzzletea new hitori tiny
puzzletea new hitori medium
puzzletea new hitori expert
```

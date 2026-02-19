# Shikaku

Divide the grid into rectangles so each contains exactly one number equal to its area.

![Shikaku gameplay](../vhs/shikaku.gif)

## How to Play

The grid contains numbered clues. Your goal is to partition the entire grid
into non-overlapping rectangles where each rectangle contains exactly one
clue and the clue's value equals the rectangle's area.

1. Navigate to a numbered clue and press `Enter` or `Space` to select it.
2. Press arrow keys to expand the rectangle in each direction.
3. Press `Shift+Arrow` to shrink a side if you overshot.
4. Press `Enter` or `Space` to confirm when the preview is green.
5. Press `Escape` to cancel, or `Backspace` to delete.

## Controls

### Navigation Mode

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `Enter` / `Space` | Select clue for expansion |
| `Backspace` | Delete rectangle at cursor |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

### Expansion Mode (after selecting a clue)

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Expand rectangle in that direction |
| Shift+Arrow / WASD / HJKL | Shrink rectangle from that side |
| `Enter` / `Space` | Confirm placement (green preview only) |
| `Escape` | Cancel, discard preview |
| `Backspace` | Delete existing rectangle, return to nav |

## Modes

| Mode | Grid | Max Rect | Description |
|------|------|----------|-------------|
| Mini 5x5 | 5x5 | 5 | Gentle introduction |
| Easy 7x7 | 7x7 | 8 | Straightforward puzzles |
| Medium 8x8 | 8x8 | 12 | Moderate challenge |
| Hard 10x10 | 10x10 | 15 | Requires careful planning |
| Expert 12x12 | 12x12 | 20 | Maximum challenge |

## Quick Start

```bash
puzzletea new shikaku mini-5x5
puzzletea new shikaku easy-7x7
puzzletea new shikaku hard-10x10
```

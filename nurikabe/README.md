# Nurikabe

Build islands from numbered clues while keeping the sea connected.

## How to Play

Each number is an island clue. Mark cells as island or sea until all rules are true:

1. Every island contains exactly one clue.
2. Island size equals clue value.
3. All sea cells are one connected region.
4. No 2x2 sea block is allowed.

Clue cells are fixed and always island cells.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `x` | Set sea |
| `z` | Set island |
| `Backspace` | Clear to unknown |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Mouse

| Action | Effect |
|--------|--------|
| Left click / drag | Toggle sea/unknown from click target, then paint that target while dragging |
| Right click / drag | Set island cells |

## Modes

| Mode | Grid | Description |
|------|------|-------------|
| Mini | 5x5 | Gentle introduction |
| Easy | 7x7 | Balanced logic |
| Medium | 9x9 | Moderate deduction |
| Hard | 11x11 | Lower clue density |
| Expert | 12x12 | Sparse clues and longer chains |

## Quick Start

```bash
puzzletea new nurikabe mini
puzzletea new nurikabe medium
puzzletea new nurikabe expert
puzzletea new islands easy       # alias
```

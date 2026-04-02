# Netwalk

Rotate network tiles until every active connection reaches the server in one loop-free tree.

![Netwalk gameplay](../vhs/netwalk.gif)

## Quick Start

```bash
puzzletea new netwalk "Mini 5x5"
puzzletea new netwalk "Easy 7x7"
puzzletea new network medium
```

## Rules

- Every active tile contains a fixed connector pattern that can only be rotated.
- All connectors must match neighboring connectors exactly.
- No connector may point off the board or into an empty cell.
- The final network must be a single connected tree rooted at the server.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Space` | Rotate clockwise |
| `Backspace` | Rotate counter-clockwise |
| `Enter` | Toggle lock |

## Modes

| Mode | Board |
|------|-------|
| `Mini 5x5` | Small starter tree |
| `Easy 7x7` | Moderate branch count |
| `Medium 9x9` | More global interaction |
| `Hard 11x11` | Longer branch chains |
| `Expert 13x13` | Largest network |

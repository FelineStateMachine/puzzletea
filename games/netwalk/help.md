# Netwalk

Rotate network tiles until every active tile connects back to the server in one clean network.

## Rules

- Each active tile contains fixed connectors that may only be **rotated**.
- Every connector must meet a matching connector on the neighboring tile.
- Connectors may not point off the board or into empty cells.
- The puzzle is solved when every active tile connects to the **server** and the finished network has **no loops**.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Space` | Rotate current tile clockwise |
| `Backspace` | Rotate current tile counter-clockwise |
| `Enter` | Toggle lock on current tile |
| `Mouse left-click` | Rotate clicked tile |
| `Mouse right-click` | Toggle lock on clicked tile |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Escape` | Return to main menu |

## Tips

- **Start at the server.** Expand outward from the known connected region and fix obvious dead ends first.
- **Watch the borders.** Edge and corner tiles have fewer legal orientations because they cannot point off the board.
- **Lock confirmed tiles.** Once a branch is clearly correct, lock it to avoid undoing progress while tracing the rest of the network.

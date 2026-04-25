# Fillomino

Fill the board so that each orthogonally connected region contains exactly as many cells as its number.

## Rules

- Every numbered region must be orthogonally connected.
- The size of each region must equal the number written in its cells.
- Regions with the same number may exist multiple times, but they cannot touch orthogonally.
- Given cells are locked. The puzzle is solved when every cell is filled and every region is valid.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Mouse left-click` | Focus a cell |
| `1`-`9` | Place number |
| `Backspace` / `Delete` | Clear cell |
| `Ctrl+H` | Toggle full help |
| `Ctrl+R` | Reset puzzle |
| `Escape` | Return to main menu |

## Tips

- Start from givens that force a small region, like `1`, `2`, or `3`.
- If a connected group already has its full size, every orthogonal neighbour must belong to another region.
- If a group cannot possibly grow to its required size, something inside or next to it is wrong.

# Ripple Effect

Fill each cage with the digits `1` through the cage size, then keep equal digits far enough apart so their ripples do not collide.

## Rules

- Every cage of size `n` must contain the digits `1..n` exactly once.
- Equal digits cannot appear too close in the same row or column.
- A digit `n` blocks the next `n-1` cells in each orthogonal direction from also being `n`.
- Given cells are locked. The puzzle is solved when every cell is filled and no cage or ripple rule is broken.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `1`-`9` | Place number |
| `Backspace` / `Delete` | Clear cell |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Tips

- Finish the smallest cages first; they have the fewest possibilities.
- When a digit is placed, count outward in the row and column to eliminate matching values nearby.
- If a cage already contains a digit, that value cannot appear again in the same cage even if the ripple rule would allow it.

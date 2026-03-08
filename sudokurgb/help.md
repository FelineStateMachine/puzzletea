# Sudoku RGB

Fill a 9x9 grid with RGB symbols so every row, column, and box matches the same multiset.

## Rules

- Each **row** must contain `{1,1,1,2,2,2,3,3,3}`.
- Each **column** must contain `{1,1,1,2,2,2,3,3,3}`.
- Each **3x3 box** must contain `{1,1,1,2,2,2,3,3,3}`.
- `1`, `2`, and `3` map to `▲`, `■`, and `●`.
- Pre-filled cells cannot be changed. Cells only conflict when a row, column,
  or box contains **more than three** of one value.

Inspired by Ripeto.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Mouse left-click` | Focus a cell |
| `1` / `2` / `3` | Fill cell with `▲` / `■` / `●` |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Escape` | Return to main menu |

## Tips

- **Count aggressively.** Every house wants exactly three of each value, so
  once a row already has three `▲`, the rest of that row cannot be `▲`.
- **Cross-check rows and boxes.** A value capped out in both the row and the
  box removes options quickly.
- **Use conflicts as quota warnings.** Red highlights mean a house already has
  too many copies of a value.

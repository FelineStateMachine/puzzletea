# Sudoku RGB

Fill a 9x9 grid with RGB symbols so every row, column, and box matches the same multiset.

## Rules

- Each **row** must contain `{1,1,1,2,2,2,3,3,3}`.
- Each **column** must contain `{1,1,1,2,2,2,3,3,3}`.
- Each **3x3 box** must contain `{1,1,1,2,2,2,3,3,3}`.
- `1`, `2`, and `3` map to `▲`, `■`, and `●`.
- The RGB hint counts show the live total for each row and column.
- Pre-filled cells cannot be changed. Row and column over-quota conflicts are
  shown on the hint counts. Only 3x3 box conflicts highlight cell backgrounds.

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
  once a row already has three `▲`, the solved hint chip marks that value.
- **Cross-check rows and boxes.** Use the row and column RGB counts to narrow
  options before you look for box-only conflicts.
- **Watch where errors appear.** Hint-count errors mean a row or column is
  over quota; red cell backgrounds mean the problem is inside a 3x3 box.

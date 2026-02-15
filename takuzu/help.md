# Takuzu

Fill the grid with two symbols so that every row and column obeys three simple rules.

## Rules

- **No triples**: Three consecutive identical symbols in a row or column are not allowed.
- **Equal count**: Each row and column must contain exactly the same number of ● and ○.
- **Unique lines**: No two rows may be identical, and no two columns may be identical.

Pre-filled cells (shown in **bold**) are locked and cannot be changed. The puzzle is
solved when every cell is filled and all three rules are satisfied.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `z` / `0` | Place ● (filled) |
| `x` / `1` | Place ○ (open) |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Tips

- **Sandwich rule**: If two identical symbols are next to each other, the cells on
  either side *must* be the opposite symbol. For example, ● ● means the neighbours
  are both ○.
- **Count what you have**: When a row or column already has its full quota of one
  symbol, every remaining empty cell in that line must be the other symbol.
- **Avoid the third**: Before placing a symbol, glance one and two cells ahead in
  each direction. If placing it would create three in a row, it has to be the
  other symbol.

# Hitori

Shade cells to eliminate duplicate numbers in every row and column.

## Rules

- **No duplicates**: No number may appear more than once in any row or
  column among the unshaded (white) cells.
- **No adjacent shading**: No two shaded cells may share an edge.
  Diagonal contact is fine.
- **White connectivity**: All unshaded cells must form a single
  orthogonally connected region.

The puzzle is solved when all three rules hold simultaneously.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `x` | Toggle shade on cell |
| `z` | Toggle circle (confirmed white) |
| `Backspace` | Clear cell to unmarked |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Tips

- **Circle the unique numbers first.** If a number appears only once in
  its row *and* column, it can never be shaded. Circling these cells
  gives you safe anchors to reason from.
- **Watch for forced shading.** When a number appears twice in a row or
  column and one copy is already circled, the other copy must be shaded.
- **Protect connectivity.** Before shading a cell, check whether it
  would split the white region into disconnected pieces. Cells that act
  as bridges between unshaded areas should almost always stay white.

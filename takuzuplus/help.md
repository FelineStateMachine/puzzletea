# Takuzu+

Fill the grid with two symbols so that the normal Takuzu rules hold, while also obeying fixed relation clues.

## Rules

- **No triples**: Three consecutive identical symbols in a row or column are not allowed.
- **Equal count**: Each row and column must contain exactly the same number of ● and ○.
- **Unique lines**: No two rows may be identical, and no two columns may be identical.
- **Relation clues**: `=` means two adjacent cells are the same. `x` means two adjacent cells are opposite.

Pre-filled cells are locked. Relation clues are also fixed and cannot be edited.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Mouse left-click` | Move to a cell, or cycle the current editable cell |
| `z` / `0` | Place ● (filled) |
| `x` / `1` | Place ○ (open) |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Escape` | Return to main menu |

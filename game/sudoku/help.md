# Sudoku

Fill a 9x9 grid with digits so every row, column, and box is complete.

## Rules

- Each **row** must contain the digits 1 through 9 exactly once.
- Each **column** must contain the digits 1 through 9 exactly once.
- Each **3x3 box** must contain the digits 1 through 9 exactly once.
- Pre-filled cells (shown in **bold**) cannot be changed. Conflicting
  cells are highlighted in red.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `1`-`9` | Fill cell with digit |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Tips

- **Scan the most-filled groups first.** Rows, columns, or boxes that
  already have seven or eight digits filled leave very few possibilities
  for the remaining cells, making them the easiest to solve.
- **Look for naked singles.** If a cell has only one digit that doesn't
  conflict with its row, column, and box, that digit is the only option.
  Check the most constrained cells before guessing anywhere.
- **Use the conflict highlighting.** When you place a digit, conflicting
  cells turn red immediately. Catch mistakes early so they don't cascade
  across the board.

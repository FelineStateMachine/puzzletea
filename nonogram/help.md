# Nonogram

Fill cells to match row and column hints, revealing a hidden pattern.

## Rules

- Each row and column has numeric **hints** that tell you how many
  consecutive filled cells appear in that line, in order.
- **Filled** cells count toward satisfying a hint.
  **Marked** cells are visual reminders that a cell should stay empty.
- Hints turn **green** when their row or column is correctly satisfied.
- The puzzle is solved when every row and column matches its hints.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `z` | Fill cell |
| `x` | Mark cell |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Tips

- **Start with the largest hints.** If a hint is close to the row or
  column length, many cells are forced to be filled regardless of
  alignment. Solve these first to unlock easier deductions elsewhere.
- **Mark cells you know are empty.** Marks cost nothing and prevent
  mistakes later. After filling a complete group, mark the cells on
  either side so you don't accidentally extend it.
- **Look for rows or columns whose hints nearly fill the line.** When
  the sum of hints plus the required gaps between them equals (or
  nearly equals) the line length, most cells can be determined
  immediately.

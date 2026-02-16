# Word Search

Find hidden words in a letter grid.

## Rules

- Words are hidden in the grid horizontally, vertically, and diagonally --
  including **backwards** in all directions (8 total).
- Found words are highlighted in green and crossed off the word list.
- The puzzle is solved when every word has been found.

To select a word, move to its first letter and press `Enter` or `Space`,
navigate to the last letter, then press `Enter` or `Space` to confirm.
Press `Escape` to cancel a selection.

## Controls

### Keyboard

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Enter` / `Space` | Select start or confirm end |
| `Escape` | Cancel current selection |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

### Mouse

| Action | Effect |
|--------|--------|
| Left click + drag | Select a word by dragging from first to last letter |
| Left click | Set the start of a selection (release and click end to confirm) |
| Right click | Cancel current selection |

Drag along a horizontal, vertical, or diagonal line to select a word.
If the selection matches a hidden word (in either direction), it is
automatically found on release.

## Tips

- **Start with uncommon letters.** Letters like Q, Z, X, and J appear far
  less often in the grid, so they are easy to spot and narrow down where a
  word must be.
- **Study the word list first.** Knowing exactly which words you need helps
  you recognise patterns in the grid instead of scanning aimlessly.
- **Check all 8 directions.** Words can run left, right, up, down, and along
  all four diagonals -- don't forget to look backwards and diagonally.

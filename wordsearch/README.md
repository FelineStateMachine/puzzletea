# Word Search

Find hidden words in a letter grid.

![Word Search gameplay](../vhs/wordsearch.gif)

## How to Play

Words are hidden in the grid horizontally, vertically, and diagonally.
Select the first letter of a word, then navigate to the last letter and
confirm. Found words are highlighted in green and crossed off the word list.
The puzzle is solved when all words are found.

1. Move the cursor to the start of a word and press `Enter` or `Space`.
2. Navigate to the end of the word.
3. Press `Enter` or `Space` to confirm, or `Escape` to cancel.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `Enter` / `Space` | Select start or confirm end |
| `Escape` | Cancel current selection |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Words | Directions | Description |
|------|------|-------|------------|-------------|
| Easy | 10x10 | 6 | 3 | Right, Down, DownRight |
| Medium | 15x15 | 10 | 6 | + DownLeft, Left, Up |
| Hard | 20x20 | 15 | 8 | All 8 directions |

## Quick Start

```bash
puzzletea new word-search easy-10x10
puzzletea new word-search medium-15x15
puzzletea new word-search hard-20x20
```

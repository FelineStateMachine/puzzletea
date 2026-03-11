# Spell Puzzle

Connect letters from the bank to spell valid words.

## Rules

- Every level has a fixed bank of allowed letters.
- Valid board words reveal onto the crossword layout.
- Valid non-board words count as bonus words.
- Invalid words display `not a word`.
- Repeating a found board word or counted bonus word displays `word already found`.
- The puzzle is solved when every board word has been revealed.

## Controls

### Keyboard

| Key | Action |
|-----|--------|
| `A-Z` | Type letters that exist in the current bank |
| `1` | Shuffle the visible bank order |
| `Enter` | Submit the traced word |
| `Backspace` | Delete the last traced letter |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Escape` | Return to main menu |

### Mouse

| Action | Effect |
|--------|--------|
| Left click and drag across the bank | Build and submit a word on release |

## Tips

- Longer words often unlock multiple crossings at once.
- Bonus words are worth finding, but they do not solve the board for you.
- Reused letters must come from distinct tiles in the bank.

# Spell Puzzle

Connect letters from a fixed bank to reveal board words in a crossword layout.
Valid extra words are counted as bonus words.

## How to Play

Build a word from the allowed letters. If the word is on the board, it reveals
in the crossword. If it is a valid TWL06 word but not on the board, it counts
as a bonus word instead.

## Controls

| Key | Action |
|-----|--------|
| `A-Z` | Type letters from the active bank |
| `1` | Shuffle the visible bank order |
| `Enter` | Submit the active word |
| `Backspace` | Delete the last letter |
| Mouse drag | Build and submit a word from the bank |

## Modes

| Mode | Bank Letters | Board Words | Bonus Minimum |
|------|--------------|-------------|---------------|
| Beginner | 6 | 4 | 3 |
| Easy | 7 | 6 | 4 |
| Medium | 8 | 8 | 6 |
| Hard | 9 | 9 | 8 |

## Elo Difficulty

Named modes are compatibility presets. Use `--difficulty <0..3000>` to target a
specific Elo difficulty for generated puzzles.

## Quick Start

```bash
puzzletea new spell-puzzle beginner
puzzletea new spell easy
```

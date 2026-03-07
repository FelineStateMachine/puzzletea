# Takuzu+

Fill the grid with two symbols while obeying the normal Takuzu rules plus fixed relation clues.

![Takuzu+ gameplay](../vhs/takuzu.gif)

## How to Play

Fill every empty cell with either a filled circle (●) or an open circle (○)
so that the grid satisfies the standard Takuzu constraints and the relation
clues placed between neighboring cells.

1. **No triples**: No three consecutive identical symbols in any row or column.
2. **Equal count**: Each row and column must contain exactly half ● and half ○.
3. **Unique lines**: No two rows may be identical; no two columns may be identical.
4. **Relation clues**: `=` means the two cells must match. `x` means they must differ.

Pre-filled cells and relation clues are fixed. The puzzle is solved when every
cell is filled and all four rules are satisfied.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| Mouse left-click | Move to a cell, or cycle the current editable cell |
| `z` / `0` | Place ● |
| `x` / `1` | Place ○ |
| `Backspace` | Clear cell |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Escape` | Return to main menu |

## Modes

| Mode | Grid | Clues | Notes |
|------|------|-------|-------|
| Beginner | 6x6 | ~50% | Relation clues heavily support early deduction |
| Easy | 6x6 | ~40% | Balanced Takuzu and relation clue logic |
| Medium | 8x8 | ~40% | Mixed deductions across rows, columns, and clues |
| Tricky | 10x10 | ~38% | Uniqueness and clue interactions matter |
| Hard | 10x10 | ~32% | Longer deduction chains |
| Very Hard | 12x12 | ~30% | Sparse givens with tighter clue reading |
| Extreme | 14x14 | ~28% | Largest board and deepest logic |

## Quick Start

```bash
puzzletea new takuzu-plus beginner
puzzletea new takuzu-plus medium
puzzletea new takuzu-plus extreme
puzzletea new binario+ hard       # alias
```

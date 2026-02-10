# Takuzu

Fill the grid with two symbols following three simple rules.

![Takuzu gameplay](../vhs/takuzu.gif)

## How to Play

Fill every empty cell with either a filled circle (●) or an open circle (○)
so that the grid satisfies three rules:

1. **No triple**: No three consecutive identical symbols in any row or column.
2. **Equal count**: Each row and column must contain exactly N/2 of each symbol.
3. **Unique lines**: No two rows may be identical; no two columns may be identical.

Pre-filled cells (shown in bold) cannot be changed. The puzzle is solved when
every cell is filled and all three rules are satisfied.

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| `z` / `0` | Place ● |
| `x` / `1` | Place ○ |
| `Backspace` | Clear cell |
| `Ctrl+H` | Toggle full help |
| `Ctrl+E` | Toggle debug overlay |
| `Ctrl+N` | Return to main menu |

## Modes

| Mode | Grid | Clues | Techniques Required |
|------|------|-------|---------------------|
| Beginner | 6x6 | ~50% | Doubles and sandwich patterns |
| Easy | 6x6 | ~40% | Counting required |
| Medium | 8x8 | ~40% | Moderate deduction |
| Tricky | 10x10 | ~38% | Uniqueness rule needed |
| Hard | 10x10 | ~32% | Long deduction chains |
| Very Hard | 12x12 | ~30% | Deep logic required |
| Extreme | 14x14 | ~28% | Maximum challenge |

## Quick Start

```bash
puzzletea new takuzu beginner
puzzletea new takuzu medium
puzzletea new takuzu extreme
puzzletea new binairo hard       # alias
```

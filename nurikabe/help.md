# Nurikabe

Shade cells as sea so every numbered island has the correct size.

## Rules

- The numbered cells are island clues and are always island cells.
- Each island must contain exactly one clue.
- The size of each island must match its clue value.
- Sea cells must form one orthogonally connected region.
- A 2x2 block of sea cells is not allowed.

The puzzle is solved when every sea cell is marked and all rules hold. Any
remaining undecided non-sea cells count as island.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `x` | Set sea |
| `z` | Set island |
| `Backspace` | Clear to unknown |
| `Ctrl+H` | Toggle full help |
| `Ctrl+R` | Reset puzzle |
| `Escape` | Return to main menu |

## Mouse

| Action | Effect |
|--------|--------|
| Left click | Move cursor, or set sea on the current cell |
| Left drag | Paint sea from the current cell |
| Right click / drag | Set island cells |

## Tips

- Start with clue `1` islands; all neighbors must be sea.
- Avoid 2x2 sea early, it removes many wrong branches.
- Track unfinished islands and reserve room to reach clue size.

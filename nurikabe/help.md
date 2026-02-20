# Nurikabe

Shade cells as sea so every numbered island has the correct size.

## Rules

- The numbered cells are island clues and are always island cells.
- Each island must contain exactly one clue.
- The size of each island must match its clue value.
- Sea cells must form one orthogonally connected region.
- A 2x2 block of sea cells is not allowed.

The puzzle is solved when all cells are decided and all rules hold.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `x` | Set sea |
| `z` | Set island |
| `Backspace` | Clear to unknown |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Mouse

| Action | Effect |
|--------|--------|
| Left click / drag | Toggle sea/unknown from click target, then paint that target while dragging |
| Right click / drag | Set island cells |

## Tips

- Start with clue `1` islands; all neighbors must be sea.
- Avoid 2x2 sea early, it removes many wrong branches.
- Track unfinished islands and reserve room to reach clue size.

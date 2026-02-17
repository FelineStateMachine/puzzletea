# Shikaku

Divide the grid into rectangles so that each rectangle contains exactly one number equal to its area.

## Rules

- The grid must be completely covered by non-overlapping rectangles.
- Each rectangle must contain **exactly one** numbered clue.
- The number on the clue must equal the **area** of its rectangle.

The puzzle is solved when all three rules hold simultaneously.

## Controls

### Keyboard — Navigation Mode

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Enter` / `Space` | Select clue to expand |
| `Backspace` | Delete rectangle at cursor |

### Keyboard — Expansion Mode

After selecting a clue, you enter Expansion Mode. Each side of the
rectangle is controlled independently.

| Key | Action |
|-----|--------|
| `Arrow` / `wasd` / `hjkl` | Expand rectangle in that direction |
| `Shift+Arrow` / `WASD` / `HJKL` | Shrink rectangle from that side |
| `Enter` / `Space` | Confirm placement (if valid) |
| `Escape` | Cancel, discard preview |
| `Backspace` | Delete existing rectangle, return to nav |

### Mouse

| Action | Effect |
|--------|--------|
| Left click + drag | Draw a rectangle from any cell |
| Release | Place the rectangle if it contains exactly one clue, the area matches, and there are no overlaps |
| Right click | Delete rectangle at cursor |

You can start a drag from **any cell**, not just a numbered clue.
While dragging, the preview turns **green** when valid or **red**
otherwise. If the rectangle is invalid on release, you stay in
Expansion Mode so you can fine-tune with the keyboard.

---

| Key | Action |
|-----|--------|
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+N` | Return to main menu |

## Tips

- **Start with 1s and 2s.** Small clues have very few possible rectangles, making them easy to place first.
- **Use the edges.** Clues near corners and borders have fewer valid rectangle configurations, so they constrain the puzzle early.
- **Watch for forced placements.** If a clue has only one possible rectangle that avoids overlapping placed rectangles, it must go there.

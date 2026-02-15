# Lights Out

Toggle lights in a cross pattern to turn them all off.

## Rules

- The grid starts with some lights **on** and some **off**.
- Toggling a cell flips that cell *and* its four orthogonal neighbors
  (up, down, left, right).
- Cells on edges and corners have fewer neighbors, so fewer lights flip.
- The goal is to turn **every light off**.

## Controls

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Move cursor |
| `Enter` / `Space` | Toggle light |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+R` | Reset puzzle |
| `Ctrl+N` | Return to main menu |

## Tips

- **Toggling the same cell twice cancels out.** Every move is its own
  undo, so if you make a mistake, just press the same cell again.
- **Work row by row from top to bottom.** Turn off all the lights in the
  current row by toggling cells in the row below it, then move on. This
  keeps solved rows from getting scrambled.
- **Plan around edges and corners.** Corner cells only affect two
  neighbors and edge cells only affect three, which limits your options.
  Tackle these constrained areas first when you can.

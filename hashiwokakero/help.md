# Hashiwokakero

Connect numbered islands with bridges to form a single connected group.

## Rules

- Each island displays a number indicating exactly how many bridges must connect to it.
- Bridges run **horizontally** or **vertically** between adjacent islands.
- Each pair of islands can be connected by a **single** or **double** bridge, but bridges may never cross.
- The puzzle is solved when every island has the correct number of bridges and all islands form one connected group.

## Controls

### Navigation Mode

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Jump to nearest island |
| `Enter` / `Space` | Select island |

### Bridge Mode

After selecting an island, you enter Bridge Mode.

| Key | Action |
|-----|--------|
| `Arrows` / `wasd` / `hjkl` | Build or cycle bridge in direction |
| `Enter` / `Space` / `Escape` | Deselect island |

Pressing a direction once builds a single bridge, twice upgrades to a
double bridge, and a third time removes it.

---

| Key | Action |
|-----|--------|
| `Ctrl+R` | Reset puzzle |
| `Ctrl+H` | Toggle help bar |
| `Ctrl+N` | Return to main menu |

## Tips

- **Start with extremes.** Islands showing their maximum value (like an `8` in the interior or a `4` in a corner) have only one valid configuration -- fill those in first.
- **Look for forced connections.** An island with only one neighboring island in each required direction has no choice; those bridges are guaranteed.
- **Work from the edges inward.** Corner and border islands have fewer possible neighbors, making them easier to solve early and giving you constraints to propagate.

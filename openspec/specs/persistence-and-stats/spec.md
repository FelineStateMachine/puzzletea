## Purpose
Specify saved game persistence, status, resume behavior, and progression statistics.

## Requirements

### Requirement: SQLite Save Store
PuzzleTea MUST persist game records, save payloads, run metadata, difficulty metadata, and statuses in SQLite.

#### Scenario: Create record
- **WHEN** a new puzzle is spawned
- **THEN** PuzzleTea stores the serialized game state with game type, display title, status, and run metadata

### Requirement: Resume And Abandonment
Saved records SHALL be importable through the owning game adapter and resumable by name, including deterministic abandoned records when appropriate.

#### Scenario: Resume abandoned deterministic puzzle
- **WHEN** a deterministic save is found for a seed-derived name
- **THEN** PuzzleTea reactivates the record instead of creating another puzzle

### Requirement: Stats And XP
Stats MUST aggregate profile level, category XP, victory counts, daily streak state, weekly progress, and difficulty-derived XP bonuses.

#### Scenario: Complete Elo puzzle
- **WHEN** an Elo-generated puzzle is completed
- **THEN** PuzzleTea applies XP using recorded actual Elo when available, otherwise target Elo or legacy mode weighting

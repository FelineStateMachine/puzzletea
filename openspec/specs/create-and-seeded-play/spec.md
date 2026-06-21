## Purpose
Describe the interactive Create workflow, variant selection, Elo targeting, and deterministic seed behavior.

## Requirements

### Requirement: Create Uses Variants And Elo
The Create screen SHALL present selectable variant leaves, require a valid target Elo from 0 through 3000, and spawn the selected variant using Elo generation.

#### Scenario: Generate from one checked variant
- **WHEN** exactly one variant leaf is checked and the user generates
- **THEN** PuzzleTea uses that variant, the entered Elo, and the optional seed if supplied

#### Scenario: Generate from multiple checked variants
- **WHEN** multiple variant leaves are checked
- **THEN** PuzzleTea chooses one checked leaf uniformly and disables deterministic seed entry for that intentionally random selection

### Requirement: Seeded Create Resume
The Create workflow MUST resume an existing deterministic save when the same seed, game, variant, and Elo identify a prior record.

#### Scenario: Reuse create seed
- **WHEN** a single selected variant has a nonblank seed matching an existing deterministic record
- **THEN** PuzzleTea imports and resumes that record instead of creating a duplicate

### Requirement: Global Seed Selection
The `--set-seed` flow SHALL deterministically select game, legacy seeded mode, name, and puzzle content from a seed string.

#### Scenario: Launch by global seed
- **WHEN** the user runs `puzzletea new --set-seed <seed>`
- **THEN** PuzzleTea rejects explicit game arguments and resumes or creates the deterministic puzzle for that seed

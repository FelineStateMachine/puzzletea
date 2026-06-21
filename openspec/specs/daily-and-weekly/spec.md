## Purpose
Describe deterministic daily puzzles and weekly gauntlet behavior.

## Requirements

### Requirement: Daily Puzzle Determinism
Daily puzzles SHALL be generated from the date and registered daily-capable seeded spawners so the same date produces the same puzzle selection.

#### Scenario: Open today's daily
- **WHEN** the user starts the daily puzzle for a date
- **THEN** PuzzleTea creates or resumes the record named for that deterministic daily puzzle

### Requirement: Daily Streak Progression
Daily completions MUST update streak and XP metadata according to saved completion history.

#### Scenario: Complete consecutive dailies
- **WHEN** the player completes daily puzzles on consecutive eligible dates
- **THEN** PuzzleTea records a continuing daily streak

### Requirement: Weekly Gauntlet
Weekly gauntlets SHALL expose a deterministic 99-puzzle ISO-week ladder, unlock current-week slots sequentially, and allow completed historical slots to be reviewed.

#### Scenario: Current week unlock
- **WHEN** the user completes the active current-week puzzle
- **THEN** the next weekly slot becomes available up to slot 99

#### Scenario: Past week review
- **WHEN** the user opens a past week
- **THEN** PuzzleTea permits review of completed saves without unlocking new historical puzzles

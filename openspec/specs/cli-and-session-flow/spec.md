## Purpose
Describe PuzzleTea command entry points and session lifecycle behavior.

## Requirements

### Requirement: Root Command And Direct Launch
The CLI SHALL load active configuration, resolve root-level aliases, and either launch the interactive menu or dispatch to the requested command.

#### Scenario: Launch menu without arguments
- **WHEN** the user runs `puzzletea`
- **THEN** PuzzleTea starts the interactive Bubble Tea application using the configured store and theme

#### Scenario: Use root aliases
- **WHEN** the user supplies root aliases such as `--new` or `--continue`
- **THEN** PuzzleTea routes to the equivalent command flow

### Requirement: New Continue List And Test Commands
The CLI MUST expose `new`, `continue`, `list`, and `test` command flows for creating games, resuming named saves, listing saves, and running smokeable game startup checks.

#### Scenario: Resume by name
- **WHEN** the user runs `puzzletea continue <name>`
- **THEN** PuzzleTea imports the saved game record and opens the session with completion state preserved

#### Scenario: List saved games
- **WHEN** the user runs `puzzletea list`
- **THEN** PuzzleTea lists resumable saved games and includes abandoned games only when requested

### Requirement: Session Lifecycle
Interactive sessions SHALL autosave game state to SQLite, update status on completion or abandonment, and return to the configured application state after spawn or import failures.

#### Scenario: Complete a puzzle
- **WHEN** a game reports solved during an active session
- **THEN** PuzzleTea records completion, preserves the save payload, and applies progression metadata

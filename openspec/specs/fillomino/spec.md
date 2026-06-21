## Purpose
Document the shipped Fillomino built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Fillomino SHALL be registered as a built-in game with canonical description "Grow the numbered regions to their exact sizes.", aliases `polyomino, regions`, variants `Fillomino (default 1300 Elo)`, and daily legacy modes `easy 6x6, medium 8x8, hard 10x10`.

#### Scenario: Resolve Fillomino
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Fillomino entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Fillomino MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Fillomino by Elo
- **WHEN** the user selects the Fillomino variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Fillomino puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Fillomino SHALL preserve legacy named modes as compatibility presets that map to the Fillomino variant and preset Elo values: Mini 5x5 (300 Elo), Easy 6x6 (700 Elo), Medium 8x8 (1300 Elo), Hard 10x10 (2000 Elo), Expert 12x12 (2600 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Fillomino mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Fillomino variant and the registered preset Elo

### Requirement: Puzzle Rules
The Fillomino runtime MUST let players grow numbered regions so every connected region has area equal to its number, with matching adjacent regions merged and contradictory region sizes rejected by solved-state checks.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Fillomino reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Fillomino SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard cursor entry is required, and mouse click-to-focus support MUST be available.

#### Scenario: Resume saved Fillomino
- **WHEN** a saved Fillomino payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Fillomino print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Fillomino
- **WHEN** Fillomino is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

## Purpose
Document the shipped Hashiwokakero built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Hashiwokakero SHALL be registered as a built-in game with canonical description "Connect the islands with bridges.", aliases `hashi, bridges`, variants `Hashiwokakero (default 1300 Elo)`, and daily legacy modes `easy 9x9, medium 7x7`.

#### Scenario: Resolve Hashiwokakero
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Hashiwokakero entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Hashiwokakero MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Hashiwokakero by Elo
- **WHEN** the user selects the Hashiwokakero variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Hashiwokakero puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Hashiwokakero SHALL preserve legacy named modes as compatibility presets that map to the Hashiwokakero variant and preset Elo values: Easy 7x7 (250 Elo), Medium 7x7 (500 Elo), Hard 7x7 (800 Elo), Easy 9x9 (700 Elo), Medium 9x9 (1100 Elo), Hard 9x9 (1500 Elo), Easy 11x11 (1300 Elo), Medium 11x11 (1700 Elo), Hard 11x11 (2100 Elo), Easy 13x13 (1900 Elo), Medium 13x13 (2400 Elo), Hard 13x13 (2900 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Hashiwokakero mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Hashiwokakero variant and the registered preset Elo

### Requirement: Puzzle Rules
The Hashiwokakero runtime MUST let players connect islands with horizontal or vertical bridges so each island reaches its required bridge count, bridges never cross, at most two bridges connect a pair, and the final bridge graph is connected.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Hashiwokakero reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Hashiwokakero SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard bridge editing is required, and mouse click-to-focus support MUST be available.

#### Scenario: Resume saved Hashiwokakero
- **WHEN** a saved Hashiwokakero payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Hashiwokakero print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Hashiwokakero
- **WHEN** Hashiwokakero is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

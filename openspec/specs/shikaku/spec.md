## Purpose
Document the shipped Shikaku built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Shikaku SHALL be registered as a built-in game with canonical description "Divide the grid into rectangles with set sizes.", aliases `rectangles`, variants `Shikaku (default 1400 Elo)`, and daily legacy modes `easy 7x7, medium 8x8`.

#### Scenario: Resolve Shikaku
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Shikaku entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Shikaku MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Shikaku by Elo
- **WHEN** the user selects the Shikaku variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Shikaku puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Shikaku SHALL preserve legacy named modes as compatibility presets that map to the Shikaku variant and preset Elo values: Mini 5x5 (300 Elo), Easy 7x7 (800 Elo), Medium 8x8 (1400 Elo), Hard 10x10 (2100 Elo), Expert 12x12 (2700 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Shikaku mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Shikaku variant and the registered preset Elo

### Requirement: Puzzle Rules
The Shikaku runtime MUST let players partition the grid into rectangles, each rectangle containing exactly one clue and having area equal to that clue.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Shikaku reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Shikaku SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard rectangle editing is required, and mouse drag support MUST be available.

#### Scenario: Resume saved Shikaku
- **WHEN** a saved Shikaku payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Shikaku print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Shikaku
- **WHEN** Shikaku is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

## Purpose
Document the shipped Ripple Effect built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Ripple Effect SHALL be registered as a built-in game with canonical description "Fill the cages with sequential numbers without violating ripple distance.", aliases `ripple`, variants `Ripple Effect (default 1500 Elo)`, and daily legacy modes `easy 6x6, medium 7x7, hard 8x8`.

#### Scenario: Resolve Ripple Effect
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Ripple Effect entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Ripple Effect MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Ripple Effect by Elo
- **WHEN** the user selects the Ripple Effect variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Ripple Effect puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Ripple Effect SHALL preserve legacy named modes as compatibility presets that map to the Ripple Effect variant and preset Elo values: Mini 5x5 (400 Elo), Easy 6x6 (900 Elo), Medium 7x7 (1500 Elo), Hard 8x8 (2200 Elo), Expert 9x9 (2800 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Ripple Effect mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Ripple Effect variant and the registered preset Elo

### Requirement: Puzzle Rules
The Ripple Effect runtime MUST let players place digits in rooms so each room contains each digit once and matching digits in any row or column are separated by more cells than the digit value.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Ripple Effect reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Ripple Effect SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard digit entry is required for rooms and cursor movement.

#### Scenario: Resume saved Ripple Effect
- **WHEN** a saved Ripple Effect payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Ripple Effect print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Ripple Effect
- **WHEN** Ripple Effect is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

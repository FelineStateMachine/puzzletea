## Purpose
Document the shipped Hitori built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Hitori SHALL be registered as a built-in game with canonical description "Shade the cells to eliminate duplicates.", aliases `none`, variants `Hitori (default 1800 Elo)`, and daily legacy modes `easy, medium`.

#### Scenario: Resolve Hitori
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Hitori entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Hitori MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Hitori by Elo
- **WHEN** the user selects the Hitori variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Hitori puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Hitori SHALL preserve legacy named modes as compatibility presets that map to the Hitori variant and preset Elo values: Mini (300 Elo), Easy (700 Elo), Medium (1300 Elo), Tricky (1800 Elo), Hard (2300 Elo), Expert (2800 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Hitori mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Hitori variant and the registered preset Elo

### Requirement: Puzzle Rules
The Hitori runtime MUST let players shade cells so each row and column has no duplicate unshaded numbers, shaded cells are never orthogonally adjacent, and all unshaded cells remain connected.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Hitori reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Hitori SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard shading controls are required, and mouse click-to-focus support MUST be available.

#### Scenario: Resume saved Hitori
- **WHEN** a saved Hitori payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Hitori print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Hitori
- **WHEN** Hitori is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

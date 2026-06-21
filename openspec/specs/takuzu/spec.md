## Purpose
Document the shipped Takuzu built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Takuzu SHALL be registered as a built-in game with canonical description "Fill the grid with filled and empty. No 3 in a row.", aliases `binairo, binary`, variants `Takuzu (default 1600 Elo)`, and daily legacy modes `medium, tricky`.

#### Scenario: Resolve Takuzu
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Takuzu entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Takuzu MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Takuzu by Elo
- **WHEN** the user selects the Takuzu variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Takuzu puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Takuzu SHALL preserve legacy named modes as compatibility presets that map to the Takuzu variant and preset Elo values: Beginner (300 Elo), Easy (700 Elo), Medium (1100 Elo), Tricky (1600 Elo), Hard (2100 Elo), Very Hard (2500 Elo), Extreme (2900 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Takuzu mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Takuzu variant and the registered preset Elo

### Requirement: Puzzle Rules
The Takuzu runtime MUST let players fill an even-sized grid with two symbols, avoid three identical symbols in a row or column, balance each row and column, and avoid duplicate completed rows or columns.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Takuzu reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Takuzu SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard symbol entry is required, and mouse click-to-focus support MUST be available.

#### Scenario: Resume saved Takuzu
- **WHEN** a saved Takuzu payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Takuzu print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Takuzu
- **WHEN** Takuzu is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

## Purpose
Document the shipped Nonogram built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Nonogram SHALL be registered as a built-in game with canonical description "Fill the cells to match tomographic hints.", aliases `none`, variants `Nonogram (default 1600 Elo)`, and daily legacy modes `standard, classic`.

#### Scenario: Resolve Nonogram
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Nonogram entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Nonogram MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Nonogram by Elo
- **WHEN** the user selects the Nonogram variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Nonogram puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Nonogram SHALL preserve legacy named modes as compatibility presets that map to the Nonogram variant and preset Elo values: Mini (100 Elo), Pocket (450 Elo), Teaser (800 Elo), Standard (1000 Elo), Classic (1300 Elo), Tricky (1600 Elo), Large (1900 Elo), Grand (2200 Elo), Epic (2600 Elo), Massive (2900 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Nonogram mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Nonogram variant and the registered preset Elo

### Requirement: Puzzle Rules
The Nonogram runtime MUST let players mark filled and empty cells so every row and column exactly matches its tomographic run-length hints.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Nonogram reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Nonogram SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard marking is required, and mouse drag support MUST be available.

#### Scenario: Resume saved Nonogram
- **WHEN** a saved Nonogram payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Nonogram print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Nonogram
- **WHEN** Nonogram is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

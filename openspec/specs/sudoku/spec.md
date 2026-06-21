## Purpose
Document the shipped Sudoku built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Sudoku SHALL be registered as a built-in game with canonical description "Fill the 9x9 grid following sudoku rules.", aliases `none`, variants `Sudoku (default 1800 Elo)`, and daily legacy modes `easy, medium`.

#### Scenario: Resolve Sudoku
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Sudoku entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Sudoku MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Sudoku by Elo
- **WHEN** the user selects the Sudoku variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Sudoku puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Sudoku SHALL preserve legacy named modes as compatibility presets that map to the Sudoku variant and preset Elo values: Beginner (0 Elo), Easy (600 Elo), Medium (1200 Elo), Hard (1800 Elo), Expert (2400 Elo), Diabolical (3000 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Sudoku mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Sudoku variant and the registered preset Elo

### Requirement: Puzzle Rules
The Sudoku runtime MUST let players fill a 9x9 grid so every row, column, and 3x3 box contains digits 1 through 9 exactly once while preserving givens.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Sudoku reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Sudoku SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard digit entry is required, and mouse click-to-focus support MUST be available.

#### Scenario: Resume saved Sudoku
- **WHEN** a saved Sudoku payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Sudoku print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Sudoku
- **WHEN** Sudoku is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

## Purpose
Document the shipped Sudoku RGB built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Sudoku RGB SHALL be registered as a built-in game with canonical description "Fill the board with RGB symbols so each row, column, and 3x3 box contains {1,1,1,2,2,2,3,3,3}. [1,2,3] maps to [triangle,square,filled]. Inspired by Sudoku Ripeto.", aliases `rgb sudoku, ripeto, sudoku ripeto`, variants `Sudoku RGB (default 1800 Elo)`, and daily legacy modes `easy, medium`.

#### Scenario: Resolve Sudoku RGB
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Sudoku RGB entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Sudoku RGB MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Sudoku RGB by Elo
- **WHEN** the user selects the Sudoku RGB variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Sudoku RGB puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Sudoku RGB SHALL preserve legacy named modes as compatibility presets that map to the Sudoku RGB variant and preset Elo values: Beginner (0 Elo), Easy (600 Elo), Medium (1200 Elo), Hard (1800 Elo), Expert (2400 Elo), Diabolical (3000 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Sudoku RGB mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Sudoku RGB variant and the registered preset Elo

### Requirement: Puzzle Rules
The Sudoku RGB runtime MUST let players fill a 9x9 grid with the three RGB symbols so each row, column, and 3x3 box contains exactly three of each symbol while preserving givens.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Sudoku RGB reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Sudoku RGB SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard symbol entry is required, and mouse click-to-focus support MUST be available.

#### Scenario: Resume saved Sudoku RGB
- **WHEN** a saved Sudoku RGB payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Sudoku RGB print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Sudoku RGB
- **WHEN** Sudoku RGB is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

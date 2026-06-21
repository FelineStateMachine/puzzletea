## Purpose
Document the shipped Word Search built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Word Search SHALL be registered as a built-in game with canonical description "Find the hidden words in a letter grid.", aliases `words, wordsearch, ws`, variants `Word Search (default 1500 Elo)`, and daily legacy modes `easy 10x10`.

#### Scenario: Resolve Word Search
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Word Search entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Word Search MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Word Search by Elo
- **WHEN** the user selects the Word Search variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Word Search puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Word Search SHALL preserve legacy named modes as compatibility presets that map to the Word Search variant and preset Elo values: Easy 10x10 (600 Elo), Medium 15x15 (1500 Elo), Hard 20x20 (2400 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Word Search mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Word Search variant and the registered preset Elo

### Requirement: Puzzle Rules
The Word Search runtime MUST let players select hidden dictionary words in straight lines on the grid and solve only when all listed words have been found.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Word Search reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Word Search SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard selection is required, and mouse drag support MUST be available.

#### Scenario: Resume saved Word Search
- **WHEN** a saved Word Search payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Word Search print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Word Search
- **WHEN** Word Search is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

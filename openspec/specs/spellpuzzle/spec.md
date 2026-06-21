## Purpose
Document the shipped Spell Puzzle built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Spell Puzzle SHALL be registered as a built-in game with canonical description "Connect letters to fill a crossword with bonus anagrams.", aliases `spell, spellpuzzle`, variants `Spell Puzzle (default 1500 Elo)`, and daily legacy modes `beginner`.

#### Scenario: Resolve Spell Puzzle
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Spell Puzzle entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Spell Puzzle MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Spell Puzzle by Elo
- **WHEN** the user selects the Spell Puzzle variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Spell Puzzle puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Spell Puzzle SHALL preserve legacy named modes as compatibility presets that map to the Spell Puzzle variant and preset Elo values: Beginner (0 Elo), Easy (600 Elo), Medium (1500 Elo), Hard (2400 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Spell Puzzle mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Spell Puzzle variant and the registered preset Elo

### Requirement: Puzzle Rules
The Spell Puzzle runtime MUST let players connect wheel letters to fill crossword board words and track valid bonus anagrams without requiring bonus words for puzzle completion.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Spell Puzzle reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Spell Puzzle SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard word entry is required, and mouse drag support MUST be available.

#### Scenario: Resume saved Spell Puzzle
- **WHEN** a saved Spell Puzzle payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Spell Puzzle print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Spell Puzzle
- **WHEN** Spell Puzzle is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

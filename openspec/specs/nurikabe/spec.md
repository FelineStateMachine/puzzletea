## Purpose
Document the shipped Nurikabe built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Nurikabe SHALL be registered as a built-in game with canonical description "Split the land while keeping one connected sea.", aliases `islands, sea`, variants `Nurikabe (default 1400 Elo)`, and daily legacy modes `easy, medium`.

#### Scenario: Resolve Nurikabe
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Nurikabe entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Nurikabe MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Nurikabe by Elo
- **WHEN** the user selects the Nurikabe variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Nurikabe puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Nurikabe SHALL preserve legacy named modes as compatibility presets that map to the Nurikabe variant and preset Elo values: Mini (300 Elo), Easy (800 Elo), Medium (1400 Elo), Hard (2100 Elo), Expert (2700 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Nurikabe mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Nurikabe variant and the registered preset Elo

### Requirement: Puzzle Rules
The Nurikabe runtime MUST let players build numbered islands with exact island sizes, keep islands separated orthogonally, avoid 2x2 sea blocks, and maintain one connected sea.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Nurikabe reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Nurikabe SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard cell editing is required, and mouse drag support MUST be available.

#### Scenario: Resume saved Nurikabe
- **WHEN** a saved Nurikabe payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Nurikabe print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Nurikabe
- **WHEN** Nurikabe is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

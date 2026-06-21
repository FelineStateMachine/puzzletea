## Purpose
Document the shipped Netwalk built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Netwalk SHALL be registered as a built-in game with canonical description "Rotate network tiles until every computer connects to the server.", aliases `network`, variants `Netwalk (default 1500 Elo)`, and daily legacy modes `easy 7x7, medium 9x9`.

#### Scenario: Resolve Netwalk
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Netwalk entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Netwalk MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Netwalk by Elo
- **WHEN** the user selects the Netwalk variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Netwalk puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Netwalk SHALL preserve legacy named modes as compatibility presets that map to the Netwalk variant and preset Elo values: Mini 5x5 (400 Elo), Easy 7x7 (900 Elo), Medium 9x9 (1500 Elo), Hard 11x11 (2200 Elo), Expert 13x13 (2800 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Netwalk mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Netwalk variant and the registered preset Elo

### Requirement: Puzzle Rules
The Netwalk runtime MUST let players rotate network tiles until every computer is connected to the server through matching pipe edges and there are no dangling required connections in the solved network.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Netwalk reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Netwalk SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard rotation is required, and mouse click-to-rotate support MUST be available.

#### Scenario: Resume saved Netwalk
- **WHEN** a saved Netwalk payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
The Netwalk print adapter SHALL render puzzle and solution data for JSONL-to-PDF workflows.

#### Scenario: Export Netwalk
- **WHEN** Netwalk is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

## Purpose
Document the shipped Lights Out built-in game behavior.

## Requirements

### Requirement: Registry Metadata
Lights Out SHALL be registered as a built-in game with canonical description "Turn the lights off.", aliases `lights`, variants `Lights Out (default 2000 Elo)`, and daily legacy modes `medium, hard`.

#### Scenario: Resolve Lights Out
- **WHEN** the registry resolves the canonical name or any alias
- **THEN** it returns the Lights Out entry with help text, import behavior, variant metadata, and legacy mode compatibility

### Requirement: Variant Elo Creation
Lights Out MUST support the current variant plus target Elo creation path, using the variant default Elo when no explicit target is supplied.

#### Scenario: Spawn Lights Out by Elo
- **WHEN** the user selects the Lights Out variant and a target Elo from 0 through 3000
- **THEN** PuzzleTea spawns a deterministic Lights Out puzzle and records target, actual, and confidence metadata when generation reports it

### Requirement: Legacy Mode Compatibility
Lights Out SHALL preserve legacy named modes as compatibility presets that map to the Lights Out variant and preset Elo values: Easy (0 Elo), Medium (1000 Elo), Hard (2000 Elo), Extreme (3000 Elo).

#### Scenario: Use legacy mode title
- **WHEN** the user supplies a legacy Lights Out mode title through CLI or deterministic daily metadata
- **THEN** PuzzleTea resolves it to the Lights Out variant and the registered preset Elo

### Requirement: Puzzle Rules
The Lights Out runtime MUST let players toggle a selected light and its orthogonal neighbors, and the puzzle is solved only when every light is off.

#### Scenario: Check solved state
- **WHEN** the player fills the board or word state
- **THEN** Lights Out reports solved only when all rule constraints are satisfied

### Requirement: Save Load And Interaction
Lights Out SHALL serialize enough state for resume/import, reset to its initial puzzle, expose debug and full-help data, and support its shipped interaction model. Keyboard toggling is required, and mouse click-to-toggle support MUST be available.

#### Scenario: Resume saved Lights Out
- **WHEN** a saved Lights Out payload is imported
- **THEN** the restored game preserves givens, player entries, cursor/selection state where applicable, and solved-state behavior

### Requirement: Print Export Support
Lights Out MUST remain excluded from printable export until a print adapter is provided.

#### Scenario: Export Lights Out
- **WHEN** Lights Out is included in a JSONL or PDF export flow
- **THEN** PuzzleTea follows the registered print support behavior for that game

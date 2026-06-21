## Purpose
Define the separation between pure catalog metadata and concrete built-in game runtime registration.

## Requirements

### Requirement: Pure Catalog Index
The catalog package SHALL index puzzle definitions, aliases, daily metadata, variants, and legacy modes without importing concrete game implementations.

#### Scenario: Build catalog
- **WHEN** definitions are passed to the catalog builder
- **THEN** duplicate canonical names, duplicate aliases, invalid Elo values, missing daily modes, and invalid legacy mode targets are rejected

### Requirement: Concrete Runtime Registry
The registry package MUST compose built-in game entries, imports, help text, print adapters, variants, and legacy mode spawners.

#### Scenario: Resolve game name
- **WHEN** a canonical game name or alias is supplied
- **THEN** the registry returns the concrete entry for that built-in game

### Requirement: Variant First Creation
Puzzle creation SHALL prefer registered variants with target Elo, while preserving legacy named modes as compatibility aliases to variant Elo presets.

#### Scenario: Resolve legacy mode
- **WHEN** a legacy mode title is supplied to CLI creation
- **THEN** PuzzleTea resolves it to the target variant and preset Elo rather than treating modes as the primary creation model

### Requirement: Daily Eligibility
The registry SHALL expose daily-capable entries from metadata and skip duplicate game IDs so daily selection has at most one concrete spawner per game.

#### Scenario: Build daily entries
- **WHEN** daily entries are requested
- **THEN** the registry returns seeded spawners tied to registered daily mode metadata

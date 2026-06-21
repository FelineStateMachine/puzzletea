## Purpose
Describe printable JSONL puzzle pack generation from `puzzletea new --export`.

## Requirements

### Requirement: Export Uses Variants
JSONL export SHALL generate puzzles from printable registered variants and target Elo, while accepting legacy mode text only as a compatibility resolver to a variant preset.

#### Scenario: Export one variant
- **WHEN** the user runs `puzzletea new <game> <selection> --export <n>`
- **THEN** PuzzleTea resolves the selection to a variant and writes `n` JSONL records

### Requirement: JSONL Record Schema
Each JSONL record MUST include pack metadata, puzzle index, generated puzzle name, game type, variant display title, optional Elo metadata, and a valid JSON save payload.

#### Scenario: Validate save payload
- **WHEN** an exported puzzle serializes its save
- **THEN** PuzzleTea rejects the export if the save payload is not valid JSON

### Requirement: Deterministic Export Seeding
Export generation SHALL use `--with-seed` to deterministically choose mixed variants, export puzzle seeds, and generated puzzle names.

#### Scenario: Repeat seeded export
- **WHEN** the same export command and seed are run again
- **THEN** PuzzleTea produces the same variant sequence and deterministic puzzle seeds

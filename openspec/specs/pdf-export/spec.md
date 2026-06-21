## Purpose
Describe rendering JSONL puzzle packs into printable PDF output.

## Requirements

### Requirement: Print Adapter Registration
The PDF export pipeline MUST use built-in print adapters registered by `builtinprint` and reject games without paper rendering support.

#### Scenario: Unsupported game
- **WHEN** a JSONL record references a game without a print adapter
- **THEN** PDF export reports the unsupported puzzle instead of silently rendering invalid output

### Requirement: PDF Layout
The PDF renderer SHALL support half-letter interiors, duplex-booklet sheet layout, cover/title pages, metadata, instructions, puzzle pages, answer pages, and automatic page-count padding.

#### Scenario: Duplex booklet
- **WHEN** `--sheet-layout duplex-booklet` is selected
- **THEN** PuzzleTea emits landscape letter sheets with two portrait half-letter pages per side and cover stock placement preserved

### Requirement: Difficulty Rendering
PDF ordering and difficulty display MUST prefer actual Elo, then target Elo, then legacy mode fallback when Elo metadata is unavailable.

#### Scenario: Render difficulty
- **WHEN** an exported record contains actual Elo
- **THEN** the PDF uses that value for ordering and difficulty display

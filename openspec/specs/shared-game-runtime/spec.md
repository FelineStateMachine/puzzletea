## Purpose
Specify shared interfaces and helper expectations for built-in game runtimes.

## Requirements

### Requirement: Gamer Interface
Every active puzzle game MUST implement initialization, update, view, reset, title setting, solved-state, save serialization, debug info, and full help bindings.

#### Scenario: Run game session
- **WHEN** the app activates a game model
- **THEN** the model can initialize, render, accept Bubble Tea messages, save state, and report completion through the shared interface

### Requirement: Spawner Interfaces
Runtime entries SHALL expose variant Elo spawners for current creation and MAY expose legacy mode, seeded, cancellable, or import behavior as needed by CLI, daily, weekly, and export flows.

#### Scenario: Spawn variant by Elo
- **WHEN** a variant is selected with target Elo
- **THEN** the owning game creates a deterministic puzzle and returns difficulty confidence metadata when available

### Requirement: Shared Helpers
Grid games SHALL use shared cursor, key, border, dynamic grid, and style helpers where applicable so movement, rendering, and save/import behavior remain consistent.

#### Scenario: Move cursor
- **WHEN** a grid game handles movement keys
- **THEN** cursor bounds and wrapping behavior follow the shared helper semantics used by that game

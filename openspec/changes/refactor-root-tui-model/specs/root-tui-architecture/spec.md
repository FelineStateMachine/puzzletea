## ADDED Requirements

### Requirement: Root update dispatch remains behavior-preserving
The root TUI model SHALL preserve existing user-facing behavior while organizing `Update` as a top-level dispatcher for app messages, global shortcuts, active game updates, screen updates, and screen actions.

#### Scenario: Existing app flows remain available
- **WHEN** a user navigates through main menu, play menu, create, continue, daily, weekly, seeded play, export, options, help, stats, theme selection, or an active puzzle
- **THEN** the app SHALL expose the same commands, navigation outcomes, persistence behavior, and view transitions as before the refactor

#### Scenario: Root update delegates by message category
- **WHEN** the root model receives async completion messages, window-size messages, key messages, screen messages, or active game messages
- **THEN** dispatch SHALL route each message to the component that owns the relevant behavior without requiring screens to mutate root model internals directly

### Requirement: Route and screen management is isolated
The app SHALL isolate route and screen mechanics from workflow behavior so current route, cached screens, screen initialization, active screen lookup, window dimensions, and active-screen resizing are managed through a dedicated app-internal boundary.

#### Scenario: Screen cache survives refactor
- **WHEN** a screen is updated, resized, left, or re-entered according to the existing app flow
- **THEN** the app SHALL preserve the same screen state retention or reinitialization behavior that existed before the refactor

#### Scenario: Route management does not own workflows
- **WHEN** a route transition requires creating, resuming, spawning, exporting, saving, applying a theme, or updating persistent state
- **THEN** the route/screen management boundary SHALL delegate that workflow behavior to the owning controller or handler

### Requirement: Screens emit typed app intents
App screens SHALL remain dependency-light interaction models that update local component state and emit typed app intents for root-level behavior.

#### Scenario: Screen action applies app behavior
- **WHEN** a screen interaction requires opening another route, starting generation, submitting export, changing theme, opening help, or loading a record
- **THEN** the screen SHALL return a typed action or command that is applied by the root app layer or its controllers

#### Scenario: Screens avoid direct app dependency ownership
- **WHEN** a screen is constructed or updated
- **THEN** it SHALL NOT require direct ownership of store, config, session lifecycle, or persistence dependencies unless an existing behavior cannot be preserved otherwise

### Requirement: Session lifecycle has a single owner
The app SHALL centralize active game lifecycle behavior behind session-oriented code.

#### Scenario: Starting a generated game
- **WHEN** normal mode selection, create flow, seeded play, daily play, or weekly play starts puzzle generation
- **THEN** session lifecycle code SHALL own generation state, cancellation, stale completion handling, game activation, save record creation, and transition into game view

#### Scenario: Resuming or reviewing a saved game
- **WHEN** the user resumes a saved game or opens a read-only weekly review
- **THEN** session lifecycle code SHALL own import, read-only handling, active game ID setup, completion-saved state, help/window-size propagation, and transition into game view

#### Scenario: Leaving an active game
- **WHEN** the user exits or quits an active game
- **THEN** session lifecycle code SHALL own save-state persistence, status updates, active game cleanup, and return-state behavior

#### Scenario: Completing an active game
- **WHEN** an active game becomes solved
- **THEN** session lifecycle code SHALL persist final save data and completed status once, preserving existing error notice behavior when persistence fails

### Requirement: Global key behavior is owned by active domain
Global keyboard behavior SHALL be routed to the app domain that owns the current state.

#### Scenario: Pending generation keys
- **WHEN** the app is showing pending puzzle generation and the user presses escape or quit
- **THEN** session lifecycle code SHALL preserve existing cancellation, return-route, and quit behavior

#### Scenario: Pending export keys
- **WHEN** the app is showing pending export and the user presses escape or quit
- **THEN** export workflow code SHALL preserve existing cancellation, return-route, and quit behavior

#### Scenario: Active game keys
- **WHEN** the app is showing an active puzzle and the user presses escape, quit, reset, full-help, debug, or enter on a solved weekly puzzle
- **THEN** active game/session handling SHALL preserve existing save, abandon, reset, help propagation, debug, and weekly advancement behavior

#### Scenario: Normal screen keys
- **WHEN** the app is showing a non-game, non-pending screen and the user presses global quit, debug, or full-help keys
- **THEN** shell-level handling SHALL preserve existing quit, debug toggle, and full-help toggle behavior

## Why

The root Bubble Tea model has accumulated several responsibilities: route dispatch, screen caching, global key handling, active game lifecycle, spawn/export cancellation, persistence, and workflow-specific setup for create/daily/seeded/weekly play. This makes changes to TUI flow harder to reason about and raises the risk of regressions when adding new screens or session behavior.

## What Changes

- Reorganize the root TUI model so `Update` remains a thin dispatcher rather than the owner of every workflow.
- Clarify ownership boundaries for route/screen management, global shell state, active game/session lifecycle, and workflow-specific orchestration.
- Consolidate active game opening, spawning, resuming, saving, and completion persistence behind session-oriented APIs.
- Preserve existing user-facing behavior for menus, create flow, seeded play, daily/weekly play, export, help, debug, theme selection, persistence, and active puzzle controls.
- Add focused tests around the refactored routing/session behavior where current coverage depends on root model internals.

## Capabilities

### New Capabilities
- `root-tui-architecture`: Internal architecture requirements for the root Bubble Tea model, screen routing, workflow ownership, and behavior-preserving refactors.

### Modified Capabilities

None. This change is intended to preserve existing spec-level behavior.

## Impact

- Affected code: `app/model.go`, `app/update.go`, `app/view.go`, `app/screen_core.go`, `app/session_controller.go`, and workflow files such as `app/handle_create.go`, `app/handle_seed.go`, `app/handle_daily.go`, `app/weekly.go`, and `app/export.go`.
- Affected tests: focused `app` package tests for update routing, screen actions, session lifecycle, spawn cancellation/completion, save-on-exit, and weekly advancement.
- APIs: no public CLI, game registry, persistence schema, or game package API changes are planned.
- Dependencies: no new runtime dependencies are planned.

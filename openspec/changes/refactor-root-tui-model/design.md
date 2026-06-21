## Context

The `app` package contains the root Bubble Tea model for PuzzleTea. It already uses a useful `screenModel` and `screenAction` pattern: menu-like screens update local component state and return actions that the root model applies. However, the root model still owns several distinct concerns at once:

- The current route and cached screen instances.
- Window sizing and active screen resizing.
- Global key behavior for normal screens, active puzzles, pending generation, and pending export.
- Active puzzle lifecycle, including spawn start/completion, save/load, completion persistence, and read-only weekly review.
- Workflow setup for create, normal mode selection, daily, seeded, weekly, export, help, stats, and theme flows.
- Shell state such as notices, debug rendering, theme/help toggles, store/config dependencies, and the shared spinner.

The refactor should keep the public Bubble Tea surface stable while making those ownership boundaries explicit enough that future UI flows can be added without expanding `model.Update` or scattering session persistence rules.

## Goals / Non-Goals

**Goals:**

- Keep `model.Update` focused on top-level dispatch: async messages, global shell keys, active game delegation, screen updates, and action application.
- Make route/screen management responsible for current route, cached screen models, window sizing, active screen lookup, initialization, and resize behavior.
- Make session lifecycle code the single owner of game spawning, spawn cancellation, spawn completion, saved-record import, active game update, completion persistence, and save-on-exit.
- Keep screen models dependency-light: screens should continue to emit typed actions instead of directly reaching into the store, config, session controller, or root model internals.
- Preserve current behavior for all app flows.
- Add focused tests around routing/session behavior that could regress during the refactor.

**Non-Goals:**

- No changes to game package APIs, `game.Gamer`, registry/catalog behavior, persistence schema, or CLI flags.
- No visual redesign of menus, screens, exports, or puzzle views.
- No replacement of Bubble Tea or Bubbles components.
- No broad package split unless a smaller `app`-internal type cannot express the boundary cleanly.

## Decisions

### Keep One Bubble Tea Root Model

The app should continue exposing a single root Bubble Tea model. Internally, the root can delegate to smaller app-owned types, but it should remain the integration point for Bubble Tea messages and views.

Alternative considered: make every app screen a full independent Bubble Tea model with store/config/session dependencies. That would reduce the root's visible size, but it would scatter persistence and lifecycle behavior across many screens. The current `screenAction` pattern is better: screens own component interaction, while app-level behavior stays centralized.

### Introduce an App-Internal Route/Screen Manager

Route and screen mechanics should be isolated from workflow behavior. A route/screen manager can own:

- Current `viewState`.
- Cached `map[viewState]screenModel`.
- Active screen lookup.
- Screen initialization through the existing registry.
- Window dimensions and active-screen resize.
- Route transitions that only change display state.

The manager should not know how to create puzzles, save games, export packs, or apply theme changes. It should receive enough shell state to construct screens, then return updated screen state to the root.

Alternative considered: leave these methods on `model` but move them into more files. That would make files shorter without changing ownership, so it would not materially improve reasoning.

### Keep Screens as Intent Emitters

Screens should continue returning `screenAction` values from `Update`. Actions should remain typed app intents, not generic strings. This keeps screen tests focused on interaction and allows root/controller tests to cover app behavior.

Where actions currently perform substantial workflow setup, the action should become a small adapter into the appropriate controller. For example, mode selection should delegate session start details instead of constructing all spawn metadata inline.

Alternative considered: have screens call controller methods directly. That would make individual actions smaller but would require injecting app dependencies into screens and would weaken the current separation.

### Make Session Lifecycle the Owner of Active Game State

The session controller should own:

- Starting normal, Elo, and seeded generation.
- Cancelling pending generation.
- Ignoring stale spawn completions.
- Activating generated or imported games.
- Applying help/window-size messages to newly activated games.
- Updating active games.
- Persisting completed puzzles.
- Saving or clearing games on escape/quit.
- Handling read-only game opens.

Workflow handlers may decide *which* puzzle to open or generate, but once that decision is made they should enter session code through higher-level methods rather than duplicating spawn request construction and save/resume rules.

Alternative considered: keep workflow-specific spawn setup in handler files. That keeps local context near each flow, but the duplicated pattern of "check record, maybe resume, otherwise spawn, then persist" makes session behavior harder to audit.

### Split Global Key Handling by Route Ownership

Global key handling should route to the domain that owns the current state:

- Pending generation: session controller handles escape/quit behavior and spawn cancellation.
- Pending export: export controller/handler handles escape/quit behavior and export cancellation.
- Active game: session controller handles save-on-escape, abandon-on-quit, reset, help toggle propagation, debug update, and weekly advancement entry points.
- Normal screens: shell/global keys handle quit, debug toggle, and full-help toggle.

This keeps `handleGlobalKey` from accumulating per-route workflow details.

Alternative considered: keep a single root `handleGlobalKey` switch. It is straightforward today, but it is already mixing shell, export, session, and weekly behavior.

### Preserve Behavior Before Improving Shape Further

The refactor should be incremental and behavior-preserving. It is acceptable to leave some workflow-specific logic on the root temporarily if moving it would enlarge the diff or weaken tests. The important outcome is clearer ownership for the root update loop and session lifecycle.

## Risks / Trade-offs

- **Risk: Moving state between structs can accidentally drop screen component state.** → Keep the existing screen cache behavior and add tests around returning to screens after resize/navigation where practical.
- **Risk: Session controller methods can become another large object.** → Limit it to active game/spawn/save/load lifecycle. Keep create/daily/seed/weekly selection logic outside unless it directly affects session invariants.
- **Risk: Behavior-preserving refactors can hide regressions in keyboard handling.** → Add focused tests for escape/quit/reset/help/debug paths in generating, export-running, active-game, and normal-screen states.
- **Risk: Export remains large and may not fit the same controller shape immediately.** → Treat export as a separate workflow boundary and avoid mixing it into session lifecycle.
- **Risk: Over-abstracting routes could obscure simple app flows.** → Prefer small concrete types and typed methods over generic event buses or reflection-based routing.

## Migration Plan

1. Add tests that pin the current high-risk behavior before moving ownership.
2. Extract route/screen management with no behavior changes.
3. Move pending generation and active game key behavior into session-oriented methods.
4. Move duplicated spawn/open/resume behavior behind session-oriented entry points where the call sites become clearer.
5. Re-run `just fmt`, `just lint`, and `just test-short`.

Rollback is straightforward because there are no data migrations or public API changes: revert the refactor commits if behavior diverges.

## Open Questions

- Should create/daily/seed/weekly decision logic live in one deterministic-play workflow helper, or remain in their current files while calling improved session APIs?
- Should export get a small controller in this change, or only enough cleanup to remove export-running special cases from root global key handling?

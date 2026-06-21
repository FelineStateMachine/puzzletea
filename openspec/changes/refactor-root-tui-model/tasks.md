## 1. Baseline Behavior Coverage

- [ ] 1.1 Add or update focused tests for root `Update` dispatch of async spawn/export completion, window resize, active game messages, screen messages, and screen actions.
- [ ] 1.2 Add or update tests for global key behavior in normal screens, pending generation, pending export, and active game states.
- [ ] 1.3 Add or update tests for session lifecycle behavior: spawn cancellation, stale spawn completion, generated game activation, saved-record import, read-only review, save-on-exit, and solved-game completion persistence.
- [ ] 1.4 Add or update tests that pin existing screen cache and resize behavior across representative route transitions.

## 2. Route and Screen Boundary

- [ ] 2.1 Introduce an app-internal route/screen management boundary for current `viewState`, cached screens, active screen lookup, screen initialization, window dimensions, and resize behavior.
- [ ] 2.2 Move `activeScreen`, `initScreen`, `resizeActiveScreen`, and window-size handling behind the route/screen boundary without changing behavior.
- [ ] 2.3 Update root model initialization to construct and use the route/screen boundary while preserving initial main-menu and initial-game behavior.
- [ ] 2.4 Keep `screenModel` and `screenAction` semantics intact and verify existing screen tests still exercise typed app intents.

## 3. Session Lifecycle Boundary

- [ ] 3.1 Expand session-oriented methods so generated game start, pending generation state, cancellation, stale completion handling, and activation are owned by session lifecycle code.
- [ ] 3.2 Move saved-record import, read-only review setup, active game ID setup, help/window-size propagation, and completion-saved state behind session lifecycle code.
- [ ] 3.3 Move save-on-exit, abandon-on-quit, active game cleanup, and solved-game completion persistence behind session lifecycle code.
- [ ] 3.4 Update create, mode-select, seed, daily, and weekly call sites to use the clarified session lifecycle entry points where doing so reduces duplicated spawn/open/resume rules.

## 4. Global Key and Workflow Dispatch

- [ ] 4.1 Refactor root global key handling so pending generation keys delegate to session lifecycle behavior.
- [ ] 4.2 Refactor root global key handling so pending export keys delegate to export workflow behavior.
- [ ] 4.3 Refactor active-game key behavior so save, quit, reset, help propagation, debug updates, and weekly advancement route through the owning session/workflow code.
- [ ] 4.4 Keep normal-screen shell keys for quit, debug toggle, and full-help toggle behavior-preserving and easy to audit.

## 5. Root Model Cleanup

- [ ] 5.1 Reduce `model.Update` to a clear top-level dispatch flow: app-level messages, global keys, active game delegation, screen update, action application.
- [ ] 5.2 Move state type definitions or helper methods into files that match their ownership boundary, avoiding file-only churn that does not clarify responsibilities.
- [ ] 5.3 Remove obsolete wrappers or duplicated helper paths created during the refactor.
- [ ] 5.4 Review `app/model.go`, `app/update.go`, `app/screen_core.go`, and `app/session_controller.go` for remaining responsibility leaks and document any intentionally deferred cleanup.

## 6. Validation

- [ ] 6.1 Run `just fmt`.
- [ ] 6.2 Run `just lint`.
- [ ] 6.3 Run `just test-short`.
- [ ] 6.4 Manually review the change against `root-tui-architecture` requirements and confirm there are no user-facing behavior changes.

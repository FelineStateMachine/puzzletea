# Graph Analysis: refactor-root-tui-model

> Generated from GitNexus knowledge graph index on branch `chore/app-model-decomp`.
> Graph stats: 8,349 nodes | 33,696 edges | 490 clusters | 300 flows.
> No circular imports detected (`gitnexus check --cycles`).

## 1. Current Dispatch Topology

### Root `Update` (`app/update.go:10–68`)

The graph confirms `Update` is a flat dispatcher with 6 outgoing call targets spanning 4 concerns:

| Message / Branch | Handler | File | Domain |
|---|---|---|---|
| `spawnCompleteMsg` / `game.SpawnCompleteMsg` | `model.handleSpawnComplete` | `app/spawn.go:85` | Session |
| `exportCompleteMsg` | `model.handleExportComplete` | `app/export.go` | Export |
| `exportSubmitAction` | `model.handleExportSubmit` | `app/export.go` | Export |
| `backAction` | `msg.applyToModel(m)` | `app/screen_core.go` | Screen action |
| `tea.WindowSizeMsg` | `model.handleWindowSize` | `app/update.go:80` | Shell/route |
| All other keys | `model.handleGlobalKey` | `app/update.go:89` | Mixed (see §3) |
| `gameView` delegation | `newSessionController(&m).updateActiveGame` | `app/session_controller.go:169` | Session |
| Screen update | `m.activeScreen().Update` → `handleScreenAction` | `app/screen_core.go` / `app/update.go:179` | Screen |

**Trace**: `Update → newSessionController → sessionController` is 2 hops (CALLS edges, confidence 0.85). The controller is reconstructed on every `Update` call that needs it — `handleSpawnComplete`, `updateActiveGame`, and the free-function wrappers all call `newSessionController(&m)` inline.

### `handleWindowSize` (`app/update.go:80–87`)

Accesses `m.width`, `m.height`, `m.state`, and calls `resizeActiveScreen()`. For `gameView`, skips screen resize (game handles its own sizing via `sessionController.activateGame` which pushes `tea.WindowSizeMsg` to the `Gamer`).

### `resizeActiveScreen` (`app/update.go:71–78`)

Reads `m.screens[m.state]`, calls `screen.Resize(m.width, m.height)`, writes back. This is the only screen-cache mutation path besides `initScreen`.

---

## 2. Screen Route Management

### `screenModel` interface (`app/screen_core.go:234–238`)

```
Resize(width, height int) screenModel
Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction)
View(notice noticeState) string
```

### Screen registry (`app/screen_core.go:242–288`)

`screenRegistry` maps all 15 `viewState` values to `screenFactory` functions. Factories read from `model` fields (`m.nav`, `m.seed`, `m.create`, `m.cont`, `m.weekly`, `m.help`, `m.stats`, `m.theme`, `m.spinner`) — screens are constructed from shell state, confirming the design's statement that screens are "dependency-light" (they receive data via construction, not via injected references).

### `activeScreen()` (`app/screen_core.go:290–291`)

Called from 2 sites (graph-confirmed):
- `model.Update` (`app/update.go:45`) — for screen message dispatch
- `model.viewContent` (`app/view.go`) — for rendering

### `initScreen(state)` (`app/screen_core.go:297–307`)

13 incoming callers (graph-confirmed):
- `InitialModel`, `InitialModelWithGame` — initialization
- 8 `screenAction.applyToModel` implementations — route transitions
- `sessionController.startSpawn`, `startEloSpawn`, `startSeededSpawn` — generating screen setup
- `enterSeedInputView` (`app/seed_input.go`)
- `handleGlobalKey` (weekly escape path) — re-inits weekly screen after save

This confirms `initScreen` is the single entry point for screen creation, but it's called from both route-transition code (actions) and workflow code (session controller, global key handler) — the ownership overlap the refactor targets.

---

## 3. Global Key Ownership Leaks

`handleGlobalKey` (`app/update.go:89–169`, 80 lines) directly accesses:

| Access | Property | Owner (proposed) |
|---|---|---|
| `m.session.spawn` | `sessionState.spawn` | Session lifecycle |
| `m.session.returnState` | `sessionState.returnState` | Session lifecycle |
| `m.session.game` | `sessionState.game` | Session lifecycle |
| `m.help.showFull` | `helpState.showFull` | Shell |
| `m.debug.enabled` | `debugState.enabled` | Shell |

**Outgoing calls** span 4 domains (graph-confirmed):

- **Session**: `cancelActiveSpawn` (via `model.cancelActiveSpawn` wrapper), `saveCurrentGame` (free function wrapper), `activeSpawnReturnState`
- **Export**: `cancelActiveExport`
- **Route/screen**: `initScreen`, `resizeActiveScreen`
- **Weekly workflow**: `refreshWeeklyBrowser`, `advanceSolvedWeekly`
- **Game**: `Gamer.Reset`, `Gamer.Update` (HelpToggleMsg propagation)

### Wrapper indirection (graph-identified)

Three free-function / method wrappers exist solely because `handleGlobalKey` operates on `model` by value while `sessionController` needs a pointer:

```
handleGlobalKey → saveCurrentGame (spawn.go:91) → sessionController.saveCurrentGame
handleGlobalKey → model.cancelActiveSpawn (spawn.go:81) → sessionController.cancelActiveSpawn
handleSpawnComplete → model.handleSpawnComplete (spawn.go:85) → sessionController.handleSpawnComplete
```

Each creates a throwaway `sessionController{model: &m}`. The refactor should either:
- Have `handleGlobalKey` construct a single controller and delegate, or
- Move the key handling into the controller itself for game/generating states.

---

## 4. Session Controller Surface

`sessionController` (`app/session_controller.go:14–16`) already owns 12 methods:

| Method | Lines | Responsibility |
|---|---|---|
| `beginSpawnContext` | 23–31 | Job ID increment, cancel prev, set generating |
| `cancelActiveSpawn` | 33–43 | Cancel context, clear spawn, bump job ID |
| `startSpawn` | 45–52 | Begin → set spawn → state=generating → initScreen → cmd |
| `startEloSpawn` | 54–61 | Same pattern, Elo variant |
| `startSeededSpawn` | 63–70 | Same pattern, seeded variant |
| `handleSpawnComplete` | 72–123 | Job ID guard, stale check, activate, create record |
| `activateGame` | 138–148 | Help/ws propagation, set session fields, state=gameView |
| `importRecord` | 150–167 | Import, read-only handling, activate |
| `updateActiveGame` | 169–177 | Delegate to Gamer.Update, debug info, persist |
| `persistCompletionIfSolved` | 179–197 | Guarded save + status update |
| `saveCurrentGame` | 199–222 | Save state + status, clear active game |
| `clearActiveGame` | 224–230 | Reset session fields to defaults |

**Process flows through the controller** (graph-confirmed):
- `Update → SessionController` (proc_200, 4 steps)
- `HandleGlobalKey → SessionController` (proc_234, 4 steps)
- `HandleDailyPuzzle → SessionController` (proc_118, 5 steps)

The controller already covers most of the session lifecycle. The gap is that `handleGlobalKey` and `handleSpawnComplete` bypass it through wrappers instead of delegating directly.

---

## 5. Impact Analysis

### Blast radius for `model` struct (`gitnexus impact`)

- **Risk**: MEDIUM
- **Impacted symbols**: 11 (6 direct + 3 depth-2 + 2 depth-3)
- **Processes affected**: 2 (both `Update` flows in `screen_navigation.go`)
- **Modules affected**: App (8 hits, direct), Ui (1 hit, direct)

**Direct dependents** (depth 1):
- `InitialModel` (`app/model.go`) — constructor
- `InitialModelWithGame` (`app/model.go`) — constructor
- `seedInputScreen.Update` (`app/screen_navigation.go`)
- `gameSelectScreen.Resize` (`app/screen_navigation.go`) — 5 processes
- `helpDetailScreen.Resize` (`app/screen_support.go`)
- `statsScreen.Resize` (`app/screen_support.go`)

**Depth-2**: `runGameProgram` (`cmd/session.go`), `gameSelectScreen.resizeSelf`, `cmd/root.go`

### Key insight

The impact is contained within the `app` package + `cmd` entry points. No `game`, `store`, `catalog`, or `registry` package code depends on `model` internals. This confirms the design's non-goal: no public API or package boundary changes needed.

---

## 6. `screenAction` Pattern

The graph identifies 8 action types implementing `screenAction`:

| Action | File | What it does |
|---|---|---|
| `backAction` | `app/screen_core.go` | Generic back navigation |
| `openPlayMenuAction` | `app/screen_core.go` | → playMenuView |
| `openOptionsMenuAction` | `app/screen_core.go` | → optionsMenuView |
| `openCreateAction` | `app/screen_core.go` | → createView |
| `openContinueAction` | `app/screen_core.go` | → continueView |
| `gameSelectEnterAction` | `app/screen_core.go` | → gameSelectView |
| `weeklyShiftAction` | `app/screen_core.go` | → weeklyView |
| `openHelpSelectAction` | `app/screen_core.go` | → helpSelectView |

All actions call `initScreen` to transition. None directly access `store`, `config`, or `sessionController` — confirming the design's "screens emit typed intents" pattern is already in place.

`exportSubmitAction` (`app/screen_core.go:232`) is the exception: it's a `tea.Msg`, not a `screenAction`, and is handled directly in root `Update` (line 22-24). This is the export-workflow special case the design notes.

---

## 7. File Inventory (affected code)

| File | Lines | Role |
|---|---|---|
| `app/model.go` | 281 | Model struct, viewState, state types, constructors |
| `app/update.go` | 181 | Update dispatch, handleGlobalKey, handleWindowSize, resizeActiveScreen |
| `app/screen_core.go` | 307 | screenModel interface, registry, activeScreen, initScreen, actions |
| `app/session_controller.go` | 230 | Session lifecycle (12 methods) |
| `app/spawn.go` | 94 | Spawn commands + 3 wrapper functions |
| `app/view.go` | 214 | View rendering |
| `app/export.go` | 1720 | Export workflow (screen + handlers) |
| `app/handle_create.go` | 114 | Create flow |
| `app/handle_seed.go` | 66 | Seed flow |
| `app/handle_daily.go` | 48 | Daily flow |
| `app/weekly.go` | 319 | Weekly flow + browser |

**Total**: ~3,574 lines in the affected surface.

---

## 8. Refactor Recommendations (graph-derived)

### 8.1 Route/screen manager

Extract from `model`:
- `state viewState` → `manager.current`
- `screens map[viewState]screenModel` → `manager.cache`
- `activeScreen()` → `manager.active()`
- `initScreen(state)` → `manager.init(state, shell)`
- `resizeActiveScreen()` → `manager.resizeActive()`
- `handleWindowSize` → `manager.handleWindowSize(msg, isGameView)`

**Graph evidence**: `activeScreen` has 2 callers (`Update`, `viewContent`), `initScreen` has 13 callers, `resizeActiveScreen` has 3 callers (`handleWindowSize`, `handleGlobalKey` ×2). Low blast radius.

### 8.2 Session key delegation

Move `handleGlobalKey` branches for `generatingView` and `gameView` into `sessionController`:

- `generatingView` escape/quit → `sessionController.handleGeneratingKey(keyMsg) (model, tea.Cmd, bool)`
- `gameView` escape/quit/reset/debug/fullhelp/enter → `sessionController.handleGameKey(keyMsg) (model, tea.Cmd, bool)`

This eliminates:
- The 3 wrapper functions in `spawn.go` (lines 81–94)
- Direct access to `m.session.*`, `m.help.showFull`, `m.debug.enabled` from `handleGlobalKey`
- The `saveCurrentGame` free function

### 8.3 Export key delegation

Move `exportRunningView` escape/quit from `handleGlobalKey` into export workflow code:
- `exportController.handleRunningKey(keyMsg) (model, tea.Cmd, bool)`

Or simpler: an `exportRunningKeyHandler` function in `app/export.go` that `handleGlobalKey` calls for `exportRunningView` state.

### 8.4 Test coverage priorities (graph-derived high-risk paths)

Based on process flow counts and caller counts, these paths have the highest regression risk:

1. **`initScreen` with 13 callers** — any change to screen cache behavior risks all route transitions
2. **`handleGlobalKey` game escape path** — calls `saveCurrentGame` + `refreshWeeklyBrowser` + `initScreen(weeklyView)` in sequence; touches session, weekly, and route domains
3. **`sessionController.handleSpawnComplete`** — 3 process flows depend on it; job ID guard + stale check + activate + record creation
4. **`sessionController.activateGame`** — called from `handleSpawnComplete` and `importRecord`; help/ws propagation order matters
5. **`resizeActiveScreen`** — called from 3 sites; nil-map-read safety depends on `screens` being initialized

---

## 9. Open Questions Resolved by Graph

### Q1: Should create/daily/seed/weekly decision logic live in one workflow helper?

**Graph evidence**: Each handler (`handle_create.go`, `handle_seed.go`, `handle_daily.go`, `weekly.go`) calls `sessionController.startSpawn` / `startEloSpawn` / `startSeededSpawn` independently. The call patterns are similar but the decision logic (which spawner, which spawn source, which return state) is workflow-specific. 

**Recommendation**: Keep decision logic in current files. The session controller already owns the spawn mechanics. Consolidating decision logic would add a layer without reducing duplication, since each workflow constructs a different `spawnRequest` with different fields.

### Q2: Should export get a small controller?

**Graph evidence**: `export.go` is 1,720 lines — the largest file in the affected surface. `handleGlobalKey` only touches it via `cancelActiveExport`. `Update` touches it via `handleExportSubmit` and `handleExportComplete`.

**Recommendation**: Minimal cleanup. Extract the `exportRunningView` key handling into `app/export.go` as a function. Do not build a full export controller in this change — the export workflow's size warrants its own dedicated refactor.

---

*Analysis performed with GitNexus v1.6.8 CLI. Graph index built on `chore/app-model-decomp` branch.*

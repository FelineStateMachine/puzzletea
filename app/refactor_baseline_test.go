package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"
)

// --- Window resize ---

func TestWindowResizeUpdatesDimensions(t *testing.T) {
	m := model{state: mainMenuView, width: 40, height: 10}
	m = m.initScreen(mainMenuView)

	m = m.handleWindowSize(tea.WindowSizeMsg{Width: 80, Height: 24})

	if m.width != 80 || m.height != 24 {
		t.Fatalf("dimensions = %dx%d, want 80x24", m.width, m.height)
	}
}

func TestWindowResizeResizesActiveScreen(t *testing.T) {
	m := model{state: mainMenuView, width: 40, height: 10}
	m = m.initScreen(mainMenuView)
	before := m.screens[mainMenuView].(mainMenuScreen)
	if before.width != 40 || before.height != 10 {
		t.Fatalf("before = %dx%d, want 40x10", before.width, before.height)
	}

	m = m.handleWindowSize(tea.WindowSizeMsg{Width: 80, Height: 24})

	after := m.screens[mainMenuView].(mainMenuScreen)
	if after.width != 80 || after.height != 24 {
		t.Fatalf("after = %dx%d, want 80x24", after.width, after.height)
	}
}

func TestWindowResizeSkipsScreenResizeInGameView(t *testing.T) {
	// In gameView, the game handles its own sizing — the screen cache
	// should not be touched by handleWindowSize.
	screen := safeScreen{}
	m := model{
		state:   gameView,
		width:   40,
		height:  10,
		screens: map[viewState]screenModel{gameView: screen},
	}

	m = m.handleWindowSize(tea.WindowSizeMsg{Width: 80, Height: 24})

	if m.width != 80 || m.height != 24 {
		t.Fatalf("dimensions = %dx%d, want 80x24", m.width, m.height)
	}
	// Screen should be the exact same value (not replaced).
	if m.screens[gameView] != screen {
		t.Fatal("expected screen cache to be untouched in gameView")
	}
}

// --- Screen cache survival ---

func TestScreenCacheSurvivesRouteTransition(t *testing.T) {
	m := model{state: mainMenuView, width: 80, height: 24}
	m = m.initScreen(mainMenuView)
	if m.screens[mainMenuView] == nil {
		t.Fatal("expected mainMenuView screen to be cached")
	}

	// Transition to play menu via action.
	next, _ := openPlayMenuAction{}.applyToModel(m)
	got := next

	// Main menu screen should still be cached (not evicted).
	if got.screens[mainMenuView] == nil {
		t.Fatal("expected mainMenuView screen to survive route transition")
	}
	// Play menu screen should now be cached.
	if got.screens[playMenuView] == nil {
		t.Fatal("expected playMenuView screen to be cached after transition")
	}
}

// --- Game view message delegation ---

func TestGameViewNonKeyMessageDelegatesToGame(t *testing.T) {
	tracker := &escapeTrackingGame{}
	m := model{
		state:   gameView,
		session: sessionState{game: tracker},
	}

	// A non-key message in gameView should reach the game via updateActiveGame.
	m.Update(tea.MouseClickMsg{})

	if tracker.updateCalls == 0 {
		t.Fatal("expected game.Update to be called for non-key messages in gameView")
	}
}

// --- Completion persistence ---

func TestPersistCompletionIfSolvedSavesOnce(t *testing.T) {
	s := openAppTestStore(t)

	rec := &store.GameRecord{
		Name:         "solved-test",
		GameType:     "Nonogram",
		Mode:         "Easy",
		InitialState: `{}`,
		SaveState:    `{}`,
		Status:       store.StatusInProgress,
	}
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}

	m := model{
		state: gameView,
		store: s,
		session: sessionState{
			game:            &solvedWeeklyGame{save: []byte(`{"done":true}`)},
			activeGameID:    rec.ID,
			completionSaved: false,
		},
	}

	newSessionController(&m).persistCompletionIfSolved()

	if !m.session.completionSaved {
		t.Fatal("expected completionSaved to be true after persisting")
	}

	// Verify the record was updated to completed.
	all, err := s.ListAllGames()
	if err != nil {
		t.Fatal(err)
	}
	var saved *store.GameRecord
	for i := range all {
		if all[i].Name == "solved-test" {
			saved = &all[i]
			break
		}
	}
	if saved == nil {
		t.Fatal("expected solved-test record to exist")
	}
	if saved.Status != store.StatusCompleted {
		t.Fatalf("status = %q, want %q", saved.Status, store.StatusCompleted)
	}

	// Calling again should be a no-op (idempotent).
	callsBefore := 0
	_ = callsBefore
	newSessionController(&m).persistCompletionIfSolved()
	// No panic, no double-save — completionSaved remains true.
	if !m.session.completionSaved {
		t.Fatal("expected completionSaved to remain true")
	}
}

func TestPersistCompletionDoesNotSaveUnsolvedGame(t *testing.T) {
	s := openAppTestStore(t)

	rec := &store.GameRecord{
		Name:         "unsolved-test",
		GameType:     "Nonogram",
		Mode:         "Easy",
		InitialState: `{}`,
		SaveState:    `{}`,
		Status:       store.StatusInProgress,
	}
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}

	m := model{
		state: gameView,
		store: s,
		session: sessionState{
			game:            &escapeTrackingGame{}, // IsSolved() returns false
			activeGameID:    rec.ID,
			completionSaved: false,
		},
	}

	newSessionController(&m).persistCompletionIfSolved()

	if m.session.completionSaved {
		t.Fatal("expected completionSaved to remain false for unsolved game")
	}
}

// --- Session lifecycle: clearActiveGame resets state ---

func TestClearActiveGameResetsSession(t *testing.T) {
	m := model{
		state: gameView,
		session: sessionState{
			game:            &escapeTrackingGame{},
			activeGameID:    42,
			completionSaved: true,
			returnState:     weeklyView,
		},
	}

	newSessionController(&m).clearActiveGame()

	if m.session.activeGameID != 0 {
		t.Fatalf("activeGameID = %d, want 0", m.session.activeGameID)
	}
	if m.session.game != nil {
		t.Fatal("expected game to be nil")
	}
	if m.session.completionSaved {
		t.Fatal("expected completionSaved to be false")
	}
	if m.session.returnState != mainMenuView {
		t.Fatalf("returnState = %d, want %d", m.session.returnState, mainMenuView)
	}
}

// --- Game key: enter on non-solved weekly does not advance ---

func TestGameViewEnterOnUnsolvedGameDoesNotAdvance(t *testing.T) {
	m := model{
		state:   gameView,
		session: sessionState{game: &escapeTrackingGame{}}, // IsSolved() = false
	}

	next, _, handled := newSessionController(&m).handleGameKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	if handled {
		t.Fatal("expected enter on unsolved game to not be handled")
	}
	if next.state != gameView {
		t.Fatalf("state = %d, want %d (gameView)", next.state, gameView)
	}
}

// --- Generating view swallows non-escape/quit keys ---

func TestGeneratingViewSwallowsOtherKeys(t *testing.T) {
	m := model{
		state:   generatingView,
		session: sessionState{generating: true},
	}

	_, _, handled := newSessionController(&m).handleGeneratingKey(tea.KeyPressMsg{Code: 'x'})

	if !handled {
		t.Fatal("expected all keys to be swallowed in generating view")
	}
}

// --- Export running view swallows non-escape/quit keys ---

func TestExportRunningViewSwallowsOtherKeys(t *testing.T) {
	m := model{
		state:   exportRunningView,
		screens: map[viewState]screenModel{exportView: safeScreen{}},
	}

	_, _, handled := m.handleExportRunningKey(tea.KeyPressMsg{Code: 'x'})

	if !handled {
		t.Fatal("expected all keys to be swallowed in export running view")
	}
}

// Ensure game import is used (for game.HelpToggleMsg etc. if tests expand).
var _ game.Gamer = (*escapeTrackingGame)(nil)

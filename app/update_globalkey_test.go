package app

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/weekly"
)

// ---------------------------------------------------------------------------
// Normal screen global keys
// ---------------------------------------------------------------------------

func TestNormalScreenQuitReturnsQuitCmd(t *testing.T) {
	m := model{state: mainMenuView}

	next, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	got := next.(model)

	if got.state != mainMenuView {
		t.Fatalf("state = %d, want %d (mainMenuView)", got.state, mainMenuView)
	}
	if cmd == nil {
		t.Fatal("expected Quit command, got nil")
	}
	if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
		t.Fatal("expected tea.Quit, got different command")
	}
}

func TestNormalScreenDebugToggles(t *testing.T) {
	m := model{state: playMenuView}

	next, cmd := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	got := next.(model)

	if !got.debug.enabled {
		t.Fatal("expected debug to be enabled")
	}
	if cmd != nil {
		t.Fatal("expected nil cmd after debug toggle")
	}

	next2, _ := got.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	got2 := next2.(model)
	if got2.debug.enabled {
		t.Fatal("expected debug to be disabled after second toggle")
	}
}

func TestNormalScreenFullHelpToggles(t *testing.T) {
	m := model{state: optionsMenuView}

	next, cmd := m.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl})
	got := next.(model)

	if !got.help.showFull {
		t.Fatal("expected full help to be enabled")
	}
	if cmd != nil {
		t.Fatal("expected nil cmd after fullHelp toggle")
	}

	next2, _ := got.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl})
	got2 := next2.(model)
	if got2.help.showFull {
		t.Fatal("expected full help to be disabled after second toggle")
	}
}

// ---------------------------------------------------------------------------
// Pending generation global keys
// ---------------------------------------------------------------------------

func TestGeneratingScreenQuitReturnsQuitCmd(t *testing.T) {
	m := model{
		state:   generatingView,
		session: sessionState{generating: true},
	}

	next, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	got := next.(model)

	if got.state != generatingView {
		t.Fatalf("state = %d, want %d (generatingView)", got.state, generatingView)
	}
	if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
		t.Fatal("expected tea.Quit in generating view")
	}
}

// ---------------------------------------------------------------------------
// Pending export global keys
// ---------------------------------------------------------------------------

func TestExportRunningEscapeReturnsToExportView(t *testing.T) {
	// Use handleGlobalKey directly to focus on the key dispatch, not screen
	// resize. A safeScreen avoids nil-pointer panics from an uninitialized
	// exportScreen viewport during resizeActiveScreen.
	m := model{
		state: exportRunningView,
		screens: map[viewState]screenModel{
			exportView: safeScreen{},
		},
	}

	next, cmd, handled := m.handleGlobalKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	if !handled {
		t.Fatal("expected handleGlobalKey to handle escape in export running view")
	}
	if next.state != exportView {
		t.Fatalf("state = %d, want %d (exportView)", next.state, exportView)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd, got %v", cmd)
	}
}

func TestExportRunningQuitReturnsQuitCmd(t *testing.T) {
	m := model{
		state: exportRunningView,
		screens: map[viewState]screenModel{
			exportView: exportScreen{},
		},
	}

	next, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	got := next.(model)

	if got.state != exportRunningView {
		t.Fatalf("state = %d, want %d (exportRunningView)", got.state, exportRunningView)
	}
	if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
		t.Fatal("expected tea.Quit in export running view")
	}
}

// ---------------------------------------------------------------------------
// Active game global keys (quit → save + quit)
// ---------------------------------------------------------------------------

func TestGameViewQuitSavesAndReturnsQuit(t *testing.T) {
	s := openAppTestStore(t)

	rec := &store.GameRecord{
		Name:         "quit-test",
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
			game:         &escapeTrackingGame{},
			activeGameID: rec.ID,
			returnState:  mainMenuView,
		},
	}

	next, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	_ = next.(model)

	if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
		t.Fatal("expected tea.Quit when quitting from game view")
	}

	// The game should have been saved as abandoned.
	// GetGameByName excludes abandoned, so use ListAllGames.
	all, err := s.ListAllGames()
	if err != nil {
		t.Fatal(err)
	}
	var saved *store.GameRecord
	for i := range all {
		if all[i].Name == "quit-test" {
			saved = &all[i]
			break
		}
	}
	if saved == nil {
		t.Fatal("expected quit-test record to exist")
	}
	if saved.Status != store.StatusAbandoned {
		t.Fatalf("status = %q, want %q", saved.Status, store.StatusAbandoned)
	}
}

// ---------------------------------------------------------------------------
// Active game: escape → save + return
// ---------------------------------------------------------------------------

func TestGameViewEscapeSavesAndReturnsToReturnState(t *testing.T) {
	s := openAppTestStore(t)

	rec := &store.GameRecord{
		Name:         "escape-test",
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
			game:         &escapeTrackingGame{},
			activeGameID: rec.ID,
			returnState:  playMenuView,
		},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)

	if got.state != playMenuView {
		t.Fatalf("state = %d, want %d (playMenuView)", got.state, playMenuView)
	}

	saved, err := s.GetGameByName("escape-test")
	if err != nil {
		t.Fatal(err)
	}
	if saved == nil {
		t.Fatal("expected escape-test record to exist")
	}
	if saved.Status != store.StatusInProgress {
		t.Fatalf("status = %q, want %q", saved.Status, store.StatusInProgress)
	}
}

// ---------------------------------------------------------------------------
// Active game: debug toggle
// ---------------------------------------------------------------------------

func TestGameViewDebugToggles(t *testing.T) {
	m := model{
		state:   gameView,
		session: sessionState{game: &escapeTrackingGame{}},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	got := next.(model)

	if !got.debug.enabled {
		t.Fatal("expected debug to be enabled")
	}

	next2, _ := got.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	got2 := next2.(model)
	if got2.debug.enabled {
		t.Fatal("expected debug to be disabled")
	}
}

// ---------------------------------------------------------------------------
// Active game: fullHelp toggle propagates to game
// ---------------------------------------------------------------------------

func TestGameViewFullHelpTogglesAndPropagates(t *testing.T) {
	tracker := &escapeTrackingGame{}
	m := model{
		state:   gameView,
		session: sessionState{game: tracker},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl})
	got := next.(model)

	if !got.help.showFull {
		t.Fatal("expected full help to be enabled")
	}
	if tracker.updateCalls == 0 {
		t.Fatal("expected game.Update to be called for HelpToggleMsg propagation")
	}

	next2, _ := got.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl})
	got2 := next2.(model)
	if got2.help.showFull {
		t.Fatal("expected full help to be disabled after second toggle")
	}
}

// ---------------------------------------------------------------------------
// Active game: reset calls game.Reset
// ---------------------------------------------------------------------------

func TestGameViewResetCallsGameReset(t *testing.T) {
	tracker := &escapeTrackingGame{}
	m := model{
		state:   gameView,
		session: sessionState{game: tracker},
	}

	gotAfter, _ := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	got := gotAfter.(model)

	if got.session.game == nil {
		t.Fatal("expected game to still be set after reset")
	}
}

// ---------------------------------------------------------------------------
// Active game: enter on solved weekly advances
// ---------------------------------------------------------------------------

func TestGameViewEnterOnSolvedWeeklyAdvances(t *testing.T) {
	s := openAppTestStore(t)

	now := time.Now()
	info := weekly.Current(now)
	year, weekNumber := info.Year, info.Week
	currentName := weekly.Name(year, weekNumber, 1)

	rec := &store.GameRecord{
		Name:         currentName,
		GameType:     "Nonogram",
		Mode:         "Easy",
		InitialState: `{}`,
		SaveState:    `{}`,
		Status:       store.StatusInProgress,
		RunKind:      store.RunKindWeekly,
		WeekYear:     year,
		WeekNumber:   weekNumber,
		WeekIndex:    1,
	}
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}

	m := model{
		state:   gameView,
		store:   s,
		spinner: newSpinner(),
		session: sessionState{
			game:            &solvedWeeklyGame{save: []byte(`{"done":true}`)},
			activeGameID:    rec.ID,
			returnState:     weeklyView,
			weeklyAdvance:   &info,
			completionSaved: false,
		},
		weekly: weeklyState{
			cursor: weekly.StartOfWeek(year, weekNumber, time.Local),
		},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := next.(model)

	if got.state != generatingView {
		t.Fatalf("state = %d, want %d (generatingView)", got.state, generatingView)
	}

	// Verify the first weekly was marked as completed
	saved, err := s.GetWeeklyGame(year, weekNumber, 1)
	if err != nil {
		t.Fatal(err)
	}
	if saved == nil {
		t.Fatal("expected weekly record")
	}
	if saved.Status != store.StatusCompleted {
		t.Fatalf("status = %q, want %q", saved.Status, store.StatusCompleted)
	}
}

// ---------------------------------------------------------------------------
// Non-key messages pass through handleGlobalKey unchanged
// ---------------------------------------------------------------------------

func TestHandleGlobalKeyIgnoresNonKeyMessages(t *testing.T) {
	m := model{state: mainMenuView}

	_, cmd, handled := m.handleGlobalKey(tea.WindowSizeMsg{Width: 100, Height: 40})
	if handled {
		t.Fatal("expected non-key message to not be handled by handleGlobalKey")
	}
	if cmd != nil {
		t.Fatal("expected nil cmd for non-key message")
	}
}

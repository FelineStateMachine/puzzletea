package app

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/resolve"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/sudoku"
)

type escapeTrackingGame struct {
	updateCalls int
	sawEscape   bool
}

func (g *escapeTrackingGame) GetDebugInfo() string {
	return ""
}

func (g *escapeTrackingGame) GetFullHelp() [][]key.Binding {
	return nil
}

func (g *escapeTrackingGame) GetSave() ([]byte, error) {
	return nil, nil
}

func (g *escapeTrackingGame) IsSolved() bool {
	return false
}

func (g *escapeTrackingGame) Reset() game.Gamer {
	return g
}

func (g *escapeTrackingGame) SetTitle(string) game.Gamer {
	return g
}

func (g *escapeTrackingGame) Init() tea.Cmd {
	return nil
}

func (g *escapeTrackingGame) View() string {
	return ""
}

func (g *escapeTrackingGame) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	g.updateCalls++
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok && keyMsg.Code == tea.KeyEscape {
		g.sawEscape = true
	}
	return g, nil
}

func TestGameViewEscapePassesThroughToGame(t *testing.T) {
	tracker := &escapeTrackingGame{}
	m := model{
		state: gameView,
		game:  tracker,
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)

	if got.state != gameView {
		t.Fatalf("expected to remain in gameView, got %d", got.state)
	}
	if tracker.updateCalls != 1 {
		t.Fatalf("expected game Update to be called once, got %d", tracker.updateCalls)
	}
	if !tracker.sawEscape {
		t.Fatal("expected game to receive escape key")
	}
}

func TestImportAndActivateRecordSuccessFlag(t *testing.T) {
	unknown := model{state: playMenuView}
	if _, ok := unknown.importAndActivateRecord(store.GameRecord{GameType: "NoSuchGameType"}); ok {
		t.Fatal("expected unknown game type import to fail")
	}

	const gameType = "TestImportAndActivateRecordGame"
	prev, hadPrev := game.Registry[gameType]
	game.Register(gameType, func(_ []byte) (game.Gamer, error) {
		return &escapeTrackingGame{}, nil
	})
	t.Cleanup(func() {
		if hadPrev {
			game.Registry[gameType] = prev
			return
		}
		delete(game.Registry, gameType)
	})

	rec := store.GameRecord{
		ID:        42,
		Name:      "resume-me",
		GameType:  gameType,
		SaveState: "{}",
		Status:    store.StatusCompleted,
	}

	next, ok := (model{
		state:        playMenuView,
		width:        80,
		height:       24,
		showFullHelp: true,
	}).importAndActivateRecord(rec)
	if !ok {
		t.Fatal("expected known game type import to succeed")
	}
	if next.state != gameView {
		t.Fatalf("state = %d, want %d (gameView)", next.state, gameView)
	}
	if next.activeGameID != rec.ID {
		t.Fatalf("activeGameID = %d, want %d", next.activeGameID, rec.ID)
	}
	if !next.completionSaved {
		t.Fatal("expected completionSaved to be true for completed record")
	}
	if next.game == nil {
		t.Fatal("expected game to be activated")
	}
}

func TestHandleSeedConfirmDoesNotResumeStatusWhenImportFails(t *testing.T) {
	s := openAppTestStore(t)
	seed := "import-failure-seed"
	nameRNG := resolve.RNGFromString("name:" + seed)
	name := seed + " - " + namegen.GenerateSeeded(nameRNG)
	rec := &store.GameRecord{
		Name:         name,
		GameType:     "NoSuchGameType",
		Mode:         "Any",
		InitialState: "{}",
		SaveState:    "{}",
		Status:       store.StatusAbandoned,
	}
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}

	ti := textinput.New()
	ti.SetValue(seed)
	m := model{
		state:     seedInputView,
		store:     s,
		seedInput: ti,
	}

	next, _ := m.handleSeedConfirm()
	got := next.(model)
	if got.state != seedInputView {
		t.Fatalf("state = %d, want %d (seedInputView)", got.state, seedInputView)
	}

	saved, err := s.GetDailyGame(name)
	if err != nil {
		t.Fatal(err)
	}
	if saved == nil {
		t.Fatal("expected seeded record to exist")
	}
	if saved.Status != store.StatusAbandoned {
		t.Fatalf("status = %q, want %q", saved.Status, store.StatusAbandoned)
	}
}

func TestHandleDailyPuzzleDoesNotResumeStatusWhenImportFails(t *testing.T) {
	s := openAppTestStore(t)
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -1),
		now,
		now.AddDate(0, 0, 1),
	}

	seen := map[string]bool{}
	for _, d := range dates {
		name := daily.Name(d)
		if seen[name] {
			continue
		}
		seen[name] = true
		rec := &store.GameRecord{
			Name:         name,
			GameType:     "NoSuchGameType",
			Mode:         "Any",
			InitialState: "{}",
			SaveState:    "{}",
			Status:       store.StatusAbandoned,
		}
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
	}

	m := model{
		state: playMenuView,
		store: s,
	}

	next, _ := m.handleDailyPuzzle()
	got := next.(model)
	if got.state != playMenuView {
		t.Fatalf("state = %d, want %d (playMenuView)", got.state, playMenuView)
	}

	for name := range seen {
		saved, err := s.GetDailyGame(name)
		if err != nil {
			t.Fatal(err)
		}
		if saved == nil {
			t.Fatalf("expected daily record %q to exist", name)
		}
		if saved.Status != store.StatusAbandoned {
			t.Fatalf("record %q status = %q, want %q", name, saved.Status, store.StatusAbandoned)
		}
	}
}

func openAppTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestCtrlYYanksSupportedPuzzle(t *testing.T) {
	previousCopy := copyToClipboard
	t.Cleanup(func() { copyToClipboard = previousCopy })

	var copied string
	copyToClipboard = func(s string) error {
		copied = s
		return nil
	}

	m := model{
		state: gameView,
		game:  sudoku.Model{},
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	if cmd != nil {
		t.Fatal("expected nil command when native clipboard copy succeeds")
	}
	if copied == "" {
		t.Fatal("expected markdown snippet to be copied")
	}
	if !strings.Contains(copied, "Given Grid") {
		t.Fatalf("expected sudoku markdown snippet, got:\n%s", copied)
	}
}

func TestCtrlYYankUnsupportedGameIsNoOp(t *testing.T) {
	previousCopy := copyToClipboard
	t.Cleanup(func() { copyToClipboard = previousCopy })

	calls := 0
	copyToClipboard = func(string) error {
		calls++
		return nil
	}

	m := model{
		state: gameView,
		game:  lightsout.Model{},
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	if cmd != nil {
		t.Fatal("expected nil command for unsupported game yank")
	}
	if calls != 0 {
		t.Fatalf("clipboard should not be called, got %d calls", calls)
	}
}

func TestCtrlYYankOutsideGameViewIsNoOp(t *testing.T) {
	previousCopy := copyToClipboard
	t.Cleanup(func() { copyToClipboard = previousCopy })

	calls := 0
	copyToClipboard = func(string) error {
		calls++
		return nil
	}

	m := model{
		state: mainMenuView,
		game:  sudoku.Model{},
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	if cmd != nil {
		t.Fatal("expected nil command outside game view")
	}
	if calls != 0 {
		t.Fatalf("clipboard should not be called outside game view, got %d calls", calls)
	}
}

func TestCtrlYYankFallsBackToOSC52(t *testing.T) {
	previousCopy := copyToClipboard
	t.Cleanup(func() { copyToClipboard = previousCopy })

	copyToClipboard = func(string) error {
		return errors.New("clipboard unavailable")
	}

	m := model{
		state: gameView,
		game:  sudoku.Model{},
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected fallback clipboard command")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("expected fallback clipboard message")
	}
}

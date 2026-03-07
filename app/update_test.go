package app

import (
	"path/filepath"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"
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
		state:   gameView,
		session: sessionState{game: tracker},
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

	loadedGame, err := lightsout.New(3, 3)
	if err != nil {
		t.Fatal(err)
	}
	save, err := loadedGame.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	rec := store.GameRecord{
		ID:        42,
		Name:      "resume-me",
		GameType:  "Lights Out",
		SaveState: string(save),
		Status:    store.StatusCompleted,
	}

	next, ok := (model{
		state:  playMenuView,
		width:  80,
		height: 24,
		help:   helpState{showFull: true},
	}).importAndActivateRecord(rec)
	if !ok {
		t.Fatal("expected known game type import to succeed")
	}
	if next.state != gameView {
		t.Fatalf("state = %d, want %d (gameView)", next.state, gameView)
	}
	if next.session.activeGameID != rec.ID {
		t.Fatalf("activeGameID = %d, want %d", next.session.activeGameID, rec.ID)
	}
	if !next.session.completionSaved {
		t.Fatal("expected completionSaved to be true for completed record")
	}
	if next.session.game == nil {
		t.Fatal("expected game to be activated")
	}
}

func TestHandleSeedConfirmDoesNotResumeStatusWhenImportFails(t *testing.T) {
	s := openAppTestStore(t)
	seed := "import-failure-seed"
	name := sessionflow.SeededName(seed)
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
		state: seedInputView,
		store: s,
		nav:   navigationState{seedInput: ti},
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

func TestSeedInputSelectorCyclesAndPersistsDefault(t *testing.T) {
	options := buildSeedModeOptions(catalog.All)
	if len(options) < 2 {
		t.Fatal("expected at least one seeded game option in addition to Random")
	}

	m := model{
		state: seedInputView,
		nav: navigationState{
			seedModeOptions: options,
			seedFocus:       seedFocusMode,
		},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	got := next.(model)

	if got.nav.seedModeIndex != 1 {
		t.Fatalf("seedModeIndex = %d, want 1", got.nav.seedModeIndex)
	}
	if got.nav.lastSeedModeKey != options[1].key {
		t.Fatalf("lastSeedModeKey = %q, want %q", got.nav.lastSeedModeKey, options[1].key)
	}

	playMenu := ui.NewMainMenu(playMenuItems)
	for range 3 {
		playMenu.CursorDown()
	}

	reopened, _ := (model{
		state: playMenuView,
		nav: navigationState{
			playMenu:        playMenu,
			lastSeedModeKey: got.nav.lastSeedModeKey,
		},
	}).handlePlayMenuEnter()
	reopenedModel := reopened.(model)

	if reopenedModel.state != seedInputView {
		t.Fatalf("state = %d, want %d (seedInputView)", reopenedModel.state, seedInputView)
	}
	if reopenedModel.currentSeedMode().key != options[1].key {
		t.Fatalf("reopened seed mode key = %q, want %q", reopenedModel.currentSeedMode().key, options[1].key)
	}
}

func TestHandleSeedConfirmUsesSelectedSpecificMode(t *testing.T) {
	s := openAppTestStore(t)
	options := buildSeedModeOptions(catalog.All)
	if len(options) < 2 {
		t.Fatal("expected at least one seeded game option in addition to Random")
	}

	ti := textinput.New()
	ti.SetValue("specific-mode-seed")

	m := model{
		state: seedInputView,
		store: s,
		nav: navigationState{
			seedInput:       ti,
			seedModeOptions: options,
			seedModeIndex:   1,
		},
	}

	next, _ := m.handleSeedConfirm()
	got := next.(model)

	if got.state != generatingView {
		t.Fatalf("state = %d, want %d (generatingView)", got.state, generatingView)
	}
	if got.session.spawn == nil {
		t.Fatal("expected spawn request to be populated")
	}
	if got.session.spawn.gameType != options[1].gameType {
		t.Fatalf("gameType = %q, want %q", got.session.spawn.gameType, options[1].gameType)
	}
	if got.session.spawn.modeTitle == "" {
		t.Fatal("expected a deterministic mode to be selected within the chosen game")
	}
	if got.session.spawn.name == sessionflow.SeededName("specific-mode-seed") {
		t.Fatal("specific seeded game should not reuse the random seeded name")
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

func TestGameSelectEscapeClearsAppliedFilterBeforeLeavingView(t *testing.T) {
	l := ui.InitCategoryList(GameCategories, "Select Category")
	l.SetFilterText("tak")
	l.SetFilterState(list.FilterApplied)

	m := model{
		state: gameSelectView,
		nav: navigationState{
			gameSelectList: l,
		},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)

	if got.state != gameSelectView {
		t.Fatalf("state = %d, want %d (gameSelectView)", got.state, gameSelectView)
	}
	if got.nav.gameSelectList.FilterState() != list.Unfiltered {
		t.Fatalf("filter state = %s, want %s", got.nav.gameSelectList.FilterState(), list.Unfiltered)
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

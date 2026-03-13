package app

import (
	"path/filepath"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/registry"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"
	"github.com/FelineStateMachine/puzzletea/weekly"
)

type escapeTrackingGame struct {
	updateCalls int
	sawEscape   bool
}

type solvedWeeklyGame struct {
	save []byte
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

func (g *solvedWeeklyGame) GetDebugInfo() string {
	return ""
}

func (g *solvedWeeklyGame) GetFullHelp() [][]key.Binding {
	return nil
}

func (g *solvedWeeklyGame) GetSave() ([]byte, error) {
	return g.save, nil
}

func (g *solvedWeeklyGame) IsSolved() bool {
	return true
}

func (g *solvedWeeklyGame) Reset() game.Gamer {
	return g
}

func (g *solvedWeeklyGame) SetTitle(string) game.Gamer {
	return g
}

func (g *solvedWeeklyGame) Init() tea.Cmd {
	return nil
}

func (g *solvedWeeklyGame) View() string {
	return ""
}

func (g *solvedWeeklyGame) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	return g, nil
}

func TestGameViewEscapeReturnsToMainMenu(t *testing.T) {
	tracker := &escapeTrackingGame{}
	m := model{
		state:   gameView,
		session: sessionState{game: tracker},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)

	if got.state != mainMenuView {
		t.Fatalf("expected to return to mainMenuView, got %d", got.state)
	}
	if tracker.updateCalls != 0 {
		t.Fatalf("expected game Update not to be called, got %d", tracker.updateCalls)
	}
	if tracker.sawEscape {
		t.Fatal("expected escape to be handled by the app before reaching the game")
	}
}

func TestImportAndActivateRecordSuccessFlag(t *testing.T) {
	unknown := model{state: playMenuView}
	nextUnknown, ok := unknown.importAndActivateRecord(store.GameRecord{
		Name:     "broken-save",
		GameType: "NoSuchGameType",
	})
	if ok {
		t.Fatal("expected unknown game type import to fail")
	}
	if nextUnknown.notice.message == "" {
		t.Fatal("expected failed import to surface a user-visible notice")
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
	if next.notice.message != "" {
		t.Fatalf("notice = %q, want empty after successful import", next.notice.message)
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
	options := buildSeedModeOptions(registry.Definitions())
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

	reopenedModel, _ := (model{
		state: playMenuView,
		nav: navigationState{
			lastSeedModeKey: got.nav.lastSeedModeKey,
		},
	}).enterSeedInputView()

	if reopenedModel.state != seedInputView {
		t.Fatalf("state = %d, want %d (seedInputView)", reopenedModel.state, seedInputView)
	}
	if reopenedModel.currentSeedMode().key != options[1].key {
		t.Fatalf("reopened seed mode key = %q, want %q", reopenedModel.currentSeedMode().key, options[1].key)
	}
}

func TestHandleSeedConfirmUsesSelectedSpecificMode(t *testing.T) {
	s := openAppTestStore(t)
	options := buildSeedModeOptions(registry.Definitions())
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

func TestBuildCurrentWeeklyRowsStartsAtOne(t *testing.T) {
	rows := buildCurrentWeeklyRows(2026, 1, nil)
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	if rows[0].Index != 1 || !rows[0].Playable || rows[0].Status != store.StatusNew {
		t.Fatalf("row[0] = %+v, want index=1 playable status=new", rows[0])
	}
}

func TestBuildCurrentWeeklyRowsUnlocksSequentially(t *testing.T) {
	games := []store.GameRecord{
		{Name: weekly.Name(2026, 1, 1), GameType: "Sudoku", Mode: "Easy", Status: store.StatusCompleted},
		{Name: weekly.Name(2026, 1, 2), GameType: "Sudoku", Mode: "Easy", Status: store.StatusInProgress},
		{Name: weekly.Name(2026, 1, 4), GameType: "Sudoku", Mode: "Easy", Status: store.StatusCompleted},
	}

	rows := buildCurrentWeeklyRows(2026, 1, games)
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Index != 2 || !rows[0].Playable {
		t.Fatalf("row[0] = %+v, want playable #2", rows[0])
	}
	if rows[1].Index != 1 || !rows[1].ReadOnly {
		t.Fatalf("row[1] = %+v, want read-only #1", rows[1])
	}
}

func TestBuildReviewWeeklyRowsOnlyIncludesCompleted(t *testing.T) {
	games := []store.GameRecord{
		{Name: weekly.Name(2026, 1, 1), GameType: "Sudoku", Mode: "Easy", Status: store.StatusCompleted},
		{Name: weekly.Name(2026, 1, 2), GameType: "Sudoku", Mode: "Easy", Status: store.StatusInProgress},
		{Name: weekly.Name(2026, 1, 3), GameType: "Sudoku", Mode: "Easy", Status: store.StatusCompleted},
	}

	rows := buildReviewWeeklyRows(games)
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Index != 3 || rows[1].Index != 1 {
		t.Fatalf("rows = %+v, want descending completed rows", rows)
	}
}

func TestImportAndActivateWeeklyReviewRecordSkipsSaveBinding(t *testing.T) {
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
		Name:      weekly.Name(2026, 1, 1),
		GameType:  "Lights Out",
		SaveState: string(save),
		Status:    store.StatusCompleted,
	}

	next, ok := (model{
		state:  weeklyView,
		width:  80,
		height: 24,
	}).importAndActivateRecordWithOptions(rec, gameOpenOptions{
		readOnly:    true,
		returnState: weeklyView,
	})
	if !ok {
		t.Fatal("expected import to succeed")
	}
	if next.session.activeGameID != 0 {
		t.Fatalf("activeGameID = %d, want 0", next.session.activeGameID)
	}
	if next.session.returnState != weeklyView {
		t.Fatalf("returnState = %d, want %d", next.session.returnState, weeklyView)
	}
}

func TestEnterOnSolvedLatestWeeklyCompletesAndQueuesNext(t *testing.T) {
	s := openAppTestStore(t)
	now := time.Now()
	year, weekNumber := now.ISOWeek()
	currentName := weekly.Name(year, weekNumber, 1)

	rec := &store.GameRecord{
		Name:         currentName,
		GameType:     "Lights Out",
		Mode:         "Easy",
		InitialState: "{}",
		SaveState:    "{}",
		Status:       store.StatusInProgress,
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
			weeklyAdvance:   &weekly.Info{Year: year, Week: weekNumber, Index: 1},
			completionSaved: false,
		},
		nav: navigationState{
			weeklyCursor: weekly.StartOfWeek(year, weekNumber, time.Local),
		},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := next.(model)

	if got.state != generatingView {
		t.Fatalf("state = %d, want %d (generatingView)", got.state, generatingView)
	}
	if got.session.spawn == nil {
		t.Fatal("expected spawn request")
	}
	if got.session.spawn.name != weekly.Name(year, weekNumber, 2) {
		t.Fatalf("spawn name = %q, want %q", got.session.spawn.name, weekly.Name(year, weekNumber, 2))
	}

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

func TestMoveWeeklyWeekDoesNotAdvancePastCurrentWeek(t *testing.T) {
	s := openAppTestStore(t)
	current := weekly.Current(time.Now())
	currentCursor := weekly.StartOfWeek(current.Year, current.Week, time.Local)

	m := model{
		store: s,
		nav: navigationState{
			weeklyCursor: currentCursor,
		},
	}

	got := m.moveWeeklyWeek(1)
	if !got.nav.weeklyCursor.Equal(currentCursor) {
		t.Fatalf("weeklyCursor = %v, want %v", got.nav.weeklyCursor, currentCursor)
	}
}

func TestCurrentWeeklyMenuIndexTracksNextPlayableSlot(t *testing.T) {
	s := openAppTestStore(t)
	year, weekNumber := time.Now().ISOWeek()

	for index := 1; index <= 2; index++ {
		rec := &store.GameRecord{
			Name:         weekly.Name(year, weekNumber, index),
			GameType:     "Nonogram",
			Mode:         "Easy",
			InitialState: "{}",
			SaveState:    "{}",
			Status:       store.StatusCompleted,
		}
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
	}

	m := model{store: s}
	if got := m.currentWeeklyMenuIndex(); got != 3 {
		t.Fatalf("currentWeeklyMenuIndex() = %d, want 3", got)
	}
}

func TestBuildPlayMenuItemsShowsCurrentWeeklyChallenge(t *testing.T) {
	now := time.Date(2026, time.March, 7, 12, 0, 0, 0, time.Local)

	items := buildPlayMenuItems(now, 7)
	if got := items[3].Desc; got != "Week 10-2026 # 7" {
		t.Fatalf("weekly desc = %q, want %q", got, "Week 10-2026 # 7")
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

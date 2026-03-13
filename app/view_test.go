package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"
)

func TestGameViewRequestsMouseCellMotion(t *testing.T) {
	m := model{
		state:   gameView,
		session: sessionState{game: &escapeTrackingGame{}},
	}

	v := m.View()
	if v.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("MouseMode = %v, want %v", v.MouseMode, tea.MouseModeCellMotion)
	}
	if !v.KeyboardEnhancements.ReportEventTypes {
		t.Fatal("expected keyboard event type reporting to remain enabled")
	}
}

func TestPlayMenuViewIncludesNotice(t *testing.T) {
	m := model{
		state:  playMenuView,
		width:  80,
		height: 24,
		nav: navigationState{
			playMenu: ui.NewMainMenu(nil),
		},
		notice: noticeState{level: noticeLevelError, message: "Could not load puzzle"},
	}

	if got := m.viewContent(); !strings.Contains(got, "Could not load puzzle") {
		t.Fatalf("viewContent() missing notice, got %q", got)
	}
}

func TestContinueViewStaysAtNaturalTableWidth(t *testing.T) {
	s := openAppTestStore(t)
	if err := s.CreateGame(&store.GameRecord{
		Name:     "shadow-trail",
		GameType: "Sudoku",
		Mode:     "Medium",
		Status:   store.StatusInProgress,
	}); err != nil {
		t.Fatal(err)
	}
	table, games := ui.InitContinueTable(s, 30)
	m := model{
		state:  continueView,
		width:  140,
		height: 30,
		cont:   continueState{table: table, games: games},
	}
	m = m.resizeActiveScreen()

	if got, want := m.cont.table.Width(), ui.ContinueTableWidth(); got != want {
		t.Fatalf("continue table width = %d, want %d", got, want)
	}
}

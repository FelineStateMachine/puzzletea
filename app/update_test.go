package app

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
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

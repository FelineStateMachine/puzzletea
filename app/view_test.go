package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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

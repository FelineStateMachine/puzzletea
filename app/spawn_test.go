package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

func TestHandleSpawnCompleteIgnoresStaleJob(t *testing.T) {
	called := false
	m := model{
		state:      generatingView,
		generating: true,
		spawnJobID: 7,
		spawnCancel: func() {
			called = true
		},
	}

	next, _ := m.handleSpawnComplete(6, game.SpawnCompleteMsg{Err: nil})
	got := next.(model)

	if !got.generating {
		t.Fatal("expected stale completion to be ignored")
	}
	if called {
		t.Fatal("stale completion should not cancel active spawn")
	}
}

func TestGeneratingEscapeCancelsActiveSpawn(t *testing.T) {
	called := false
	m := model{
		state:       generatingView,
		generating:  true,
		seedPending: true,
		spawnJobID:  3,
		spawnCancel: func() {
			called = true
		},
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)

	if !called {
		t.Fatal("expected spawn cancellation callback to run")
	}
	if got.generating {
		t.Fatal("expected generating to be false after escape")
	}
	if got.state != playMenuView {
		t.Fatalf("expected playMenuView after seeded generation escape, got %d", got.state)
	}
	if got.spawnCancel != nil {
		t.Fatal("expected spawnCancel to be cleared")
	}
}

package app

import (
	"context"
	"log"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

const spawnTimeout = 10 * time.Second

type spawnCompleteMsg struct {
	jobID  int64
	result game.SpawnCompleteMsg
}

// spawnCmd returns a tea.Cmd that runs Spawn() off the main goroutine.
func spawnCmd(spawner game.Spawner, ctx context.Context, jobID int64) tea.Cmd {
	return func() tea.Msg {
		var (
			g   game.Gamer
			err error
		)
		if cancellable, ok := spawner.(game.CancellableSpawner); ok {
			g, err = cancellable.SpawnContext(ctx)
		} else {
			g, err = spawner.Spawn()
		}
		return spawnCompleteMsg{
			jobID:  jobID,
			result: game.SpawnCompleteMsg{Game: g, Err: err},
		}
	}
}

// spawnSeededCmd returns a tea.Cmd that runs SpawnSeeded() off the main goroutine.
// The caller must not use rng after this call — *rand.Rand is not goroutine-safe
// and ownership is transferred to the spawned goroutine.
func spawnSeededCmd(spawner game.SeededSpawner, rng *rand.Rand, ctx context.Context, jobID int64) tea.Cmd {
	return func() tea.Msg {
		var (
			g   game.Gamer
			err error
		)
		if cancellable, ok := spawner.(game.CancellableSeededSpawner); ok {
			g, err = cancellable.SpawnSeededContext(ctx, rng)
		} else {
			g, err = spawner.SpawnSeeded(rng)
		}
		return spawnCompleteMsg{
			jobID:  jobID,
			result: game.SpawnCompleteMsg{Game: g, Err: err},
		}
	}
}

func (m *model) beginSpawnContext() (context.Context, int64) {
	m.cancelActiveSpawn()
	m.session.spawnJobID++
	jobID := m.session.spawnJobID
	ctx, cancel := context.WithTimeout(context.Background(), spawnTimeout)
	m.session.spawnCancel = cancel
	m.session.generating = true
	return ctx, jobID
}

func (m *model) cancelActiveSpawn() {
	if m.session.spawnCancel != nil {
		m.session.spawnCancel()
		m.session.spawnCancel = nil
	}
	if m.session.generating {
		// Invalidate late completion messages from a canceled job.
		m.session.spawnJobID++
	}
	m.session.generating = false
	m.session.spawn = nil
}

func (m model) handleSpawnComplete(jobID int64, msg game.SpawnCompleteMsg) (tea.Model, tea.Cmd) {
	if jobID != m.session.spawnJobID {
		return m, nil
	}

	m.session.generating = false
	if m.session.spawnCancel != nil {
		m.session.spawnCancel()
		m.session.spawnCancel = nil
	}
	request := m.session.spawn
	m.session.spawn = nil

	// If the user navigated away while generating, discard the result.
	if m.state != generatingView {
		return m, nil
	}

	if msg.Err != nil {
		log.Printf("failed to spawn game: %v", msg.Err)
		if request != nil {
			m.state = request.returnState
		}
		return m, nil
	}

	if request == nil {
		log.Printf("missing spawn request metadata")
		return m, nil
	}
	m = m.activateGame(msg.Game.SetTitle(request.name), 0, false)

	// Capture initial state and create DB record.
	rec, err := sessionflow.CreateRecord(
		m.store,
		m.session.game,
		request.name,
		request.gameType,
		request.modeTitle,
	)
	if err != nil {
		log.Printf("failed to create game record: %v", err)
	} else {
		m.session.activeGameID = rec.ID
	}
	return m, nil
}

// saveCurrentGame saves the current game state to the DB if a game is active.
func saveCurrentGame(m model, status store.GameStatus) model {
	if m.session.game == nil || m.session.activeGameID == 0 {
		return m
	}
	saveData, err := m.session.game.GetSave()
	if err != nil {
		log.Printf("failed to get save data: %v", err)
		return m
	}
	if err := m.store.UpdateSaveState(m.session.activeGameID, string(saveData)); err != nil {
		log.Printf("failed to update save state: %v", err)
	}
	// Don't overwrite a completed status when navigating away.
	if !(m.session.completionSaved && status != store.StatusCompleted) {
		if err := m.store.UpdateStatus(m.session.activeGameID, status); err != nil {
			log.Printf("failed to update game status: %v", err)
		}
	}
	m.session.activeGameID = 0
	m.session.game = nil
	return m
}

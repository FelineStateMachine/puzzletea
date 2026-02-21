package app

import (
	"context"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
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
	m.spawnJobID++
	jobID := m.spawnJobID
	ctx, cancel := context.WithTimeout(context.Background(), spawnTimeout)
	m.spawnCancel = cancel
	m.generating = true
	return ctx, jobID
}

func (m *model) cancelActiveSpawn() {
	if m.spawnCancel != nil {
		m.spawnCancel()
		m.spawnCancel = nil
	}
	if m.generating {
		// Invalidate late completion messages from a canceled job.
		m.spawnJobID++
	}
	m.generating = false
}

func (m model) handleSpawnComplete(jobID int64, msg game.SpawnCompleteMsg) (tea.Model, tea.Cmd) {
	if jobID != m.spawnJobID {
		return m, nil
	}

	m.generating = false
	if m.spawnCancel != nil {
		m.spawnCancel()
		m.spawnCancel = nil
	}
	isDaily := m.dailyPending
	isSeed := m.seedPending
	m.dailyPending = false
	m.seedPending = false

	// If the user navigated away while generating, discard the result.
	if m.state != generatingView {
		return m, nil
	}

	if msg.Err != nil {
		log.Printf("failed to spawn game: %v", msg.Err)
		if isDaily || isSeed {
			m.state = playMenuView
		} else {
			m.state = modeSelectView
		}
		return m, nil
	}

	var name, gameType, modeTitle string
	if isDaily {
		name = m.dailyName
		gameType = m.dailyGameType
		modeTitle = m.dailyModeTitle
	} else if isSeed {
		name = m.seedName
		gameType = m.seedGameType
		modeTitle = m.seedModeTitle
	} else {
		name = GenerateUniqueName(m.store)
		gameType = m.selectedCategory.Name
		modeTitle = m.selectedModeTitle
	}
	m = m.activateGame(msg.Game.SetTitle(name), 0, false)

	// Capture initial state and create DB record.
	initialState, err := m.game.GetSave()
	if err != nil {
		log.Printf("failed to get initial save: %v", err)
		return m, nil
	}
	rec := &store.GameRecord{
		Name:         name,
		GameType:     gameType,
		Mode:         modeTitle,
		InitialState: string(initialState),
		SaveState:    string(initialState),
		Status:       store.StatusNew,
	}
	if err := m.store.CreateGame(rec); err != nil {
		log.Printf("failed to create game record: %v", err)
	} else {
		m.activeGameID = rec.ID
	}
	return m, nil
}

// saveCurrentGame saves the current game state to the DB if a game is active.
func saveCurrentGame(m model, status store.GameStatus) model {
	if m.game == nil || m.activeGameID == 0 {
		return m
	}
	saveData, err := m.game.GetSave()
	if err != nil {
		log.Printf("failed to get save data: %v", err)
		return m
	}
	if err := m.store.UpdateSaveState(m.activeGameID, string(saveData)); err != nil {
		log.Printf("failed to update save state: %v", err)
	}
	// Don't overwrite a completed status when navigating away.
	if !(m.completionSaved && status != store.StatusCompleted) {
		if err := m.store.UpdateStatus(m.activeGameID, status); err != nil {
			log.Printf("failed to update game status: %v", err)
		}
	}
	m.activeGameID = 0
	m.game = nil
	return m
}

// GenerateUniqueName generates a unique adjective-noun name, retrying if needed.
func GenerateUniqueName(s *store.Store) string {
	for range 100 {
		name := namegen.Generate()
		exists, err := s.NameExists(name)
		if err != nil || !exists {
			return name
		}
	}
	// Fallback: append a number
	for i := 1; ; i++ {
		name := namegen.Generate() + "-" + strconv.Itoa(i)
		exists, err := s.NameExists(name)
		if err != nil || !exists {
			return name
		}
	}
}

package main

import (
	"log"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

// spawnCmd returns a tea.Cmd that runs Spawn() off the main goroutine.
func spawnCmd(spawner game.Spawner) tea.Cmd {
	return func() tea.Msg {
		g, err := spawner.Spawn()
		return game.SpawnCompleteMsg{Game: g, Err: err}
	}
}

// spawnSeededCmd returns a tea.Cmd that runs SpawnSeeded() off the main goroutine.
// The caller must not use rng after this call — *rand.Rand is not goroutine-safe
// and ownership is transferred to the spawned goroutine.
func spawnSeededCmd(spawner game.SeededSpawner, rng *rand.Rand) tea.Cmd {
	return func() tea.Msg {
		g, err := spawner.SpawnSeeded(rng)
		return game.SpawnCompleteMsg{Game: g, Err: err}
	}
}

func (m model) handleSpawnComplete(msg game.SpawnCompleteMsg) (tea.Model, tea.Cmd) {
	m.generating = false
	isDaily := m.dailyPending
	m.dailyPending = false

	// If the user navigated away while generating, discard the result.
	if m.state != generatingView {
		return m, nil
	}

	if msg.Err != nil {
		log.Printf("failed to spawn game: %v", msg.Err)
		if isDaily {
			m.state = mainMenuView
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
		m.game = msg.Game.SetTitle(name)
	} else {
		name = generateUniqueName(m.store)
		gameType = m.selectedCategory.Name
		modeTitle = m.mode.Title()
		m.game = msg.Game.SetTitle(name)
	}
	m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
	m.state = gameView
	m.completionSaved = false

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
	_ = m.store.UpdateSaveState(m.activeGameID, string(saveData))
	// Don't overwrite a completed status when navigating away.
	if !(m.completionSaved && status != store.StatusCompleted) {
		_ = m.store.UpdateStatus(m.activeGameID, status)
	}
	m.activeGameID = 0
	m.game = nil
	return m
}

// generateUniqueName generates a unique adjective-noun name, retrying if needed.
func generateUniqueName(s *store.Store) string {
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

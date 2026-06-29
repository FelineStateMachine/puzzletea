package app

import (
	"github.com/FelineStateMachine/puzzletea/game"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spawnCompleteMsg:
		cmd := newSessionController(&m).handleSpawnComplete(msg.jobID, msg.result)
		return m, cmd
	case game.SpawnCompleteMsg:
		cmd := newSessionController(&m).handleSpawnComplete(m.session.spawnJobID, msg)
		return m, cmd
	case exportCompleteMsg:
		next, cmd := m.handleExportComplete(msg)
		return next, cmd
	case exportSubmitAction:
		next, cmd := m.handleExportSubmit()
		return next, cmd
	case backAction:
		return msg.applyToModel(m)
	case tea.WindowSizeMsg:
		m = m.handleWindowSize(msg)
		if m.state != gameView {
			return m, nil
		}
	}

	next, cmd, handled := m.handleGlobalKey(msg)
	if handled {
		return next, cmd
	}
	m = next

	if m.state == gameView {
		gameCmd := newSessionController(&m).updateActiveGame(msg)
		return m, gameCmd
	}

	screen := m.activeScreen()
	if screen == nil {
		return m, nil
	}

	// Clear stale notices when the user interacts with the export form.
	if m.state == exportView {
		switch msg.(type) {
		case tea.KeyPressMsg, tea.MouseClickMsg:
			m = m.clearNotice()
		}
	}

	nextScreen, screenCmd, action := screen.Update(msg)
	if m.screens == nil {
		m.screens = make(map[viewState]screenModel)
	}
	m.screens[m.state] = nextScreen
	if action == nil {
		return m, screenCmd
	}

	next, actionCmd := m.handleScreenAction(action)
	return next, tea.Batch(screenCmd, actionCmd)
}

// handleGlobalKey dispatches key messages to the domain that owns the current
// state. Pending generation and active game keys are handled by the session
// controller; pending export keys by the export workflow; normal screen keys
// by shell-level handling. Non-key messages are never consumed here.
func (m model) handleGlobalKey(msg tea.Msg) (model, tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}

	if m.state == generatingView {
		return newSessionController(&m).handleGeneratingKey(keyMsg)
	}

	if m.state == exportRunningView {
		return m.handleExportRunningKey(keyMsg)
	}

	if m.state == gameView {
		return newSessionController(&m).handleGameKey(keyMsg)
	}

	// Normal screen shell keys.
	switch {
	case key.Matches(keyMsg, rootKeys.Quit):
		return m, tea.Quit, true
	case key.Matches(keyMsg, rootKeys.Debug):
		m.debug.enabled = !m.debug.enabled
		return m, nil, true
	case key.Matches(keyMsg, rootKeys.FullHelp):
		m.help.showFull = !m.help.showFull
		return m, nil, true
	default:
		return m, nil, false
	}
}

func (m model) handleScreenAction(action screenAction) (model, tea.Cmd) {
	return action.applyToModel(m)
}

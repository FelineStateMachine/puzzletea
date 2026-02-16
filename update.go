package main

import (
	"log"
	"time"

	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case game.SpawnCompleteMsg:
		return m.handleSpawnComplete(msg)

	case tea.WindowSizeMsg:
		h, v := ui.RootStyle.GetFrameSize()
		w, ht := msg.Width-h, msg.Height-v
		m.width = w
		m.height = ht
		menuW, menuH := min(w, 64), min(ht, 24)
		m.mainMenuList.SetSize(menuW, menuH)
		m.gameSelectList.SetSize(menuW, menuH)
		if m.state == modeSelectView {
			m.modeSelectList.SetSize(menuW, menuH)
		}
		if m.state == continueView {
			m.continueTable.SetWidth(w)
			m.continueTable.SetHeight(ht)
		}
		if m.state == helpSelectView {
			m.helpSelectList.SetSize(menuW, menuH)
		}
		if m.state == helpDetailView {
			m.helpViewport.Width = w
			m.helpViewport.Height = ht - 2
		}

	case tea.KeyMsg:
		// During generation, only allow Escape (to cancel) and Ctrl+C (to quit).
		if m.state == generatingView {
			switch msg.Type {
			case tea.KeyEscape:
				if m.dailyPending {
					m.dailyPending = false
					m.state = mainMenuView
				} else {
					m.state = modeSelectView
				}
				return m, nil
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.Type {
		case tea.KeyCtrlN:
			m = saveCurrentGame(m, store.StatusInProgress)
			m.state = mainMenuView
			m.debug = false
		case tea.KeyEnter:
			if m.state != gameView {
				return m.handleEnter()
			}
		case tea.KeyEscape:
			return m.handleEscape()
		case tea.KeyCtrlC:
			m = saveCurrentGame(m, store.StatusAbandoned)
			return m, tea.Quit
		case tea.KeyCtrlE:
			m.debug = !m.debug
		case tea.KeyCtrlH:
			m.showFullHelp = !m.showFullHelp
			if m.state == gameView && m.game != nil {
				m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
			}
		case tea.KeyCtrlR:
			if m.state == gameView && m.game != nil {
				m.game = m.game.Reset()
			}
		}
	}

	switch m.state {
	case mainMenuView:
		m.mainMenuList, cmd = m.mainMenuList.Update(msg)
	case generatingView:
		m.spinner, cmd = m.spinner.Update(msg)
	case gameView:
		m.game, cmd = m.game.Update(msg)
		if m.debug {
			m.debuginfo = m.renderDebugInfo()
		}
		if !m.completionSaved && m.game.IsSolved() {
			m.completionSaved = true
			saveData, err := m.game.GetSave()
			if err == nil {
				_ = m.store.UpdateSaveState(m.activeGameID, string(saveData))
			}
			_ = m.store.UpdateStatus(m.activeGameID, store.StatusCompleted)
		}
	case gameSelectView:
		m.gameSelectList, cmd = m.gameSelectList.Update(msg)
	case modeSelectView:
		m.modeSelectList, cmd = m.modeSelectList.Update(msg)
	case continueView:
		m.continueTable, cmd = m.continueTable.Update(msg)
	case helpSelectView:
		m.helpSelectList, cmd = m.helpSelectList.Update(msg)
	case helpDetailView:
		m.helpViewport, cmd = m.helpViewport.Update(msg)
	}

	return m, cmd
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case mainMenuView:
		return m.handleMainMenuEnter()
	case gameSelectView:
		return m.handleGameSelectEnter()
	case modeSelectView:
		return m.handleModeSelectEnter()
	case continueView:
		return m.handleContinueEnter()
	case helpSelectView:
		return m.handleHelpSelectEnter()
	}
	return m, nil
}

func (m model) handleMainMenuEnter() (tea.Model, tea.Cmd) {
	item, ok := m.mainMenuList.SelectedItem().(ui.MenuItem)
	if !ok {
		return m, nil
	}
	switch item.Title() {
	case "Daily Puzzle":
		return m.handleDailyPuzzle()
	case "Generate":
		m.state = gameSelectView
	case "Continue":
		m.continueTable, m.continueGames = ui.InitContinueTable(m.store, m.height)
		m.state = continueView
	case "Guides":
		m.helpSelectList = ui.InitList(GameCategories, "How to Play")
		m.helpSelectList.SetSize(min(m.width, 64), min(m.height, 24))
		m.state = helpSelectView
	case "Quit":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleDailyPuzzle() (tea.Model, tea.Cmd) {
	today := time.Now()
	rng := daily.RNG(today)
	name := daily.Name(today, rng)

	// Check if a daily game already exists for today.
	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		log.Printf("failed to check daily game: %v", err)
		return m, nil
	}

	if rec != nil {
		// Daily exists - resume or review it.
		importFn, ok := game.Registry[rec.GameType]
		if !ok {
			log.Printf("unknown game type for daily: %s", rec.GameType)
			return m, nil
		}
		g, err := importFn([]byte(rec.SaveState))
		if err != nil {
			log.Printf("failed to import daily game: %v", err)
			return m, nil
		}
		m.game = g.SetTitle(rec.Name)
		m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
		m.activeGameID = rec.ID
		m.state = gameView
		m.completionSaved = rec.Status == store.StatusCompleted

		// Resume abandoned dailies by resetting their status.
		if rec.Status == store.StatusAbandoned {
			_ = m.store.UpdateStatus(rec.ID, store.StatusInProgress)
		}
		return m, nil
	}

	// No existing daily - generate a new one.
	spawner, gameType, modeTitle := daily.Mode(rng)
	m.dailyPending = true
	m.dailyName = name
	m.dailyGameType = gameType
	m.dailyModeTitle = modeTitle
	m.state = generatingView
	m.generating = true
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng))
}

func (m model) handleGameSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.gameSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.selectedCategory = cat
	m.modeSelectList = ui.InitList(cat.Modes, cat.Name+" - Select Mode")
	m.modeSelectList.SetSize(m.gameSelectList.Width(), m.gameSelectList.Height())
	m.state = modeSelectView
	return m, nil
}

func (m model) handleModeSelectEnter() (tea.Model, tea.Cmd) {
	m.mode, _ = m.modeSelectList.SelectedItem().(game.Mode)
	spawner, ok := m.mode.(game.Spawner)
	if !ok {
		return m, nil
	}
	m.state = generatingView
	m.generating = true
	return m, tea.Batch(m.spinner.Tick, spawnCmd(spawner))
}

func (m model) handleContinueEnter() (tea.Model, tea.Cmd) {
	idx := m.continueTable.Cursor()
	if idx < 0 || idx >= len(m.continueGames) {
		return m, nil
	}
	rec := m.continueGames[idx]
	importFn, ok := game.Registry[rec.GameType]
	if !ok {
		return m, nil
	}
	g, err := importFn([]byte(rec.SaveState))
	if err != nil {
		log.Printf("failed to import game: %v", err)
		return m, nil
	}
	m.game = g.SetTitle(rec.Name)
	m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
	m.activeGameID = rec.ID
	m.state = gameView
	m.completionSaved = rec.Status == store.StatusCompleted
	return m, nil
}

func (m model) handleHelpSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.helpSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.helpCategory = cat

	if m.helpRenderer == nil || m.helpRendererWidth != m.width {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width),
		)
		if err != nil {
			log.Printf("failed to create help renderer: %v", err)
			return m, nil
		}
		m.helpRenderer = renderer
		m.helpRendererWidth = m.width
	}
	rendered, err := m.helpRenderer.Render(cat.Help)
	if err != nil {
		log.Printf("failed to render help: %v", err)
		return m, nil
	}

	m.helpViewport = viewport.New(m.width, m.height-2)
	m.helpViewport.SetContent(rendered)
	m.state = helpDetailView
	return m, nil
}

func (m model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.state {
	case generatingView:
		m.state = modeSelectView
	case modeSelectView:
		m.state = gameSelectView
	case gameSelectView, continueView:
		m.state = mainMenuView
	case helpDetailView:
		m.state = helpSelectView
	case helpSelectView:
		m.state = mainMenuView
	}
	return m, nil
}

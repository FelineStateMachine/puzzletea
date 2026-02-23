package app

import (
	"log"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/resolve"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
)

const (
	helpPanelInsetX         = 2
	helpPanelInsetY         = 1
	helpPanelHorizontalTrim = 6
	helpPanelVerticalTrim   = 8
)

func helpViewportSize(width, height int) (int, int) {
	panelWidth := max(width-(helpPanelInsetX*2), 1)
	panelHeight := max(height-(helpPanelInsetY*2), 1)
	contentWidth := max(panelWidth-helpPanelHorizontalTrim, 1)
	contentHeight := max(panelHeight-helpPanelVerticalTrim, 1)
	return contentWidth, contentHeight
}

func statsViewportSize(width, height int, cards []stats.Card) (int, int) {
	contentWidth, _ := helpViewportSize(width, height)
	panelHeight := max(height-(helpPanelInsetY*2), 1)
	contentHeight := max(panelHeight-stats.StaticHeight(cards), 1)
	return contentWidth, contentHeight
}

func (m model) updateHelpDetailViewport() model {
	helpWidth, helpHeight := helpViewportSize(m.width, m.height)
	if m.helpRenderer == nil || m.helpRendererWidth != helpWidth {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(helpWidth),
		)
		if err != nil {
			log.Printf("failed to create help renderer: %v", err)
			m.helpRenderer = nil
			m.helpRendererWidth = 0
		} else {
			m.helpRenderer = renderer
			m.helpRendererWidth = helpWidth
		}
	}

	rendered := m.helpCategory.Help
	if m.helpRenderer != nil {
		out, err := m.helpRenderer.Render(m.helpCategory.Help)
		if err != nil {
			log.Printf("failed to render help: %v", err)
		} else {
			rendered = out
		}
	}

	m.helpViewport = viewport.New(
		viewport.WithWidth(helpWidth),
		viewport.WithHeight(helpHeight),
	)
	m.helpViewport.SetContent(rendered)
	return m
}

func (m model) updateStatsViewport() model {
	statsWidth, statsHeight := statsViewportSize(m.width, m.height, m.statsCards)
	m.statsViewport.SetWidth(statsWidth)
	m.statsViewport.SetHeight(statsHeight)
	m.statsViewport.SetContent(stats.RenderCardGrid(m.statsCards, statsWidth))
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spawnCompleteMsg:
		return m.handleSpawnComplete(msg.jobID, msg.result)
	case game.SpawnCompleteMsg:
		// Backward compatibility for callers still using the old message type.
		return m.handleSpawnComplete(m.spawnJobID, msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		menuW := min(m.width, 64)
		m.gameSelectList.SetSize(menuW, min(m.height, ui.ListHeight(m.gameSelectList)))
		if m.state == seedInputView {
			m.seedInput.SetWidth(min(m.width, 48))
		}
		if m.state == modeSelectView {
			m.modeSelectList.SetSize(menuW, min(m.height, ui.ListHeight(m.modeSelectList)))
		}
		if m.state == continueView {
			m.continueTable.SetWidth(m.width)
			visibleRows := min(len(m.continueGames), ui.MaxTableRows)
			m.continueTable.SetHeight(min(m.height, visibleRows))
		}
		if m.state == helpSelectView {
			m.helpSelectList.SetSize(menuW, min(m.height, ui.ListHeight(m.helpSelectList)))
		}
		if m.state == helpDetailView {
			m = m.updateHelpDetailViewport()
		}
		if m.state == themeSelectView {
			const maxVisibleItems = 8
			listW := min(menuW, theme.MaxNameLen+4)
			m.themeList.SetSize(listW, min(m.height, maxVisibleItems*3))
		}
		if m.state == statsView {
			m = m.updateStatsViewport()
		}

	case tea.KeyPressMsg:
		// During generation, only allow Escape (to cancel) and Ctrl+C (to quit).
		if m.state == generatingView {
			switch {
			case key.Matches(msg, rootKeys.Escape):
				m.cancelActiveSpawn()
				if m.dailyPending || m.seedPending {
					m.dailyPending = false
					m.seedPending = false
					m.state = playMenuView
				} else {
					m.state = modeSelectView
				}
				return m, nil
			case key.Matches(msg, rootKeys.Quit):
				return m, tea.Quit
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, rootKeys.MainMenu):
			m = saveCurrentGame(m, store.StatusInProgress)
			m.state = mainMenuView
			m.debug = false
		case key.Matches(msg, rootKeys.Enter):
			if m.state != gameView {
				return m.handleEnter()
			}
		case key.Matches(msg, rootKeys.Escape):
			if m.state != gameView {
				return m.handleEscape()
			}
		case key.Matches(msg, rootKeys.Quit):
			m = saveCurrentGame(m, store.StatusAbandoned)
			return m, tea.Quit
		case key.Matches(msg, rootKeys.Debug):
			m.debug = !m.debug
		case key.Matches(msg, rootKeys.FullHelp):
			m.showFullHelp = !m.showFullHelp
			if m.state == gameView && m.game != nil {
				m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
			}
		case key.Matches(msg, rootKeys.ResetGame):
			if m.state == gameView && m.game != nil {
				m.game = m.game.Reset()
			}
		}
	}

	switch m.state {
	case mainMenuView:
		updateMainMenuCursor(msg, &m.mainMenu)
	case playMenuView:
		updateMainMenuCursor(msg, &m.playMenu)
	case optionsMenuView:
		updateMainMenuCursor(msg, &m.optionsMenu)
	case seedInputView:
		m.seedInput, cmd = m.seedInput.Update(msg)
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
				if err := m.store.UpdateSaveState(m.activeGameID, string(saveData)); err != nil {
					log.Printf("failed to save completion state: %v", err)
				}
			}
			if err := m.store.UpdateStatus(m.activeGameID, store.StatusCompleted); err != nil {
				log.Printf("failed to mark game completed: %v", err)
			}
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
	case statsView:
		m.statsViewport, cmd = m.statsViewport.Update(msg)
	case themeSelectView:
		prev := m.themeList.Index()
		m.themeList, cmd = m.themeList.Update(msg)
		// Live preview: apply theme when cursor moves.
		if m.themeList.Index() != prev {
			if item, ok := m.themeList.SelectedItem().(ui.MenuItem); ok {
				_ = theme.Apply(item.ItemTitle)
				ui.UpdateThemeListStyles(&m.themeList)
			}
		}
	}

	return m, cmd
}

func updateMainMenuCursor(msg tea.Msg, menu *ui.MainMenu) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}
	switch keyMsg.String() {
	case "up", "k":
		menu.CursorUp()
	case "down", "j":
		menu.CursorDown()
	}
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case mainMenuView:
		return m.handleMainMenuEnter()
	case playMenuView:
		return m.handlePlayMenuEnter()
	case optionsMenuView:
		return m.handleOptionsMenuEnter()
	case seedInputView:
		return m.handleSeedConfirm()
	case gameSelectView:
		return m.handleGameSelectEnter()
	case modeSelectView:
		return m.handleModeSelectEnter()
	case continueView:
		return m.handleContinueEnter()
	case helpSelectView:
		return m.handleHelpSelectEnter()
	case themeSelectView:
		return m.handleThemeConfirm()
	}
	return m, nil
}

func (m model) handleMainMenuEnter() (tea.Model, tea.Cmd) {
	item := m.mainMenu.Selected()
	switch item.Title() {
	case "Play":
		m.playMenu = ui.NewMainMenu(playMenuItems)
		m.state = playMenuView
	case "Stats":
		return m.handleStatsEnter()
	case "Options":
		m.optionsMenu = ui.NewMainMenu(optionsMenuItems)
		m.state = optionsMenuView
	case "Quit":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handlePlayMenuEnter() (tea.Model, tea.Cmd) {
	item := m.playMenu.Selected()
	switch item.Title() {
	case "Daily Puzzle":
		return m.handleDailyPuzzle()
	case "Generate":
		m.state = gameSelectView
	case "Continue":
		m.continueTable, m.continueGames = ui.InitContinueTable(m.store, m.height)
		m.state = continueView
	case "Enter Seed":
		ti := textinput.New()
		ti.Placeholder = "any word or phrase"
		ti.CharLimit = 64
		ti.SetWidth(min(m.width, 48))
		m.seedInput = ti
		m.state = seedInputView
		return m, m.seedInput.Focus()
	}
	return m, nil
}

func (m model) handleOptionsMenuEnter() (tea.Model, tea.Cmd) {
	item := m.optionsMenu.Selected()
	switch item.Title() {
	case "Theme":
		return m.handleThemeEnter()
	case "Guides":
		m.helpSelectList = ui.InitList(GameCategories, "How to Play")
		m.helpSelectList.SetSize(min(m.width, 64), min(m.height, ui.ListHeight(m.helpSelectList)))
		m.state = helpSelectView
	}
	return m, nil
}

func (m model) handleSeedConfirm() (tea.Model, tea.Cmd) {
	seed := strings.TrimSpace(m.seedInput.Value())
	if seed == "" {
		return m, nil
	}

	// Prevent seeded puzzles from mimicking daily puzzle names by
	// silently lowercasing a "Daily" prefix so titles never start
	// with the real daily prefix "Daily ".
	if strings.HasPrefix(strings.ToLower(seed), "daily") {
		seed = strings.ToLower(seed[:len("daily")]) + seed[len("daily"):]
	}

	// Deterministic name from seed, using an independent sub-RNG so
	// namegen changes don't affect mode selection or puzzle content.
	nameRNG := resolve.RNGFromString("name:" + seed)
	name := seed + " - " + namegen.GenerateSeeded(nameRNG)

	// Check if a game with this seed-derived name already exists.
	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		log.Printf("failed to check seeded game: %v", err)
		return m, nil
	}

	if rec != nil {
		// Seed game exists - resume it.
		var resumed bool
		m, resumed = m.importAndActivateRecord(*rec)

		// Resume abandoned seeded games by resetting their status.
		if resumed && rec.Status == store.StatusAbandoned {
			if err := m.store.UpdateStatus(rec.ID, store.StatusInProgress); err != nil {
				log.Printf("failed to mark seeded game in progress: %v", err)
			}
		}
		return m, nil
	}

	// No existing game — generate a new seeded puzzle.
	spawner, gameType, modeTitle, err := resolve.SeededMode(seed, GameCategories)
	if err != nil {
		log.Printf("failed to select seeded mode: %v", err)
		return m, nil
	}

	rng := resolve.RNGFromString(seed)
	ctx, jobID := m.beginSpawnContext()
	m.seedPending = true
	m.seedName = name
	m.seedGameType = gameType
	m.seedModeTitle = modeTitle
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng, ctx, jobID))
}

func (m model) handleDailyPuzzle() (tea.Model, tea.Cmd) {
	today := time.Now()
	name := daily.Name(today)

	// Check if a daily game already exists for today.
	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		log.Printf("failed to check daily game: %v", err)
		return m, nil
	}

	if rec != nil {
		// Daily exists - resume or review it.
		var resumed bool
		m, resumed = m.importAndActivateRecord(*rec)

		// Resume abandoned dailies by resetting their status.
		if resumed && rec.Status == store.StatusAbandoned {
			if err := m.store.UpdateStatus(rec.ID, store.StatusInProgress); err != nil {
				log.Printf("failed to mark daily game in progress: %v", err)
			}
		}
		return m, nil
	}

	// No existing daily - generate a new one.
	// Mode selection uses rendezvous hashing (independent of RNG).
	// RNG is reserved purely for puzzle content generation.
	spawner, gameType, modeTitle := daily.Mode(today)
	if spawner == nil {
		log.Printf("no daily mode available for %s", today.Format("2006-01-02"))
		return m, nil
	}
	rng := daily.RNG(today)
	ctx, jobID := m.beginSpawnContext()
	m.dailyPending = true
	m.dailyName = name
	m.dailyGameType = gameType
	m.dailyModeTitle = modeTitle
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng, ctx, jobID))
}

func (m model) handleGameSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.gameSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.selectedCategory = cat
	m.modeSelectList = ui.InitList(cat.Modes, cat.Name+" - Select Mode")
	m.modeSelectList.SetSize(min(m.width, 64), min(m.height, ui.ListHeight(m.modeSelectList)))
	m.state = modeSelectView
	return m, nil
}

func (m model) handleModeSelectEnter() (tea.Model, tea.Cmd) {
	item := m.modeSelectList.SelectedItem()
	mode, ok := item.(game.Mode)
	if !ok {
		return m, nil
	}
	spawner, ok := item.(game.Spawner)
	if !ok {
		return m, nil
	}
	m.selectedModeTitle = mode.Title()
	ctx, jobID := m.beginSpawnContext()
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnCmd(spawner, ctx, jobID))
}

func (m model) handleContinueEnter() (tea.Model, tea.Cmd) {
	idx := m.continueTable.Cursor()
	if idx < 0 || idx >= len(m.continueGames) {
		return m, nil
	}
	rec := m.continueGames[idx]
	m, _ = m.importAndActivateRecord(rec)
	return m, nil
}

func (m model) handleHelpSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.helpSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.helpCategory = cat
	m = m.updateHelpDetailViewport()
	m.state = helpDetailView
	return m, nil
}

func (m model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.state {
	case playMenuView:
		m.state = mainMenuView
	case optionsMenuView:
		m.state = mainMenuView
	case seedInputView:
		m.state = playMenuView
	case generatingView:
		m.cancelActiveSpawn()
		m.state = modeSelectView
	case modeSelectView:
		m.state = gameSelectView
	case gameSelectView, continueView:
		m.state = playMenuView
	case helpDetailView:
		m.state = helpSelectView
	case helpSelectView:
		m.state = optionsMenuView
	case statsView:
		m.state = mainMenuView
	case themeSelectView:
		// Revert to the previously saved theme.
		_ = theme.Apply(m.previousTheme)
		m.state = optionsMenuView
	}
	return m, nil
}

func (m model) handleThemeEnter() (tea.Model, tea.Cmd) {
	m.previousTheme = m.cfg.Theme

	names := theme.ThemeNames()
	items := make([]list.Item, len(names))
	for i, n := range names {
		desc := "dark theme"
		if n == theme.DefaultThemeName {
			desc = "built-in earth-tone palette"
		} else if t := theme.LookupTheme(n); t != nil && !t.Meta.IsDark {
			desc = "light theme"
		}
		items[i] = ui.MenuItem{ItemTitle: n, Desc: desc}
	}

	// Cap list height to show ~8 items (like the continue table).
	// Default delegate: height=2, spacing=1 → 3 lines per item.
	const maxVisibleItems = 8
	listH := min(m.height, maxVisibleItems*3)
	listW := min(m.width, theme.MaxNameLen+4)

	m.themeList = ui.InitThemeList(items, listW, listH)

	// Position cursor on the currently active theme.
	for i, item := range items {
		if mi, ok := item.(ui.MenuItem); ok && mi.ItemTitle == m.previousTheme {
			m.themeList.Select(i)
			break
		}
	}
	// If current theme is default (empty string), select index 0.
	if m.previousTheme == "" {
		m.themeList.Select(0)
	}

	m.state = themeSelectView
	return m, nil
}

func (m model) handleThemeConfirm() (tea.Model, tea.Cmd) {
	item, ok := m.themeList.SelectedItem().(ui.MenuItem)
	if !ok {
		return m, nil
	}

	themeName := item.ItemTitle
	if themeName == theme.DefaultThemeName {
		themeName = ""
	}

	// Apply the theme (already live-previewed, but ensure it's set).
	_ = theme.Apply(item.ItemTitle)

	// Persist the selection.
	m.cfg.Theme = themeName
	if err := m.cfg.Save(config.DefaultPath()); err != nil {
		log.Printf("failed to save config: %v", err)
	}

	m.state = mainMenuView
	return m, nil
}

func (m model) handleStatsEnter() (tea.Model, tea.Cmd) {
	catStats, err := m.store.GetCategoryStats()
	if err != nil {
		log.Printf("failed to get category stats: %v", err)
		return m, nil
	}
	modeStats, err := m.store.GetModeStats()
	if err != nil {
		log.Printf("failed to get mode stats: %v", err)
		return m, nil
	}
	streakDates, err := m.store.GetDailyStreakDates()
	if err != nil {
		log.Printf("failed to get daily streak dates: %v", err)
		return m, nil
	}
	now := time.Now()
	currentDaily := false
	rec, err := m.store.GetDailyGame(daily.Name(now))
	if err != nil {
		log.Printf("failed to check current daily: %v", err)
	} else {
		currentDaily = rec != nil
	}

	m.statsCards = stats.BuildCards(catStats, modeStats)
	m.statsProfile = stats.BuildProfileBanner(catStats, modeStats, streakDates, currentDaily)

	statsWidth, statsHeight := statsViewportSize(m.width, m.height, m.statsCards)
	m.statsViewport = viewport.New(
		viewport.WithWidth(statsWidth),
		viewport.WithHeight(statsHeight),
	)
	m.statsViewport.SetContent(stats.RenderCardGrid(m.statsCards, statsWidth))
	m.state = statsView
	return m, nil
}

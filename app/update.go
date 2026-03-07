package app

import (
	"log"
	"time"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/resolve"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
)

const (
	helpPanelInsetX         = 2
	helpPanelInsetY         = 1
	helpPanelHorizontalTrim = 6
	helpPanelVerticalTrim   = 8
	categoryPanelChrome     = 8
	categoryBodyMaxWidth    = 86
	categoryBodyMaxHeight   = 16
	categoryMinListWidth    = 24
	categoryMaxListWidth    = 30
	categoryGapWidth        = 2
	categoryDetailTrimX     = 6
	categoryDetailTrimY     = 4
	categoryStackGapHeight  = 1
	categoryMinSideBySideW  = 72
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

type categoryPickerMetrics struct {
	bodyWidth    int
	bodyHeight   int
	listWidth    int
	listHeight   int
	detailWidth  int
	detailHeight int
	stacked      bool
}

func categoryPickerSize(width, height int) categoryPickerMetrics {
	bodyWidth := min(width, categoryBodyMaxWidth)
	bodyHeight := min(max(height-categoryPanelChrome, 1), categoryBodyMaxHeight)

	if bodyWidth < categoryMinSideBySideW {
		listHeight := min(bodyHeight, categoryPickerListHeight())
		detailHeight := max(bodyHeight-listHeight-categoryStackGapHeight, 1)
		if detailHeight == 1 && bodyHeight > 1 {
			listHeight = max(bodyHeight-categoryStackGapHeight-detailHeight, 1)
		}
		return categoryPickerMetrics{
			bodyWidth:    bodyWidth,
			bodyHeight:   bodyHeight,
			listWidth:    bodyWidth,
			listHeight:   listHeight,
			detailWidth:  bodyWidth,
			detailHeight: detailHeight,
			stacked:      true,
		}
	}

	listWidth := min(categoryMaxListWidth, max(categoryMinListWidth, bodyWidth/3))
	detailWidth := max(bodyWidth-listWidth-categoryGapWidth, 1)
	return categoryPickerMetrics{
		bodyWidth:    bodyWidth,
		bodyHeight:   bodyHeight,
		listWidth:    listWidth,
		listHeight:   bodyHeight,
		detailWidth:  detailWidth,
		detailHeight: bodyHeight,
	}
}

func selectedCategoryName(item list.Item) string {
	cat, ok := item.(game.Category)
	if !ok {
		return ""
	}
	return cat.Name
}

func activeFilterList(m model) *list.Model {
	switch m.state {
	case gameSelectView:
		return &m.nav.gameSelectList
	case themeSelectView:
		return &m.theme.list
	default:
		return nil
	}
}

func (m model) updateCategoryDetailViewport() model {
	metrics := categoryPickerSize(m.width, m.height)
	contentWidth := max(metrics.detailWidth-categoryDetailTrimX, 1)
	contentHeight := max(metrics.detailHeight-categoryDetailTrimY, 1)

	if m.nav.categoryDetail.Width() == 0 || m.nav.categoryDetail.Height() == 0 {
		m.nav.categoryDetail = viewport.New(
			viewport.WithWidth(contentWidth),
			viewport.WithHeight(contentHeight),
		)
	}
	m.nav.categoryDetail.SetWidth(contentWidth)
	m.nav.categoryDetail.SetHeight(contentHeight)
	m.nav.categoryDetail.FillHeight = true

	cat, ok := m.nav.gameSelectList.SelectedItem().(game.Category)
	if !ok {
		m.nav.categoryDetail.SetContent("")
		return m
	}

	m.nav.categoryDetail.SetContent(renderCategoryDetailContent(cat, contentWidth))
	m.nav.categoryDetail.GotoTop()
	return m
}

func (m model) updateHelpDetailViewport() model {
	helpWidth, helpHeight := helpViewportSize(m.width, m.height)
	if m.help.renderer == nil || m.help.rendererWidth != helpWidth {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(helpWidth),
		)
		if err != nil {
			log.Printf("failed to create help renderer: %v", err)
			m.help.renderer = nil
			m.help.rendererWidth = 0
		} else {
			m.help.renderer = renderer
			m.help.rendererWidth = helpWidth
		}
	}

	rendered := m.help.category.Help
	if m.help.renderer != nil {
		out, err := m.help.renderer.Render(m.help.category.Help)
		if err != nil {
			log.Printf("failed to render help: %v", err)
		} else {
			rendered = out
		}
	}

	m.help.viewport = viewport.New(
		viewport.WithWidth(helpWidth),
		viewport.WithHeight(helpHeight),
	)
	m.help.viewport.SetContent(rendered)
	return m
}

func (m model) updateStatsViewport() model {
	statsWidth, statsHeight := statsViewportSize(m.width, m.height, m.stats.cards)
	m.stats.viewport.SetWidth(statsWidth)
	m.stats.viewport.SetHeight(statsHeight)
	m.stats.viewport.SetContent(stats.RenderCardGrid(m.stats.cards, statsWidth))
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spawnCompleteMsg:
		return m.handleSpawnComplete(msg.jobID, msg.result)
	case game.SpawnCompleteMsg:
		// Backward compatibility for callers still using the old message type.
		return m.handleSpawnComplete(m.session.spawnJobID, msg)
	case tea.WindowSizeMsg:
		m = m.handleWindowSize(msg)
	case tea.KeyPressMsg:
		next, keyCmd, done := m.handleGlobalKey(msg)
		if done {
			return next, keyCmd
		}
		m = next
	}

	switch m.state {
	case mainMenuView:
		updateMainMenuCursor(msg, &m.nav.mainMenu)
	case playMenuView:
		updateMainMenuCursor(msg, &m.nav.playMenu)
	case optionsMenuView:
		updateMainMenuCursor(msg, &m.nav.optionsMenu)
	case seedInputView:
		m, cmd = m.handleSeedInputUpdate(msg)
	case generatingView:
		m.spinner, cmd = m.spinner.Update(msg)
	case gameView:
		m.session.game, cmd = m.session.game.Update(msg)
		if m.debug.enabled {
			m.debug.info = m.renderDebugInfo()
		}
		m = m.persistCompletionIfSolved()
	case gameSelectView:
		prev := selectedCategoryName(m.nav.gameSelectList.SelectedItem())
		m.nav.gameSelectList, cmd = m.nav.gameSelectList.Update(msg)
		if selectedCategoryName(m.nav.gameSelectList.SelectedItem()) != prev {
			m = m.updateCategoryDetailViewport()
		}
	case modeSelectView:
		m.nav.modeSelectList, cmd = m.nav.modeSelectList.Update(msg)
	case continueView:
		m.nav.continueTable, cmd = m.nav.continueTable.Update(msg)
	case helpSelectView:
		m.nav.helpSelectList, cmd = m.nav.helpSelectList.Update(msg)
	case helpDetailView:
		m.help.viewport, cmd = m.help.viewport.Update(msg)
	case statsView:
		m.stats.viewport, cmd = m.stats.viewport.Update(msg)
	case themeSelectView:
		prev := m.theme.list.Index()
		m.theme.list, cmd = m.theme.list.Update(msg)
		if m.theme.list.Index() != prev {
			if item, ok := m.theme.list.SelectedItem().(ui.MenuItem); ok {
				_ = theme.Apply(item.ItemTitle)
				ui.UpdateThemeListStyles(&m.theme.list)
			}
		}
	}

	return m, cmd
}

func (m model) handleWindowSize(msg tea.WindowSizeMsg) model {
	m.width = msg.Width
	m.height = msg.Height
	menuW := min(m.width, 64)
	metrics := categoryPickerSize(m.width, m.height)
	m.nav.gameSelectList.SetSize(metrics.listWidth, metrics.listHeight)
	m = m.updateCategoryDetailViewport()
	if m.state == seedInputView {
		m.nav.seedInput.SetWidth(min(m.width, 48))
	}
	if m.state == modeSelectView {
		m.nav.modeSelectList.SetSize(menuW, min(m.height, ui.ListHeight(m.nav.modeSelectList)))
	}
	if m.state == continueView {
		m.nav.continueTable.SetWidth(m.width)
		visibleRows := min(len(m.nav.continueGames), ui.MaxTableRows)
		m.nav.continueTable.SetHeight(min(m.height, visibleRows))
	}
	if m.state == helpSelectView {
		m.nav.helpSelectList.SetSize(menuW, min(m.height, ui.ListHeight(m.nav.helpSelectList)))
	}
	if m.state == helpDetailView {
		m = m.updateHelpDetailViewport()
	}
	if m.state == themeSelectView {
		const maxVisibleItems = 8
		listW := min(menuW, theme.MaxNameLen+4)
		m.theme.list.SetSize(listW, min(m.height, maxVisibleItems*3))
	}
	if m.state == statsView {
		m = m.updateStatsViewport()
	}
	return m
}

func (m model) handleGlobalKey(msg tea.KeyPressMsg) (model, tea.Cmd, bool) {
	if m.state == generatingView {
		switch {
		case key.Matches(msg, rootKeys.Escape):
			returnState := m.activeSpawnReturnState()
			m.cancelActiveSpawn()
			m.state = returnState
			return m, nil, true
		case key.Matches(msg, rootKeys.Quit):
			return m, tea.Quit, true
		}
		return m, nil, true
	}

	if l := activeFilterList(m); l != nil {
		switch {
		case l.SettingFilter() && l.FilterValue() == "" && key.Matches(msg, rootKeys.Enter):
			l.ResetFilter()
			if m.state == gameSelectView {
				m = m.updateCategoryDetailViewport()
			}
			return m, nil, true
		case l.SettingFilter() && key.Matches(msg, rootKeys.Enter):
			return m, nil, false
		case l.FilterState() != list.Unfiltered && key.Matches(msg, rootKeys.Escape):
			return m, nil, false
		}
	}

	switch {
	case m.state == gameSelectView && msg.String() == "pgup":
		m.nav.categoryDetail.PageUp()
		return m, nil, true
	case m.state == gameSelectView && msg.String() == "pgdown":
		m.nav.categoryDetail.PageDown()
		return m, nil, true
	case m.state == gameView && key.Matches(msg, rootKeys.Escape):
		m = saveCurrentGame(m, store.StatusInProgress)
		m.state = mainMenuView
		m.debug.enabled = false
		return m, nil, true
	case key.Matches(msg, rootKeys.Enter):
		if m.state != gameView {
			next, cmd := m.handleEnter()
			return next.(model), cmd, true
		}
	case key.Matches(msg, rootKeys.Escape):
		if m.state != gameView {
			next, cmd := m.handleEscape()
			return next.(model), cmd, true
		}
	case key.Matches(msg, rootKeys.Quit):
		m = saveCurrentGame(m, store.StatusAbandoned)
		return m, tea.Quit, true
	case key.Matches(msg, rootKeys.Debug):
		m.debug.enabled = !m.debug.enabled
	case key.Matches(msg, rootKeys.FullHelp):
		m.help.showFull = !m.help.showFull
		if m.state == gameView && m.session.game != nil {
			m.session.game, _ = m.session.game.Update(game.HelpToggleMsg{Show: m.help.showFull})
		}
	case key.Matches(msg, rootKeys.ResetGame):
		if m.state == gameView && m.session.game != nil {
			m.session.game = m.session.game.Reset()
		}
	}

	return m, nil, false
}

func (m model) activeSpawnReturnState() viewState {
	if m.session.spawn == nil {
		return modeSelectView
	}
	return m.session.spawn.returnState
}

func (m model) persistCompletionIfSolved() model {
	if m.session.game == nil || m.session.completionSaved || !m.session.game.IsSolved() {
		return m
	}

	m.session.completionSaved = true
	saveData, err := m.session.game.GetSave()
	if err == nil {
		if err := m.store.UpdateSaveState(m.session.activeGameID, string(saveData)); err != nil {
			log.Printf("failed to save completion state: %v", err)
		}
	}
	if err := m.store.UpdateStatus(m.session.activeGameID, store.StatusCompleted); err != nil {
		log.Printf("failed to mark game completed: %v", err)
	}
	return m
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
	item := m.nav.mainMenu.Selected()
	switch item.Title() {
	case "Play":
		m.nav.playMenu = ui.NewMainMenu(playMenuItems)
		m.state = playMenuView
	case "Stats":
		return m.handleStatsEnter()
	case "Options":
		m.nav.optionsMenu = ui.NewMainMenu(optionsMenuItems)
		m.state = optionsMenuView
	case "Quit":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handlePlayMenuEnter() (tea.Model, tea.Cmd) {
	item := m.nav.playMenu.Selected()
	switch item.Title() {
	case "Daily Puzzle":
		return m.handleDailyPuzzle()
	case "Generate":
		m.state = gameSelectView
		m = m.updateCategoryDetailViewport()
	case "Continue":
		m.nav.continueTable, m.nav.continueGames = ui.InitContinueTable(m.store, m.height)
		m.state = continueView
	case "Enter Seed":
		return m.enterSeedInputView()
	}
	return m, nil
}

func (m model) handleOptionsMenuEnter() (tea.Model, tea.Cmd) {
	item := m.nav.optionsMenu.Selected()
	switch item.Title() {
	case "Theme":
		return m.handleThemeEnter()
	case "Guides":
		m.nav.helpSelectList = ui.InitList(GameCategories, "How to Play")
		m.nav.helpSelectList.SetSize(min(m.width, 64), min(m.height, ui.ListHeight(m.nav.helpSelectList)))
		m.state = helpSelectView
	}
	return m, nil
}

func (m model) handleSeedConfirm() (tea.Model, tea.Cmd) {
	seed := sessionflow.NormalizeSeed(m.nav.seedInput.Value())
	if seed == "" {
		return m, nil
	}

	selectedMode := m.currentSeedMode()
	name := sessionflow.SeededName(seed)
	if selectedMode.key != "" {
		name = sessionflow.SeededNameForGame(seed, selectedMode.gameType)
	}

	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		log.Printf("failed to check seeded game: %v", err)
		return m, nil
	}
	if rec != nil {
		var resumed bool
		m, resumed = m.importAndActivateRecord(*rec)
		if resumed {
			if err := sessionflow.ResumeAbandonedDeterministicRecord(m.store, rec); err != nil {
				log.Printf("%v", err)
			}
		}
		return m, nil
	}

	var spawner game.SeededSpawner
	var gameType string
	modeTitle := ""
	if selectedMode.key == "" {
		spawner, gameType, modeTitle, err = resolve.SeededMode(seed, catalog.All)
		if err != nil {
			log.Printf("failed to select seeded mode: %v", err)
			return m, nil
		}
	} else {
		spawner, gameType, modeTitle, err = resolve.SeededModeForGame(seed, selectedMode.gameType, catalog.All)
		if err != nil {
			log.Printf("failed to select seeded mode for %s: %v", selectedMode.gameType, err)
			return m, nil
		}
	}

	rng := resolve.RNGFromString(seed)
	ctx, jobID := m.beginSpawnContext()
	m.session.spawn = &spawnRequest{
		source:      spawnSourceSeed,
		name:        name,
		gameType:    gameType,
		modeTitle:   modeTitle,
		returnState: playMenuView,
	}
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng, ctx, jobID))
}

func (m model) handleDailyPuzzle() (tea.Model, tea.Cmd) {
	today := time.Now()
	name := daily.Name(today)

	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		log.Printf("failed to check daily game: %v", err)
		return m, nil
	}
	if rec != nil {
		var resumed bool
		m, resumed = m.importAndActivateRecord(*rec)
		if resumed {
			if err := sessionflow.ResumeAbandonedDeterministicRecord(m.store, rec); err != nil {
				log.Printf("%v", err)
			}
		}
		return m, nil
	}

	spawner, gameType, modeTitle := daily.Mode(today)
	if spawner == nil {
		log.Printf("no daily mode available for %s", today.Format("2006-01-02"))
		return m, nil
	}

	rng := daily.RNG(today)
	ctx, jobID := m.beginSpawnContext()
	m.session.spawn = &spawnRequest{
		source:      spawnSourceDaily,
		name:        name,
		gameType:    gameType,
		modeTitle:   modeTitle,
		returnState: playMenuView,
	}
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng, ctx, jobID))
}

func (m model) handleGameSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.nav.gameSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.nav.selectedCategory = cat
	m.nav.modeSelectList = ui.InitList(cat.Modes, cat.Name+" - Select Mode")
	m.nav.modeSelectList.SetSize(min(m.width, 64), min(m.height, ui.ListHeight(m.nav.modeSelectList)))
	m.state = modeSelectView
	return m, nil
}

func (m model) handleModeSelectEnter() (tea.Model, tea.Cmd) {
	item := m.nav.modeSelectList.SelectedItem()
	mode, ok := item.(game.Mode)
	if !ok {
		return m, nil
	}
	spawner, ok := item.(game.Spawner)
	if !ok {
		return m, nil
	}
	m.nav.selectedModeTitle = mode.Title()
	ctx, jobID := m.beginSpawnContext()
	m.session.spawn = &spawnRequest{
		source:      spawnSourceNormal,
		name:        sessionflow.GenerateUniqueName(m.store),
		gameType:    m.nav.selectedCategory.Name,
		modeTitle:   m.nav.selectedModeTitle,
		returnState: modeSelectView,
	}
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnCmd(spawner, ctx, jobID))
}

func (m model) handleContinueEnter() (tea.Model, tea.Cmd) {
	idx := m.nav.continueTable.Cursor()
	if idx < 0 || idx >= len(m.nav.continueGames) {
		return m, nil
	}
	rec := m.nav.continueGames[idx]
	m, _ = m.importAndActivateRecord(rec)
	return m, nil
}

func (m model) handleHelpSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.nav.helpSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.help.category = cat
	m = m.updateHelpDetailViewport()
	m.state = helpDetailView
	return m, nil
}

func (m model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.state {
	case playMenuView, optionsMenuView, statsView:
		m.state = mainMenuView
	case seedInputView, gameSelectView, continueView:
		m.state = playMenuView
	case generatingView:
		returnState := m.activeSpawnReturnState()
		m.cancelActiveSpawn()
		m.state = returnState
	case modeSelectView:
		m.state = gameSelectView
		m = m.updateCategoryDetailViewport()
	case helpDetailView:
		m.state = helpSelectView
	case helpSelectView, themeSelectView:
		if m.state == themeSelectView {
			_ = theme.Apply(m.theme.previous)
		}
		m.state = optionsMenuView
	}
	return m, nil
}

func (m model) handleThemeEnter() (tea.Model, tea.Cmd) {
	m.theme.previous = m.cfg.Theme

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

	const maxVisibleItems = 8
	listH := min(m.height, maxVisibleItems*3)
	listW := min(m.width, theme.MaxNameLen+4)

	m.theme.list = ui.InitThemeList(items, listW, listH)
	for i, item := range items {
		if mi, ok := item.(ui.MenuItem); ok && mi.ItemTitle == m.theme.previous {
			m.theme.list.Select(i)
			break
		}
	}
	if m.theme.previous == "" {
		m.theme.list.Select(0)
	}

	m.state = themeSelectView
	return m, nil
}

func (m model) handleThemeConfirm() (tea.Model, tea.Cmd) {
	item, ok := m.theme.list.SelectedItem().(ui.MenuItem)
	if !ok {
		return m, nil
	}

	themeName := item.ItemTitle
	if themeName == theme.DefaultThemeName {
		themeName = ""
	}

	_ = theme.Apply(item.ItemTitle)
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

	m.stats.cards = stats.BuildCards(catStats, modeStats)
	m.stats.profile = stats.BuildProfileBanner(catStats, modeStats, streakDates, currentDaily)

	statsWidth, statsHeight := statsViewportSize(m.width, m.height, m.stats.cards)
	m.stats.viewport = viewport.New(
		viewport.WithWidth(statsWidth),
		viewport.WithHeight(statsHeight),
	)
	m.stats.viewport.SetContent(stats.RenderCardGrid(m.stats.cards, statsWidth))
	m.state = statsView
	return m, nil
}

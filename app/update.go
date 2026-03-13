package app

import (
	"log"

	"github.com/FelineStateMachine/puzzletea/game"
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

func helpSelectListSize(width, height int, l list.Model) (int, int) {
	contentWidth, contentHeight := helpViewportSize(width, height)
	listWidth := min(contentWidth, 64)
	listHeight := min(contentHeight, ui.ListHeight(l))
	return listWidth, listHeight
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
	entry, ok := selectedCategoryEntry(item)
	if !ok {
		return ""
	}
	return entry.Definition.Name
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

	entry, ok := selectedCategoryEntry(m.nav.gameSelectList.SelectedItem())
	if !ok {
		m.nav.categoryDetail.SetContent("")
		return m
	}

	m.nav.categoryDetail.SetContent(renderCategoryDetailContent(entry, contentWidth))
	m.nav.categoryDetail.GotoTop()
	return m
}

func (m model) updateHelpDetailViewport() model {
	helpWidth, helpHeight := helpViewportSize(m.width, m.height)
	palette := theme.Current()
	themeKey := helpMarkdownThemeKey(palette)
	if m.help.renderer == nil || m.help.rendererWidth != helpWidth || m.help.rendererTheme != themeKey {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStyles(helpMarkdownStyle(palette)),
			glamour.WithWordWrap(helpWidth),
			glamour.WithChromaFormatter("terminal16m"),
		)
		if err != nil {
			log.Printf("failed to create help renderer: %v", err)
			m.help.renderer = nil
			m.help.rendererWidth = 0
			m.help.rendererTheme = ""
		} else {
			m.help.renderer = renderer
			m.help.rendererWidth = helpWidth
			m.help.rendererTheme = themeKey
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
	m.stats.viewport.SetContent(ui.RenderStatsCardGrid(m.stats.cards, statsWidth))
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd, handled := m.handleFrameworkMsg(msg)
	if handled {
		return next, cmd
	}

	m = next
	return m.updateActiveState(msg)
}

func (m model) handleWindowSize(msg tea.WindowSizeMsg) model {
	m.width = msg.Width
	m.height = msg.Height
	menuW := min(m.width, 64)
	metrics := categoryPickerSize(m.width, m.height)
	m.nav.gameSelectList.SetSize(metrics.listWidth, metrics.listHeight)
	m = m.updateCategoryDetailViewport()
	return m.resizeActiveState(menuW)
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
	case m.state == weeklyView && (msg.String() == "left" || msg.String() == "h"):
		m = m.moveWeeklyWeek(-1)
		return m, nil, true
	case m.state == weeklyView && (msg.String() == "right" || msg.String() == "l"):
		m = m.moveWeeklyWeek(1)
		return m, nil, true
	case m.state == gameView && key.Matches(msg, rootKeys.Enter):
		next, cmd, handled := m.advanceSolvedWeekly()
		if handled {
			return next, cmd, true
		}
	case m.state == gameSelectView && msg.String() == "pgup":
		m.nav.categoryDetail.PageUp()
		return m, nil, true
	case m.state == gameSelectView && msg.String() == "pgdown":
		m.nav.categoryDetail.PageDown()
		return m, nil, true
	case m.state == gameView && key.Matches(msg, rootKeys.Escape):
		returnState := m.session.returnState
		m = saveCurrentGame(m, store.StatusInProgress)
		m.state = returnState
		if returnState == weeklyView {
			m = m.refreshWeeklyBrowser()
		}
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
	case weeklyView:
		return m.handleWeeklyEnter()
	case helpSelectView:
		return m.handleHelpSelectEnter()
	case themeSelectView:
		return m.handleThemeConfirm()
	}
	return m, nil
}

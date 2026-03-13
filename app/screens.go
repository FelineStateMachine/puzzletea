package app

import (
	"github.com/FelineStateMachine/puzzletea/registry"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type screenAction interface{ isScreenAction() }

type openPlayMenuAction struct{}

func (openPlayMenuAction) isScreenAction() {}

type openStatsAction struct{}

func (openStatsAction) isScreenAction() {}

type openOptionsMenuAction struct{}

func (openOptionsMenuAction) isScreenAction() {}

type quitAction struct{}

func (quitAction) isScreenAction() {}

type openGameSelectAction struct{}

func (openGameSelectAction) isScreenAction() {}

type openContinueAction struct{}

func (openContinueAction) isScreenAction() {}

type openDailyAction struct{}

func (openDailyAction) isScreenAction() {}

type openWeeklyAction struct{}

func (openWeeklyAction) isScreenAction() {}

type openSeedInputAction struct{}

func (openSeedInputAction) isScreenAction() {}

type backAction struct {
	target viewState
}

func (backAction) isScreenAction() {}

type gameSelectEnterAction struct{}

func (gameSelectEnterAction) isScreenAction() {}

type modeSelectEnterAction struct{}

func (modeSelectEnterAction) isScreenAction() {}

type continueEnterAction struct{}

func (continueEnterAction) isScreenAction() {}

type weeklyShiftAction struct {
	delta int
}

func (weeklyShiftAction) isScreenAction() {}

type weeklyEnterAction struct{}

func (weeklyEnterAction) isScreenAction() {}

type helpSelectEnterAction struct{}

func (helpSelectEnterAction) isScreenAction() {}

type openThemeSelectAction struct{}

func (openThemeSelectAction) isScreenAction() {}

type openHelpSelectAction struct{}

func (openHelpSelectAction) isScreenAction() {}

type previewThemeAction struct {
	name string
}

func (previewThemeAction) isScreenAction() {}

type confirmThemeAction struct{}

func (confirmThemeAction) isScreenAction() {}

type seedConfirmAction struct{}

func (seedConfirmAction) isScreenAction() {}

type screenModel interface {
	State() viewState
	Resize(width, height int) screenModel
	Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction)
	View(notice noticeState) string
	Apply(m model) model
}

type mainMenuScreen struct {
	width  int
	height int
	menu   ui.MainMenu
}

func (s mainMenuScreen) State() viewState { return mainMenuView }

func (s mainMenuScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	return s
}

func (s mainMenuScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	updateMainMenuCursor(msg, &s.menu)
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil, nil
	}
	if key.Matches(keyMsg, rootKeys.Enter) {
		switch s.menu.Selected().Title() {
		case "Play":
			return s, nil, openPlayMenuAction{}
		case "Stats":
			return s, nil, openStatsAction{}
		case "Options":
			return s, nil, openOptionsMenuAction{}
		case "Quit":
			return s, nil, quitAction{}
		}
	}
	return s, nil, nil
}

func (s mainMenuScreen) View(notice noticeState) string {
	return centerContentWithNotice(s.width, s.height, notice, s.menu.View())
}

func (s mainMenuScreen) Apply(m model) model {
	m.state = mainMenuView
	m.nav.mainMenu = s.menu
	return m
}

type playMenuScreen struct {
	width  int
	height int
	menu   ui.MainMenu
}

func (s playMenuScreen) State() viewState { return playMenuView }

func (s playMenuScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	return s
}

func (s playMenuScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	updateMainMenuCursor(msg, &s.menu)
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil, nil
	}
	switch {
	case key.Matches(keyMsg, rootKeys.Enter):
		switch s.menu.Selected().Title() {
		case "Create":
			return s, nil, openGameSelectAction{}
		case "Continue":
			return s, nil, openContinueAction{}
		case "Daily":
			return s, nil, openDailyAction{}
		case "Weekly":
			return s, nil, openWeeklyAction{}
		case "Seeded":
			return s, nil, openSeedInputAction{}
		}
	case key.Matches(keyMsg, rootKeys.Escape):
		return s, nil, backAction{target: mainMenuView}
	}
	return s, nil, nil
}

func (s playMenuScreen) View(notice noticeState) string {
	return centerContentWithNotice(s.width, s.height, notice, s.menu.ViewAsPanel("Play"))
}

func (s playMenuScreen) Apply(m model) model {
	m.state = playMenuView
	m.nav.playMenu = s.menu
	return m
}

type optionsMenuScreen struct {
	width  int
	height int
	menu   ui.MainMenu
}

func (s optionsMenuScreen) State() viewState { return optionsMenuView }

func (s optionsMenuScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	return s
}

func (s optionsMenuScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	updateMainMenuCursor(msg, &s.menu)
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil, nil
	}
	switch {
	case key.Matches(keyMsg, rootKeys.Enter):
		switch s.menu.Selected().Title() {
		case "Theme":
			return s, nil, openThemeSelectAction{}
		case "Guides":
			return s, nil, openHelpSelectAction{}
		}
	case key.Matches(keyMsg, rootKeys.Escape):
		return s, nil, backAction{target: mainMenuView}
	}
	return s, nil, nil
}

func (s optionsMenuScreen) View(notice noticeState) string {
	items := s.menu.RenderItems() + "\n\n" + ui.DimItemStyle().Render("- Dami")
	return renderPanelView(s.width, s.height, notice, "Options", items, "↑/↓ navigate • enter select • esc back")
}

func (s optionsMenuScreen) Apply(m model) model {
	m.state = optionsMenuView
	m.nav.optionsMenu = s.menu
	return m
}

type seedInputScreen struct {
	width  int
	height int
	seed   seedState
}

func (s seedInputScreen) State() viewState { return seedInputView }

func (s seedInputScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	s.seed.input.SetWidth(min(width, 48))
	return s
}

func (s seedInputScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Enter):
			if normalizeSeedValue(s.seed.input.Value()) != "" {
				return s, nil, seedConfirmAction{}
			}
			return s, nil, nil
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: playMenuView}
		}
	}

	m := model{seed: s.seed}
	next, cmd := m.handleSeedInputUpdate(msg)
	s.seed = next.seed
	return s, cmd, nil
}

func (s seedInputScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		seed:   s.seed,
		notice: notice,
	}
	return m.renderPanel(
		"Enter Seed",
		m.seedInputBody(),
		"↑/↓ change field • ←/→ game • enter confirm • esc back",
	)
}

func (s seedInputScreen) Apply(m model) model {
	m.state = seedInputView
	m.seed = s.seed
	return m
}

type gameSelectScreen struct {
	width  int
	height int
	list   list.Model
	detail viewport.Model
}

func (s gameSelectScreen) State() viewState { return gameSelectView }

func (s gameSelectScreen) Resize(width, height int) screenModel {
	m := model{
		width:  width,
		height: height,
		nav: navigationState{
			gameSelectList: s.list,
			categoryDetail: s.detail,
		},
	}
	metrics := categoryPickerSize(width, height)
	m.nav.gameSelectList.SetSize(metrics.listWidth, metrics.listHeight)
	m = m.updateCategoryDetailViewport()
	s.width = width
	s.height = height
	s.list = m.nav.gameSelectList
	s.detail = m.nav.categoryDetail
	return s
}

func (s gameSelectScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case s.list.SettingFilter() && s.list.FilterValue() == "" && key.Matches(keyMsg, rootKeys.Enter):
			s.list.ResetFilter()
			s = s.resizeSelf()
			return s, nil, nil
		case s.list.SettingFilter() && key.Matches(keyMsg, rootKeys.Enter):
		case s.list.FilterState() != list.Unfiltered && key.Matches(keyMsg, rootKeys.Escape):
		case key.Matches(keyMsg, rootKeys.Enter):
			return s, nil, gameSelectEnterAction{}
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: playMenuView}
		case keyMsg.String() == "pgup":
			s.detail.PageUp()
			return s, nil, nil
		case keyMsg.String() == "pgdown":
			s.detail.PageDown()
			return s, nil, nil
		}
	}

	prev := selectedCategoryName(s.list.SelectedItem())
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	if selectedCategoryName(s.list.SelectedItem()) != prev {
		s = s.resizeSelf()
	}
	return s, cmd, nil
}

func (s gameSelectScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		nav: navigationState{
			gameSelectList: s.list,
			categoryDetail: s.detail,
		},
	}
	return m.gameSelectViewContent()
}

func (s gameSelectScreen) Apply(m model) model {
	m.state = gameSelectView
	m.nav.gameSelectList = s.list
	m.nav.categoryDetail = s.detail
	return m
}

func (s gameSelectScreen) resizeSelf() gameSelectScreen {
	return s.Resize(s.width, s.height).(gameSelectScreen)
}

type modeSelectScreen struct {
	width  int
	height int
	entry  registry.Entry
	list   list.Model
}

func (s modeSelectScreen) State() viewState { return modeSelectView }

func (s modeSelectScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	s.list.SetSize(min(width, 64), min(height, ui.ListHeight(s.list)))
	return s
}

func (s modeSelectScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Enter):
			return s, nil, modeSelectEnterAction{}
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: gameSelectView}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd, nil
}

func (s modeSelectScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		nav: navigationState{
			selectedCategory: s.entry,
			modeSelectList:   s.list,
		},
	}
	return m.renderPanel(
		m.nav.selectedCategory.Definition.Name+" — Select Mode",
		m.nav.modeSelectList.View(),
		"↑/↓ navigate • enter select • esc back",
	)
}

func (s modeSelectScreen) Apply(m model) model {
	m.state = modeSelectView
	m.nav.selectedCategory = s.entry
	m.nav.modeSelectList = s.list
	return m
}

type continueScreen struct {
	width  int
	height int
	cont   continueState
}

func (s continueScreen) State() viewState { return continueView }

func (s continueScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	s.cont.table.SetWidth(min(width, ui.ContinueTableWidth()))
	visibleRows := min(len(s.cont.games), ui.MaxTableRows)
	s.cont.table.SetHeight(min(height, visibleRows))
	return s
}

func (s continueScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Enter):
			return s, nil, continueEnterAction{}
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: playMenuView}
		}
	}

	var cmd tea.Cmd
	s.cont.table, cmd = s.cont.table.Update(msg)
	return s, cmd, nil
}

func (s continueScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		cont:   s.cont,
	}
	return m.renderContinueView()
}

func (s continueScreen) Apply(m model) model {
	m.state = continueView
	m.cont = s.cont
	return m
}

type weeklyScreen struct {
	width  int
	height int
	weekly weeklyState
}

func (s weeklyScreen) State() viewState { return weeklyView }

func (s weeklyScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	s.weekly.table.SetWidth(min(width, ui.WeeklyTableWidth()))
	visibleRows := min(len(s.weekly.rows), ui.MaxTableRows)
	s.weekly.table.SetHeight(min(height, visibleRows))
	return s
}

func (s weeklyScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: playMenuView}
		case key.Matches(keyMsg, rootKeys.Enter):
			return s, nil, weeklyEnterAction{}
		case keyMsg.String() == "left" || keyMsg.String() == "h":
			return s, nil, weeklyShiftAction{delta: -1}
		case keyMsg.String() == "right" || keyMsg.String() == "l":
			return s, nil, weeklyShiftAction{delta: 1}
		}
	}

	var cmd tea.Cmd
	s.weekly.table, cmd = s.weekly.table.Update(msg)
	return s, cmd, nil
}

func (s weeklyScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		weekly: s.weekly,
	}
	return centerContentWithNotice(s.width, s.height, notice, m.weeklyViewContent())
}

func (s weeklyScreen) Apply(m model) model {
	m.state = weeklyView
	m.weekly = s.weekly
	return m
}

type helpSelectScreen struct {
	width  int
	height int
	help   helpState
}

func (s helpSelectScreen) State() viewState { return helpSelectView }

func (s helpSelectScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	listWidth, listHeight := helpSelectListSize(width, height, s.help.selectList)
	s.help.selectList.SetSize(listWidth, listHeight)
	return s
}

func (s helpSelectScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Enter):
			return s, nil, helpSelectEnterAction{}
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: optionsMenuView}
		}
	}

	var cmd tea.Cmd
	s.help.selectList, cmd = s.help.selectList.Update(msg)
	return s, cmd, nil
}

func (s helpSelectScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		help:   s.help,
	}
	return m.renderPanel(
		"How to Play",
		m.help.selectList.View(),
		"↑/↓ navigate • enter select • esc back",
	)
}

func (s helpSelectScreen) Apply(m model) model {
	m.state = helpSelectView
	m.help.selectList = s.help.selectList
	return m
}

type helpDetailScreen struct {
	width  int
	height int
	help   helpState
}

func (s helpDetailScreen) State() viewState { return helpDetailView }

func (s helpDetailScreen) Resize(width, height int) screenModel {
	m := model{
		width:  width,
		height: height,
		help:   s.help,
	}
	m = m.updateHelpDetailViewport()
	s.width = width
	s.height = height
	s.help = m.help
	return s
}

func (s helpDetailScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok && key.Matches(keyMsg, rootKeys.Escape) {
		return s, nil, backAction{target: helpSelectView}
	}

	var cmd tea.Cmd
	s.help.viewport, cmd = s.help.viewport.Update(msg)
	return s, cmd, nil
}

func (s helpDetailScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		help:   s.help,
	}
	return m.renderPanel(
		m.help.category.Definition.Name+" — Guide",
		m.help.viewport.View(),
		"↑/↓ scroll • esc back",
	)
}

func (s helpDetailScreen) Apply(m model) model {
	m.state = helpDetailView
	m.help = s.help
	return m
}

type statsScreen struct {
	width  int
	height int
	stats  statsState
}

func (s statsScreen) State() viewState { return statsView }

func (s statsScreen) Resize(width, height int) screenModel {
	m := model{
		width:  width,
		height: height,
		stats:  s.stats,
	}
	m = m.updateStatsViewport()
	s.width = width
	s.height = height
	s.stats = m.stats
	return s
}

func (s statsScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok && key.Matches(keyMsg, rootKeys.Escape) {
		return s, nil, backAction{target: mainMenuView}
	}

	var cmd tea.Cmd
	s.stats.viewport, cmd = s.stats.viewport.Update(msg)
	return s, cmd, nil
}

func (s statsScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		stats:  s.stats,
	}
	return m.renderStatsView()
}

func (s statsScreen) Apply(m model) model {
	m.state = statsView
	m.stats = s.stats
	return m
}

type themeSelectScreen struct {
	width  int
	height int
	theme  themeState
}

func (s themeSelectScreen) State() viewState { return themeSelectView }

func (s themeSelectScreen) Resize(width, height int) screenModel {
	const maxVisibleItems = 8
	menuW := min(width, 64)
	listW := min(menuW, theme.MaxNameLen+4)
	s.width = width
	s.height = height
	s.theme.list.SetSize(listW, min(height, maxVisibleItems*3))
	return s
}

func (s themeSelectScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case s.theme.list.SettingFilter() && s.theme.list.FilterValue() == "" && key.Matches(keyMsg, rootKeys.Enter):
			s.theme.list.ResetFilter()
			return s, nil, nil
		case s.theme.list.SettingFilter() && key.Matches(keyMsg, rootKeys.Enter):
		case s.theme.list.FilterState() != list.Unfiltered && key.Matches(keyMsg, rootKeys.Escape):
		case key.Matches(keyMsg, rootKeys.Enter):
			return s, nil, confirmThemeAction{}
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: optionsMenuView}
		}
	}

	prev := s.theme.list.Index()
	var cmd tea.Cmd
	s.theme.list, cmd = s.theme.list.Update(msg)
	if s.theme.list.Index() != prev {
		if item, ok := s.theme.list.SelectedItem().(ui.MenuItem); ok {
			return s, cmd, previewThemeAction{name: item.ItemTitle}
		}
	}
	return s, cmd, nil
}

func (s themeSelectScreen) View(notice noticeState) string {
	m := model{
		width:  s.width,
		height: s.height,
		notice: notice,
		theme:  s.theme,
	}
	return m.themeSelectViewContent()
}

func (s themeSelectScreen) Apply(m model) model {
	m.state = themeSelectView
	m.theme = s.theme
	return m
}

type generatingScreen struct {
	width   int
	height  int
	spinner spinner.Model
}

func (s generatingScreen) State() viewState { return generatingView }

func (s generatingScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	return s
}

func (s generatingScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd, nil
}

func (s generatingScreen) View(notice noticeState) string {
	content := s.spinner.View() + " Generating puzzle..."
	box := ui.GeneratingFrame().Render(appendNoticeContent(s.width, notice, content))
	return ui.CenterView(s.width, s.height, box)
}

func (s generatingScreen) Apply(m model) model {
	m.state = generatingView
	m.spinner = s.spinner
	return m
}

func normalizeSeedValue(input string) string {
	return sessionflow.NormalizeSeed(input)
}

func (m model) activeScreen() screenModel {
	switch m.state {
	case mainMenuView:
		return mainMenuScreen{
			width:  m.width,
			height: m.height,
			menu:   m.nav.mainMenu,
		}
	case playMenuView:
		return playMenuScreen{
			width:  m.width,
			height: m.height,
			menu:   m.nav.playMenu,
		}
	case optionsMenuView:
		return optionsMenuScreen{
			width:  m.width,
			height: m.height,
			menu:   m.nav.optionsMenu,
		}
	case seedInputView:
		return seedInputScreen{
			width:  m.width,
			height: m.height,
			seed:   m.seed,
		}
	case gameSelectView:
		return gameSelectScreen{
			width:  m.width,
			height: m.height,
			list:   m.nav.gameSelectList,
			detail: m.nav.categoryDetail,
		}
	case modeSelectView:
		return modeSelectScreen{
			width:  m.width,
			height: m.height,
			entry:  m.nav.selectedCategory,
			list:   m.nav.modeSelectList,
		}
	case continueView:
		return continueScreen{
			width:  m.width,
			height: m.height,
			cont:   m.cont,
		}
	case weeklyView:
		return weeklyScreen{
			width:  m.width,
			height: m.height,
			weekly: m.weekly,
		}
	case helpSelectView:
		return helpSelectScreen{
			width:  m.width,
			height: m.height,
			help:   m.help,
		}
	case helpDetailView:
		return helpDetailScreen{
			width:  m.width,
			height: m.height,
			help:   m.help,
		}
	case statsView:
		return statsScreen{
			width:  m.width,
			height: m.height,
			stats:  m.stats,
		}
	case themeSelectView:
		return themeSelectScreen{
			width:  m.width,
			height: m.height,
			theme:  m.theme,
		}
	case generatingView:
		return generatingScreen{
			width:   m.width,
			height:  m.height,
			spinner: m.spinner,
		}
	default:
		return nil
	}
}

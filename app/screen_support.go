package app

import (
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

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

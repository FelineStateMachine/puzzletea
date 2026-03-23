package app

import (
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type mainMenuScreen struct {
	width  int
	height int
	menu   ui.MainMenu
}

func (s mainMenuScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	return s
}

func (s mainMenuScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	updateMainMenuCursor(msg, &s.menu)
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok || !key.Matches(keyMsg, rootKeys.Enter) {
		return s, nil, nil
	}

	switch s.menu.SelectedAction() {
	case mainMenuActionPlay:
		return s, nil, openPlayMenuAction{}
	case mainMenuActionExport:
		return s, nil, openExportAction{}
	case mainMenuActionStats:
		return s, nil, openStatsAction{}
	case mainMenuActionOptions:
		return s, nil, openOptionsMenuAction{}
	case mainMenuActionQuit:
		return s, nil, quitAction{}
	default:
		return s, nil, nil
	}
}

func (s mainMenuScreen) View(notice noticeState) string {
	return centerContentWithNotice(s.width, s.height, notice, s.menu.View())
}

type playMenuScreen struct {
	width  int
	height int
	menu   ui.MainMenu
}

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
		switch s.menu.SelectedAction() {
		case playMenuActionCreate:
			return s, nil, openGameSelectAction{}
		case playMenuActionContinue:
			return s, nil, openContinueAction{}
		case playMenuActionDaily:
			return s, nil, openDailyAction{}
		case playMenuActionWeekly:
			return s, nil, openWeeklyAction{}
		case playMenuActionSeeded:
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

type optionsMenuScreen struct {
	width  int
	height int
	menu   ui.MainMenu
}

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
		switch s.menu.SelectedAction() {
		case optionsMenuActionTheme:
			return s, nil, openThemeSelectAction{}
		case optionsMenuActionGuides:
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

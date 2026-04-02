package app

import (
	"github.com/FelineStateMachine/puzzletea/registry"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type seedInputScreen struct {
	width  int
	height int
	seed   seedState
}

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
			if sessionflow.NormalizeSeed(s.seed.input.Value()) != "" {
				return s, nil, seedConfirmAction{seed: s.seed}
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
	return renderPanelView(s.width, s.height, notice,
		"Enter Seed",
		seedInputBody(s.seed),
		"↑/↓ change field • ←/→ game • enter confirm • esc back",
	)
}

type gameSelectScreen struct {
	width  int
	height int
	list   list.Model
	detail viewport.Model
}

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
			entry, ok := selectedCategoryEntry(s.list.SelectedItem())
			if !ok {
				return s, nil, nil
			}
			return s, nil, gameSelectEnterAction{entry: entry}
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
	return gameSelectViewContent(s.width, s.height, s.list, s.detail, notice)
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
			item := unwrapModeDisplayItem(s.list.SelectedItem())
			mode, ok := item.(registry.ModeEntry)
			if !ok {
				return s, nil, nil
			}
			return s, nil, modeSelectEnterAction{entry: s.entry, mode: mode}
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: gameSelectView}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd, nil
}

func (s modeSelectScreen) View(notice noticeState) string {
	return renderPanelView(s.width, s.height, notice,
		s.entry.Definition.Name+" — Select Mode",
		s.list.View(),
		"↑/↓ navigate • enter select • esc back",
	)
}

type continueScreen struct {
	width  int
	height int
	cont   continueState
}

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
			idx := s.cont.table.Cursor()
			if idx >= 0 && idx < len(s.cont.games) {
				return s, nil, continueEnterAction{record: s.cont.games[idx]}
			}
			return s, nil, nil
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, backAction{target: playMenuView}
		}
	}

	var cmd tea.Cmd
	s.cont.table, cmd = s.cont.table.Update(msg)
	return s, cmd, nil
}

func (s continueScreen) View(notice noticeState) string {
	return renderContinueView(s.cont, s.width, s.height, notice)
}

type weeklyScreen struct {
	width  int
	height int
	weekly weeklyState
}

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
			return s, nil, weeklyEnterAction{weekly: s.weekly}
		case keyMsg.String() == "left" || keyMsg.String() == "h":
			return s, nil, weeklyShiftAction{delta: -1, weekly: s.weekly}
		case keyMsg.String() == "right" || keyMsg.String() == "l":
			return s, nil, weeklyShiftAction{delta: 1, weekly: s.weekly}
		}
	}

	var cmd tea.Cmd
	s.weekly.table, cmd = s.weekly.table.Update(msg)
	return s, cmd, nil
}

func (s weeklyScreen) View(notice noticeState) string {
	return centerContentWithNotice(s.width, s.height, notice, weeklyViewContent(s.weekly))
}

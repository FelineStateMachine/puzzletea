package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
)

func (m model) currentWeeklyMenuIndex() int {
	if m.store == nil {
		return 1
	}

	year, week := time.Now().ISOWeek()
	highestCompleted, err := m.store.GetCurrentWeeklyHighestCompletedIndex(year, week)
	if err != nil {
		return 1
	}
	if highestCompleted >= weeklyEntryCount {
		return weeklyEntryCount
	}
	if highestCompleted < 1 {
		return 1
	}
	return highestCompleted + 1
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

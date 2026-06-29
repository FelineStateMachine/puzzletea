package app

import tea "charm.land/bubbletea/v2"

// safeScreen is a minimal screenModel for tests that need a non-nil screen
// in the cache but don't care about rendering. It avoids nil-pointer panics
// that zero-value real screens (like exportScreen) would trigger on Resize.
type safeScreen struct{}

func (s safeScreen) Resize(width, height int) screenModel { return s }
func (s safeScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	return s, nil, nil
}
func (s safeScreen) View(notice noticeState) string { return "" }

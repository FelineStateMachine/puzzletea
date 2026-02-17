package app

import "github.com/charmbracelet/glamour"

func initDebugRenderer() *glamour.TermRenderer {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
		glamour.WithChromaFormatter("terminal16m"),
	)
	if err != nil {
		return nil
	}
	return renderer
}

func (m model) renderDebugInfo() string {
	if m.debugRenderer == nil {
		return m.game.GetDebugInfo()
	}
	s, err := m.debugRenderer.Render(m.game.GetDebugInfo())
	if err != nil {
		return m.game.GetDebugInfo()
	}
	return s
}

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
	if m.debug.renderer == nil {
		return m.session.game.GetDebugInfo()
	}
	s, err := m.debug.renderer.Render(m.session.game.GetDebugInfo())
	if err != nil {
		return m.session.game.GetDebugInfo()
	}
	return s
}

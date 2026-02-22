package app

import (
	"errors"

	"github.com/FelineStateMachine/puzzletea/markdownexport"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

var copyToClipboard = clipboard.WriteAll

func (m model) yankPuzzleCmd() tea.Cmd {
	if m.state != gameView || m.game == nil {
		return nil
	}

	gameType, err := markdownexport.DetectGameType(m.game)
	if err != nil || !markdownexport.SupportsGameType(gameType) {
		return nil
	}

	save, err := m.game.GetSave()
	if err != nil {
		return nil
	}

	snippet, err := markdownexport.RenderPuzzleSnippet(gameType, "", save)
	if err != nil {
		if errors.Is(err, markdownexport.ErrUnsupportedGame) {
			return nil
		}
		return nil
	}

	if err := copyToClipboard(snippet); err == nil {
		return nil
	}

	return tea.SetClipboard(snippet)
}

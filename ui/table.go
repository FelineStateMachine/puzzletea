package ui

import (
	"log"

	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// FormatStatus converts a GameStatus enum to a human-readable display string.
func FormatStatus(s store.GameStatus) string {
	switch s {
	case store.StatusNew:
		return "New"
	case store.StatusInProgress:
		return "In Progress"
	case store.StatusCompleted:
		return "Completed"
	default:
		return string(s)
	}
}

// InitContinueTable loads saved games from the store and builds the table
// widget. Returns the table, the loaded games, and the column definitions.
func InitContinueTable(s *store.Store, height int) (table.Model, []store.GameRecord) {
	games, err := s.ListGames()
	if err != nil {
		log.Printf("failed to list games: %v", err)
		games = nil
	}

	columns := []table.Column{
		{Title: "Name", Width: 34},
		{Title: "Game", Width: 15},
		{Title: "Mode", Width: 15},
		{Title: "Status", Width: 12},
		{Title: "Last Updated", Width: 20},
	}

	rows := make([]table.Row, len(games))
	for i, g := range games {
		rows[i] = table.Row{
			g.Name,
			g.GameType,
			g.Mode,
			FormatStatus(g.Status),
			g.UpdatedAt.Local().Format("Jan 02 15:04"),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(max(height-2, 1)),
	)

	st := table.DefaultStyles()
	st.Header = st.Header.
		Bold(true).
		Foreground(MenuTableHeader).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(MenuDim)
	st.Selected = st.Selected.
		Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
		Background(MenuAccent).
		Bold(true)
	st.Cell = st.Cell.
		Foreground(MenuText)
	t.SetStyles(st)

	return t, games
}

package ui

import (
	"fmt"
	"log"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
)

// MaxTableRows is the maximum number of rows visible in the continue table.
const MaxTableRows = 20

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

	tableWidth := 0
	for _, c := range columns {
		tableWidth += c.Width
	}
	// Account for column gaps (2 chars between each column).
	tableWidth += (len(columns) - 1) * 2

	visibleRows := min(len(rows), MaxTableRows)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(tableWidth),
		table.WithHeight(min(max(height-2, 1), visibleRows)),
	)
	t.SetStyles(defaultTableStyles())

	return t, games
}

// InitWeeklyTable builds the weekly browser table from pre-rendered rows.
func InitWeeklyTable(rows []table.Row, height int) table.Model {
	columns := []table.Column{
		{Title: "Puzzle", Width: 8},
		{Title: "Game", Width: 15},
		{Title: "Mode", Width: 18},
		{Title: "Bonus XP", Width: 10},
		{Title: "Status", Width: 12},
	}

	tableWidth := 0
	for _, c := range columns {
		tableWidth += c.Width
	}
	tableWidth += (len(columns) - 1) * 2

	tableHeight := min(max(height-2, 1), MaxTableRows)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(tableWidth),
		table.WithHeight(tableHeight),
	)
	t.SetStyles(defaultTableStyles())
	return t
}

func defaultTableStyles() table.Styles {
	p := theme.Current()

	st := table.DefaultStyles()
	st.Header = st.Header.
		Bold(true).
		Foreground(p.Accent).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(p.Border)
	st.Selected = st.Selected.
		Foreground(p.AccentText).
		Background(p.Accent).
		Bold(true)
	st.Cell = st.Cell.
		Foreground(p.FG)
	return st
}

// TablePagination returns a pagination string like "1-20 of 45" for the
// continue table, or an empty string if all rows fit on one page.
func TablePagination(t table.Model) string {
	total := len(t.Rows())
	if total <= MaxTableRows {
		return ""
	}
	cursor := t.Cursor()
	height := t.Height()

	// Estimate the visible window based on cursor and table height.
	start := cursor - height + 1
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
	}

	return FooterHint().Render(fmt.Sprintf("%d–%d of %d", start+1, end, total))
}

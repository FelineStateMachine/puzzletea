package game

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Shared cursor color tokens, used both in style construction and for
// composite styling where a cursor color is applied on top of another style.
var (
	CursorFG     = compat.AdaptiveColor{Light: lipgloss.Color("255"), Dark: lipgloss.Color("235")}
	CursorBG     = compat.AdaptiveColor{Light: lipgloss.Color("130"), Dark: lipgloss.Color("173")}
	CursorWarmBG = compat.AdaptiveColor{Light: lipgloss.Color("130"), Dark: lipgloss.Color("214")}
	ConflictFG   = compat.AdaptiveColor{Light: lipgloss.Color("160"), Dark: lipgloss.Color("167")}
	ConflictBG   = compat.AdaptiveColor{Light: lipgloss.Color("224"), Dark: lipgloss.Color("52")}
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("255"), Dark: lipgloss.Color("255")}).
			Background(CursorBG).
			Padding(0, 1)

	solvedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("78")})

	// CursorStyle highlights the cursor position with an accent background.
	// Used by lightsout, sudoku, wordsearch.
	CursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(CursorFG).
			Background(CursorBG)

	// CursorWarmStyle highlights the cursor with a warmer background.
	// Used by hitori, takuzu, nonogram.
	CursorWarmStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(CursorFG).
			Background(CursorWarmBG)

	// CursorSolvedStyle highlights the cursor position on a solved grid.
	// Shared across all puzzle types.
	CursorSolvedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(CursorFG).
				Background(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("28")})

	// StatusBarStyle is the shared style for the status/help bar below each puzzle grid.
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("244"), Dark: lipgloss.Color("244")}).
			MarginTop(1)
)

// TitleBarView renders a title bar with the game name, mode name, and optional solved badge.
func TitleBarView(gameName, modeName string, solved bool) string {
	title := titleStyle.Render(gameName + " - " + modeName)
	if solved {
		badge := solvedBadgeStyle.Render("  SOLVED")
		return title + badge + "\nctrl+n to play again"
	}
	return title + "\n"
}

// DebugHeader returns the markdown heading and property table header for debug info.
// rows is a list of [key, value] pairs for the "Game State" table.
func DebugHeader(title string, rows [][2]string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n## Game State\n\n", title)
	sb.WriteString("| Property | Value |\n| :--- | :--- |\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "| %s | %s |\n", r[0], r[1])
	}
	return sb.String()
}

// DebugTable returns a markdown table section with a heading and arbitrary columns.
// headers is the column names, rows is a list of row values.
func DebugTable(heading string, headers []string, rows [][]string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n## %s\n\n", heading)

	sb.WriteString("|")
	for _, h := range headers {
		fmt.Fprintf(&sb, " %s |", h)
	}
	sb.WriteString("\n|")
	for range headers {
		sb.WriteString(" :--- |")
	}
	sb.WriteString("\n")

	for _, row := range rows {
		sb.WriteString("|")
		for _, cell := range row {
			fmt.Fprintf(&sb, " %s |", cell)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

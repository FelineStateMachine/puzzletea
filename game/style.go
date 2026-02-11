package game

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "173"}).
			Padding(0, 1)

	solvedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "78"})
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

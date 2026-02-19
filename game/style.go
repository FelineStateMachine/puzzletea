package game

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/theme"
)

// Shared cursor / conflict color accessors. Puzzle packages that compose
// styles on top of the cursor colors read these at render time.

func CursorFG() color.Color   { return theme.Current().AccentText }
func CursorBG() color.Color   { return theme.Current().AccentBG }
func ConflictFG() color.Color { return theme.Current().Error }
func ConflictBG() color.Color { return theme.Current().ErrorBG }
func SolvedFG() color.Color   { return theme.Current().SolvedFG }

// CursorStyle highlights the cursor position with an accent background.
func CursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(CursorFG()).
		Background(CursorBG())
}

// CursorSolvedStyle highlights the cursor position on a solved grid.
func CursorSolvedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(CursorFG()).
		Background(theme.Current().SuccessBG)
}

// Bracket markers for cursor identification. These are prepended/appended
// to cell content so the cursor is identifiable independent of BG contrast.
const (
	CursorLeft  = "\u25b8" // ▸
	CursorRight = "\u25c2" // ◂
)

// StatusBarStyle returns the shared style for the status/help bar below each puzzle grid.
func StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextDim).
		MarginTop(1)
}

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().AccentText).
		Background(CursorBG()).
		Padding(0, 1)
}

func solvedBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().SuccessBorder)
}

// TitleBarView renders a title bar with the game name, mode name, and optional solved badge.
func TitleBarView(gameName, modeName string, solved bool) string {
	title := titleStyle().Render(gameName + " - " + modeName)
	if solved {
		badge := solvedBadgeStyle().Render("  SOLVED")
		return title + badge + "\nctrl+n to play again"
	}
	return lipgloss.NewStyle().PaddingBottom(1).Render(title)
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

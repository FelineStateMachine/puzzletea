package spellpuzzle

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const boardCellWidth = game.DynamicGridCellWidth

func renderView(m Model) string {
	title := game.TitleBarView("Spell Puzzle", m.modeTitle, m.solved)
	main := renderMainView(m)
	bank := renderBankRow(m)
	if m.solved {
		content := lipgloss.JoinVertical(lipgloss.Center, main, bank)
		return game.ComposeGameView(title, content)
	}
	trace := renderTraceRow(m)
	status := renderStatusRow(m)

	return game.ComposeGameViewRows(
		title,
		main,
		game.StaticRow(bank),
		game.StaticRow(trace),
		game.StaticRow(status),
	)
}

func renderMainView(m Model) string {
	boardBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border).
		Padding(0, 1)
	sidebarBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border).
		Padding(0, 1)

	boardView := boardBorder.Render(renderBoard(m))
	sidebarView := sidebarBorder.Render(renderSidebar(m))
	return lipgloss.JoinHorizontal(lipgloss.Top, boardView, "  ", sidebarView)
}

func renderBoard(m Model) string {
	if m.board.Width == 0 || m.board.Height == 0 {
		return ""
	}

	gridColors := game.DefaultBorderColors()
	gridColors.BackgroundBG = nil
	gridColors.SolvedBG = nil

	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:     m.board.Width,
		Height:    m.board.Height,
		CellWidth: boardCellWidth,
		Solved:    m.solved,
		Colors:    &gridColors,
		Cell: func(x, y int) string {
			return renderBoardCell(m.board.Cells[y][x])
		},
		ZoneAt: func(x, y int) int {
			cell := m.board.Cells[y][x]
			if !cell.Occupied {
				return 0
			}
			return 1
		},
		HasVerticalEdge: func(x, y int) bool {
			return m.board.hasVerticalEdge(x, y)
		},
		HasHorizontalEdge: func(x, y int) bool {
			return m.board.hasHorizontalEdge(x, y)
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			fill, solved := m.board.bridgeFillState(bridge)
			if !fill {
				return nil
			}
			if solved {
				return theme.Current().SuccessBG
			}
			return theme.Current().BG
		},
	})
}

func renderBoardCell(cell boardCell) string {
	if !cell.Occupied {
		return lipgloss.NewStyle().
			Width(boardCellWidth).
			Align(lipgloss.Center).
			Render(" ")
	}

	style := lipgloss.NewStyle().Width(boardCellWidth).Align(lipgloss.Center)
	display := " · "
	switch {
	case cell.Revealed:
		style = style.Foreground(theme.Current().SolvedFG).Background(theme.Current().SuccessBG)
		display = fmt.Sprintf(" %c ", cell.Letter)
	default:
		style = style.Foreground(theme.Current().TextDim).Background(theme.Current().BG)
	}
	return style.Render(display)
}

func renderSidebar(m Model) string {
	var builder strings.Builder
	header := lipgloss.NewStyle().Foreground(theme.Current().Accent).Bold(true)
	value := lipgloss.NewStyle().Foreground(theme.Current().Info)
	builder.WriteString(header.Render("Allowed Letters"))
	builder.WriteString("\n")
	builder.WriteString(value.Render(strings.Join(strings.Fields(fmt.Sprint(string(m.bank))), "")))
	builder.WriteString("\n\n")
	builder.WriteString(header.Render("Words Left"))
	builder.WriteString("\n")
	builder.WriteString(value.Render(fmt.Sprintf("%d", m.countWordsLeft())))
	builder.WriteString("\n\n")
	builder.WriteString(header.Render("Bonus Words"))
	builder.WriteString("\n")
	builder.WriteString(value.Render(fmt.Sprintf("%d/%d", len(m.bonusWords), m.countBonusWordsTotal())))
	builder.WriteString("\n\n")
	builder.WriteString(header.Render("Recent Bonus"))
	builder.WriteString("\n")
	recent := m.recentBonusWords(5)
	if len(recent) == 0 {
		builder.WriteString(lipgloss.NewStyle().Foreground(theme.Current().TextDim).Render("none yet"))
	} else {
		for _, word := range recent {
			builder.WriteString(value.Render(word))
			builder.WriteString("\n")
		}
	}
	return strings.TrimRight(builder.String(), "\n")
}

func renderBankRow(m Model) string {
	paletteColors := theme.Current().ThemeColors()
	tiles := make([]string, 0, len(m.bank))
	for i, letter := range m.bank {
		style := lipgloss.NewStyle().Width(3).Align(lipgloss.Center)
		display := fmt.Sprintf(" %c ", letter)
		switch {
		case m.tileUsage(i) > 0:
			style = style.Foreground(theme.Current().SolvedFG).Background(theme.Current().SuccessBG)
		case m.bankCursor == i:
			style = style.Foreground(game.CursorFG())
			display = game.CursorLeft + string(letter) + game.CursorRight
		case m.hoverTile == i:
			style = style.Foreground(theme.Current().FG).Background(theme.Current().HighlightBG)
		default:
			bg := bankTileBackground(paletteColors, m.bankColorSlots, i)
			style = style.Foreground(theme.TextOnBG(bg)).Background(bg)
		}
		tiles = append(tiles, style.Render(display))
	}
	return strings.Join(tiles, " ")
}

func bankTileBackground(colors []color.Color, slots []int, index int) color.Color {
	if len(colors) == 0 {
		return theme.Current().Surface
	}
	slot := index
	if index >= 0 && index < len(slots) {
		slot = slots[index]
	}
	return colors[slot%len(colors)]
}

func renderTraceRow(m Model) string {
	trace := m.traceWord()
	if trace == "" {
		trace = "build a word from the letter bank"
	}
	return game.StatusBarStyle().Render("trace: " + trace)
}

func renderStatusRow(m Model) string {
	if m.feedback.Message == "" {
		if m.showFullHelp {
			return game.StatusBarStyle().Render("type letters from the bank  1: shuffle  enter: submit  bkspc: delete  mouse: drag across letters")
		}
		return game.StatusBarStyle().Render("type: add letters  1: shuffle  enter: submit  bkspc: delete")
	}

	style := game.StatusBarStyle()
	if m.feedback.Kind == feedbackError {
		style = style.Foreground(theme.Current().Error)
	} else {
		style = style.Foreground(theme.Current().SuccessBorder)
	}
	return style.Render(m.feedback.Message)
}

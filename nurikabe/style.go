package nurikabe

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

type cellVisual struct {
	text     string
	fg       color.Color
	bg       color.Color
	bridgeBG color.Color
	bold     bool
	kind     cellVisualKind
}

type cellVisualKind int

const (
	cellVisualLand cellVisualKind = iota
	cellVisualSea
	cellVisualCursor
	cellVisualSolved
	cellVisualConflict
)

func resolveCellVisual(m Model, x, y int) cellVisual {
	p := theme.Current()
	landBg := theme.Blend(p.BG, p.Success, 0.45)
	seaBg := theme.Blend(p.BG, p.Secondary, 0.24)
	c := m.marks[y][x]
	clue := m.clues[y][x]
	isCursor := x == m.cursor.X && y == m.cursor.Y
	conflict := m.conflicts[y][x]
	inSeaSquare := isSeaSquareCell(m.marks, x, y)
	visual := cellVisual{
		text:     "   ",
		fg:       theme.TextOnBG(landBg),
		bg:       landBg,
		bridgeBG: landBg,
		kind:     cellVisualLand,
	}

	switch {
	case clue > 0:
		visual.text = fmt.Sprintf("%2d ", clue)
		if clue < 10 {
			visual.text = fmt.Sprintf(" %d ", clue)
		}
		visual.fg = p.Info
		visual.bold = true
	case c == seaCell:
		visual.text = " ~ "
		if inSeaSquare {
			visual.text = " @ "
		}
		visual.bg = seaBg
		visual.bridgeBG = seaBg
		visual.fg = theme.TextOnBG(seaBg)
		visual.kind = cellVisualSea
	case c == islandCell:
		visual.text = " \u00b7 "
	}

	switch {
	case conflict:
		visual.bg = game.ConflictBG()
		visual.bridgeBG = game.ConflictBG()
		visual.fg = game.ConflictFG()
		visual.kind = cellVisualConflict
	case m.solved:
		visual.bg = p.SuccessBG
		visual.bridgeBG = p.SuccessBG
		visual.fg = theme.TextOnBG(p.SuccessBG)
		visual.kind = cellVisualSolved
	case isCursor:
		visual.bg = game.CursorBG()
		visual.fg = game.CursorFG()
		visual.kind = cellVisualCursor
	}

	return visual
}

func renderCellVisual(visual cellVisual) string {
	style := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(visual.fg).
		Background(visual.bg)
	if visual.bold {
		style = style.Bold(true)
	}
	return style.Render(visual.text)
}

func colorsMatch(left, right color.Color) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	lr, lg, lb, la := left.RGBA()
	rr, rg, rb, ra := right.RGBA()
	return lr == rr && lg == rg && lb == rb && la == ra
}

func dominantBridgeBackground(visuals []cellVisual) color.Color {
	for _, kind := range []cellVisualKind{
		cellVisualConflict,
		cellVisualSolved,
		cellVisualCursor,
	} {
		for _, visual := range visuals {
			if visual.kind == kind {
				return visual.bridgeBG
			}
		}
	}
	return nil
}

func blendBridgeBackgrounds(visuals []cellVisual) color.Color {
	if len(visuals) == 0 {
		return nil
	}

	var rSum, gSum, bSum, aSum uint32
	for _, visual := range visuals {
		r, g, b, a := visual.bridgeBG.RGBA()
		rSum += uint32(r >> 8)
		gSum += uint32(g >> 8)
		bSum += uint32(b >> 8)
		aSum += uint32(a >> 8)
	}

	count := uint32(len(visuals))
	return color.NRGBA{
		R: uint8(rSum / count),
		G: uint8(gSum / count),
		B: uint8(bSum / count),
		A: uint8(aSum / count),
	}
}

func bridgeFill(m Model, bridge game.DynamicGridBridge) color.Color {
	if bridge.Count == 0 {
		return nil
	}

	visuals := make([]cellVisual, 0, bridge.Count)
	for i := 0; i < bridge.Count; i++ {
		cell := bridge.Cells[i]
		visuals = append(visuals, resolveCellVisual(m, cell.X, cell.Y))
	}

	bg := visuals[0].bridgeBG
	allMatch := true
	for _, visual := range visuals[1:] {
		if !colorsMatch(bg, visual.bridgeBG) {
			allMatch = false
			break
		}
	}
	if allMatch {
		return bg
	}

	if dominant := dominantBridgeBackground(visuals); dominant != nil {
		return dominant
	}

	return blendBridgeBackgrounds(visuals)
}

func gridView(m Model) string {
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.width,
		Height: m.height,
		Solved: m.solved,
		Cell: func(x, y int) string {
			return renderCellVisual(resolveCellVisual(m, x, y))
		},
		HasVerticalEdge: func(x, _ int) bool {
			return x <= 0 || x >= m.width
		},
		HasHorizontalEdge: func(_, y int) bool {
			return y <= 0 || y >= m.height
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(m, bridge)
		},
	})
}

// isSeaSquareCell reports whether (x,y) is part of any 2x2 sea block.
func isSeaSquareCell(marks grid, x, y int) bool {
	if len(marks) == 0 || len(marks[0]) == 0 {
		return false
	}
	if y < 0 || y >= len(marks) || x < 0 || x >= len(marks[0]) {
		return false
	}
	if marks[y][x] != seaCell {
		return false
	}

	startX := max(0, x-1)
	endX := min(len(marks[0])-2, x)
	startY := max(0, y-1)
	endY := min(len(marks)-2, y)

	for yy := startY; yy <= endY; yy++ {
		for xx := startX; xx <= endX; xx++ {
			if marks[yy][xx] == seaCell &&
				marks[yy][xx+1] == seaCell &&
				marks[yy+1][xx] == seaCell &&
				marks[yy+1][xx+1] == seaCell {
				return true
			}
		}
	}

	return false
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move x/LMB: sea  z/RMB: island  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("x/LMB: sea  z/RMB: island  bkspc: clear")
}

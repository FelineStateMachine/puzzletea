package nurikabe

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const (
	cellWidth   = game.DynamicGridCellWidth
	neutralZone = -1
)

type renderGridState struct {
	zones      [][]int
	zoneColors map[int]color.Color
}

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
	cellVisualNeutral cellVisualKind = iota
	cellVisualSea
	cellVisualCursor
	cellVisualSolved
	cellVisualConflict
)

func buildRenderGridState(m Model) renderGridState {
	zones := make([][]int, m.height)
	for y := range m.height {
		zones[y] = make([]int, m.width)
		for x := range m.width {
			zones[y][x] = neutralZone
		}
	}

	_, idx := islandComponents(m.marks, m.clues)
	if idx == nil {
		return renderGridState{zones: zones}
	}

	themeColors := theme.Current().ThemeColors()
	zoneColors := make(map[int]color.Color)
	for y := range m.height {
		for x := range m.width {
			zone := idx[y][x]
			if zone < 0 {
				continue
			}
			zones[y][x] = zone
			if len(themeColors) == 0 {
				continue
			}
			if _, ok := zoneColors[zone]; !ok {
				zoneColors[zone] = themeColors[zone%len(themeColors)]
			}
		}
	}

	return renderGridState{
		zones:      zones,
		zoneColors: zoneColors,
	}
}

func zoneBackground(renderState renderGridState, zone int) color.Color {
	if zone < 0 {
		return nil
	}
	return renderState.zoneColors[zone]
}

func resolveCellVisual(m Model, x, y int) cellVisual {
	return resolveCellVisualWithState(m, buildRenderGridState(m), x, y)
}

func resolveCellVisualWithState(m Model, renderState renderGridState, x, y int) cellVisual {
	p := theme.Current()
	seaBg := theme.Blend(p.BG, p.Secondary, 0.24)
	c := m.marks[y][x]
	clue := m.clues[y][x]
	isCursor := x == m.cursor.X && y == m.cursor.Y
	conflict := m.conflicts[y][x]
	inSeaSquare := isSeaSquareCell(m.marks, x, y)
	islandBg := zoneBackground(renderState, renderState.zones[y][x])
	visual := cellVisual{
		text:     "   ",
		fg:       p.FG,
		bg:       p.BG,
		bridgeBG: p.BG,
		kind:     cellVisualNeutral,
	}

	switch {
	case clue > 0:
		visual.text = fmt.Sprintf("%2d ", clue)
		if clue < 10 {
			visual.text = fmt.Sprintf(" %d ", clue)
		}
		if islandBg != nil {
			visual.bg = islandBg
			visual.bridgeBG = islandBg
			visual.fg = theme.TextOnBG(islandBg)
		} else {
			visual.fg = p.Info
		}
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
		if islandBg != nil {
			visual.bg = islandBg
			visual.bridgeBG = islandBg
			visual.fg = theme.TextOnBG(islandBg)
		}
	case m.solved && c == unknownCell:
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
	if isCursor {
		visual.text = cursorWrappedText(visual.text)
	}

	return visual
}

func cursorWrappedText(text string) string {
	runes := []rune(text)
	if len(runes) != cellWidth {
		return text
	}
	if runes[0] != ' ' || runes[cellWidth-1] != ' ' {
		return text
	}
	return game.CursorLeft + string(runes[1]) + game.CursorRight
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

func dominantBridgeBackground(visuals []cellVisual) color.Color {
	for _, kind := range []cellVisualKind{
		cellVisualConflict,
		cellVisualSolved,
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
	return bridgeFillWithState(m, buildRenderGridState(m), bridge)
}

func bridgeFillWithState(m Model, renderState renderGridState, bridge game.DynamicGridBridge) color.Color {
	if bridge.Count == 0 {
		return nil
	}

	visuals := make([]cellVisual, 0, bridge.Count)
	for i := 0; i < bridge.Count; i++ {
		cell := bridge.Cells[i]
		visuals = append(visuals, resolveCellVisualWithState(m, renderState, cell.X, cell.Y))
	}

	bg := visuals[0].bridgeBG
	allMatch := true
	for _, visual := range visuals[1:] {
		if !game.SameColor(bg, visual.bridgeBG) {
			allMatch = false
			break
		}
	}
	if allMatch {
		if dominant := dominantBridgeBackground(visuals); dominant != nil {
			return dominant
		}
		if bridge.Uniform && bridge.Zone >= 0 {
			return nil
		}
		return bg
	}

	if dominant := dominantBridgeBackground(visuals); dominant != nil {
		return dominant
	}

	return blendBridgeBackgrounds(visuals)
}

func gridView(m Model) string {
	renderState := buildRenderGridState(m)

	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.width,
		Height: m.height,
		Solved: m.solved,
		Cell: func(x, y int) string {
			return renderCellVisual(resolveCellVisualWithState(m, renderState, x, y))
		},
		ZoneAt: func(x, y int) int {
			return renderState.zones[y][x]
		},
		ZoneFill: func(zone int) color.Color {
			return zoneBackground(renderState, zone)
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFillWithState(m, renderState, bridge)
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

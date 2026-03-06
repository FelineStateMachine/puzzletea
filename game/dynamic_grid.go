package game

import (
	"image"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

const DynamicGridCellWidth = 3

const (
	dynamicGridCellStepX = DynamicGridCellWidth + 1
	dynamicGridCellStepY = 2
)

type DynamicGridMetrics struct {
	Width     int
	Height    int
	CellWidth int
}

type DynamicGridBridgeKind int

const (
	DynamicGridBridgeVertical DynamicGridBridgeKind = iota
	DynamicGridBridgeHorizontal
	DynamicGridBridgeJunction
)

type DynamicGridBridge struct {
	Kind    DynamicGridBridgeKind
	X       int
	Y       int
	Cells   [4]image.Point
	Count   int
	Zone    int
	Uniform bool
}

type DynamicGridSpec struct {
	Width      int
	Height     int
	Solved     bool
	Colors     *GridBorderColors
	Cell       func(x, y int) string
	ZoneAt     func(x, y int) int
	ZoneFill   func(zone int) color.Color
	BridgeFill func(bridge DynamicGridBridge) color.Color
}

func RenderDynamicGrid(spec DynamicGridSpec) string {
	rows := make([]string, 0, spec.Height*dynamicGridCellStepY+1)
	rows = append(rows, dynamicGridBoundaryRow(spec, 0))
	for y := range spec.Height {
		rows = append(rows, dynamicGridContentRow(spec, y))
		rows = append(rows, dynamicGridBoundaryRow(spec, y+1))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func DynamicGridOrigin(termW, termH int, fullView, title, grid string) (x, y int) {
	viewWidth := lipgloss.Width(fullView)
	viewHeight := lipgloss.Height(fullView)
	if viewWidth > termW {
		viewWidth = termW
	}
	if viewHeight > termH {
		viewHeight = termH
	}

	centerX := max((termW-viewWidth)/2, 0)
	centerY := max((termH-viewHeight)/2, 0)

	titleHeight := strings.Count(title, "\n") + 1
	gridWidth := lipgloss.Width(grid)
	gridPadLeft := max((viewWidth-gridWidth)/2, 0)

	return centerX + 1 + gridPadLeft, centerY + 1 + titleHeight
}

func DynamicGridScreenToCell(
	metrics DynamicGridMetrics,
	originX, originY, screenX, screenY int,
	includeSeparators bool,
) (col, row int, ok bool) {
	if metrics.Width <= 0 || metrics.Height <= 0 {
		return 0, 0, false
	}

	cellWidth := metrics.CellWidth
	if cellWidth <= 0 {
		cellWidth = DynamicGridCellWidth
	}

	lx := screenX - originX
	ly := screenY - originY
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	maxX := (metrics.Width-1)*(cellWidth+1) + (cellWidth - 1)
	maxY := (metrics.Height - 1) * dynamicGridCellStepY
	if lx > maxX || ly > maxY {
		return 0, 0, false
	}

	if includeSeparators {
		col = min((lx+cellWidth/2)/(cellWidth+1), metrics.Width-1)
		row = min((ly+1)/dynamicGridCellStepY, metrics.Height-1)
	} else {
		col = lx / (cellWidth + 1)
		row = ly / dynamicGridCellStepY
		if lx%(cellWidth+1) >= cellWidth || ly%dynamicGridCellStepY != 0 {
			return 0, 0, false
		}
	}

	if col < 0 || col >= metrics.Width || row < 0 || row >= metrics.Height {
		return 0, 0, false
	}
	return col, row, true
}

func dynamicGridContentRow(spec DynamicGridSpec, y int) string {
	var b strings.Builder
	for x := range spec.Width {
		bridge := dynamicGridVerticalBridge(spec, x, y)
		b.WriteString(dynamicGridRenderBorderChar(
			verticalEdgeRune(dynamicGridHasVerticalEdge(spec, x, y)),
			dynamicGridColors(spec),
			spec.Solved,
			dynamicGridBridgeBackground(spec, bridge),
		))
		if spec.Cell != nil {
			b.WriteString(spec.Cell(x, y))
		}
	}

	bridge := dynamicGridVerticalBridge(spec, spec.Width, y)
	b.WriteString(dynamicGridRenderBorderChar(
		verticalEdgeRune(dynamicGridHasVerticalEdge(spec, spec.Width, y)),
		dynamicGridColors(spec),
		spec.Solved,
		dynamicGridBridgeBackground(spec, bridge),
	))
	return b.String()
}

func dynamicGridBoundaryRow(spec DynamicGridSpec, y int) string {
	var b strings.Builder
	for x := 0; x <= spec.Width; x++ {
		runeValue := dynamicGridJunctionRune(spec, x, y)
		b.WriteString(dynamicGridRenderBorderChar(
			runeValue,
			dynamicGridColors(spec),
			spec.Solved,
			dynamicGridBridgeBackground(spec, dynamicGridJunctionBridge(spec, x, y, runeValue)),
		))
		if x == spec.Width {
			continue
		}

		ch := ' '
		if dynamicGridHasHorizontalEdge(spec, x, y) {
			ch = '─'
		}
		b.WriteString(strings.Repeat(
			dynamicGridRenderBorderChar(
				ch,
				dynamicGridColors(spec),
				spec.Solved,
				dynamicGridBridgeBackground(spec, dynamicGridHorizontalBridge(spec, x, y)),
			),
			DynamicGridCellWidth,
		))
	}
	return b.String()
}

func dynamicGridColors(spec DynamicGridSpec) GridBorderColors {
	if spec.Colors != nil {
		return *spec.Colors
	}
	return DefaultBorderColors()
}

func dynamicGridRenderBorderChar(ch rune, colors GridBorderColors, solved bool, bg color.Color) string {
	fg := colors.BorderFG
	if solved {
		fg = colors.SolvedBorderFG
	}
	if bg == nil {
		bg = colors.BackgroundBG
		if solved {
			bg = colors.SolvedBG
		}
	}
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Render(string(ch))
}

func dynamicGridHasHorizontalEdge(spec DynamicGridSpec, x, y int) bool {
	switch {
	case y <= 0, y >= spec.Height:
		return true
	default:
		return dynamicGridZoneAt(spec, x, y-1) != dynamicGridZoneAt(spec, x, y)
	}
}

func dynamicGridHasVerticalEdge(spec DynamicGridSpec, x, y int) bool {
	switch {
	case x <= 0, x >= spec.Width:
		return true
	default:
		return dynamicGridZoneAt(spec, x-1, y) != dynamicGridZoneAt(spec, x, y)
	}
}

func dynamicGridJunctionRune(spec DynamicGridSpec, x, y int) rune {
	north := y > 0 && dynamicGridHasVerticalEdge(spec, x, y-1)
	south := y < spec.Height && dynamicGridHasVerticalEdge(spec, x, y)
	west := x > 0 && dynamicGridHasHorizontalEdge(spec, x-1, y)
	east := x < spec.Width && dynamicGridHasHorizontalEdge(spec, x, y)

	switch {
	case north && south && west && east:
		return '┼'
	case north && south && west:
		return '┤'
	case north && south && east:
		return '├'
	case west && east && north:
		return '┴'
	case west && east && south:
		return '┬'
	case south && east:
		return '┌'
	case south && west:
		return '┐'
	case north && east:
		return '└'
	case north && west:
		return '┘'
	case north || south:
		return '│'
	case west || east:
		return '─'
	default:
		return ' '
	}
}

func dynamicGridBridgeBackground(spec DynamicGridSpec, bridge DynamicGridBridge) color.Color {
	if bridge.Count == 0 {
		return nil
	}
	if spec.BridgeFill != nil {
		if bg := spec.BridgeFill(bridge); bg != nil {
			return bg
		}
	}
	if bridge.Uniform && spec.ZoneFill != nil {
		return spec.ZoneFill(bridge.Zone)
	}
	return nil
}

func dynamicGridVerticalBridge(spec DynamicGridSpec, x, y int) DynamicGridBridge {
	if dynamicGridHasVerticalEdge(spec, x, y) || x <= 0 || x >= spec.Width {
		return DynamicGridBridge{
			Kind: DynamicGridBridgeVertical,
			X:    x,
			Y:    y,
		}
	}

	return dynamicGridBridge(
		spec,
		DynamicGridBridgeVertical,
		x,
		y,
		image.Pt(x-1, y),
		image.Pt(x, y),
	)
}

func dynamicGridHorizontalBridge(spec DynamicGridSpec, x, y int) DynamicGridBridge {
	if dynamicGridHasHorizontalEdge(spec, x, y) || y <= 0 || y >= spec.Height {
		return DynamicGridBridge{
			Kind: DynamicGridBridgeHorizontal,
			X:    x,
			Y:    y,
		}
	}

	return dynamicGridBridge(
		spec,
		DynamicGridBridgeHorizontal,
		x,
		y,
		image.Pt(x, y-1),
		image.Pt(x, y),
	)
}

func dynamicGridJunctionBridge(spec DynamicGridSpec, x, y int, runeValue rune) DynamicGridBridge {
	if runeValue != ' ' {
		return DynamicGridBridge{
			Kind: DynamicGridBridgeJunction,
			X:    x,
			Y:    y,
		}
	}
	if x <= 0 || y <= 0 || x >= spec.Width || y >= spec.Height {
		return DynamicGridBridge{
			Kind: DynamicGridBridgeJunction,
			X:    x,
			Y:    y,
		}
	}

	return dynamicGridBridge(
		spec,
		DynamicGridBridgeJunction,
		x,
		y,
		image.Pt(x-1, y-1),
		image.Pt(x, y-1),
		image.Pt(x-1, y),
		image.Pt(x, y),
	)
}

func dynamicGridBridge(spec DynamicGridSpec, kind DynamicGridBridgeKind, x, y int, cells ...image.Point) DynamicGridBridge {
	bridge := DynamicGridBridge{
		Kind:    kind,
		X:       x,
		Y:       y,
		Count:   len(cells),
		Uniform: len(cells) > 0,
	}

	for i := range cells {
		bridge.Cells[i] = cells[i]
		zone := dynamicGridZoneAt(spec, cells[i].X, cells[i].Y)
		if i == 0 {
			bridge.Zone = zone
			continue
		}
		if zone != bridge.Zone {
			bridge.Uniform = false
		}
	}
	return bridge
}

func dynamicGridZoneAt(spec DynamicGridSpec, x, y int) int {
	if spec.ZoneAt == nil {
		return 0
	}
	return spec.ZoneAt(x, y)
}

func verticalEdgeRune(hasEdge bool) rune {
	if hasEdge {
		return '│'
	}
	return ' '
}

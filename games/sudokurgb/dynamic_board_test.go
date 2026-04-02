package sudokurgb

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestCellViewOmitsCrosshairAndBoxTint(t *testing.T) {
	g := newGrid(nil)
	g[2][2].v = 2
	g[1][4].v = 3
	m := Model{
		cursor:   game.Cursor{X: 1, Y: 1},
		grid:     g,
		analysis: analyzeGrid(g),
		keys:     DefaultKeyMap,
	}

	gotSameBox := cellView(m, 2, 2, false)
	wantSameBox := userCellStyle(2).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("■")
	if gotSameBox != wantSameBox {
		t.Fatalf("same-box cellView() = %q, want %q", gotSameBox, wantSameBox)
	}

	gotSameRow := cellView(m, 4, 1, false)
	wantSameRow := userCellStyle(3).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("●")
	if gotSameRow != wantSameRow {
		t.Fatalf("same-row cellView() = %q, want %q", gotSameRow, wantSameRow)
	}
}

func TestCellViewUsesCursorBackground(t *testing.T) {
	g := newGrid(nil)
	g[1][1].v = 2
	m := Model{
		cursor:   game.Cursor{X: 1, Y: 1},
		grid:     g,
		analysis: analyzeGrid(g),
		keys:     DefaultKeyMap,
	}

	gotFilled := cellView(m, 1, 1, false)
	wantFilled := valueCursorStyle(2).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render(game.CursorLeft + "■" + game.CursorRight)
	if gotFilled != wantFilled {
		t.Fatalf("filled cursor cellView() = %q, want %q", gotFilled, wantFilled)
	}

	m.grid[1][1].v = 0
	gotEmpty := cellView(m, 1, 1, false)
	wantEmpty := game.CursorStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render(game.CursorLeft + "·" + game.CursorRight)
	if gotEmpty != wantEmpty {
		t.Fatalf("empty cursor cellView() = %q, want %q", gotEmpty, wantEmpty)
	}
}

func TestProvidedCellViewUsesGivenTint(t *testing.T) {
	g := newGrid([]cell{{x: 0, y: 0, v: 2}})
	m := Model{
		grid:         g,
		keys:         DefaultKeyMap,
		cursor:       game.Cursor{X: 8, Y: 8},
		analysis:     analyzeGrid(g),
		providedGrid: buildProvidedGrid([]cell{{x: 0, y: 0, v: 2}}),
	}

	got := cellView(m, 0, 0, false)
	want := lipgloss.NewStyle().
		Bold(true).
		Foreground(symbolColor(2)).
		Background(theme.GivenTint(game.DefaultBorderColors().BackgroundBG)).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("■")
	if got != want {
		t.Fatalf("provided cellView() = %q, want %q", got, want)
	}
}

func TestBuildBoardBlockRendersHintBands(t *testing.T) {
	g := newGrid(nil)
	g[0][0].v = 1
	g[0][1].v = 2
	g[0][2].v = 3
	m := Model{
		grid:     g,
		analysis: analyzeGrid(g),
		keys:     DefaultKeyMap,
	}

	layout := buildBoardBlock(m, false)
	lines := strings.Split(ansi.Strip(layout.Block), "\n")
	if got, want := layout.HintHeight, colHintHeight; got != want {
		t.Fatalf("HintHeight = %d, want %d", got, want)
	}
	if got, want := layout.HintWidth, rowHintWidth; got != want {
		t.Fatalf("HintWidth = %d, want %d", got, want)
	}
	if got := []rune(lines[layout.HintHeight])[layout.HintWidth]; got != '┌' {
		t.Fatalf("grid top border should start after hints, got %q", got)
	}
	if strings.TrimSpace(lines[layout.HintHeight][:layout.HintWidth]) != "" {
		t.Fatalf("row hints should not occupy the border row, got %q", lines[layout.HintHeight][:layout.HintWidth])
	}
	if strings.TrimSpace(lines[layout.HintHeight+1][:layout.HintWidth]) == "" {
		t.Fatalf("row hints missing on first content row: %q", lines[layout.HintHeight+1][:layout.HintWidth])
	}
	if strings.TrimSpace(lines[0][layout.HintWidth:]) == "" {
		t.Fatalf("column hints missing from top band: %q", lines[0])
	}
}

func TestHintRenderingUsesSolvedAndErrorStyles(t *testing.T) {
	analysis := boardAnalysis{}
	analysis.rowCounts[0][1] = houseQuota
	analysis.rowCounts[0][2] = 2
	analysis.rowCounts[0][3] = 4
	analysis.rowOverQuota[0][3] = true
	analysis.colCounts[0][2] = houseQuota

	rowHints := rowHintView(analysis, false)
	wantSolved := hintSolvedStyle(1).Render("3")
	if !strings.Contains(rowHints, wantSolved) {
		t.Fatalf("rowHintView() = %q, want solved chip %q", rowHints, wantSolved)
	}
	wantError := hintErrorStyle().Render("4")
	if !strings.Contains(rowHints, wantError) {
		t.Fatalf("rowHintView() = %q, want error chip %q", rowHints, wantError)
	}

	colHints := colHintView(analysis, false)
	wantColSolved := hintSolvedStyle(2).Render("3")
	if !strings.Contains(colHints, wantColSolved) {
		t.Fatalf("colHintView() = %q, want solved chip %q", colHints, wantColSolved)
	}
}

func TestSolvedBoardUsesSuccessBackground(t *testing.T) {
	p := theme.Current()
	g := validCompleteGrid()
	m := Model{
		grid:     g,
		analysis: analyzeGrid(g),
		keys:     DefaultKeyMap,
		cursor:   game.Cursor{X: 8, Y: 8},
	}

	gotCell := cellView(m, 0, 0, true)
	wantCell := userCellStyle(1).
		Bold(true).
		Foreground(p.SolvedFG).
		Background(p.SuccessBG).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("▲")
	if gotCell != wantCell {
		t.Fatalf("solved cellView() = %q, want %q", gotCell, wantCell)
	}

	gotRowHints := rowHintView(m.analysis, true)
	wantRowHint := solvedBoardStyle().Render("3")
	if !strings.Contains(gotRowHints, wantRowHint) {
		t.Fatalf("solved rowHintView() = %q, want success background %q", gotRowHints, wantRowHint)
	}

	gotColHints := colHintView(m.analysis, true)
	if !strings.Contains(gotColHints, wantRowHint) {
		t.Fatalf("solved colHintView() = %q, want success background %q", gotColHints, wantRowHint)
	}

	wantSpacer := solvedBoardStyle().Render(" ")
	if !strings.Contains(gotRowHints, wantSpacer) {
		t.Fatalf("solved rowHintView() = %q, want solved spacer %q", gotRowHints, wantSpacer)
	}
}

func TestCellViewUsesErrorBackgroundForBoxConflictOnly(t *testing.T) {
	g := newGrid(nil)
	g[0][0].v = 3
	g[0][1].v = 3
	g[1][2].v = 3
	g[2][0].v = 3

	m := Model{
		grid:     g,
		analysis: analyzeGrid(g),
		keys:     DefaultKeyMap,
		cursor:   game.Cursor{X: 8, Y: 8},
	}

	got := cellView(m, 0, 0, false)
	want := boxConflictCellStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("●")
	if got != want {
		t.Fatalf("box conflict cellView() = %q, want %q", got, want)
	}

	rowOnly := newGrid(nil)
	for x := 0; x < 4; x++ {
		rowOnly[0][x].v = 1
	}
	m.grid = rowOnly
	m.analysis = analyzeGrid(rowOnly)
	got = cellView(m, 0, 0, false)
	want = userCellStyle(1).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("▲")
	if got != want {
		t.Fatalf("row-only conflict cellView() = %q, want %q", got, want)
	}
}

func TestUserCellViewUsesDefaultBoardBackground(t *testing.T) {
	g := newGrid(nil)
	g[0][0].v = 2
	m := Model{
		grid:     g,
		analysis: analyzeGrid(g),
		keys:     DefaultKeyMap,
		cursor:   game.Cursor{X: 8, Y: 8},
	}

	got := cellView(m, 0, 0, false)
	want := userCellStyle(2).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("■")
	if got != want {
		t.Fatalf("user cellView() = %q, want %q", got, want)
	}
}

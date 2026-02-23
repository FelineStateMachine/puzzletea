package pdfexport

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type nurikabeSave struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Clues  string `json:"clues"`
}

type nonogramSave struct {
	State    string  `json:"state"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	RowHints [][]int `json:"row-hints"`
	ColHints [][]int `json:"col-hints"`
}

type hashiSave struct {
	Width   int           `json:"width"`
	Height  int           `json:"height"`
	Islands []hashiIsland `json:"islands"`
}

type hashiIsland struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Required int `json:"required"`
}

type shikakuSave struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Clues  []shikakuClue `json:"clues"`
}

type shikakuClue struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Value int `json:"value"`
}

type hitoriSave struct {
	Size    int    `json:"size"`
	Numbers string `json:"numbers"`
}

type takuzuSave struct {
	Size     int    `json:"size"`
	State    string `json:"state"`
	Provided string `json:"provided"`
}

type sudokuSave struct {
	Provided []sudokuCell `json:"provided"`
}

type sudokuCell struct {
	X int `json:"x"`
	Y int `json:"y"`
	V int `json:"v"`
}

type wordSearchSave struct {
	Width  int              `json:"width"`
	Height int              `json:"height"`
	Grid   string           `json:"grid"`
	Words  []wordSearchWord `json:"words"`
}

type wordSearchWord struct {
	Text string `json:"text"`
}

func ParseNonogramPrintData(saveData []byte) (*NonogramData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save nonogramSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode nonogram save: %w", err)
	}

	stateRows := splitNonogramStateRows(save.State)

	width := save.Width
	if width <= 0 {
		width = len(save.ColHints)
	}
	if width <= 0 {
		width = maxRuneWidth(stateRows)
	}

	height := save.Height
	if height <= 0 {
		height = len(save.RowHints)
	}
	if height <= 0 {
		height = len(stateRows)
	}

	if width <= 0 || height <= 0 {
		return nil, nil
	}

	return &NonogramData{
		Width:    width,
		Height:   height,
		RowHints: normalizeNonogramHintRows(save.RowHints, height),
		ColHints: normalizeNonogramHintRows(save.ColHints, width),
		Grid:     normalizeNonogramStateGrid(stateRows, width, height),
	}, nil
}

func ParseNurikabePrintData(saveData []byte) (*NurikabeData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save nurikabeSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode nurikabe save: %w", err)
	}

	width := save.Width
	height := save.Height
	if width <= 0 || height <= 0 {
		return nil, nil
	}

	clues, err := parseNurikabeClues(save.Clues, width, height)
	if err != nil {
		return nil, err
	}

	return &NurikabeData{
		Width:  width,
		Height: height,
		Clues:  clues,
	}, nil
}

func ParseShikakuPrintData(saveData []byte) (*ShikakuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save shikakuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode shikaku save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, nil
	}

	clues := make([][]int, save.Height)
	for y := 0; y < save.Height; y++ {
		clues[y] = make([]int, save.Width)
	}

	for _, clue := range save.Clues {
		if clue.X < 0 || clue.X >= save.Width || clue.Y < 0 || clue.Y >= save.Height {
			continue
		}
		if clue.Value <= 0 {
			continue
		}
		clues[clue.Y][clue.X] = clue.Value
	}

	return &ShikakuData{
		Width:  save.Width,
		Height: save.Height,
		Clues:  clues,
	}, nil
}

func ParseHashiPrintData(saveData []byte) (*HashiData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save hashiSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode hashiwokakero save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, nil
	}

	islands := make([]HashiIsland, 0, len(save.Islands))
	for _, island := range save.Islands {
		if island.X < 0 || island.X >= save.Width || island.Y < 0 || island.Y >= save.Height {
			continue
		}
		if island.Required <= 0 {
			continue
		}
		islands = append(islands, HashiIsland(island))
	}

	return &HashiData{
		Width:   save.Width,
		Height:  save.Height,
		Islands: islands,
	}, nil
}

func ParseHitoriPrintData(saveData []byte) (*HitoriData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save hitoriSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode hitori save: %w", err)
	}

	rows := splitNormalizedLines(save.Numbers)
	size := save.Size
	if size <= 0 {
		size = len(rows)
	}
	if size <= 0 {
		return nil, nil
	}

	numbers := make([][]string, size)
	for y := 0; y < size; y++ {
		numbers[y] = make([]string, size)
		if y >= len(rows) {
			continue
		}

		rowValues := parseHitoriRowValues(rows[y])
		for x := 0; x < size && x < len(rowValues); x++ {
			numbers[y][x] = rowValues[x]
		}
	}

	return &HitoriData{
		Size:    size,
		Numbers: numbers,
	}, nil
}

func ParseTakuzuPrintData(saveData []byte) (*TakuzuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save takuzuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode takuzu save: %w", err)
	}

	stateRows := splitNormalizedLines(save.State)
	providedRows := splitNormalizedLines(save.Provided)

	size := save.Size
	if size <= 0 {
		size = max(len(stateRows), len(providedRows))
	}
	if size <= 0 {
		return nil, nil
	}

	givens := make([][]string, size)
	for y := 0; y < size; y++ {
		givens[y] = make([]string, size)

		var stateRunes []rune
		if y < len(stateRows) {
			stateRunes = []rune(stateRows[y])
		}

		var providedRunes []rune
		if y < len(providedRows) {
			providedRunes = []rune(providedRows[y])
		}

		for x := 0; x < size; x++ {
			if x >= len(providedRunes) || providedRunes[x] != '#' {
				continue
			}
			if x >= len(stateRunes) {
				continue
			}
			if stateRunes[x] != '0' && stateRunes[x] != '1' {
				continue
			}
			givens[y][x] = string(stateRunes[x])
		}
	}

	return &TakuzuData{
		Size:          size,
		Givens:        givens,
		GroupEveryTwo: true,
	}, nil
}

func ParseSudokuPrintData(saveData []byte) (*SudokuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save sudokuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode sudoku save: %w", err)
	}

	var givens [9][9]int
	for _, cell := range save.Provided {
		if !isSudokuCellInBounds(cell.X, cell.Y) {
			continue
		}
		if cell.V < 1 || cell.V > 9 {
			continue
		}
		givens[cell.Y][cell.X] = cell.V
	}

	return &SudokuData{Givens: givens}, nil
}

func ParseWordSearchPrintData(saveData []byte) (*WordSearchData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save wordSearchSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode word search save: %w", err)
	}

	rows := strings.Split(strings.ReplaceAll(strings.ReplaceAll(save.Grid, "\r\n", "\n"), "\r", "\n"), "\n")
	if len(rows) == 1 && rows[0] == "" {
		rows = nil
	}

	width := save.Width
	for _, row := range rows {
		if n := len([]rune(row)); n > width {
			width = n
		}
	}

	height := save.Height
	if len(rows) > height {
		height = len(rows)
	}
	if width <= 0 || height <= 0 {
		return nil, nil
	}

	grid := make([][]string, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]string, width)
		runes := []rune{}
		if y < len(rows) {
			runes = []rune(rows[y])
		}
		for x := 0; x < width; x++ {
			grid[y][x] = " "
			if x >= len(runes) {
				continue
			}
			r := runes[x]
			if unicode.IsSpace(r) {
				continue
			}
			grid[y][x] = string(unicode.ToUpper(r))
		}
	}

	words := make([]string, 0, len(save.Words))
	for _, word := range save.Words {
		text := strings.ToUpper(strings.TrimSpace(word.Text))
		if text == "" {
			continue
		}
		words = append(words, text)
	}

	return &WordSearchData{
		Width:  width,
		Height: height,
		Grid:   grid,
		Words:  words,
	}, nil
}

func isSudokuCellInBounds(x, y int) bool {
	return x >= 0 && x < 9 && y >= 0 && y < 9
}

func splitNormalizedLines(raw string) []string {
	normalized := strings.ReplaceAll(strings.ReplaceAll(raw, "\r\n", "\n"), "\r", "\n")
	if strings.TrimSpace(normalized) == "" {
		return nil
	}
	return strings.Split(normalized, "\n")
}

func parseHitoriRowValues(row string) []string {
	row = strings.TrimSpace(row)
	if row == "" {
		return nil
	}

	if strings.Contains(row, " ") || strings.Contains(row, ",") {
		fields := strings.Fields(strings.ReplaceAll(row, ",", " "))
		if len(fields) > 1 {
			values := make([]string, len(fields))
			for i, field := range fields {
				values[i] = normalizeHitoriToken(field)
			}
			return values
		}
	}

	runes := []rune(row)
	values := make([]string, len(runes))
	for i, r := range runes {
		values[i] = normalizeHitoriRune(r)
	}
	return values
}

func normalizeHitoriToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" || token == "." {
		return ""
	}
	if utf8.RuneCountInString(token) == 1 {
		r, _ := utf8.DecodeRuneInString(token)
		return normalizeHitoriRune(r)
	}
	return token
}

func normalizeHitoriRune(r rune) string {
	switch {
	case r == '.':
		return ""
	case r >= '0' && r <= '9':
		return string(r)
	default:
		value := int(r - '0')
		if value >= 10 && value <= 35 {
			return fmt.Sprintf("%d", value)
		}
		return string(r)
	}
}

func parseNurikabeClues(raw string, width, height int) ([][]int, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid clue dimensions: %dx%d", width, height)
	}

	clues := make([][]int, height)
	for y := 0; y < height; y++ {
		clues[y] = make([]int, width)
	}

	rows := splitNormalizedLines(raw)
	for y := 0; y < len(rows) && y < height; y++ {
		parts := strings.Split(rows[y], ",")
		for x := 0; x < len(parts) && x < width; x++ {
			token := strings.TrimSpace(parts[x])
			if token == "" {
				continue
			}
			value, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("invalid clue value %q at (%d,%d): %w", token, x, y, err)
			}
			if value < 0 {
				return nil, fmt.Errorf("negative clue value %d at (%d,%d)", value, x, y)
			}
			clues[y][x] = value
		}
	}

	return clues, nil
}

func splitNonogramStateRows(raw string) []string {
	normalized := strings.ReplaceAll(strings.ReplaceAll(raw, "\r\n", "\n"), "\r", "\n")
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, "\n")
}

func maxRuneWidth(rows []string) int {
	maxWidth := 0
	for _, row := range rows {
		if n := len([]rune(row)); n > maxWidth {
			maxWidth = n
		}
	}
	return maxWidth
}

func normalizeNonogramHintRows(src [][]int, size int) [][]int {
	if size <= 0 {
		return nil
	}

	normalized := make([][]int, size)
	for i := 0; i < size; i++ {
		if i >= len(src) {
			normalized[i] = []int{0}
			continue
		}

		filtered := make([]int, 0, len(src[i]))
		for _, value := range src[i] {
			if value > 0 {
				filtered = append(filtered, value)
			}
		}
		if len(filtered) == 0 {
			filtered = []int{0}
		}
		normalized[i] = filtered
	}

	return normalized
}

func normalizeNonogramStateGrid(rows []string, width, height int) [][]string {
	if width <= 0 || height <= 0 {
		return nil
	}

	grid := make([][]string, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]string, width)
		for x := 0; x < width; x++ {
			grid[y][x] = " "
		}

		if y >= len(rows) {
			continue
		}

		runes := []rune(rows[y])
		for x := 0; x < width && x < len(runes); x++ {
			if runes[x] == ' ' {
				continue
			}
			grid[y][x] = string(runes[x])
		}
	}

	return grid
}

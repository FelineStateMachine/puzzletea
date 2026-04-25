package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type wordSearchSave struct {
	Width  int              `json:"width"`
	Height int              `json:"height"`
	Grid   string           `json:"grid"`
	Words  []wordSearchWord `json:"words"`
}

type wordSearchWord struct {
	Text string `json:"text"`
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
	height = max(height, len(rows))
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

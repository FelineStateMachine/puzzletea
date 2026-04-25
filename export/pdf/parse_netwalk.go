package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type netwalkSave struct {
	Size             int    `json:"size"`
	Masks            string `json:"masks"`
	InitialRotations string `json:"initial_rotations"`
	Kinds            string `json:"kinds"`
}

func ParseNetwalkPrintData(saveData []byte) (*NetwalkData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save netwalkSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode netwalk save: %w", err)
	}
	if save.Size <= 0 {
		return nil, nil
	}

	maskRows, err := parseNetwalkRows(save.Size, save.Masks)
	if err != nil {
		return nil, fmt.Errorf("parse netwalk masks: %w", err)
	}
	rotationRows, err := parseNetwalkRows(save.Size, save.InitialRotations)
	if err != nil {
		return nil, fmt.Errorf("parse netwalk rotations: %w", err)
	}
	kindRows, err := parseNetwalkRows(save.Size, save.Kinds)
	if err != nil {
		return nil, fmt.Errorf("parse netwalk kinds: %w", err)
	}

	data := &NetwalkData{
		Size:      save.Size,
		Masks:     make([][]uint8, save.Size),
		Rotations: make([][]uint8, save.Size),
		RootX:     -1,
		RootY:     -1,
	}

	for y := 0; y < save.Size; y++ {
		data.Masks[y] = make([]uint8, save.Size)
		data.Rotations[y] = make([]uint8, save.Size)
		for x := 0; x < save.Size; x++ {
			mask, ok := parseHexNibble(maskRows[y][x])
			if !ok {
				return nil, fmt.Errorf("invalid mask value %q at (%d,%d)", maskRows[y][x], x, y)
			}
			if rotationRows[y][x] < '0' || rotationRows[y][x] > '3' {
				return nil, fmt.Errorf("invalid rotation value %q at (%d,%d)", rotationRows[y][x], x, y)
			}
			data.Masks[y][x] = mask
			data.Rotations[y][x] = rotationRows[y][x] - '0'
			if kindRows[y][x] == 'S' {
				data.RootX = x
				data.RootY = y
			}
		}
	}

	if data.RootX < 0 || data.RootY < 0 {
		return nil, fmt.Errorf("netwalk print data missing root")
	}

	return data, nil
}

func parseNetwalkRows(size int, raw string) ([][]byte, error) {
	rows := splitNormalizedLines(raw)
	if len(rows) != size {
		return nil, fmt.Errorf("row count = %d, want %d", len(rows), size)
	}

	result := make([][]byte, size)
	for y, row := range rows {
		if len(row) != size {
			return nil, fmt.Errorf("row %d width = %d, want %d", y, len(row), size)
		}
		result[y] = []byte(row)
	}
	return result, nil
}

func parseHexNibble(value byte) (uint8, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return 10 + value - 'a', true
	case value >= 'A' && value <= 'F':
		return 10 + value - 'A', true
	default:
		return 0, false
	}
}

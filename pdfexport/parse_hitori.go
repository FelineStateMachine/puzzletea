package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

type hitoriSave struct {
	Size    int    `json:"size"`
	Numbers string `json:"numbers"`
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

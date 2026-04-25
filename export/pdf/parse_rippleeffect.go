package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type rippleEffectSave struct {
	Width  int                `json:"width"`
	Height int                `json:"height"`
	Givens string             `json:"givens"`
	Cages  []RippleEffectCage `json:"cages"`
}

func ParseRippleEffectPrintData(saveData []byte) (*RippleEffectData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save rippleEffectSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode ripple effect save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, nil
	}

	givens, err := parseNumberGrid(save.Givens, save.Width, save.Height)
	if err != nil {
		return nil, err
	}

	return &RippleEffectData{
		Width:  save.Width,
		Height: save.Height,
		Givens: givens,
		Cages:  append([]RippleEffectCage(nil), save.Cages...),
	}, nil
}

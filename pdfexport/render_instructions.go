package pdfexport

import (
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

const instructionWrapInsetMM = 0.0

func wrapInstructionLines(pdf *fpdf.Fpdf, width float64, rules []string) []string {
	if pdf == nil || width <= 0 {
		return nil
	}

	setInstructionStyle(pdf)

	wrapWidth := width - instructionWrapInsetMM*2
	if wrapWidth <= 0 {
		wrapWidth = width
	}

	lines := make([]string, 0, len(rules))
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		chunks := pdf.SplitLines([]byte(rule), wrapWidth)
		if len(chunks) == 0 {
			lines = append(lines, rule)
			continue
		}

		for _, chunk := range chunks {
			line := strings.TrimSpace(string(chunk))
			if line != "" {
				lines = append(lines, line)
			}
		}
	}

	return lines
}

func InstructionLineCount(pdf *fpdf.Fpdf, width float64, rules []string) int {
	return len(wrapInstructionLines(pdf, width, rules))
}

func RenderInstructions(pdf *fpdf.Fpdf, x, y, width float64, rules []string) int {
	lines := wrapInstructionLines(pdf, width, rules)
	if len(lines) == 0 {
		return 0
	}

	setInstructionStyle(pdf)
	for i, line := range lines {
		pdf.SetXY(x, y+float64(i)*InstructionLineHMM)
		pdf.CellFormat(width, InstructionLineHMM, line, "", 0, "C", false, 0, "")
	}

	return len(lines)
}

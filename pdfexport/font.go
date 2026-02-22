package pdfexport

const (
	standardCellFontMin = 4.2
	standardCellFontMax = 8.0
)

func standardCellFontSize(cellSize, scale float64) float64 {
	return clampStandardCellFontSize(cellSize * scale)
}

func clampStandardCellFontSize(fontSize float64) float64 {
	if fontSize < standardCellFontMin {
		return standardCellFontMin
	}
	if fontSize > standardCellFontMax {
		return standardCellFontMax
	}
	return fontSize
}

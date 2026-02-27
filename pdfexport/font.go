package pdfexport

const (
	standardCellFontMin = 5.2
	standardCellFontMax = 8.2
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

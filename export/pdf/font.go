package pdfexport

const (
	pdfFontSizeDelta    = 3.0
	standardCellFontMin = 5.2 + pdfFontSizeDelta
	standardCellFontMax = 8.2 + pdfFontSizeDelta
)

func standardCellFontSize(cellSize, scale float64) float64 {
	return clampStandardCellFontSize(cellSize*scale + pdfFontSizeDelta)
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

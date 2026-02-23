package pdfexport

import "github.com/go-pdf/fpdf"

func RenderNonogramPage(pdf *fpdf.Fpdf, data *NonogramData) {
	renderNonogramPage(pdf, data)
}

func RenderNurikabePage(pdf *fpdf.Fpdf, data *NurikabeData) {
	renderNurikabePage(pdf, data)
}

func RenderShikakuPage(pdf *fpdf.Fpdf, data *ShikakuData) {
	renderShikakuPage(pdf, data)
}

func RenderHashiPage(pdf *fpdf.Fpdf, data *HashiData) {
	renderHashiPage(pdf, data)
}

func RenderHitoriPage(pdf *fpdf.Fpdf, data *HitoriData) {
	renderHitoriPage(pdf, data)
}

func RenderTakuzuPage(pdf *fpdf.Fpdf, data *TakuzuData) {
	renderTakuzuPage(pdf, data)
}

func RenderSudokuPage(pdf *fpdf.Fpdf, data *SudokuData) {
	renderSudokuPage(pdf, data)
}

func RenderWordSearchPage(pdf *fpdf.Fpdf, data *WordSearchData) {
	renderWordSearchPage(pdf, data)
}

func RenderGridTablePage(pdf *fpdf.Fpdf, table *GridTable) {
	renderGridTablePage(pdf, table)
}

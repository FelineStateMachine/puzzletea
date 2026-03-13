package builtinprint

import (
	"sync"

	"github.com/FelineStateMachine/puzzletea/fillomino"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/nurikabe"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/rippleeffect"
	"github.com/FelineStateMachine/puzzletea/shikaku"
	"github.com/FelineStateMachine/puzzletea/spellpuzzle"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/sudokurgb"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/takuzuplus"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
)

var registerOnce sync.Once

func Register() {
	registerOnce.Do(func() {
		adapters := []pdfexport.PrintAdapter{
			fillomino.PDFPrintAdapter,
			hashiwokakero.PDFPrintAdapter,
			hitori.PDFPrintAdapter,
			nonogram.PDFPrintAdapter,
			nurikabe.PDFPrintAdapter,
			rippleeffect.PDFPrintAdapter,
			shikaku.PDFPrintAdapter,
			spellpuzzle.PDFPrintAdapter,
			sudoku.PDFPrintAdapter,
			sudokurgb.PDFPrintAdapter,
			takuzu.PDFPrintAdapter,
			takuzuplus.PDFPrintAdapter,
			wordsearch.PDFPrintAdapter,
		}
		for _, adapter := range adapters {
			pdfexport.RegisterPrintAdapter(adapter)
		}
	})
}

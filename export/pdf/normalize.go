package pdfexport

import "github.com/FelineStateMachine/puzzletea/puzzle"

func normalizeToken(s string) string {
	return puzzle.NormalizeName(s)
}

package game

import "github.com/FelineStateMachine/puzzletea/pdfexport"

type PrintAdapter = pdfexport.PrintAdapter

func RegisterPrintAdapter(adapter PrintAdapter) {
	pdfexport.RegisterPrintAdapter(adapter)
}

func LookupPrintAdapter(gameType string) (PrintAdapter, bool) {
	return pdfexport.LookupPrintAdapter(gameType)
}

func HasPrintAdapter(gameType string) bool {
	return pdfexport.HasPrintAdapter(gameType)
}

func IsNilPrintPayload(payload any) bool {
	return pdfexport.IsNilPrintPayload(payload)
}

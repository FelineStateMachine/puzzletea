package pdfexport

import (
	"reflect"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

type PrintAdapter interface {
	CanonicalGameType() string
	Aliases() []string
	BuildPDFPayload(save []byte) (any, error)
	RenderPDFBody(pdf *fpdf.Fpdf, payload any) error
}

var printAdapterRegistry = map[string]PrintAdapter{}

func RegisterPrintAdapter(adapter PrintAdapter) {
	canonical := normalizeGameTypeToken(adapter.CanonicalGameType())
	if canonical == "" {
		return
	}

	printAdapterRegistry[canonical] = adapter
	for _, alias := range adapter.Aliases() {
		normalized := normalizeGameTypeToken(alias)
		if normalized == "" {
			continue
		}
		printAdapterRegistry[normalized] = adapter
	}
}

func LookupPrintAdapter(gameType string) (PrintAdapter, bool) {
	adapter, ok := printAdapterRegistry[normalizeGameTypeToken(gameType)]
	return adapter, ok
}

func HasPrintAdapter(gameType string) bool {
	_, ok := LookupPrintAdapter(gameType)
	return ok
}

func IsNilPrintPayload(payload any) bool {
	if payload == nil {
		return true
	}

	v := reflect.ValueOf(payload)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func normalizeGameTypeToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}

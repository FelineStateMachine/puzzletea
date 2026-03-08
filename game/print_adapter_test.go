package game

import (
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

type testPrintAdapter struct {
	canonical string
	aliases   []string
}

func (a testPrintAdapter) CanonicalGameType() string           { return a.canonical }
func (a testPrintAdapter) Aliases() []string                   { return a.aliases }
func (a testPrintAdapter) BuildPDFPayload([]byte) (any, error) { return nil, nil }
func (a testPrintAdapter) RenderPDFBody(*fpdf.Fpdf, any) error { return nil }

func TestPrintAdapterRegistryLookupAndAliases(t *testing.T) {
	adapter := testPrintAdapter{
		canonical: "Test Word Search",
		aliases:   []string{"testwordsearch", "test-word-search"},
	}
	RegisterPrintAdapter(adapter)

	if !HasPrintAdapter("Test Word Search") {
		t.Fatal("expected canonical lookup to be supported")
	}
	if !HasPrintAdapter("test_word_search") {
		t.Fatal("expected underscore alias lookup to be supported")
	}
	if !HasPrintAdapter("testwordsearch") {
		t.Fatal("expected compact alias lookup to be supported")
	}
	if HasPrintAdapter("lights out") {
		t.Fatal("expected unknown type to be unsupported")
	}
}

func TestRegisterPrintAdapterSkipsBlankCanonical(t *testing.T) {
	RegisterPrintAdapter(testPrintAdapter{canonical: "  "})
	if HasPrintAdapter("should-not-exist") {
		t.Fatal("blank canonical adapter should not create lookups")
	}
}

func TestIsNilPrintPayload(t *testing.T) {
	var ptr *int
	if !IsNilPrintPayload(ptr) {
		t.Fatal("expected typed nil pointer to be treated as nil payload")
	}
	if IsNilPrintPayload(5) {
		t.Fatal("expected concrete payload to be non-nil")
	}
}

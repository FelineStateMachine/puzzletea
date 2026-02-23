package game

import (
	"testing"

	"github.com/go-pdf/fpdf"
)

type testPrintAdapter struct {
	canonical string
	aliases   []string
}

func (a testPrintAdapter) CanonicalGameType() string { return a.canonical }
func (a testPrintAdapter) Aliases() []string         { return a.aliases }
func (a testPrintAdapter) RenderMarkdownSnippet([]byte) (string, error) {
	return "", nil
}
func (a testPrintAdapter) BuildPDFPayload([]byte, string) (any, error) { return nil, nil }
func (a testPrintAdapter) RenderPDFBody(*fpdf.Fpdf, any) error         { return nil }

func TestPrintAdapterRegistryLookupAndAliases(t *testing.T) {
	original := printAdapterRegistry
	printAdapterRegistry = map[string]PrintAdapter{}
	t.Cleanup(func() { printAdapterRegistry = original })

	adapter := testPrintAdapter{
		canonical: "Word Search",
		aliases:   []string{"wordsearch", "word-search"},
	}
	RegisterPrintAdapter(adapter)

	if !HasPrintAdapter("Word Search") {
		t.Fatal("expected canonical lookup to be supported")
	}
	if !HasPrintAdapter("word_search") {
		t.Fatal("expected underscore alias lookup to be supported")
	}
	if !HasPrintAdapter("wordsearch") {
		t.Fatal("expected compact alias lookup to be supported")
	}
	if HasPrintAdapter("lights out") {
		t.Fatal("expected unknown type to be unsupported")
	}
}

func TestRegisterPrintAdapterSkipsBlankCanonical(t *testing.T) {
	original := printAdapterRegistry
	printAdapterRegistry = map[string]PrintAdapter{}
	t.Cleanup(func() { printAdapterRegistry = original })

	RegisterPrintAdapter(testPrintAdapter{canonical: "  "})
	if len(printAdapterRegistry) != 0 {
		t.Fatalf("registry size = %d, want 0", len(printAdapterRegistry))
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

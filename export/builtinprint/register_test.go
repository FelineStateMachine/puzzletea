package builtinprint

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/export/pdf"
	"github.com/FelineStateMachine/puzzletea/registry"
)

func TestRegisterBuiltins(t *testing.T) {
	Register()

	for _, entry := range registry.Entries() {
		hasAdapter := pdfexport.HasPrintAdapter(entry.Definition.Name)
		if entry.Print == nil && hasAdapter {
			t.Fatalf("unexpected print adapter for %q", entry.Definition.Name)
		}
		if entry.Print != nil && !hasAdapter {
			t.Fatalf("expected print adapter for %q", entry.Definition.Name)
		}
	}
}

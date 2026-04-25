// Package builtinprint registers the built-in print adapters exposed by the
// game registry into the export pipeline.
package builtinprint

import (
	"sync"

	"github.com/FelineStateMachine/puzzletea/export/pdf"
	"github.com/FelineStateMachine/puzzletea/registry"
)

var registerOnce sync.Once

func Register() {
	registerOnce.Do(func() {
		for _, entry := range registry.Entries() {
			if entry.Print == nil {
				continue
			}
			pdfexport.RegisterPrintAdapter(entry.Print)
		}
	})
}

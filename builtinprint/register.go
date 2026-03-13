package builtinprint

import (
	"sync"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
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

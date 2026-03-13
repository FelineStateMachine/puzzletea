package app

import (
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/weekly"
)

type gameOpenOptions struct {
	readOnly    bool
	returnState viewState
	weeklyInfo  *weekly.Info
}

func (m model) importAndActivateRecord(rec store.GameRecord) (model, bool) {
	return m.importAndActivateRecordWithOptions(rec, gameOpenOptions{returnState: mainMenuView})
}

func (m model) importAndActivateRecordWithOptions(rec store.GameRecord, options gameOpenOptions) (model, bool) {
	ok := newSessionController(&m).importRecord(rec, options)
	return m, ok
}

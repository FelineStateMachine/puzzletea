package fillomino

import "testing"

func TestImportRejectsEmptyProvidedCell(t *testing.T) {
	_, err := ImportModel([]byte(`{"width":2,"height":2,"state":". .\n. .","provided":"#.\n..","mode_title":"Test"}`))
	if err == nil {
		t.Fatal("expected import error")
	}
}

package main_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestPuzzlePackageStaysPure(t *testing.T) {
	assertPackageDoesNotImport(t, "puzzle", []string{
		"charm.land/bubbletea",
		"charm.land/bubbles",
		"charm.land/lipgloss",
		"github.com/FelineStateMachine/puzzletea/theme",
		"github.com/FelineStateMachine/puzzletea/pdfexport",
	})
}

func TestStatsPackageAvoidsRenderingImports(t *testing.T) {
	assertPackageDoesNotImport(t, "stats", []string{
		"charm.land/bubbletea",
		"charm.land/bubbles",
		"charm.land/lipgloss",
		"github.com/FelineStateMachine/puzzletea/theme",
		"github.com/FelineStateMachine/puzzletea/ui",
	})
}

func TestStorePackageDoesNotImportSchedulePackages(t *testing.T) {
	assertPackageDoesNotImport(t, "store", []string{
		"github.com/FelineStateMachine/puzzletea/daily",
		"github.com/FelineStateMachine/puzzletea/weekly",
	})
}

func TestCatalogPackageDoesNotImportConcreteGames(t *testing.T) {
	assertPackageDoesNotImport(t, "catalog", []string{
		"github.com/FelineStateMachine/puzzletea/fillomino",
		"github.com/FelineStateMachine/puzzletea/hashiwokakero",
		"github.com/FelineStateMachine/puzzletea/hitori",
		"github.com/FelineStateMachine/puzzletea/lightsout",
		"github.com/FelineStateMachine/puzzletea/nonogram",
		"github.com/FelineStateMachine/puzzletea/nurikabe",
		"github.com/FelineStateMachine/puzzletea/rippleeffect",
		"github.com/FelineStateMachine/puzzletea/shikaku",
		"github.com/FelineStateMachine/puzzletea/sudoku",
		"github.com/FelineStateMachine/puzzletea/sudokurgb",
		"github.com/FelineStateMachine/puzzletea/takuzu",
		"github.com/FelineStateMachine/puzzletea/takuzuplus",
		"github.com/FelineStateMachine/puzzletea/wordsearch",
	})
}

func assertPackageDoesNotImport(t *testing.T, dir string, forbidden []string) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imp := range file.Imports {
			value := strings.Trim(imp.Path.Value, `"`)
			for _, prefix := range forbidden {
				if strings.HasPrefix(value, prefix) {
					t.Fatalf("%s imports forbidden package %q", path, value)
				}
			}
		}
	}
}

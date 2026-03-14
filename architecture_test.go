package main_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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

func TestGamePackageDoesNotImportPDFExport(t *testing.T) {
	assertPackageDoesNotImport(t, "game", []string{
		"github.com/FelineStateMachine/puzzletea/pdfexport",
	})
}

func TestCatalogPackageDoesNotImportConcreteGames(t *testing.T) {
	assertPackageDoesNotImport(t, "catalog", concreteGameImportPaths(t))
}

func TestBuiltinPrintDoesNotImportConcreteGames(t *testing.T) {
	assertPackageDoesNotImport(t, "builtinprint", concreteGameImportPaths(t))
}

func TestSessionPackageDoesNotUseNameDerivedRunMetadata(t *testing.T) {
	assertFilesDoNotContain(t, "session", []string{
		"RunKindForName(",
		"RunDateForName(",
		"SeedTextForName(",
		"WeeklyIdentityForName(",
	})
}

func TestStoreCreateGameDoesNotUseNameDerivedRunMetadata(t *testing.T) {
	assertFileDoesNotContain(t, filepath.Join("store", "store.go"), []string{
		"RunKindForName(",
		"RunDateForName(",
		"SeedTextForName(",
		"WeeklyIdentityForName(",
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

func concreteGameImportPaths(t testing.TB) []string {
	t.Helper()

	pattern := filepath.Join(repoRoot(t), "*", "gamemode.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob concrete game packages: %v", err)
	}

	importPaths := make([]string, 0, len(matches))
	for _, match := range matches {
		dir := filepath.Base(filepath.Dir(match))
		importPaths = append(importPaths, "github.com/FelineStateMachine/puzzletea/"+dir)
	}
	slices.Sort(importPaths)
	return importPaths
}

func repoRoot(t testing.TB) string {
	t.Helper()

	_, path, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Dir(path))
}

func assertFileDoesNotContain(t *testing.T, path string, forbidden []string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	content := string(data)
	for _, pattern := range forbidden {
		if strings.Contains(content, pattern) {
			t.Fatalf("%s contains forbidden pattern %q", path, pattern)
		}
	}
}

func assertFilesDoNotContain(t *testing.T, dir string, forbidden []string) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(data)
		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				t.Fatalf("%s contains forbidden pattern %q", path, pattern)
			}
		}
	}
}

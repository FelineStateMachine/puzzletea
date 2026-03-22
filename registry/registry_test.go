package registry

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestResolveNormalizesSpacingAndAliases(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "word   search", want: "Word Search"},
		{name: "hashi", want: "Hashiwokakero"},
		{name: "network", want: "Netwalk"},
		{name: "polyomino", want: "Fillomino"},
	}

	for _, tt := range tests {
		entry, ok := Resolve(tt.name)
		if !ok {
			t.Fatalf("Resolve(%q) = false, want true", tt.name)
		}
		if got := entry.Definition.Name; got != tt.want {
			t.Fatalf("Resolve(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestModeSeededFlagMatchesSpawnerAvailability(t *testing.T) {
	for _, entry := range Entries() {
		for _, mode := range entry.Modes {
			if got, want := mode.Definition.Seeded, mode.Seeded != nil; got != want {
				t.Fatalf("%s/%s seeded flag = %v, want %v",
					entry.Definition.Name, mode.Definition.Title, got, want)
			}
		}
	}
}

func TestEntriesStayAlignedWithDefinitions(t *testing.T) {
	definitions := Definitions()
	entries := Entries()
	if got, want := len(entries), len(definitions); got != want {
		t.Fatalf("len(Entries()) = %d, want %d", got, want)
	}

	for _, def := range definitions {
		entry, ok := Lookup(def.Name)
		if !ok {
			t.Fatalf("Lookup(%q) = false", def.Name)
		}
		if entry.Import == nil {
			t.Fatalf("%s missing import function", def.Name)
		}
		if got, want := len(entry.Modes), len(def.Modes); got != want {
			t.Fatalf("%s mode count = %d, want %d", def.Name, got, want)
		}
	}
}

func TestRegistryAllMatchesConcreteGamePackages(t *testing.T) {
	got := registryAllImportPaths(t)
	want := concreteGameImportPaths(t)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registry all import paths = %v, want %v", got, want)
	}
}

func TestPrintAdaptersStayAlignedWithRegistryEntries(t *testing.T) {
	for _, entry := range Entries() {
		if entry.Definition.Name == "Lights Out" {
			if entry.Print != nil {
				t.Fatal("Lights Out should not expose a print adapter")
			}
			continue
		}
		if entry.Print == nil {
			t.Fatalf("%s missing print adapter", entry.Definition.Name)
		}
	}
}

func registryAllImportPaths(t *testing.T) []string {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "registry.go", nil, 0)
	if err != nil {
		t.Fatalf("parse registry.go: %v", err)
	}

	imports := make(map[string]string)
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		name := importName(path, imp.Name)
		imports[name] = path
	}

	paths := registryAllSelectorImportPaths(t, file, imports)
	slices.Sort(paths)
	return paths
}

func registryAllSelectorImportPaths(t *testing.T, file *ast.File, imports map[string]string) []string {
	t.Helper()

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != "all" || len(valueSpec.Values) != 1 {
				continue
			}

			lit, ok := valueSpec.Values[0].(*ast.CompositeLit)
			if !ok {
				t.Fatalf("registry all value is %T, want composite literal", valueSpec.Values[0])
			}

			paths := make([]string, 0, len(lit.Elts))
			for _, elt := range lit.Elts {
				selector, ok := elt.(*ast.SelectorExpr)
				if !ok {
					t.Fatalf("registry all element is %T, want selector", elt)
				}
				pkg, ok := selector.X.(*ast.Ident)
				if !ok {
					t.Fatalf("registry all selector receiver is %T, want identifier", selector.X)
				}

				path, ok := imports[pkg.Name]
				if !ok {
					t.Fatalf("registry all package %q missing from imports", pkg.Name)
				}
				paths = append(paths, path)
			}
			return paths
		}
	}

	t.Fatal("registry all declaration not found")
	return nil
}

func importName(path string, name *ast.Ident) string {
	if name != nil {
		return name.Name
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
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
	return filepath.Clean(filepath.Join(filepath.Dir(path), ".."))
}

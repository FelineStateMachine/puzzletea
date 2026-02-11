package namegen

import (
	"strings"
	"testing"
)

func TestGenerate_Format(t *testing.T) {
	for range 50 {
		name := Generate()

		if strings.Count(name, "-") < 1 {
			t.Errorf("expected at least one hyphen in %q", name)
		}

		parts := strings.SplitN(name, "-", 2)
		if len(parts) != 2 {
			t.Fatalf("expected 2 parts, got %d for %q", len(parts), name)
		}
		if parts[0] == "" {
			t.Errorf("adjective part is empty in %q", name)
		}
		if parts[1] == "" {
			t.Errorf("noun part is empty in %q", name)
		}
		if name != strings.ToLower(name) {
			t.Errorf("expected all lowercase, got %q", name)
		}
		if strings.Contains(name, " ") {
			t.Errorf("expected no spaces in %q", name)
		}
	}
}

func TestGenerate_ValidWords(t *testing.T) {
	adjSet := make(map[string]bool, len(adjectives))
	for _, a := range adjectives {
		adjSet[a] = true
	}
	nounSet := make(map[string]bool, len(nouns))
	for _, n := range nouns {
		nounSet[n] = true
	}

	for range 50 {
		name := Generate()
		parts := strings.SplitN(name, "-", 2)

		if !adjSet[parts[0]] {
			t.Errorf("adjective %q not in adjectives list", parts[0])
		}
		if !nounSet[parts[1]] {
			t.Errorf("noun %q not in nouns list", parts[1])
		}
	}
}

func TestGenerate_Variety(t *testing.T) {
	seen := make(map[string]bool)
	for range 20 {
		seen[Generate()] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected at least 2 distinct names from 20 generations, got %d", len(seen))
	}
}

func TestWordLists(t *testing.T) {
	t.Run("adjectives not empty", func(t *testing.T) {
		if len(adjectives) == 0 {
			t.Fatal("adjectives list is empty")
		}
	})

	t.Run("nouns not empty", func(t *testing.T) {
		if len(nouns) == 0 {
			t.Fatal("nouns list is empty")
		}
	})

	t.Run("no duplicate adjectives", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, a := range adjectives {
			if seen[a] {
				t.Errorf("duplicate adjective: %q", a)
			}
			seen[a] = true
		}
	})

	t.Run("no duplicate nouns", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, n := range nouns {
			if seen[n] {
				t.Errorf("duplicate noun: %q", n)
			}
			seen[n] = true
		}
	})
}

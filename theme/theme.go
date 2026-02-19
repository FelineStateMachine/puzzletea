// Package theme provides color theming for PuzzleTea. It loads terminal
// color schemes from an embedded JSON file and exposes a semantic [Palette]
// derived from the active theme's 16-color ANSI palette.
package theme

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Theme represents a terminal color scheme with the standard 16 ANSI colors
// plus background, foreground, and cursor colors.
type Theme struct {
	Name                string `json:"name"`
	Black               string `json:"black"`
	Red                 string `json:"red"`
	Green               string `json:"green"`
	Yellow              string `json:"yellow"`
	Blue                string `json:"blue"`
	Purple              string `json:"purple"`
	Cyan                string `json:"cyan"`
	White               string `json:"white"`
	BrightBlack         string `json:"brightBlack"`
	BrightRed           string `json:"brightRed"`
	BrightGreen         string `json:"brightGreen"`
	BrightYellow        string `json:"brightYellow"`
	BrightBlue          string `json:"brightBlue"`
	BrightPurple        string `json:"brightPurple"`
	BrightCyan          string `json:"brightCyan"`
	BrightWhite         string `json:"brightWhite"`
	Background          string `json:"background"`
	Foreground          string `json:"foreground"`
	CursorColor         string `json:"cursorColor,omitempty"`
	SelectionBackground string `json:"selectionBackground,omitempty"`
	Meta                struct {
		IsDark bool `json:"isDark"`
	} `json:"meta"`
}

// Palette returns the semantic color palette derived from this theme.
func (t Theme) Palette() Palette {
	return derivePalette(t)
}

// MaxNameLen is the length of the longest theme name (including the default).
// Useful for sizing UI elements like the theme picker list.
var MaxNameLen int

var (
	mu        sync.RWMutex
	current   Palette
	allThemes []Theme
	byName    map[string]int // lowercase name -> index in allThemes
)

func init() {
	allThemes = parseEmbeddedThemes()
	byName = make(map[string]int, len(allThemes)+1)
	MaxNameLen = len(DefaultThemeName)
	for i, t := range allThemes {
		byName[strings.ToLower(t.Name)] = i
		if len(t.Name) > MaxNameLen {
			MaxNameLen = len(t.Name)
		}
	}
	current = defaultPalette()
}

// Current returns the active palette. Safe for concurrent use.
func Current() Palette {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Apply activates the named theme. An empty name applies the built-in default.
// Returns an error if the theme name is not found.
func Apply(name string) error {
	mu.Lock()
	defer mu.Unlock()

	if name == "" || strings.EqualFold(name, DefaultThemeName) {
		current = defaultPalette()
		return nil
	}

	idx, ok := byName[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("unknown theme %q", name)
	}
	current = allThemes[idx].Palette()
	return nil
}

// AllThemes returns every available theme.
func AllThemes() []Theme { return allThemes }

// ThemeNames returns the list of all theme names, with the default first.
func ThemeNames() []string {
	names := make([]string, 0, len(allThemes)+1)
	names = append(names, DefaultThemeName)
	for _, t := range allThemes {
		names = append(names, t.Name)
	}
	return names
}

// LookupTheme returns the theme with the given name, or nil if not found.
func LookupTheme(name string) *Theme {
	idx, ok := byName[strings.ToLower(name)]
	if !ok {
		return nil
	}
	t := allThemes[idx]
	return &t
}

func parseEmbeddedThemes() []Theme {
	var themes []Theme
	if err := json.Unmarshal(themesJSON, &themes); err != nil {
		panic(fmt.Sprintf("theme: failed to parse embedded themes.json: %v", err))
	}
	return themes
}

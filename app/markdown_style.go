package app

import (
	"fmt"
	"image/color"

	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/glamour/ansi"
)

func helpMarkdownStyle(p theme.Palette) ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.FG),
			},
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.TextDim),
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
			Margin:      uintPtr(1),
		},
		Paragraph: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.FG),
			},
			Margin: uintPtr(1),
		},
		List: ansi.StyleList{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: colorPtr(p.FG),
				},
				Margin: uintPtr(1),
			},
			LevelIndent: 2,
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.Accent),
				Bold:  boolPtr(true),
			},
			Margin: uintPtr(1),
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.Accent),
				Bold:  boolPtr(true),
			},
			Margin: uintPtr(1),
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.AccentSoft),
				Bold:  boolPtr(true),
			},
			Margin: uintPtr(1),
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.Info),
				Bold:  boolPtr(true),
			},
			Margin: uintPtr(1),
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.Secondary),
				Bold:  boolPtr(true),
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.Linked),
				Bold:  boolPtr(true),
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.TextDim),
				Bold:  boolPtr(true),
			},
		},
		Text: ansi.StylePrimitive{
			Color: colorPtr(p.FG),
		},
		Emph: ansi.StylePrimitive{
			Color:  colorPtr(p.Secondary),
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Color: colorPtr(p.AccentSoft),
			Bold:  boolPtr(true),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color: colorPtr(p.Border),
		},
		Item: ansi.StylePrimitive{
			Color: colorPtr(p.Accent),
		},
		Enumeration: ansi.StylePrimitive{
			Color: colorPtr(p.Accent),
		},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(p.Accent),
			},
			Ticked:   "[x]",
			Unticked: "[ ]",
		},
		Link: ansi.StylePrimitive{
			Color:     colorPtr(p.Highlight),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color:     colorPtr(p.Highlight),
			Underline: boolPtr(true),
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           colorPtr(p.AccentSoft),
				BackgroundColor: colorPtr(p.Surface),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color:           colorPtr(p.FG),
					BackgroundColor: colorPtr(p.Surface),
				},
				Margin: uintPtr(1),
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: colorPtr(p.FG),
				},
				Margin: uintPtr(1),
			},
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionTerm: ansi.StylePrimitive{
			Color: colorPtr(p.Accent),
			Bold:  boolPtr(true),
		},
		DefinitionDescription: ansi.StylePrimitive{
			Color: colorPtr(p.FG),
		},
		Strikethrough: ansi.StylePrimitive{
			Color: colorPtr(p.TextDim),
		},
	}
}

func helpMarkdownThemeKey(p theme.Palette) string {
	return fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s:%s:%s:%s",
		hexColor(p.BG),
		hexColor(p.FG),
		hexColor(p.Surface),
		hexColor(p.Accent),
		hexColor(p.AccentSoft),
		hexColor(p.Info),
		hexColor(p.Highlight),
		hexColor(p.Border),
		hexColor(p.TextDim),
	)
}

func hexColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func colorPtr(c color.Color) *string {
	s := hexColor(c)
	return &s
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func uintPtr(v uint) *uint {
	return &v
}

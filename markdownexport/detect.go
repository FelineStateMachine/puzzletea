package markdownexport

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/nurikabe"
	"github.com/FelineStateMachine/puzzletea/shikaku"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
)

func DetectGameType(g game.Gamer) (string, error) {
	switch g.(type) {
	case hashiwokakero.Model, *hashiwokakero.Model:
		return "Hashiwokakero", nil
	case hitori.Model, *hitori.Model:
		return "Hitori", nil
	case lightsout.Model, *lightsout.Model:
		return "Lights Out", nil
	case nonogram.Model, *nonogram.Model:
		return "Nonogram", nil
	case nurikabe.Model, *nurikabe.Model:
		return "Nurikabe", nil
	case shikaku.Model, *shikaku.Model:
		return "Shikaku", nil
	case sudoku.Model, *sudoku.Model:
		return "Sudoku", nil
	case takuzu.Model, *takuzu.Model:
		return "Takuzu", nil
	case wordsearch.Model, *wordsearch.Model:
		return "Word Search", nil
	default:
		return "", fmt.Errorf("cannot detect game type for %T", g)
	}
}

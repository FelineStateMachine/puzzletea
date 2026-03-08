package session

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/resolve"
	"github.com/FelineStateMachine/puzzletea/store"
)

// NormalizeSeed keeps seeded names distinct from real dailies.
func NormalizeSeed(seed string) string {
	seed = strings.TrimSpace(seed)
	for _, reserved := range []string{"daily", "week"} {
		if strings.HasPrefix(strings.ToLower(seed), reserved) {
			return strings.ToLower(seed[:len(reserved)]) + seed[len(reserved):]
		}
	}
	return seed
}

// SeededName derives the deterministic display name for a seeded puzzle.
func SeededName(seed string) string {
	nameRNG := resolve.RNGFromString("name:" + seed)
	return seed + " - " + namegen.GenerateSeeded(nameRNG)
}

// SeededNameForGame derives the deterministic display name for a seeded
// puzzle locked to a specific game.
func SeededNameForGame(seed, gameType string) string {
	if strings.TrimSpace(gameType) == "" {
		return SeededName(seed)
	}

	scope := game.NormalizeName(gameType)
	nameRNG := resolve.RNGFromString("name:" + seed + ":" + scope)
	return fmt.Sprintf("%s [%s] - %s",
		seed,
		gameType,
		namegen.GenerateSeeded(nameRNG),
	)
}

// ImportRecord reconstructs a saved game and reapplies the record title.
func ImportRecord(rec *store.GameRecord) (game.Gamer, error) {
	if rec == nil {
		return nil, fmt.Errorf("nil game record")
	}

	g, err := registry.Import(rec.GameType, []byte(rec.SaveState))
	if err != nil {
		return nil, err
	}
	return g.SetTitle(rec.Name), nil
}

// CreateRecord persists the initial state for a newly created game.
func CreateRecord(
	s *store.Store,
	g game.Gamer,
	name string,
	gameType string,
	modeTitle string,
) (*store.GameRecord, error) {
	initialState, err := g.GetSave()
	if err != nil {
		return nil, fmt.Errorf("failed to get initial save: %w", err)
	}

	rec := &store.GameRecord{
		Name:         name,
		GameID:       string(puzzle.CanonicalGameID(gameType)),
		GameType:     gameType,
		ModeID:       string(puzzle.CanonicalModeID(modeTitle)),
		Mode:         modeTitle,
		InitialState: string(initialState),
		SaveState:    string(initialState),
		Status:       store.StatusNew,
		RunKind:      store.RunKindForName(name),
		RunDate:      store.RunDateForName(name),
		SeedText:     store.SeedTextForName(name),
	}
	if year, week, index, ok := store.WeeklyIdentityForName(name); ok {
		rec.WeekYear = year
		rec.WeekNumber = week
		rec.WeekIndex = index
	}
	if err := s.CreateGame(rec); err != nil {
		return nil, fmt.Errorf("failed to create game record: %w", err)
	}
	return rec, nil
}

// ResumeAbandonedDeterministicRecord marks an abandoned deterministic game
// as in-progress before resuming it.
func ResumeAbandonedDeterministicRecord(s *store.Store, rec *store.GameRecord) error {
	if rec == nil || rec.Status != store.StatusAbandoned {
		return nil
	}
	if err := s.UpdateStatus(rec.ID, store.StatusInProgress); err != nil {
		return fmt.Errorf("failed to mark deterministic game in progress: %w", err)
	}
	return nil
}

// GenerateUniqueName generates a unique adjective-noun name, retrying if needed.
func GenerateUniqueName(s *store.Store) string {
	for range 100 {
		name := namegen.Generate()
		exists, err := s.NameExists(name)
		if err != nil || !exists {
			return name
		}
	}
	for i := 1; ; i++ {
		name := namegen.Generate() + "-" + strconv.Itoa(i)
		exists, err := s.NameExists(name)
		if err != nil || !exists {
			return name
		}
	}
}

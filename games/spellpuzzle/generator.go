package spellpuzzle

import (
	"errors"
	"math"
	"math/rand/v2"
	"sort"
)

var errNoPuzzleFound = errors.New("spell puzzle: failed to generate puzzle")

type GeneratedPuzzle struct {
	Bank        string
	Placements  []WordPlacement
	BonusWords  []string
	AllowedWord []string
}

type crosswordCell struct {
	Letter     rune
	Horizontal bool
	Vertical   bool
}

type crosswordState struct {
	Cells      map[Position]crosswordCell
	Placements []WordPlacement
}

func GeneratePuzzle(mode SpellPuzzleMode) (GeneratedPuzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GeneratePuzzleSeeded(mode, rng)
}

func GeneratePuzzleSeeded(mode SpellPuzzleMode, rng *rand.Rand) (GeneratedPuzzle, error) {
	seeds := append([]string(nil), seedWordsByLength[mode.BankSize]...)
	if len(seeds) == 0 {
		return GeneratedPuzzle{}, errNoPuzzleFound
	}

	shuffleStrings(seeds, rng)
	maxAttempts := min(len(seeds), 800)
	for _, bank := range seeds[:maxAttempts] {
		candidates := spellableWords(bank, 3)
		if len(candidates) < mode.BoardWordCount+mode.MinBonusWords {
			continue
		}

		placements, ok := buildCrossword(bank, candidates, mode.BoardWordCount, rng)
		if !ok {
			continue
		}

		used := make(map[string]struct{}, len(placements))
		for _, placement := range placements {
			used[placement.Text] = struct{}{}
		}

		bonusWords := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			if _, ok := used[candidate]; ok {
				continue
			}
			bonusWords = append(bonusWords, candidate)
		}
		if len(bonusWords) < mode.MinBonusWords {
			continue
		}

		sort.Strings(bonusWords)
		sort.Strings(candidates)
		scrambledBank := scrambleBank(bank, placements, rng)
		return GeneratedPuzzle{
			Bank:        scrambledBank,
			Placements:  normalizePlacements(placements),
			BonusWords:  bonusWords,
			AllowedWord: candidates,
		}, nil
	}

	return GeneratedPuzzle{}, errNoPuzzleFound
}

func buildCrossword(bank string, candidates []string, want int, rng *rand.Rand) ([]WordPlacement, bool) {
	ordered := rankBoardCandidates(candidates)
	if len(ordered) < want {
		return nil, false
	}

	firstWordCount := min(len(ordered), max(4, want+2))
	firstWords := append([]string(nil), ordered[:firstWordCount]...)
	shuffleStrings(firstWords, rng)

	for _, firstWord := range firstWords {
		state := crosswordState{
			Cells: map[Position]crosswordCell{},
		}
		firstPlacement := WordPlacement{
			Text:        firstWord,
			Start:       Position{X: 0, Y: 0},
			Orientation: Horizontal,
		}
		state = applyPlacement(state, firstPlacement)
		used := map[string]struct{}{firstWord: {}}
		if solved, ok := placeRemainingWords(state, ordered, used, want, rng); ok {
			return solved.Placements, true
		}
	}

	_ = bank
	return nil, false
}

func placeRemainingWords(state crosswordState, candidates []string, used map[string]struct{}, want int, rng *rand.Rand) (crosswordState, bool) {
	if len(state.Placements) == want {
		return state, true
	}

	ranked := rankWordsForState(state, candidates, used)
	if len(ranked) == 0 {
		return crosswordState{}, false
	}

	need := want - len(state.Placements)
	if len(ranked) < need {
		return crosswordState{}, false
	}

	for _, word := range ranked {
		placements := validPlacements(state, word)
		if len(placements) == 0 {
			continue
		}
		sortPlacements(placements, state)
		limit := min(len(placements), 8)
		for _, placement := range placements[:limit] {
			nextState := applyPlacement(state, placement)
			nextUsed := cloneWordSet(used)
			nextUsed[word] = struct{}{}
			if solved, ok := placeRemainingWords(nextState, candidates, nextUsed, want, rng); ok {
				return solved, true
			}
		}
	}

	return crosswordState{}, false
}

func rankBoardCandidates(candidates []string) []string {
	ordered := append([]string(nil), candidates...)
	overlap := make(map[string]int, len(ordered))
	for _, word := range ordered {
		score := 0
		seen := map[rune]struct{}{}
		for _, letter := range word {
			if _, ok := seen[letter]; ok {
				continue
			}
			seen[letter] = struct{}{}
			for _, other := range ordered {
				if word == other {
					continue
				}
				if stringsContainsRune(other, letter) {
					score++
				}
			}
		}
		overlap[word] = score
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if len(ordered[i]) != len(ordered[j]) {
			return len(ordered[i]) > len(ordered[j])
		}
		if overlap[ordered[i]] != overlap[ordered[j]] {
			return overlap[ordered[i]] > overlap[ordered[j]]
		}
		return ordered[i] < ordered[j]
	})
	return ordered
}

func rankWordsForState(state crosswordState, candidates []string, used map[string]struct{}) []string {
	type scoredWord struct {
		Word  string
		Score int
	}

	scored := make([]scoredWord, 0, len(candidates))
	cellPositions := sortedCellPositions(state.Cells)
	for _, word := range candidates {
		if _, ok := used[word]; ok {
			continue
		}
		score := len(word) * 10
		for _, pos := range cellPositions {
			cell := state.Cells[pos]
			if stringsContainsRune(word, cell.Letter) {
				score += 3
			}
		}
		scored = append(scored, scoredWord{Word: word, Score: score})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		return scored[i].Word < scored[j].Word
	})

	result := make([]string, 0, len(scored))
	for _, item := range scored {
		result = append(result, item.Word)
	}
	return result
}

func validPlacements(state crosswordState, word string) []WordPlacement {
	if len(state.Placements) == 0 {
		return []WordPlacement{{
			Text:        word,
			Start:       Position{X: 0, Y: 0},
			Orientation: Horizontal,
		}}
	}

	result := make([]WordPlacement, 0, 16)
	seen := make(map[WordPlacement]struct{})
	for _, pos := range sortedCellPositions(state.Cells) {
		cell := state.Cells[pos]
		for idx, letter := range word {
			if letter != cell.Letter {
				continue
			}

			if cell.Horizontal && !cell.Vertical {
				placement := WordPlacement{
					Text:        word,
					Start:       Position{X: pos.X, Y: pos.Y - idx},
					Orientation: Vertical,
				}
				if _, ok := seen[placement]; !ok && placementAllowed(state, placement) {
					seen[placement] = struct{}{}
					result = append(result, placement)
				}
			}
			if cell.Vertical && !cell.Horizontal {
				placement := WordPlacement{
					Text:        word,
					Start:       Position{X: pos.X - idx, Y: pos.Y},
					Orientation: Horizontal,
				}
				if _, ok := seen[placement]; !ok && placementAllowed(state, placement) {
					seen[placement] = struct{}{}
					result = append(result, placement)
				}
			}
		}
	}
	return result
}

func placementAllowed(state crosswordState, placement WordPlacement) bool {
	before := placement.Start
	after := placement.Start
	if placement.Orientation == Horizontal {
		before.X--
		after.X += len(placement.Text)
	} else {
		before.Y--
		after.Y += len(placement.Text)
	}
	if _, ok := state.Cells[before]; ok {
		return false
	}
	if _, ok := state.Cells[after]; ok {
		return false
	}

	intersections := 0
	for idx, letter := range placement.Text {
		pos := placement.Start
		if placement.Orientation == Horizontal {
			pos.X += idx
		} else {
			pos.Y += idx
		}

		existing, occupied := state.Cells[pos]
		if occupied {
			if existing.Letter != letter {
				return false
			}
			if placement.Orientation == Horizontal && existing.Horizontal {
				return false
			}
			if placement.Orientation == Vertical && existing.Vertical {
				return false
			}
			intersections++
			continue
		}

		if placement.Orientation == Horizontal {
			if hasCell(state.Cells, Position{X: pos.X, Y: pos.Y - 1}) || hasCell(state.Cells, Position{X: pos.X, Y: pos.Y + 1}) {
				return false
			}
			continue
		}
		if hasCell(state.Cells, Position{X: pos.X - 1, Y: pos.Y}) || hasCell(state.Cells, Position{X: pos.X + 1, Y: pos.Y}) {
			return false
		}
	}

	return intersections > 0
}

func applyPlacement(state crosswordState, placement WordPlacement) crosswordState {
	next := crosswordState{
		Cells:      cloneCells(state.Cells),
		Placements: append(append([]WordPlacement(nil), state.Placements...), placement),
	}
	if len(state.Placements) == 0 {
		next.Placements = []WordPlacement{placement}
	}

	for idx, letter := range placement.Text {
		pos := placement.Start
		if placement.Orientation == Horizontal {
			pos.X += idx
		} else {
			pos.Y += idx
		}
		cell := next.Cells[pos]
		cell.Letter = letter
		if placement.Orientation == Horizontal {
			cell.Horizontal = true
		} else {
			cell.Vertical = true
		}
		next.Cells[pos] = cell
	}
	return next
}

func sortPlacements(placements []WordPlacement, state crosswordState) {
	sort.SliceStable(placements, func(i, j int) bool {
		iScore := placementScore(placements[i], state)
		jScore := placementScore(placements[j], state)
		if iScore != jScore {
			return iScore > jScore
		}
		if placements[i].Start.Y != placements[j].Start.Y {
			return placements[i].Start.Y < placements[j].Start.Y
		}
		if placements[i].Start.X != placements[j].Start.X {
			return placements[i].Start.X < placements[j].Start.X
		}
		if placements[i].Orientation != placements[j].Orientation {
			return placements[i].Orientation < placements[j].Orientation
		}
		return placements[i].Text < placements[j].Text
	})
}

func placementScore(placement WordPlacement, state crosswordState) int {
	score := 0
	minX, minY, maxX, maxY := boundsForState(state)
	for idx := range len(placement.Text) {
		pos := placement.Start
		if placement.Orientation == Horizontal {
			pos.X += idx
		} else {
			pos.Y += idx
		}
		if existing, ok := state.Cells[pos]; ok {
			if existing.Letter != 0 {
				score += 10
			}
		}
	}

	nextMinX := min(minX, placement.Start.X)
	nextMinY := min(minY, placement.Start.Y)
	end := placement.Start
	if placement.Orientation == Horizontal {
		end.X += len(placement.Text) - 1
	} else {
		end.Y += len(placement.Text) - 1
	}
	nextMaxX := max(maxX, end.X)
	nextMaxY := max(maxY, end.Y)
	area := (nextMaxX - nextMinX + 1) * (nextMaxY - nextMinY + 1)
	score -= area
	score -= int(math.Abs(float64(placement.Start.X)) + math.Abs(float64(placement.Start.Y)))
	return score
}

func boundsForState(state crosswordState) (minX, minY, maxX, maxY int) {
	first := true
	for pos := range state.Cells {
		if first {
			minX, maxX = pos.X, pos.X
			minY, maxY = pos.Y, pos.Y
			first = false
			continue
		}
		minX = min(minX, pos.X)
		minY = min(minY, pos.Y)
		maxX = max(maxX, pos.X)
		maxY = max(maxY, pos.Y)
	}
	return minX, minY, maxX, maxY
}

func cloneCells(cells map[Position]crosswordCell) map[Position]crosswordCell {
	cloned := make(map[Position]crosswordCell, len(cells))
	for pos, cell := range cells {
		cloned[pos] = cell
	}
	return cloned
}

func cloneWordSet(words map[string]struct{}) map[string]struct{} {
	cloned := make(map[string]struct{}, len(words))
	for word := range words {
		cloned[word] = struct{}{}
	}
	return cloned
}

func shuffleStrings(items []string, rng *rand.Rand) {
	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
}

func scrambleBank(bank string, placements []WordPlacement, rng *rand.Rand) string {
	runes := []rune(bank)
	if len(runes) < 2 {
		return bank
	}

	if hasSinglePermutation(runes) {
		return bank
	}

	original := string(runes)
	candidate := append([]rune(nil), runes...)
	for range 16 {
		rng.Shuffle(len(candidate), func(i, j int) {
			candidate[i], candidate[j] = candidate[j], candidate[i]
		})
		scrambled := string(candidate)
		if scrambled == original {
			continue
		}
		if matchesPlacedWord(scrambled, placements) {
			continue
		}
		return scrambled
	}

	for i := 1; i < len(runes); i++ {
		rotated := string(append(append([]rune(nil), runes[i:]...), runes[:i]...))
		if rotated == original {
			continue
		}
		if matchesPlacedWord(rotated, placements) {
			continue
		}
		return rotated
	}

	return bank
}

func hasSinglePermutation(runes []rune) bool {
	first := runes[0]
	for _, current := range runes[1:] {
		if current != first {
			return false
		}
	}
	return true
}

func matchesPlacedWord(candidate string, placements []WordPlacement) bool {
	reversed := reverseWord(candidate)
	for _, placement := range placements {
		if placement.Text == candidate || placement.Text == reversed {
			return true
		}
	}
	return false
}

func reverseWord(word string) string {
	runes := []rune(word)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func stringsContainsRune(text string, target rune) bool {
	for _, letter := range text {
		if letter == target {
			return true
		}
	}
	return false
}

func hasCell(cells map[Position]crosswordCell, pos Position) bool {
	_, ok := cells[pos]
	return ok
}

func sortedCellPositions(cells map[Position]crosswordCell) []Position {
	positions := make([]Position, 0, len(cells))
	for pos := range cells {
		positions = append(positions, pos)
	}
	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].Y != positions[j].Y {
			return positions[i].Y < positions[j].Y
		}
		return positions[i].X < positions[j].X
	})
	return positions
}

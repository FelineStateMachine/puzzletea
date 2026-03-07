package takuzuplus

import (
	"cmp"
	"image"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/FelineStateMachine/puzzletea/takuzu"
)

type cellPos struct{ x, y int }

type relationCandidate struct {
	horizontal bool
	x, y       int
}

type relationChoice struct {
	candidate     relationCandidate
	remove        cellPos
	additive      bool
	endpointClass int
	symbol        rune
	score         int
}

func generateComplete(size int) grid {
	return grid(takuzu.GenerateCompleteGrid(size))
}

func generateCompleteSeeded(size int, rng *rand.Rand) grid {
	return grid(takuzu.GenerateCompleteGridSeeded(size, rng))
}

func generatePuzzle(complete grid, size int, prefilled float64, profile relationProfile) (grid, [][]bool, relations) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return generatePuzzleSeeded(complete, size, prefilled, profile, rng)
}

func generatePuzzleSeeded(complete grid, size int, prefilled float64, profile relationProfile, rng *rand.Rand) (grid, [][]bool, relations) {
	maxAttempts := 10
	switch {
	case size >= 14:
		maxAttempts = 3
	case size >= 12:
		maxAttempts = 6
	}

	bestPenalty := math.MaxInt
	var bestPuzzle grid
	var bestProvided [][]bool
	var bestRelations relations

	for attempt := 0; attempt < maxAttempts; attempt++ {
		desiredRelations := profile.targetCount(rng)
		puzzle, provided, rels := generatePuzzleAttempt(complete, size, prefilled, profile, desiredRelations, rng)
		penalty := relationPenalty(profile, size, rels, provided)
		if penalty < bestPenalty {
			bestPenalty = penalty
			bestPuzzle = puzzle.clone()
			bestProvided = cloneProvided(provided)
			bestRelations = rels.clone()
		}
		if penalty == 0 {
			return puzzle, provided, rels
		}
	}

	return bestPuzzle, bestProvided, bestRelations
}

func generatePuzzleAttempt(
	complete grid,
	size int,
	prefilled float64,
	profile relationProfile,
	desiredRelations int,
	rng *rand.Rand,
) (grid, [][]bool, relations) {
	puzzle := complete.clone()
	provided := make([][]bool, size)
	for y := range size {
		provided[y] = make([]bool, size)
		for x := range size {
			provided[y][x] = true
		}
	}
	rels := newRelations(size)

	targetFilled := int(prefilled * float64(size*size))
	filled := size * size

	for filled > targetFilled {
		progress := false
		cells := providedCells(provided)
		rng.Shuffle(len(cells), func(i, j int) {
			cells[i], cells[j] = cells[j], cells[i]
		})

		for _, cell := range cells {
			if !provided[cell.y][cell.x] || filled <= targetFilled {
				continue
			}

			remainingRemovals := filled - targetFilled
			relationShortage := desiredRelations - countRelations(rels)
			preferRelation := countRelations(rels) < profile.MaxRelations &&
				shouldPreferRelationRemoval(relationShortage, remainingRemovals, rng)

			if preferRelation {
				if tryRelationRemoval(puzzle, complete, provided, rels, size, profile, cell, rng) {
					filled--
					progress = true
					continue
				}
			}

			if tryDirectRemoval(puzzle, provided, rels, size, cell) {
				filled--
				progress = true
				continue
			}

			if !preferRelation && countRelations(rels) < profile.MaxRelations &&
				tryRelationRemoval(puzzle, complete, provided, rels, size, profile, cell, rng) {
				filled--
				progress = true
			}
		}

		if !progress {
			break
		}
	}

	for countRelations(rels) < desiredRelations {
		if !tryPostPassRelationRemoval(puzzle, complete, provided, rels, size, profile, rng) {
			break
		}
		filled--
	}

	for countRelations(rels) < profile.MinRelations || additiveRelationsNeeded(profile, size, rels, provided) {
		if countRelations(rels) >= profile.MaxRelations {
			break
		}
		if !tryAdditiveRelation(complete, provided, rels, size, profile, rng) {
			break
		}
	}

	return puzzle, provided, rels
}

func shouldPreferRelationRemoval(relationShortage, remainingRemovals int, rng *rand.Rand) bool {
	if relationShortage <= 0 || remainingRemovals <= 0 {
		return false
	}
	if remainingRemovals <= relationShortage {
		return true
	}
	return rng.IntN(remainingRemovals) < relationShortage
}

func tryDirectRemoval(puzzle grid, provided [][]bool, rels relations, size int, cell cellPos) bool {
	saved := puzzle[cell.y][cell.x]
	puzzle[cell.y][cell.x] = emptyCell
	if countSolutions(puzzle, size, 2, rels) == 1 {
		provided[cell.y][cell.x] = false
		return true
	}
	puzzle[cell.y][cell.x] = saved
	return false
}

func tryRelationRemoval(
	puzzle, complete grid,
	provided [][]bool,
	rels relations,
	size int,
	profile relationProfile,
	remove cellPos,
	rng *rand.Rand,
) bool {
	choices := scoredRelationChoices(complete, provided, rels, size, profile, remove, false)
	if len(choices) == 0 {
		return false
	}
	shuffleTopTies(choices, rng)

	for _, choice := range choices {
		applyRelationCandidate(rels, choice.candidate, complete)
		saved := puzzle[remove.y][remove.x]
		puzzle[remove.y][remove.x] = emptyCell
		if countSolutions(puzzle, size, 2, rels) == 1 {
			provided[remove.y][remove.x] = false
			return true
		}
		puzzle[remove.y][remove.x] = saved
		clearRelationCandidate(rels, choice.candidate)
	}

	return false
}

func tryPostPassRelationRemoval(
	puzzle, complete grid,
	provided [][]bool,
	rels relations,
	size int,
	profile relationProfile,
	rng *rand.Rand,
) bool {
	cells := providedCells(provided)
	rng.Shuffle(len(cells), func(i, j int) {
		cells[i], cells[j] = cells[j], cells[i]
	})

	for _, cell := range cells {
		if tryRelationRemoval(puzzle, complete, provided, rels, size, profile, cell, rng) {
			return true
		}
	}
	return false
}

func tryAdditiveRelation(
	complete grid,
	provided [][]bool,
	rels relations,
	size int,
	profile relationProfile,
	rng *rand.Rand,
) bool {
	choices := additiveRelationChoices(complete, provided, rels, size, profile)
	if len(choices) == 0 {
		return false
	}
	shuffleTopTies(choices, rng)
	applyRelationCandidate(rels, choices[0].candidate, complete)
	return true
}

func scoredRelationChoices(
	complete grid,
	provided [][]bool,
	rels relations,
	size int,
	profile relationProfile,
	remove cellPos,
	additive bool,
) []relationChoice {
	candidates := adjacentRelationCandidates(rels, remove.x, remove.y, size)
	if len(candidates) == 0 {
		return nil
	}

	metrics := analyzeRelations(rels, provided, size)
	choices := make([]relationChoice, 0, len(candidates))
	for _, candidate := range candidates {
		ax, ay, bx, by := relationEndpoints(candidate)
		otherProvided := false
		if ax == remove.x && ay == remove.y {
			otherProvided = provided[by][bx]
		} else {
			otherProvided = provided[ay][ax]
		}

		endpoint := endpointZeroProvided
		if otherProvided {
			endpoint = endpointOneProvided
		}

		choice := relationChoice{
			candidate:     candidate,
			remove:        remove,
			additive:      additive,
			endpointClass: endpoint,
			symbol:        relationSymbol(complete, candidate),
		}
		choice.score = relationChoiceScore(choice, profile, metrics, size)
		choices = append(choices, choice)
	}

	slices.SortFunc(choices, func(a, b relationChoice) int {
		return cmp.Compare(b.score, a.score)
	})
	return choices
}

func additiveRelationChoices(
	complete grid,
	provided [][]bool,
	rels relations,
	size int,
	profile relationProfile,
) []relationChoice {
	metrics := analyzeRelations(rels, provided, size)
	choices := make([]relationChoice, 0, 2*size*(size-1))
	for _, candidate := range allUnusedRelationCandidates(rels, size) {
		ax, ay, bx, by := relationEndpoints(candidate)
		endpoint := endpointClass(boolInt(provided[ay][ax]) + boolInt(provided[by][bx]))
		choice := relationChoice{
			candidate:     candidate,
			additive:      true,
			endpointClass: endpoint,
			symbol:        relationSymbol(complete, candidate),
		}
		choice.score = relationChoiceScore(choice, profile, metrics, size)
		if endpoint == endpointTwoProvided {
			choice.score += 5
		}
		choices = append(choices, choice)
	}

	slices.SortFunc(choices, func(a, b relationChoice) int {
		return cmp.Compare(b.score, a.score)
	})
	return choices
}

func relationChoiceScore(choice relationChoice, profile relationProfile, metrics relationMetrics, size int) int {
	score := 0
	score += profile.EndpointWeights[choice.endpointClass] * 10

	band := profile.EndpointBands[choice.endpointClass]
	currentRatio := ratioForCount(metrics.EndpointCounts[choice.endpointClass], metrics.Total)
	targetMid := (band.Min + band.Max) / 2
	if currentRatio < band.Min {
		score += 18
	} else {
		score -= int(math.Round(math.Abs(currentRatio-targetMid) * 20))
	}

	if metrics.Total > 0 {
		if choice.symbol == relationSame && metrics.SameCount < metrics.DifferentCount {
			score += 6
		}
		if choice.symbol == relationDiff && metrics.DifferentCount < metrics.SameCount {
			score += 6
		}
		if choice.candidate.horizontal && metrics.HorizontalCount < metrics.VerticalCount {
			score += 4
		}
		if !choice.candidate.horizontal && metrics.VerticalCount < metrics.HorizontalCount {
			score += 4
		}
	}

	pos := relationPosition(choice.candidate)
	if profile.MinSpacing > 0 {
		minGap := minDistance(metrics.Positions, pos)
		if minGap >= profile.MinSpacing {
			score += minGap
		} else {
			score -= 12 * (profile.MinSpacing - minGap + 1)
		}
	}

	if profile.MinRegions > 0 {
		bit := relationRegionBit(size, pos)
		if metrics.RegionMask&bit == 0 {
			score += 8
		}
	}

	rowSpreadBonus := 0
	if choice.candidate.horizontal {
		rowSpreadBonus = 2
	}
	score += rowSpreadBonus

	return score
}

func additiveRelationsNeeded(profile relationProfile, size int, rels relations, provided [][]bool) bool {
	metrics := analyzeRelations(rels, provided, size)
	if metrics.Total < profile.MinRelations {
		return true
	}
	if metrics.Total == 0 {
		return false
	}
	twoBand := profile.EndpointBands[endpointTwoProvided]
	if !ratioWithinBand(metrics.EndpointCounts[endpointTwoProvided], metrics.Total, twoBand) &&
		ratioForCount(metrics.EndpointCounts[endpointTwoProvided], metrics.Total) < twoBand.Min {
		return true
	}
	return false
}

func adjacentRelationCandidates(rels relations, x, y, size int) []relationCandidate {
	candidates := make([]relationCandidate, 0, 4)
	if x > 0 && rels.horizontal[y][x-1] == relationNone {
		candidates = append(candidates, relationCandidate{horizontal: true, x: x - 1, y: y})
	}
	if x < size-1 && rels.horizontal[y][x] == relationNone {
		candidates = append(candidates, relationCandidate{horizontal: true, x: x, y: y})
	}
	if y > 0 && rels.vertical[y-1][x] == relationNone {
		candidates = append(candidates, relationCandidate{horizontal: false, x: x, y: y - 1})
	}
	if y < size-1 && rels.vertical[y][x] == relationNone {
		candidates = append(candidates, relationCandidate{horizontal: false, x: x, y: y})
	}
	return candidates
}

func allUnusedRelationCandidates(rels relations, size int) []relationCandidate {
	candidates := make([]relationCandidate, 0, 2*size*(size-1))
	for y := 0; y < size; y++ {
		for x := 0; x < size-1; x++ {
			if rels.horizontal[y][x] == relationNone {
				candidates = append(candidates, relationCandidate{horizontal: true, x: x, y: y})
			}
		}
	}
	for y := 0; y < size-1; y++ {
		for x := 0; x < size; x++ {
			if rels.vertical[y][x] == relationNone {
				candidates = append(candidates, relationCandidate{horizontal: false, x: x, y: y})
			}
		}
	}
	return candidates
}

func applyRelationCandidate(rels relations, candidate relationCandidate, complete grid) {
	if candidate.horizontal {
		rels.horizontal[candidate.y][candidate.x] = relationForValues(complete[candidate.y][candidate.x], complete[candidate.y][candidate.x+1])
		return
	}
	rels.vertical[candidate.y][candidate.x] = relationForValues(complete[candidate.y][candidate.x], complete[candidate.y+1][candidate.x])
}

func clearRelationCandidate(rels relations, candidate relationCandidate) {
	if candidate.horizontal {
		rels.horizontal[candidate.y][candidate.x] = relationNone
		return
	}
	rels.vertical[candidate.y][candidate.x] = relationNone
}

func relationEndpoints(candidate relationCandidate) (ax, ay, bx, by int) {
	if candidate.horizontal {
		return candidate.x, candidate.y, candidate.x + 1, candidate.y
	}
	return candidate.x, candidate.y, candidate.x, candidate.y + 1
}

func relationPosition(candidate relationCandidate) image.Point {
	if candidate.horizontal {
		return horizontalRelationPosition(candidate.x, candidate.y)
	}
	return verticalRelationPosition(candidate.x, candidate.y)
}

func relationSymbol(complete grid, candidate relationCandidate) rune {
	ax, ay, bx, by := relationEndpoints(candidate)
	return relationForValues(complete[ay][ax], complete[by][bx])
}

func providedCells(provided [][]bool) []cellPos {
	size := len(provided)
	cells := make([]cellPos, 0, size*size)
	for y := range size {
		for x := range size {
			if provided[y][x] {
				cells = append(cells, cellPos{x: x, y: y})
			}
		}
	}
	return cells
}

func cloneProvided(provided [][]bool) [][]bool {
	clone := make([][]bool, len(provided))
	for y := range provided {
		clone[y] = append([]bool(nil), provided[y]...)
	}
	return clone
}

func minDistance(existing []image.Point, pos image.Point) int {
	if len(existing) == 0 {
		return profileDistanceUnset()
	}
	best := math.MaxInt
	for _, prior := range existing {
		dist := abs(pos.X-prior.X) + abs(pos.Y-prior.Y)
		if dist < best {
			best = dist
		}
	}
	return best
}

func ratioForCount(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total)
}

func shuffleTopTies(choices []relationChoice, rng *rand.Rand) {
	if len(choices) < 2 {
		return
	}
	topScore := choices[0].score
	limit := 1
	for limit < len(choices) && choices[limit].score == topScore {
		limit++
	}
	rng.Shuffle(limit, func(i, j int) {
		choices[i], choices[j] = choices[j], choices[i]
	})
}

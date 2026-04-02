package netwalk

import (
	"errors"
	"math"
	"math/rand/v2"
)

const maxGenerateAttempts = 64

type generateProfile struct {
	ParentDegreeWeights    [5]int
	OrthogonalPackedWeight int
	DiagonalPackedWeight   int
	SpanGrowthWeight       int
	MinSpanRatio           float64
}

var legacyGenerateProfile = generateProfile{
	ParentDegreeWeights: [5]int{16, 16, 8, -2, -4},
}

func Generate(size, targetActive int) (Puzzle, error) {
	return GenerateSeeded(size, targetActive, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))
}

func GenerateSeeded(size, targetActive int, rng *rand.Rand) (Puzzle, error) {
	return generateSeededWithProfile(size, targetActive, legacyGenerateProfile, rng)
}

func GenerateSeededWithDensity(size int, fillRatio float64, profile generateProfile, rng *rand.Rand) (Puzzle, error) {
	return generateSeededWithProfile(size, targetActiveFromFillRatio(size, fillRatio), profile, rng)
}

func generateSeededWithProfile(size, targetActive int, profile generateProfile, rng *rand.Rand) (Puzzle, error) {
	if size <= 1 {
		return Puzzle{}, errors.New("netwalk size must be at least 2")
	}
	if targetActive < 2 {
		targetActive = 2
	}
	if targetActive > size*size {
		targetActive = size * size
	}

	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		puzzle := buildTreePuzzle(size, targetActive, profile, rng)
		scramblePuzzle(&puzzle, rng)
		if !analyzePuzzle(puzzle).solved {
			return puzzle, nil
		}
	}

	return Puzzle{}, errors.New("failed to generate netwalk puzzle")
}

type frontierEdge struct {
	from point
	to   point
}

type activeBounds struct {
	minX int
	maxX int
	minY int
	maxY int
}

func buildTreePuzzle(size, targetActive int, profile generateProfile, rng *rand.Rand) Puzzle {
	puzzle := newPuzzle(size)
	root := point{X: size / 2, Y: size / 2}
	puzzle.Root = root

	active := map[point]struct{}{root: {}}
	adjacency := map[point]directionMask{root: 0}

	for len(active) < targetActive {
		bounds := measureActiveBounds(active)
		frontier := collectFrontier(size, active)
		if len(frontier) == 0 {
			break
		}

		edge := frontier[weightedFrontierIndex(size, frontier, active, adjacency, bounds, profile, rng)]
		active[edge.to] = struct{}{}
		adjacency[edge.to] = 0

		dir := directionBetween(edge.from, edge.to)
		adjacency[edge.from] |= dir
		adjacency[edge.to] |= opposite(dir)
	}

	for p := range active {
		kind := nodeCell
		if p == root {
			kind = serverCell
		}
		puzzle.Tiles[p.Y][p.X] = tile{
			BaseMask: adjacency[p],
			Kind:     kind,
		}
	}

	return puzzle
}

func collectFrontier(size int, active map[point]struct{}) []frontierEdge {
	frontier := make([]frontierEdge, 0, len(active)*2)
	for y := range size {
		for x := range size {
			cell := point{X: x, Y: y}
			if _, ok := active[cell]; !ok {
				continue
			}
			for _, dir := range directions {
				next := point{X: cell.X + dir.dx, Y: cell.Y + dir.dy}
				if next.X < 0 || next.X >= size || next.Y < 0 || next.Y >= size {
					continue
				}
				if _, ok := active[next]; ok {
					continue
				}
				frontier = append(frontier, frontierEdge{from: cell, to: next})
			}
		}
	}
	return frontier
}

func targetActiveFromFillRatio(size int, fillRatio float64) int {
	if size <= 1 {
		return 2
	}
	target := int(math.Round(float64(size*size) * fillRatio))
	if target < 2 {
		return 2
	}
	if target > size*size {
		return size * size
	}
	return target
}

func measureActiveBounds(active map[point]struct{}) activeBounds {
	first := true
	bounds := activeBounds{}
	for p := range active {
		if first {
			bounds = activeBounds{minX: p.X, maxX: p.X, minY: p.Y, maxY: p.Y}
			first = false
			continue
		}
		if p.X < bounds.minX {
			bounds.minX = p.X
		}
		if p.X > bounds.maxX {
			bounds.maxX = p.X
		}
		if p.Y < bounds.minY {
			bounds.minY = p.Y
		}
		if p.Y > bounds.maxY {
			bounds.maxY = p.Y
		}
	}
	return bounds
}

func (b activeBounds) spanX() int {
	return b.maxX - b.minX + 1
}

func (b activeBounds) spanY() int {
	return b.maxY - b.minY + 1
}

func weightedFrontierIndex(
	size int,
	frontier []frontierEdge,
	active map[point]struct{},
	adjacency map[point]directionMask,
	bounds activeBounds,
	profile generateProfile,
	rng *rand.Rand,
) int {
	if len(frontier) == 1 {
		return 0
	}

	weights := make([]int, len(frontier))
	total := 0
	for i, edge := range frontier {
		weight := frontierWeight(size, edge, active, adjacency, bounds, profile)
		weights[i] = weight
		total += weight
	}

	pick := rng.IntN(total)
	running := 0
	for i, weight := range weights {
		running += weight
		if pick < running {
			return i
		}
	}
	return len(frontier) - 1
}

func frontierWeight(
	size int,
	edge frontierEdge,
	active map[point]struct{},
	adjacency map[point]directionMask,
	bounds activeBounds,
	profile generateProfile,
) int {
	deg := degree(adjacency[edge.from])
	if deg >= len(profile.ParentDegreeWeights) {
		deg = len(profile.ParentDegreeWeights) - 1
	}

	weight := 64 + profile.ParentDegreeWeights[deg]
	orthogonal, diagonal := packedNeighborCounts(size, edge, active)
	weight += orthogonal * profile.OrthogonalPackedWeight
	weight += diagonal * profile.DiagonalPackedWeight
	weight += spanGrowthScore(size, edge.to, bounds, profile)
	if weight < 1 {
		return 1
	}
	return weight
}

func packedNeighborCounts(
	size int,
	edge frontierEdge,
	active map[point]struct{},
) (int, int) {
	orthogonal := 0
	diagonal := 0
	for _, dir := range directions {
		next := point{X: edge.to.X + dir.dx, Y: edge.to.Y + dir.dy}
		if next == edge.from || !inBounds(size, next) {
			continue
		}
		if _, ok := active[next]; ok {
			orthogonal++
		}
	}

	for _, delta := range [][2]int{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}} {
		next := point{X: edge.to.X + delta[0], Y: edge.to.Y + delta[1]}
		if !inBounds(size, next) {
			continue
		}
		if _, ok := active[next]; ok {
			diagonal++
		}
	}
	return orthogonal, diagonal
}

func spanGrowthScore(size int, candidate point, bounds activeBounds, profile generateProfile) int {
	target := minSpanTarget(size, profile.MinSpanRatio)
	if target == 0 || profile.SpanGrowthWeight == 0 {
		return 0
	}

	score := 0
	if bounds.spanX() < target && (candidate.X < bounds.minX || candidate.X > bounds.maxX) {
		score += profile.SpanGrowthWeight
	}
	if bounds.spanY() < target && (candidate.Y < bounds.minY || candidate.Y > bounds.maxY) {
		score += profile.SpanGrowthWeight
	}
	return score
}

func minSpanTarget(size int, ratio float64) int {
	if size <= 0 || ratio <= 0 {
		return 0
	}
	target := int(math.Ceil(float64(size) * ratio))
	if target < 1 {
		return 1
	}
	if target > size {
		return size
	}
	return target
}

func inBounds(size int, p point) bool {
	return p.X >= 0 && p.X < size && p.Y >= 0 && p.Y < size
}

func scramblePuzzle(puzzle *Puzzle, rng *rand.Rand) {
	if puzzle == nil {
		return
	}

	changed := 0
	nonEmpty := 0
	for y := range puzzle.Size {
		for x := range puzzle.Size {
			t := &puzzle.Tiles[y][x]
			if !isActive(*t) {
				continue
			}
			nonEmpty++
			options := uniqueRotations(t.BaseMask)
			rotation := options[rng.IntN(len(options))]
			t.Rotation = rotation
			t.InitialRotation = rotation
			if rotation != 0 {
				changed++
			}
		}
	}

	if changed > 0 || nonEmpty == 0 {
		return
	}

	active := make([]point, 0, nonEmpty)
	for y := range puzzle.Size {
		for x := range puzzle.Size {
			if isActive(puzzle.Tiles[y][x]) {
				active = append(active, point{X: x, Y: y})
			}
		}
	}
	for _, p := range active {
		t := &puzzle.Tiles[p.Y][p.X]
		options := uniqueRotations(t.BaseMask)
		if len(options) <= 1 {
			continue
		}
		t.Rotation = options[1%len(options)]
		t.InitialRotation = t.Rotation
		return
	}
}

func directionBetween(from, to point) directionMask {
	switch {
	case to.X == from.X && to.Y == from.Y-1:
		return north
	case to.X == from.X+1 && to.Y == from.Y:
		return east
	case to.X == from.X && to.Y == from.Y+1:
		return south
	case to.X == from.X-1 && to.Y == from.Y:
		return west
	default:
		return 0
	}
}

func opposite(mask directionMask) directionMask {
	switch mask {
	case north:
		return south
	case east:
		return west
	case south:
		return north
	case west:
		return east
	default:
		return 0
	}
}

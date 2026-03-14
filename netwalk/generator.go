package netwalk

import (
	"errors"
	"math/rand/v2"
)

const maxGenerateAttempts = 64

func Generate(size, targetActive int) (Puzzle, error) {
	return GenerateSeeded(size, targetActive, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))
}

func GenerateSeeded(size, targetActive int, rng *rand.Rand) (Puzzle, error) {
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
		puzzle := buildTreePuzzle(size, targetActive, rng)
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

func buildTreePuzzle(size, targetActive int, rng *rand.Rand) Puzzle {
	puzzle := newPuzzle(size)
	root := point{X: size / 2, Y: size / 2}
	puzzle.Root = root

	active := map[point]struct{}{root: {}}
	adjacency := map[point]directionMask{root: 0}

	for len(active) < targetActive {
		frontier := collectFrontier(size, active)
		if len(frontier) == 0 {
			break
		}

		edge := frontier[weightedFrontierIndex(frontier, adjacency, root, rng)]
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

func weightedFrontierIndex(
	frontier []frontierEdge,
	adjacency map[point]directionMask,
	root point,
	rng *rand.Rand,
) int {
	if len(frontier) == 1 {
		return 0
	}

	weights := make([]int, len(frontier))
	total := 0
	for i, edge := range frontier {
		weight := 10
		deg := degree(adjacency[edge.from])
		switch {
		case deg <= 1:
			weight += 6
		case deg == 2:
			weight += 3
		case deg >= 3:
			weight -= 2
		}

		dx := edge.to.X - root.X
		if dx < 0 {
			dx = -dx
		}
		dy := edge.to.Y - root.Y
		if dy < 0 {
			dy = -dy
		}
		weight += max(0, 6-(dx+dy))
		if weight < 1 {
			weight = 1
		}
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

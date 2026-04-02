package takuzuplus

import "github.com/FelineStateMachine/puzzletea/games/takuzu"

type mrvChoice struct {
	x, y  int
	vals  [2]rune
	count int
}

func relationForValues(a, b rune) rune {
	if a == b {
		return relationSame
	}
	return relationDiff
}

func relationSatisfied(kind, a, b rune) bool {
	switch kind {
	case relationSame:
		return a == b
	case relationDiff:
		return a != b
	default:
		return true
	}
}

func relationAllowed(g grid, rels relations, x, y int, val rune) bool {
	if x > 0 {
		if rel := rels.horizontal[y][x-1]; rel != relationNone && g[y][x-1] != emptyCell && !relationSatisfied(rel, g[y][x-1], val) {
			return false
		}
	}
	if x < len(g[y])-1 {
		if rel := rels.horizontal[y][x]; rel != relationNone && g[y][x+1] != emptyCell && !relationSatisfied(rel, val, g[y][x+1]) {
			return false
		}
	}
	if y > 0 {
		if rel := rels.vertical[y-1][x]; rel != relationNone && g[y-1][x] != emptyCell && !relationSatisfied(rel, g[y-1][x], val) {
			return false
		}
	}
	if y < len(g)-1 {
		if rel := rels.vertical[y][x]; rel != relationNone && g[y+1][x] != emptyCell && !relationSatisfied(rel, val, g[y+1][x]) {
			return false
		}
	}
	return true
}

func checkRelations(g grid, rels relations) bool {
	for y, row := range rels.horizontal {
		for x, rel := range row {
			if rel == relationNone || g[y][x] == emptyCell || g[y][x+1] == emptyCell {
				continue
			}
			if !relationSatisfied(rel, g[y][x], g[y][x+1]) {
				return false
			}
		}
	}
	for y, row := range rels.vertical {
		for x, rel := range row {
			if rel == relationNone || g[y][x] == emptyCell || g[y+1][x] == emptyCell {
				continue
			}
			if !relationSatisfied(rel, g[y][x], g[y+1][x]) {
				return false
			}
		}
	}
	return true
}

func canPlaceWithRelations(g grid, size, x, y int, val rune, rels relations) bool {
	return takuzu.CanPlaceInGrid(g, size, x, y, val) && relationAllowed(g, rels, x, y, val)
}

func selectMRVCell(g grid, size int, rels relations) mrvChoice {
	choice := mrvChoice{x: -1, y: -1, count: 3}
	for y := range size {
		for x := range size {
			if g[y][x] != emptyCell {
				continue
			}

			var vals [2]rune
			count := 0
			if canPlaceWithRelations(g, size, x, y, zeroCell, rels) {
				vals[count] = zeroCell
				count++
			}
			if canPlaceWithRelations(g, size, x, y, oneCell, rels) {
				vals[count] = oneCell
				count++
			}

			if count == 0 {
				return mrvChoice{x: x, y: y, count: 0}
			}
			if count < choice.count {
				choice = mrvChoice{x: x, y: y, vals: vals, count: count}
			}
		}
	}
	if choice.x < 0 {
		return mrvChoice{x: -1, y: -1, count: 0}
	}
	return choice
}

func countSolutions(g grid, size, limit int, rels relations) int {
	choice := selectMRVCell(g, size, rels)
	if choice.x < 0 {
		if takuzu.CheckConstraintsGrid(g, size) && takuzu.HasUniqueLinesGrid(g, size) && checkRelations(g, rels) {
			return 1
		}
		return 0
	}
	if choice.count == 0 {
		return 0
	}

	total := 0
	for i := range choice.count {
		val := choice.vals[i]
		g[choice.y][choice.x] = val
		total += countSolutions(g, size, limit-total, rels)
		g[choice.y][choice.x] = emptyCell
		if total >= limit {
			return total
		}
	}
	return total
}

package hitori

// hasNoDuplicatesInRows returns true if no row contains a duplicate number
// among the non-shaded cells.
func hasNoDuplicatesInRows(numbers grid, marks [][]cellMark, size int) bool {
	for y := range size {
		seen := map[rune]bool{}
		for x := range size {
			if marks[y][x] == shaded {
				continue
			}
			num := numbers[y][x]
			if seen[num] {
				return false
			}
			seen[num] = true
		}
	}
	return true
}

// hasNoDuplicatesInCols returns true if no column contains a duplicate number
// among the non-shaded cells.
func hasNoDuplicatesInCols(numbers grid, marks [][]cellMark, size int) bool {
	for x := range size {
		seen := map[rune]bool{}
		for y := range size {
			if marks[y][x] == shaded {
				continue
			}
			num := numbers[y][x]
			if seen[num] {
				return false
			}
			seen[num] = true
		}
	}
	return true
}

// hasNoAdjacentShaded returns true if no two shaded cells share an edge.
func hasNoAdjacentShaded(marks [][]cellMark, size int) bool {
	for y := range size {
		for x := range size {
			if marks[y][x] != shaded {
				continue
			}
			if x > 0 && marks[y][x-1] == shaded {
				return false
			}
			if x < size-1 && marks[y][x+1] == shaded {
				return false
			}
			if y > 0 && marks[y-1][x] == shaded {
				return false
			}
			if y < size-1 && marks[y+1][x] == shaded {
				return false
			}
		}
	}
	return true
}

// allWhiteConnected returns true if all non-shaded cells form a single
// orthogonally connected region. Uses BFS.
func allWhiteConnected(marks [][]cellMark, size int) bool {
	// Find first non-shaded cell.
	startX, startY := -1, -1
	whiteCount := 0
	for y := range size {
		for x := range size {
			if marks[y][x] != shaded {
				whiteCount++
				if startX == -1 {
					startX, startY = x, y
				}
			}
		}
	}

	if whiteCount <= 1 {
		return true
	}

	// BFS from the first white cell.
	type pos struct{ x, y int }
	visited := make([][]bool, size)
	for y := range size {
		visited[y] = make([]bool, size)
	}

	queue := []pos{{startX, startY}}
	visited[startY][startX] = true
	visitedCount := 1

	dirs := [4]pos{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			nx, ny := curr.x+d.x, curr.y+d.y
			if nx >= 0 && nx < size && ny >= 0 && ny < size &&
				!visited[ny][nx] && marks[ny][nx] != shaded {
				visited[ny][nx] = true
				visitedCount++
				queue = append(queue, pos{nx, ny})
			}
		}
	}

	return visitedCount == whiteCount
}

// isValidSolution checks all three Hitori constraints.
func isValidSolution(numbers grid, marks [][]cellMark, size int) bool {
	return hasNoDuplicatesInRows(numbers, marks, size) &&
		hasNoDuplicatesInCols(numbers, marks, size) &&
		hasNoAdjacentShaded(marks, size) &&
		allWhiteConnected(marks, size)
}

// computeConflicts returns a grid of booleans indicating which cells are
// involved in a rule violation. A cell is flagged if it participates in a
// row/column duplicate among unshaded cells, if it is a shaded cell adjacent
// to another shaded cell, or if it is an unshaded cell disconnected from the
// largest connected white region.
func computeConflicts(numbers grid, marks [][]cellMark, size int) [][]bool {
	conflicts := make([][]bool, size)
	for y := range size {
		conflicts[y] = make([]bool, size)
	}

	markDuplicateConflicts(numbers, marks, conflicts, size)
	markAdjacentShadedConflicts(marks, conflicts, size)
	markDisconnectedWhiteConflicts(marks, conflicts, size)

	return conflicts
}

// markDuplicateConflicts flags unshaded cells that share a number with
// another unshaded cell in the same row or column.
func markDuplicateConflicts(numbers grid, marks [][]cellMark, conflicts [][]bool, size int) {
	// Check rows.
	for y := range size {
		seen := map[rune][]int{}
		for x := range size {
			if marks[y][x] == shaded {
				continue
			}
			num := numbers[y][x]
			seen[num] = append(seen[num], x)
		}
		for _, positions := range seen {
			if len(positions) > 1 {
				for _, x := range positions {
					conflicts[y][x] = true
				}
			}
		}
	}

	// Check columns.
	for x := range size {
		seen := map[rune][]int{}
		for y := range size {
			if marks[y][x] == shaded {
				continue
			}
			num := numbers[y][x]
			seen[num] = append(seen[num], y)
		}
		for _, positions := range seen {
			if len(positions) > 1 {
				for _, y := range positions {
					conflicts[y][x] = true
				}
			}
		}
	}
}

// markAdjacentShadedConflicts flags shaded cells that share an edge with
// another shaded cell.
func markAdjacentShadedConflicts(marks [][]cellMark, conflicts [][]bool, size int) {
	for y := range size {
		for x := range size {
			if marks[y][x] != shaded {
				continue
			}
			if x > 0 && marks[y][x-1] == shaded {
				conflicts[y][x] = true
				conflicts[y][x-1] = true
			}
			if y > 0 && marks[y-1][x] == shaded {
				conflicts[y][x] = true
				conflicts[y-1][x] = true
			}
		}
	}
}

// markDisconnectedWhiteConflicts flags unshaded cells that are not part of the
// largest orthogonally connected white region.
func markDisconnectedWhiteConflicts(marks [][]cellMark, conflicts [][]bool, size int) {
	// Find all connected components of unshaded cells via BFS.
	type pos struct{ x, y int }
	visited := make([][]bool, size)
	for y := range size {
		visited[y] = make([]bool, size)
	}

	dirs := [4]pos{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	var components [][]pos

	for sy := range size {
		for sx := range size {
			if marks[sy][sx] == shaded || visited[sy][sx] {
				continue
			}
			// BFS to find this component.
			var component []pos
			queue := []pos{{sx, sy}}
			visited[sy][sx] = true
			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				component = append(component, curr)
				for _, d := range dirs {
					nx, ny := curr.x+d.x, curr.y+d.y
					if nx >= 0 && nx < size && ny >= 0 && ny < size &&
						!visited[ny][nx] && marks[ny][nx] != shaded {
						visited[ny][nx] = true
						queue = append(queue, pos{nx, ny})
					}
				}
			}
			components = append(components, component)
		}
	}

	if len(components) <= 1 {
		return
	}

	// Find the largest component.
	largest := 0
	for i, comp := range components {
		if len(comp) > len(components[largest]) {
			largest = i
		}
	}

	// Mark all cells NOT in the largest component as conflicts.
	for i, comp := range components {
		if i == largest {
			continue
		}
		for _, p := range comp {
			conflicts[p.y][p.x] = true
		}
	}
}

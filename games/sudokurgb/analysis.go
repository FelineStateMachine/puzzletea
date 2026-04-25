package sudokurgb

type boardAnalysis struct {
	rowCounts        [gridSize][valueCount + 1]int
	colCounts        [gridSize][valueCount + 1]int
	rowOverQuota     [gridSize][valueCount + 1]bool
	colOverQuota     [gridSize][valueCount + 1]bool
	boxConflictCells [gridSize][gridSize]bool
}

func analyzeGrid(g grid) boardAnalysis {
	var analysis boardAnalysis

	for y := range gridSize {
		for x := range gridSize {
			value := g[y][x].v
			if value < 1 || value > valueCount {
				continue
			}
			analysis.rowCounts[y][value]++
			analysis.colCounts[x][value]++
		}
	}

	for i := range gridSize {
		for value := 1; value <= valueCount; value++ {
			analysis.rowOverQuota[i][value] = analysis.rowCounts[i][value] > houseQuota
			analysis.colOverQuota[i][value] = analysis.colCounts[i][value] > houseQuota
		}
	}

	for boxY := range 3 {
		for boxX := range 3 {
			markBoxConflicts(&analysis, g, boxX, boxY)
		}
	}

	return analysis
}

func markBoxConflicts(analysis *boardAnalysis, g grid, boxX, boxY int) {
	type pos struct {
		x int
		y int
	}

	var seen [valueCount + 1][]pos
	for dy := range 3 {
		for dx := range 3 {
			x, y := boxX*3+dx, boxY*3+dy
			value := g[y][x].v
			if value < 1 || value > valueCount {
				continue
			}
			seen[value] = append(seen[value], pos{x: x, y: y})
		}
	}

	for value := 1; value <= valueCount; value++ {
		if len(seen[value]) <= houseQuota {
			continue
		}
		for _, cell := range seen[value] {
			analysis.boxConflictCells[cell.y][cell.x] = true
		}
	}
}

func (a boardAnalysis) hasConflicts() bool {
	for i := range gridSize {
		for value := 1; value <= valueCount; value++ {
			if a.rowOverQuota[i][value] || a.colOverQuota[i][value] {
				return true
			}
		}
	}

	for y := range gridSize {
		for x := range gridSize {
			if a.boxConflictCells[y][x] {
				return true
			}
		}
	}

	return false
}

func (a boardAnalysis) rowCountsString(row int) string {
	return countTripletString(a.rowCounts[row])
}

func (a boardAnalysis) colCountsString(col int) string {
	return countTripletString(a.colCounts[col])
}

func countTripletString(counts [valueCount + 1]int) string {
	return digitString(counts[1]) + "/" + digitString(counts[2]) + "/" + digitString(counts[3])
}

func digitString(v int) string {
	return string(rune('0' + v))
}

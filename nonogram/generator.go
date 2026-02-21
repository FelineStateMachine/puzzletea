package nonogram

import (
	"context"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxAttempts      = 100
	maxPossibilities = 20000
	maxCacheEntries  = 16384
	maxCacheCells    = 500_000_000
)

type lineCacheKey struct {
	length int
	hint   string
}

type possibilitiesCacheEntry struct {
	possibilities [][]cellState
	cellCost      int
}

var (
	possibilitiesCache     = make(map[lineCacheKey]*possibilitiesCacheEntry)
	possibilitiesOrder     = make([]lineCacheKey, 0, maxCacheEntries)
	possibilitiesOrderHead = 0
	possibilitiesCacheSize int
	possibilitiesMu        sync.Mutex
)

func cacheKey(length int, hint []int) lineCacheKey {
	return lineCacheKey{
		length: length,
		hint:   encodeHint(hint),
	}
}

func encodeHint(hint []int) string {
	if len(hint) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(hint) * 3)
	for i, v := range hint {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()
}

func GenerateRandomTomography(mode NonogramMode) Hints {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GenerateRandomTomographySeeded(mode, rng)
}

func GenerateRandomTomographySeeded(mode NonogramMode, rng *rand.Rand) Hints {
	maxDim := max(mode.Width, mode.Height)
	timeout := time.Duration(maxDim) * time.Second

	for attempt := range maxAttempts {
		density := lerp(mode.Density, 0.5, float64(attempt)*0.02)
		s := generateRandomStateSeeded(mode.Height, mode.Width, density, rng)
		g := newGrid(s)
		hints := generateTomography(g)
		if !isValidPuzzle(hints, mode.Height, mode.Width) {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		count := countSolutions(hints, mode.Width, mode.Height, 2, ctx)
		cancel()
		if count == 1 {
			return hints
		}
	}
	return Hints{}
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func generateRandomState(h, w int, density float64) state {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return generateRandomStateSeeded(h, w, density, rng)
}

func generateRandomStateSeeded(h, w int, density float64, rng *rand.Rand) state {
	if h <= 0 || w <= 0 {
		return ""
	}

	density = max(0.1, min(0.9, density))

	var b strings.Builder
	b.Grow((w + 1) * h)

	for i := range h {
		rowDensity := density + (rng.Float64()-0.5)*0.3
		rowDensity = max(0.05, min(0.95, rowDensity))
		for range w {
			if rng.Float64() < rowDensity {
				b.WriteRune(filledTile)
			} else {
				b.WriteRune(emptyTile)
			}
		}
		if i < h-1 {
			b.WriteRune('\n')
		}
	}

	return state(b.String())
}

func isValidPuzzle(hints Hints, height, width int) bool {
	for _, rh := range hints.rows {
		if len(rh) == 1 && (rh[0] == 0 || rh[0] == width) {
			return false
		}
	}
	for _, ch := range hints.cols {
		if len(ch) == 1 && (ch[0] == 0 || ch[0] == height) {
			return false
		}
	}
	return true
}

type cellState int

const (
	cellUnknown cellState = iota
	cellEmpty
	cellFilled
)

type solutionGrid struct {
	cells [][]cellState
	w, h  int
}

type cellChange struct {
	x, y int
	prev cellState
}

type undoStack struct {
	changes []cellChange
}

func newSolutionGrid(w, h int) solutionGrid {
	cells := make([][]cellState, h)
	for y := range h {
		cells[y] = make([]cellState, w)
		for x := range w {
			cells[y][x] = cellUnknown
		}
	}
	return solutionGrid{cells: cells, w: w, h: h}
}

func (g *solutionGrid) isComplete() bool {
	for y := range g.h {
		for x := range g.w {
			if g.cells[y][x] == cellUnknown {
				return false
			}
		}
	}
	return true
}

func countSolutions(hints Hints, w, h, limit int, ctx context.Context) int {
	g := newSolutionGrid(w, h)
	stack := undoStack{
		changes: make([]cellChange, 0, w*h),
	}
	return countSolutionsRecursive(&g, hints, limit, ctx, &stack)
}

func countSolutionsRecursive(
	g *solutionGrid,
	hints Hints,
	limit int,
	ctx context.Context,
	stack *undoStack,
) int {
	if limit <= 0 {
		return 0
	}
	select {
	case <-ctx.Done():
		return -1
	default:
	}

	mark := stack.mark()
	if !propagateInPlace(g, hints, ctx, stack) {
		stack.revert(g, mark)
		return 0
	}

	if g.isComplete() {
		stack.revert(g, mark)
		return 1
	}

	x, y, ok := pickMostConstrainedCell(*g, hints)
	if !ok {
		stack.revert(g, mark)
		return 0
	}

	count := 0
	branchOrder := [2]cellState{cellFilled, cellEmpty}
	for _, value := range branchOrder {
		branchMark := stack.mark()
		if stack.set(g, x, y, value) {
			n := countSolutionsRecursive(g, hints, limit-count, ctx, stack)
			if n < 0 {
				stack.revert(g, mark)
				return n
			}
			count += n
			if count >= limit {
				stack.revert(g, mark)
				return count
			}
		}
		stack.revert(g, branchMark)
	}

	stack.revert(g, mark)
	return count
}

func propagateInPlace(g *solutionGrid, hints Hints, ctx context.Context, stack *undoStack) bool {
	rowBuf := make([]cellState, g.w)
	colBuf := make([]cellState, g.h)

	changed := true
	for changed {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		changed = false

		for y := range g.h {
			for x := range g.w {
				rowBuf[x] = g.cells[y][x]
			}

			newRow, valid := propagateLine(rowBuf[:g.w], hints.rows[y])
			if !valid {
				return false
			}

			for x := range g.w {
				if g.cells[y][x] == cellUnknown && newRow[x] != cellUnknown {
					if !stack.set(g, x, y, newRow[x]) {
						return false
					}
					changed = true
				}
			}
		}

		for x := range g.w {
			for y := range g.h {
				colBuf[y] = g.cells[y][x]
			}

			newCol, valid := propagateLine(colBuf[:g.h], hints.cols[x])
			if !valid {
				return false
			}

			for y := range g.h {
				if g.cells[y][x] == cellUnknown && newCol[y] != cellUnknown {
					if !stack.set(g, x, y, newCol[y]) {
						return false
					}
					changed = true
				}
			}
		}
	}

	return true
}

func (s *undoStack) mark() int {
	return len(s.changes)
}

func (s *undoStack) set(g *solutionGrid, x, y int, state cellState) bool {
	current := g.cells[y][x]
	if current == state {
		return true
	}
	if current != cellUnknown {
		return false
	}
	s.changes = append(s.changes, cellChange{x: x, y: y, prev: current})
	g.cells[y][x] = state
	return true
}

func (s *undoStack) revert(g *solutionGrid, mark int) {
	for i := len(s.changes) - 1; i >= mark; i-- {
		change := s.changes[i]
		g.cells[change.y][change.x] = change.prev
	}
	s.changes = s.changes[:mark]
}

func pickMostConstrainedCell(g solutionGrid, hints Hints) (x, y int, ok bool) {
	firstX, firstY := -1, -1
	bestNeighborScore := -1
	bestHintWeight := -1

	for yy := range g.h {
		for xx := range g.w {
			if g.cells[yy][xx] != cellUnknown {
				continue
			}
			if firstX < 0 {
				firstX, firstY = xx, yy
			}

			neighborScore := 0
			if yy > 0 && g.cells[yy-1][xx] != cellUnknown {
				neighborScore++
			}
			if yy+1 < g.h && g.cells[yy+1][xx] != cellUnknown {
				neighborScore++
			}
			if xx > 0 && g.cells[yy][xx-1] != cellUnknown {
				neighborScore++
			}
			if xx+1 < g.w && g.cells[yy][xx+1] != cellUnknown {
				neighborScore++
			}

			hintWeight := len(hints.rows[yy]) + len(hints.cols[xx])
			if neighborScore > bestNeighborScore ||
				(neighborScore == bestNeighborScore && hintWeight > bestHintWeight) {
				bestNeighborScore = neighborScore
				bestHintWeight = hintWeight
				x, y = xx, yy
				ok = true
			}
		}
	}

	if firstX < 0 {
		return 0, 0, false
	}
	if bestNeighborScore <= 0 {
		return firstX, firstY, true
	}
	return x, y, ok
}

func propagateLine(line []cellState, hint []int) ([]cellState, bool) {
	n := len(line)
	possibilities := getCachedPossibilities(n, hint)

	if len(possibilities) > maxPossibilities {
		return line, true
	}

	filtered := make([][]cellState, 0, len(possibilities))
	for _, poss := range possibilities {
		if matchesLine(poss, line) {
			filtered = append(filtered, poss)
		}
	}

	if len(filtered) == 0 {
		return nil, false
	}

	result := make([]cellState, n)
	for i := range n {
		allFilled := true
		allEmpty := true
		for _, poss := range filtered {
			if poss[i] != cellFilled {
				allFilled = false
			}
			if poss[i] != cellEmpty {
				allEmpty = false
			}
		}
		if allFilled {
			result[i] = cellFilled
		} else if allEmpty {
			result[i] = cellEmpty
		} else {
			result[i] = cellUnknown
		}
	}

	return result, true
}

func getCachedPossibilities(length int, hint []int) [][]cellState {
	key := cacheKey(length, hint)

	possibilitiesMu.Lock()
	if entry, ok := possibilitiesCache[key]; ok {
		cached := entry.possibilities
		possibilitiesMu.Unlock()
		return cached
	}
	possibilitiesMu.Unlock()

	possibilities := generateLinePossibilities(length, hint)
	cellCost := len(possibilities) * length
	if cellCost <= 0 || cellCost > maxCacheCells {
		return possibilities
	}

	possibilitiesMu.Lock()
	if entry, ok := possibilitiesCache[key]; ok {
		possibilitiesMu.Unlock()
		return entry.possibilities
	}
	possibilitiesCache[key] = &possibilitiesCacheEntry{
		possibilities: possibilities,
		cellCost:      cellCost,
	}
	possibilitiesOrder = append(possibilitiesOrder, key)
	possibilitiesCacheSize += cellCost
	trimPossibilitiesCacheLocked()
	possibilitiesMu.Unlock()

	return possibilities
}

func trimPossibilitiesCacheLocked() {
	for len(possibilitiesCache) > maxCacheEntries || possibilitiesCacheSize > maxCacheCells {
		if possibilitiesOrderHead >= len(possibilitiesOrder) {
			return
		}
		key := possibilitiesOrder[possibilitiesOrderHead]
		possibilitiesOrderHead++
		entry, ok := possibilitiesCache[key]
		if !ok {
			continue
		}
		possibilitiesCacheSize -= entry.cellCost
		delete(possibilitiesCache, key)
	}

	if possibilitiesOrderHead > 1024 && possibilitiesOrderHead*2 > len(possibilitiesOrder) {
		compacted := make([]lineCacheKey, len(possibilitiesOrder)-possibilitiesOrderHead)
		copy(compacted, possibilitiesOrder[possibilitiesOrderHead:])
		possibilitiesOrder = compacted
		possibilitiesOrderHead = 0
	}
}

func generateLinePossibilities(length int, hint []int) [][]cellState {
	if len(hint) == 0 || (len(hint) == 1 && hint[0] == 0) {
		result := make([]cellState, length)
		for i := range result {
			result[i] = cellEmpty
		}
		return [][]cellState{result}
	}

	totalFilled := 0
	for _, h := range hint {
		totalFilled += h
	}
	minLength := totalFilled + len(hint) - 1
	if minLength > length {
		return nil
	}

	return generateLinePossibilitiesRecursive(length, hint, 0, 0, make([]cellState, 0, length))
}

func generateLinePossibilitiesRecursive(length int, hint []int, hintIdx, pos int, current []cellState) [][]cellState {
	if hintIdx == len(hint) {
		for len(current) < length {
			current = append(current, cellEmpty)
		}
		return [][]cellState{append([]cellState(nil), current...)}
	}

	blockLen := hint[hintIdx]
	remainingHints := hint[hintIdx+1:]
	remainingFilled := 0
	for _, h := range remainingHints {
		remainingFilled += h
	}
	minRemaining := remainingFilled + len(remainingHints)

	maxStart := length - blockLen - minRemaining

	var results [][]cellState

	for start := pos; start <= maxStart; start++ {
		newCurrent := append([]cellState(nil), current...)

		for i := len(current); i < start; i++ {
			newCurrent = append(newCurrent, cellEmpty)
		}

		for i := 0; i < blockLen; i++ {
			newCurrent = append(newCurrent, cellFilled)
		}

		if hintIdx < len(hint)-1 {
			newCurrent = append(newCurrent, cellEmpty)
		}
		results = append(results, generateLinePossibilitiesRecursive(length, hint, hintIdx+1, len(newCurrent), newCurrent)...)
	}

	return results
}

func matchesLine(possibility, line []cellState) bool {
	if len(possibility) != len(line) {
		return false
	}
	for i, p := range possibility {
		if line[i] != cellUnknown && line[i] != p {
			return false
		}
	}
	return true
}

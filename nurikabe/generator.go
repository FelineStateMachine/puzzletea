package nurikabe

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"sync"
	"time"
)

const (
	generationHardTimeout        = 10 * time.Second
	uniquenessMinRemainingBudget = 250 * time.Millisecond
	frontierBiasSampleCount      = 7
)

type generationBudget struct {
	attempts             int
	solvabilityNodes     int
	uniquenessNodes      int
	uniquenessProbeLimit int
}

type islandProfile struct {
	baseSeaCoverage   float64
	seaJitterRatio    float64
	minSeaCoverage    float64
	maxSeaCoverage    float64
	minAverageSize    float64
	maxSingletonRatio float64
	minLargestIsland  int
	maxLargestIsland  int
}

type candidateState struct {
	width, height int
	sea           [][]bool
	labels        [][]int
	frontier      []point
	frontierPos   [][]int

	seaCount      int
	sea2x2Count   int
	nextLabel     int
	componentSize map[int]int
	sizeHistogram map[int]int
	singletons    int
	largestIsland int
}

type candidateStats struct {
	clueCount        int
	totalLand        int
	singletonCount   int
	singletonRatio   float64
	largestIsland    int
	averageIsland    float64
	sizeEntropy      float64
	normalizedSpread float64
	fragmentPenalty  float64
	clueDistance     float64
}

type candidateResult struct {
	attemptID int
	puzzle    Puzzle
	stats     candidateStats
	score     float64
	clueKey   string
}

type attemptSeed struct {
	attemptID int
	seedA     uint64
	seedB     uint64
}

type attemptOutcome struct {
	attemptID int
	candidate *candidateResult
	err       error
}

type uniquenessJob struct {
	clueKey string
	puzzle  Puzzle
}

type uniquenessOutcome struct {
	clueKey string
	count   int
	err     error
}

func Generate(mode NurikabeMode) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GenerateSeededWithContext(context.Background(), mode, rng)
}

func GenerateWithContext(ctx context.Context, mode NurikabeMode) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GenerateSeededWithContext(ctx, mode, rng)
}

func GenerateSeeded(mode NurikabeMode, rng *rand.Rand) (Puzzle, error) {
	return GenerateSeededWithContext(context.Background(), mode, rng)
}

func GenerateSeededWithContext(ctx context.Context, mode NurikabeMode, rng *rand.Rand) (Puzzle, error) {
	if rng == nil {
		return Puzzle{}, fmt.Errorf("nil RNG")
	}
	if mode.Width <= 0 || mode.Height <= 0 {
		return Puzzle{}, fmt.Errorf("invalid mode size %dx%d", mode.Width, mode.Height)
	}
	if mode.MaxIslandSize <= 0 {
		return Puzzle{}, fmt.Errorf("invalid max island size: %d", mode.MaxIslandSize)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	timeboxedCtx, cancel := context.WithTimeout(ctx, generationHardTimeout)
	defer cancel()
	deadline, _ := timeboxedCtx.Deadline()

	area := mode.Width * mode.Height
	clueTarget := int(math.Round(float64(area) * mode.ClueDensity))
	if clueTarget < 2 {
		clueTarget = 2
	}
	if clueTarget > area-2 {
		clueTarget = area - 2
	}

	profile := modeIslandProfile(mode)
	budget := modeGenerationBudget(mode)
	workerCount := candidateWorkerCount(mode)
	uniquenessWorkers := uniquenessWorkerCount(mode)

	masterA := rng.Uint64()
	masterB := rng.Uint64()

	attemptCh := make(chan attemptSeed)
	candidateCh := make(chan attemptOutcome, workerCount*2)
	uniqueCh := make(chan uniquenessJob, budget.uniquenessProbeLimit)
	uniqueResultCh := make(chan uniquenessOutcome, budget.uniquenessProbeLimit)

	var candidateWG sync.WaitGroup
	for range workerCount {
		candidateWG.Add(1)
		go func() {
			defer candidateWG.Done()
			candidateWorker(timeboxedCtx, mode, clueTarget, profile, attemptCh, candidateCh)
		}()
	}

	var uniqueWG sync.WaitGroup
	for range uniquenessWorkers {
		uniqueWG.Add(1)
		go func() {
			defer uniqueWG.Done()
			uniquenessWorker(timeboxedCtx, budget, uniqueCh, uniqueResultCh)
		}()
	}
	go func() {
		uniqueWG.Wait()
		close(uniqueResultCh)
	}()

	bestSolvable := candidateResult{}
	haveBestSolvable := false
	bestUnique := candidateResult{}
	haveBestUnique := false

	pendingUniqueByKey := map[string]candidateResult{}
	knownUniqueByKey := map[string]uniquenessOutcome{}
	uniquenessScheduled := 0

	handleUniqueResult := func(result uniquenessOutcome) {
		knownUniqueByKey[result.clueKey] = result
		candidate, ok := pendingUniqueByKey[result.clueKey]
		delete(pendingUniqueByKey, result.clueKey)
		if !ok {
			return
		}
		if result.err != nil {
			if errors.Is(result.err, context.Canceled) ||
				errors.Is(result.err, context.DeadlineExceeded) {
				return
			}
			return
		}
		if result.count == 1 && (!haveBestUnique || betterCandidate(candidate, bestUnique)) {
			bestUnique = candidate
			haveBestUnique = true
		}
	}

	drainUnique := func() {
		for {
			select {
			case result, ok := <-uniqueResultCh:
				if !ok {
					return
				}
				handleUniqueResult(result)
			default:
				return
			}
		}
	}

	attemptID := 0
outer:
	for attemptID < budget.attempts {
		if err := timeboxedCtx.Err(); err != nil {
			break
		}
		drainUnique()

		roundSize := min(workerCount, budget.attempts-attemptID)
		dispatched := 0
		for i := 0; i < roundSize; i++ {
			seedA, seedB := deriveAttemptSeeds(masterA, masterB, attemptID+i)
			job := attemptSeed{attemptID: attemptID + i, seedA: seedA, seedB: seedB}
			select {
			case attemptCh <- job:
				dispatched++
			case <-timeboxedCtx.Done():
				break outer
			}
		}

		outcomes := make([]attemptOutcome, 0, dispatched)
		for len(outcomes) < dispatched {
			select {
			case out := <-candidateCh:
				outcomes = append(outcomes, out)
			case <-timeboxedCtx.Done():
				break outer
			}
		}

		sort.Slice(outcomes, func(i, j int) bool {
			return outcomes[i].attemptID < outcomes[j].attemptID
		})

		for _, out := range outcomes {
			if out.candidate == nil {
				continue
			}
			candidate := *out.candidate
			if !haveBestSolvable || betterCandidate(candidate, bestSolvable) {
				bestSolvable = candidate
				haveBestSolvable = true
			}

			if knownResult, ok := knownUniqueByKey[candidate.clueKey]; ok {
				if knownResult.err == nil &&
					knownResult.count == 1 &&
					(!haveBestUnique || betterCandidate(candidate, bestUnique)) {
					bestUnique = candidate
					haveBestUnique = true
				}
				continue
			}

			if pendingCandidate, ok := pendingUniqueByKey[candidate.clueKey]; ok {
				if betterCandidate(candidate, pendingCandidate) {
					pendingUniqueByKey[candidate.clueKey] = candidate
				}
				continue
			}

			if shouldProbeUniqueness(candidate, haveBestSolvable, bestSolvable, uniquenessScheduled, budget, deadline) {
				job := uniquenessJob{
					clueKey: candidate.clueKey,
					puzzle:  candidate.puzzle,
				}
				pendingUniqueByKey[candidate.clueKey] = candidate
				select {
				case uniqueCh <- job:
					uniquenessScheduled++
				case <-timeboxedCtx.Done():
					break outer
				}
			}
		}

		drainUnique()
		attemptID += dispatched
	}

	close(attemptCh)
	candidateWG.Wait()
	close(uniqueCh)
	for result := range uniqueResultCh {
		handleUniqueResult(result)
	}

	if haveBestUnique {
		return bestUnique.puzzle, nil
	}
	if haveBestSolvable {
		return bestSolvable.puzzle, nil
	}

	if err := timeboxedCtx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return Puzzle{}, fmt.Errorf("nurikabe generation canceled: %w", err)
		}
		return Puzzle{}, fmt.Errorf("nurikabe generation timed out: %w", err)
	}

	return Puzzle{}, fmt.Errorf("unable to generate solvable nurikabe puzzle for mode %q within constraints", mode.Title())
}

func candidateWorker(
	ctx context.Context,
	mode NurikabeMode,
	clueTarget int,
	profile islandProfile,
	attemptCh <-chan attemptSeed,
	candidateCh chan<- attemptOutcome,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case attempt, ok := <-attemptCh:
			if !ok {
				return
			}

			rng := rand.New(rand.NewPCG(attempt.seedA, attempt.seedB))
			candidate, err := buildCandidateByCarving(mode, clueTarget, profile, rng, attempt.attemptID)
			out := attemptOutcome{attemptID: attempt.attemptID, err: err}
			if err == nil {
				candidateCopy := candidate
				out.candidate = &candidateCopy
			}

			select {
			case candidateCh <- out:
			case <-ctx.Done():
				return
			}
		}
	}
}

func uniquenessWorker(
	ctx context.Context,
	budget generationBudget,
	uniqueCh <-chan uniquenessJob,
	uniqueResultCh chan<- uniquenessOutcome,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-uniqueCh:
			if !ok {
				return
			}

			count, _, err := CountSolutionsContext(ctx, job.puzzle, 2, budget.uniquenessNodes)
			result := uniquenessOutcome{clueKey: job.clueKey, count: count, err: err}
			select {
			case uniqueResultCh <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func buildCandidateByCarving(
	mode NurikabeMode,
	clueTarget int,
	profile islandProfile,
	rng *rand.Rand,
	attemptID int,
) (candidateResult, error) {
	start := point{rng.IntN(mode.Width), rng.IntN(mode.Height)}
	state := newCandidateState(mode.Width, mode.Height, start)

	targetSea := chooseTargetSeaCells(mode, profile, rng)
	for state.seaCount < targetSea {
		candidate, ok := state.popFrontier(rng)
		if !ok {
			return candidateResult{}, fmt.Errorf("frontier exhausted")
		}
		if !state.canCarve(candidate) {
			continue
		}
		if !state.carve(candidate) {
			continue
		}
	}

	components := state.components()
	stats := state.stats(components, clueTarget)
	if !stats.meetsProfile(mode, profile, clueTarget) {
		return candidateResult{}, fmt.Errorf("distribution out of bounds")
	}

	clues := make(clueGrid, mode.Height)
	for y := range mode.Height {
		clues[y] = make([]int, mode.Width)
	}
	for _, component := range components {
		if len(component) == 0 {
			continue
		}
		clueCell := component[rng.IntN(len(component))]
		clues[clueCell.y][clueCell.x] = len(component)
	}

	puzzle := Puzzle{Width: mode.Width, Height: mode.Height, Clues: clues}
	if err := validateClues(puzzle.Clues, puzzle.Width, puzzle.Height); err != nil {
		return candidateResult{}, err
	}

	completed := state.completedGrid(clues)
	if !isSolvedGrid(completed, clues) {
		return candidateResult{}, fmt.Errorf("constructed board failed solved validation")
	}

	stats.clueCount = len(components)
	stats.clueDistance = clueDistance(stats.clueCount, clueTarget)

	candidate := candidateResult{
		attemptID: attemptID,
		puzzle:    puzzle,
		stats:     stats,
		score:     scoreCandidate(mode, stats),
		clueKey:   serializeClues(clues),
	}

	return candidate, nil
}

func newCandidateState(width, height int, start point) *candidateState {
	sea := make([][]bool, height)
	labels := make([][]int, height)
	frontierPos := make([][]int, height)
	for y := range height {
		sea[y] = make([]bool, width)
		labels[y] = make([]int, width)
		frontierPos[y] = make([]int, width)
		for x := range width {
			labels[y][x] = 1
			frontierPos[y][x] = -1
		}
	}

	state := &candidateState{
		width:         width,
		height:        height,
		sea:           sea,
		labels:        labels,
		frontier:      make([]point, 0, width*height),
		frontierPos:   frontierPos,
		componentSize: map[int]int{1: width * height},
		sizeHistogram: map[int]int{width * height: 1},
		singletons:    1,
		largestIsland: width * height,
		nextLabel:     2,
	}
	if width*height > 1 {
		state.singletons = 0
	}
	state.carve(start)
	return state
}

func (s *candidateState) inBounds(p point) bool {
	return p.x >= 0 && p.x < s.width && p.y >= 0 && p.y < s.height
}

func (s *candidateState) canCarve(p point) bool {
	if !s.inBounds(p) || s.sea[p.y][p.x] {
		return false
	}
	if s.newSeaSquaresFromCarve(p) > 0 {
		return false
	}
	return true
}

func (s *candidateState) carve(p point) bool {
	if !s.canCarve(p) {
		return false
	}

	oldLabel := s.labels[p.y][p.x]
	if oldLabel < 0 {
		return false
	}

	deltaSquares := s.newSeaSquaresFromCarve(p)
	s.sea[p.y][p.x] = true
	s.labels[p.y][p.x] = -1
	s.seaCount++
	s.sea2x2Count += deltaSquares
	s.removeFrontier(p)

	oldSize := s.componentSize[oldLabel]
	s.removeComponent(oldLabel)

	neighbors := s.landNeighborsWithLabel(p, oldLabel)
	sort.Slice(neighbors, func(i, j int) bool {
		if neighbors[i].y != neighbors[j].y {
			return neighbors[i].y < neighbors[j].y
		}
		return neighbors[i].x < neighbors[j].x
	})
	neighbors = dedupePoints(neighbors)

	switch len(neighbors) {
	case 0:
		// Entire component removed.
	case 1:
		s.setComponent(oldLabel, oldSize-1)
	default:
		s.relabelSplit(oldLabel, neighbors)
	}

	for _, d := range dirs {
		n := point{x: p.x + d.x, y: p.y + d.y}
		if !s.inBounds(n) || s.sea[n.y][n.x] {
			continue
		}
		s.addFrontier(n)
	}

	return true
}

func dedupePoints(in []point) []point {
	if len(in) < 2 {
		return in
	}
	out := in[:1]
	for i := 1; i < len(in); i++ {
		if in[i] != in[i-1] {
			out = append(out, in[i])
		}
	}
	return out
}

func (s *candidateState) landNeighborsWithLabel(p point, label int) []point {
	neighbors := make([]point, 0, 4)
	for _, d := range dirs {
		n := point{x: p.x + d.x, y: p.y + d.y}
		if !s.inBounds(n) {
			continue
		}
		if s.labels[n.y][n.x] == label {
			neighbors = append(neighbors, n)
		}
	}
	return neighbors
}

func (s *candidateState) relabelSplit(oldLabel int, seeds []point) {
	visited := make([][]bool, s.height)
	for y := range s.height {
		visited[y] = make([]bool, s.width)
	}

	componentIndex := 0
	for _, seed := range seeds {
		if visited[seed.y][seed.x] || s.labels[seed.y][seed.x] != oldLabel {
			continue
		}

		newLabel := oldLabel
		if componentIndex > 0 {
			newLabel = s.nextLabel
			s.nextLabel++
		}

		queue := []point{seed}
		visited[seed.y][seed.x] = true
		cells := make([]point, 0, 16)
		for len(queue) > 0 {
			cell := queue[0]
			queue = queue[1:]
			cells = append(cells, cell)

			for _, d := range dirs {
				n := point{x: cell.x + d.x, y: cell.y + d.y}
				if !s.inBounds(n) || visited[n.y][n.x] || s.labels[n.y][n.x] != oldLabel {
					continue
				}
				visited[n.y][n.x] = true
				queue = append(queue, n)
			}
		}

		if newLabel != oldLabel {
			for _, cell := range cells {
				s.labels[cell.y][cell.x] = newLabel
			}
		}
		s.setComponent(newLabel, len(cells))
		componentIndex++
	}
}

func (s *candidateState) newSeaSquaresFromCarve(p point) int {
	count := 0
	for dy := -1; dy <= 0; dy++ {
		for dx := -1; dx <= 0; dx++ {
			x0 := p.x + dx
			y0 := p.y + dy
			if x0 < 0 || y0 < 0 || x0+1 >= s.width || y0+1 >= s.height {
				continue
			}

			allSea := true
			for yy := y0; yy <= y0+1; yy++ {
				for xx := x0; xx <= x0+1; xx++ {
					if xx == p.x && yy == p.y {
						continue
					}
					if !s.sea[yy][xx] {
						allSea = false
						break
					}
				}
				if !allSea {
					break
				}
			}
			if allSea {
				count++
			}
		}
	}
	return count
}

func (s *candidateState) addFrontier(p point) {
	if !s.inBounds(p) || s.sea[p.y][p.x] || s.frontierPos[p.y][p.x] >= 0 {
		return
	}
	idx := len(s.frontier)
	s.frontier = append(s.frontier, p)
	s.frontierPos[p.y][p.x] = idx
}

func (s *candidateState) removeFrontier(p point) {
	if !s.inBounds(p) {
		return
	}
	idx := s.frontierPos[p.y][p.x]
	if idx < 0 || idx >= len(s.frontier) {
		return
	}

	last := len(s.frontier) - 1
	tail := s.frontier[last]
	s.frontier[idx] = tail
	s.frontierPos[tail.y][tail.x] = idx
	s.frontier = s.frontier[:last]
	s.frontierPos[p.y][p.x] = -1
}

func (s *candidateState) popFrontier(rng *rand.Rand) (point, bool) {
	if len(s.frontier) == 0 {
		return point{}, false
	}

	bestIdx := -1
	bestScore := -1
	samples := min(frontierBiasSampleCount, len(s.frontier))
	for range samples {
		idx := rng.IntN(len(s.frontier))
		score := s.frontierPriorityScore(s.frontier[idx])
		if score > bestScore || (score == bestScore && rng.IntN(2) == 0) {
			bestScore = score
			bestIdx = idx
		}
	}

	if bestIdx < 0 {
		bestIdx = rng.IntN(len(s.frontier))
	}

	picked := s.frontier[bestIdx]
	s.removeFrontier(picked)
	return picked, true
}

func (s *candidateState) frontierPriorityScore(p point) int {
	if !s.inBounds(p) || s.sea[p.y][p.x] {
		return 0
	}

	label := s.labels[p.y][p.x]
	componentSize := s.componentSize[label]
	sameLabelNeighbors := len(s.landNeighborsWithLabel(p, label))

	// Favor carving from larger components; local degree nudges toward useful splits.
	return componentSize*16 + sameLabelNeighbors*3
}

func (s *candidateState) removeComponent(label int) {
	size, ok := s.componentSize[label]
	if !ok {
		return
	}
	delete(s.componentSize, label)
	s.sizeHistogram[size]--
	if s.sizeHistogram[size] <= 0 {
		delete(s.sizeHistogram, size)
	}
	if size == 1 {
		s.singletons--
	}
	if size == s.largestIsland {
		s.recomputeLargestIsland()
	}
}

func (s *candidateState) setComponent(label, size int) {
	if size <= 0 {
		return
	}
	s.componentSize[label] = size
	s.sizeHistogram[size]++
	if size == 1 {
		s.singletons++
	}
	if size > s.largestIsland {
		s.largestIsland = size
	}
}

func (s *candidateState) recomputeLargestIsland() {
	largest := 0
	for size := range s.sizeHistogram {
		if size > largest {
			largest = size
		}
	}
	s.largestIsland = largest
}

func (s *candidateState) components() [][]point {
	byLabel := map[int][]point{}
	labels := make([]int, 0, len(s.componentSize))
	for label := range s.componentSize {
		labels = append(labels, label)
	}
	sort.Ints(labels)

	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			label := s.labels[y][x]
			if label < 0 {
				continue
			}
			byLabel[label] = append(byLabel[label], point{x: x, y: y})
		}
	}

	components := make([][]point, 0, len(labels))
	for _, label := range labels {
		component := byLabel[label]
		if len(component) == 0 {
			continue
		}
		components = append(components, component)
	}
	return components
}

func (s *candidateState) completedGrid(clues clueGrid) grid {
	g := newGrid(s.width, s.height, islandCell)
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			if s.sea[y][x] {
				g[y][x] = seaCell
				continue
			}
			if clues[y][x] > 0 {
				g[y][x] = islandCell
			}
		}
	}
	return g
}

func (s *candidateState) stats(components [][]point, clueTarget int) candidateStats {
	stats := candidateStats{}
	if len(components) == 0 {
		return stats
	}

	sizes := make([]int, 0, len(components))
	total := 0
	for _, component := range components {
		size := len(component)
		sizes = append(sizes, size)
		total += size
		if size == 1 {
			stats.singletonCount++
		}
		if size > stats.largestIsland {
			stats.largestIsland = size
		}
	}

	stats.clueCount = len(components)
	stats.totalLand = total
	stats.singletonRatio = float64(stats.singletonCount) / float64(max(1, len(components)))
	stats.averageIsland = float64(total) / float64(len(components))
	stats.sizeEntropy = componentSizeEntropy(sizes)
	stats.normalizedSpread = normalizedSpread(sizes)
	stats.fragmentPenalty = fragmentPenalty(len(components), clueTarget)
	stats.clueDistance = clueDistance(len(components), clueTarget)

	return stats
}

func componentSizeEntropy(sizes []int) float64 {
	if len(sizes) == 0 {
		return 0
	}
	freq := map[int]int{}
	for _, size := range sizes {
		freq[size]++
	}
	denom := float64(len(sizes))
	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / denom
		entropy -= p * math.Log(p)
	}
	return entropy
}

func normalizedSpread(sizes []int) float64 {
	if len(sizes) == 0 {
		return 0
	}
	mean := 0.0
	for _, size := range sizes {
		mean += float64(size)
	}
	mean /= float64(len(sizes))
	if mean == 0 {
		return 0
	}
	variance := 0.0
	for _, size := range sizes {
		d := float64(size) - mean
		variance += d * d
	}
	variance /= float64(len(sizes))
	return math.Sqrt(variance) / mean
}

func fragmentPenalty(clueCount, clueTarget int) float64 {
	if clueTarget <= 0 {
		return 0
	}
	delta := math.Abs(float64(clueCount - clueTarget))
	return delta / float64(clueTarget)
}

func clueDistance(clueCount, clueTarget int) float64 {
	if clueTarget <= 0 {
		return 0
	}
	return math.Abs(float64(clueCount-clueTarget)) / float64(clueTarget)
}

func (s candidateStats) meetsProfile(mode NurikabeMode, profile islandProfile, clueTarget int) bool {
	if s.clueCount < 2 {
		return false
	}
	if s.singletonRatio > profile.maxSingletonRatio {
		return false
	}
	if s.averageIsland < profile.minAverageSize {
		return false
	}
	if s.largestIsland < profile.minLargestIsland {
		return false
	}
	maxLargest := mode.MaxIslandSize * 2
	if profile.maxLargestIsland > 0 {
		maxLargest = min(maxLargest, profile.maxLargestIsland)
	}
	if s.largestIsland > maxLargest {
		return false
	}

	tol := max(2, clueTarget/2)
	if s.clueCount < max(2, clueTarget-tol) || s.clueCount > clueTarget+tol {
		return false
	}

	return true
}

func scoreCandidate(mode NurikabeMode, stats candidateStats) float64 {
	score := 0.0
	score += stats.sizeEntropy * 3.2
	score += stats.normalizedSpread * 1.3
	score -= stats.singletonRatio * 2.6
	score -= stats.fragmentPenalty * 1.6
	score -= stats.clueDistance * 0.9
	score += min(1.5, float64(stats.largestIsland)/float64(max(1, mode.MaxIslandSize))) * 0.4
	idealLargest := float64(max(mode.Width, mode.Height)) * 1.45
	if stats.largestIsland > int(math.Round(idealLargest)) {
		oversize := (float64(stats.largestIsland) - idealLargest) / max(1.0, idealLargest)
		score -= oversize * 1.2
	}
	return score
}

func betterCandidate(a, b candidateResult) bool {
	if d := a.score - b.score; math.Abs(d) > 1e-9 {
		return d > 0
	}
	if a.stats.singletonCount != b.stats.singletonCount {
		return a.stats.singletonCount < b.stats.singletonCount
	}
	if a.attemptID != b.attemptID {
		return a.attemptID < b.attemptID
	}
	return a.clueKey < b.clueKey
}

func shouldProbeUniqueness(
	candidate candidateResult,
	haveBestSolvable bool,
	bestSolvable candidateResult,
	scheduled int,
	budget generationBudget,
	deadline time.Time,
) bool {
	if scheduled >= budget.uniquenessProbeLimit {
		return false
	}
	if time.Until(deadline) < uniquenessMinRemainingBudget {
		return false
	}
	if scheduled < 3 {
		return true
	}
	if !haveBestSolvable {
		return true
	}
	return candidate.score >= bestSolvable.score-0.8
}

func chooseTargetSeaCells(mode NurikabeMode, profile islandProfile, rng *rand.Rand) int {
	area := mode.Width * mode.Height
	targetSea := int(math.Round(float64(area) * profile.baseSeaCoverage))
	jitter := int(math.Round(float64(area) * profile.seaJitterRatio))
	if jitter > 0 {
		targetSea += rng.IntN(2*jitter+1) - jitter
	}

	minSea := max(1, int(math.Round(float64(area)*profile.minSeaCoverage)))
	maxSea := min(area-1, int(math.Round(float64(area)*profile.maxSeaCoverage)))
	if targetSea < minSea {
		targetSea = minSea
	}
	if targetSea > maxSea {
		targetSea = maxSea
	}

	return targetSea
}

func deriveAttemptSeeds(masterA, masterB uint64, attemptID int) (uint64, uint64) {
	id := uint64(attemptID + 1)
	seedA := splitmix64(masterA ^ (id * 0x9e3779b97f4a7c15))
	seedB := splitmix64(masterB ^ (id * 0xbf58476d1ce4e5b9))
	return seedA, seedB
}

func splitmix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

func candidateWorkerCount(mode NurikabeMode) int {
	switch mode.Title() {
	case "Mini":
		return 1
	case "Easy":
		return 2
	case "Medium":
		return 2
	case "Hard":
		return 4
	case "Expert":
		return 4
	default:
		return 2
	}
}

func uniquenessWorkerCount(mode NurikabeMode) int {
	switch mode.Title() {
	case "Mini":
		return 1
	case "Easy":
		return 1
	case "Medium":
		return 2
	case "Hard":
		return 4
	case "Expert":
		return 5
	default:
		return 2
	}
}

func modeIslandProfile(mode NurikabeMode) islandProfile {
	switch mode.Title() {
	case "Mini":
		return islandProfile{
			baseSeaCoverage:   0.51,
			seaJitterRatio:    0.05,
			minSeaCoverage:    0.40,
			maxSeaCoverage:    0.62,
			minAverageSize:    1.85,
			maxSingletonRatio: 0.63,
			minLargestIsland:  3,
			maxLargestIsland:  0,
		}
	case "Easy":
		return islandProfile{
			baseSeaCoverage:   0.55,
			seaJitterRatio:    0.05,
			minSeaCoverage:    0.42,
			maxSeaCoverage:    0.64,
			minAverageSize:    1.95,
			maxSingletonRatio: 0.56,
			minLargestIsland:  4,
			maxLargestIsland:  0,
		}
	case "Medium":
		return islandProfile{
			baseSeaCoverage:   0.58,
			seaJitterRatio:    0.05,
			minSeaCoverage:    0.44,
			maxSeaCoverage:    0.65,
			minAverageSize:    1.95,
			maxSingletonRatio: 0.58,
			minLargestIsland:  5,
			maxLargestIsland:  0,
		}
	case "Hard":
		return islandProfile{
			baseSeaCoverage:   0.60,
			seaJitterRatio:    0.04,
			minSeaCoverage:    0.44,
			maxSeaCoverage:    0.67,
			minAverageSize:    1.8,
			maxSingletonRatio: 0.72,
			minLargestIsland:  4,
			maxLargestIsland:  15,
		}
	case "Expert":
		return islandProfile{
			baseSeaCoverage:   0.62,
			seaJitterRatio:    0.04,
			minSeaCoverage:    0.44,
			maxSeaCoverage:    0.68,
			minAverageSize:    1.7,
			maxSingletonRatio: 0.78,
			minLargestIsland:  4,
			maxLargestIsland:  17,
		}
	default:
		return islandProfile{
			baseSeaCoverage:   0.55,
			seaJitterRatio:    0.05,
			minSeaCoverage:    0.42,
			maxSeaCoverage:    0.65,
			minAverageSize:    2.1,
			maxSingletonRatio: 0.48,
			minLargestIsland:  4,
			maxLargestIsland:  0,
		}
	}
}

func modeGenerationBudget(mode NurikabeMode) generationBudget {
	switch mode.Title() {
	case "Mini":
		return generationBudget{attempts: 320, solvabilityNodes: 25000, uniquenessNodes: 60000, uniquenessProbeLimit: 6}
	case "Easy":
		return generationBudget{attempts: 400, solvabilityNodes: 45000, uniquenessNodes: 100000, uniquenessProbeLimit: 8}
	case "Medium":
		return generationBudget{attempts: 500, solvabilityNodes: 70000, uniquenessNodes: 140000, uniquenessProbeLimit: 10}
	case "Hard":
		return generationBudget{attempts: 420, solvabilityNodes: 70000, uniquenessNodes: 120000, uniquenessProbeLimit: 8}
	case "Expert":
		return generationBudget{attempts: 320, solvabilityNodes: 70000, uniquenessNodes: 100000, uniquenessProbeLimit: 6}
	default:
		return generationBudget{attempts: 400, solvabilityNodes: 45000, uniquenessNodes: 100000, uniquenessProbeLimit: 8}
	}
}

func generationNodeLimit(mode NurikabeMode) int {
	return modeGenerationBudget(mode).solvabilityNodes
}

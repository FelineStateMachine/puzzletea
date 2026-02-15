# Hitori Puzzle Generation

This document outlines a strategy for generating NxN Hitori puzzles with a guaranteed unique solution.

## Overview

Hitori puzzles start with a **complete grid of numbers**. The player must shade some cells black so that:
1. No number appears twice in any row or column among the unshaded (white) cells
2. No two black cells share an edge (diagonal is allowed)
3. All white cells remain orthogonally connected

The generation approach:
1. **Create a Latin Square** (valid solved state with no duplicates in rows/cols)
2. **Generate a valid cell mask** (which cells are black, obeying Hitori rules)
3. **Construct the puzzle** (add duplicates for black cells in the solution)
4. **Verify uniqueness** via backtracking solver

---

## Step 1: Create a Latin Square

A Latin Square is an NxN grid where each number (1-N) appears exactly once in each row and column. This forms the "white cell" solution.

### Algorithm

```go
func generateLatinSquare(size int) grid {
    // 1. Create first row with shuffled 1-N
    firstRow := make([]rune, size)
    nums := make([]rune, size)
    for i := range size {
        nums[i] = rune('1' + i)
    }
    rand.Shuffle(len(nums), func(i, j int) { nums[i], nums[j] = nums[j], nums[i] })
    copy(firstRow, nums)

    // 2. Generate subsequent rows by cyclic shift
    g := make(grid, size)
    for y := range size {
        g[y] = make([]rune, size)
        for x := range size {
            g[y][x] = firstRow[(x+y)%size]
        }
    }

    // 3. Shuffle columns (preserves Latin property)
    cols := make([]int, size)
    for i := range size { cols[i] = i }
    rand.Shuffle(len(cols), func(i, j int) { cols[i], cols[j] = cols[j], cols[i] })

    result := make(grid, size)
    for y := range size {
        result[y] = make([]rune, size)
        for x := range size {
            result[y][x] = g[y][cols[x]]
        }
    }

    return result
}
```

### Why this works
- Cyclic shifts guarantee each row has all N numbers
- Column shuffling preserves the Latin property (each column still has all N numbers)
- The result is a valid Hitori "white cell" solution

---

## Step 2: Generate a Valid Cell Mask

We need to decide which cells are black. The mask must satisfy:
1. **No adjacent black cells**: Black cells cannot share an edge (can touch diagonally)
2. **White cells connected**: All white cells must form one orthogonal component

### Algorithm

```go
func generateValidMask(size int, blackRatio float64) [][]bool {
    mask := make([][]bool, size)
    for y := range size {
        mask[y] = make([]bool, size)
    }

    targetBlack := int(blackRatio * float64(size*size))
    attempts := 0
    const maxAttempts = 1000

    for countBlack(mask, size) < targetBlack && attempts < maxAttempts {
        attempts++
        x, y := rand.IntN(size), rand.IntN(size)
        
        if mask[y][x] {
            continue  // Already black
        }

        // Check no adjacent black
        if hasOrthogonalNeighbor(mask, size, x, y) {
            continue  // Would violate adjacency rule
        }

        mask[y][x] = true

        // Verify white connectivity
        if !whiteCellsConnected(mask, size) {
            mask[y][x] = false  // Revert: would isolate white cells
        }
    }

    return mask
}
```

### Helper: Check Orthogonal Neighbor

```go
func hasOrthogonalNeighbor(mask [][]bool, size, x, y int) bool {
    if x > 0 && mask[y][x-1]   { return true }
    if x < size-1 && mask[y][x+1] { return true }
    if y > 0 && mask[y-1][x]   { return true }
    if y < size-1 && mask[y+1][x] { return true }
    return false
}
```

### Helper: White Cell Connectivity (BFS/DFS)

```go
func whiteCellsConnected(mask [][]bool, size int) bool {
    // Find first white cell
    startX, startY := -1, -1
    found := false
    for y := range size {
        for x := range size {
            if !mask[y][x] {
                startX, startY = x, y
                found = true
                break
            }
        }
        if found {
            break
        }
    }
    if startX == -1 {
        return true  // No white cells (all black)
    }

    // BFS to count reachable white cells
    visited := make([][]bool, size)
    for y := range size {
        visited[y] = make([]bool, size)
    }

    queue := []cellPos{{startX, startY}}
    visited[startY][startX] = true
    visitedCount := 1

    for len(queue) > 0 {
        curr := queue[len(queue)-1]
        queue = queue[:len(queue)-1]

        // Check 4 orthogonal neighbors
        for _, d := range []struct{dx, dy int}{{0,-1},{0,1},{-1,0},{1,0}} {
            nx, ny := curr.x+d.dx, curr.y+d.dy
            if nx >= 0 && nx < size && ny >= 0 && ny < size {
                if !mask[ny][nx] && !visited[ny][nx] {
                    visited[ny][nx] = true
                    visitedCount++
                    queue = append(queue, cellPos{nx, ny})
                }
            }
        }
    }

    // Count total white cells
    whiteCount := 0
    for y := range size {
        for x := range size {
            if !mask[y][x] {
                whiteCount++
            }
        }
    }

    return visitedCount == whiteCount
}
```

---

## Step 3: Construct the Puzzle

Now combine the Latin Square and mask to create the puzzle grid:
- White cells: keep their original number
- Black cells: copy a number from another white cell in the same row OR column

```go
func constructPuzzle(baseGrid grid, mask [][]bool) grid {
    size := len(baseGrid)
    puzzle := make(grid, size)
    for y := range size {
        puzzle[y] = make([]rune, size)
        for x := range size {
            if mask[y][x] {
                // Black cell: introduce duplicate from same row or column
                // Choose random white cell in row or column
                whiteCellsInRow := findWhiteCellsInRow(mask, size, y, x)
                whiteCellsInCol := findWhiteCellsInCol(mask, size, y, x)
                
                allWhite := append(whiteCellsInRow, whiteCellsInCol...)
                if len(allWhite) == 0 {
                    panic("no white cells to copy from")
                }
                
                pick := allWhite[rand.IntN(len(allWhite))]
                puzzle[y][x] = baseGrid[pick.y][pick.x]
            } else {
                puzzle[y][x] = baseGrid[y][x]
            }
        }
    }
    return puzzle
}

func findWhiteCellsInRow(mask [][]bool, size, y, excludeX int) []cellPos {
    var result []cellPos
    for x := range size {
        if x != excludeX && !mask[y][x] {
            result = append(result, cellPos{x, y})
        }
    }
    return result
}

func findWhiteCellsInCol(mask [][]bool, size, x, excludeY int) []cellPos {
    var result []cellPos
    for y := range size {
        if y != excludeY && !mask[y][x] {
            result = append(result, cellPos{x, y})
        }
    }
    return result
}
```

---

## Step 4: Verify Uniqueness via Backtracking Solver

The puzzle may have multiple valid solutions. We must verify exactly one exists.

### Algorithm

```go
func countSolutions(puzzle grid, size, limit int) int {
    // Convert puzzle to mutable state: cell can be 'white' or 'black'
    // Start with all cells marked as unknown
    
    return countSolutionsRecursive(state, 0, limit)
}

func countSolutionsRecursive(state [][]cellState, pos int, limit int) int {
    if pos == size*size {
        if isValidComplete(state) {
            return 1
        }
        return 0
    }

    x, y := pos%size, pos/size
    
    // Skip cells that are already determined (if using progressive solving)
    // For pure backtracking, try both states:
    
    // Try marking as white
    if canBeWhite(state, size, x, y) {
        state[y][x] = white
        count := countSolutionsRecursive(state, pos+1, limit)
        state[y][x] = unknown
        if count >= limit {
            return count
        }
    }

    // Try marking as black
    if canBeBlack(state, size, x, y) {
        state[y][x] = black
        if countSolutionsRecursive(state, pos+1, limit) >= limit {
            return limit
        }
        state[y][x] = unknown
    }

    return 0
}
```

### Constraint Checking Functions

```go
func canBeWhite(state [][]cellState, size, x, y int) bool {
    // Get the number from puzzle at this position
    num := puzzle[y][x]
    
    // Check: if this cell is white, no duplicate in row
    for i := range size {
        if i != x && state[y][i] == white && puzzle[y][i] == num {
            return false
        }
    }
    // Check: if this cell is white, no duplicate in column
    for i := range size {
        if i != y && state[i][x] == white && puzzle[i][x] == num {
            return false
        }
    }
    return true
}

func canBeBlack(state [][]cellState, size, x, y int) bool {
    // Check: no adjacent black cells
    if x > 0 && state[y][x-1] == black { return false }
    if x < size-1 && state[y][x+1] == black { return false }
    if y > 0 && state[y-1][x] == black { return false }
    if y < size-1 && state[y+1][x] == black { return false }
    
    // Check: removing this cell doesn't isolate white cells
    // (This is expensive - use connectivity check)
    return whiteCellsConnectedAfter(state, size, x, y)
}
```

---

## Step 5: The Generation Loop

```go
func Generate(size int, difficulty float64) (grid, [][]bool, error) {
    blackRatio := 0.2 + (1-difficulty)*0.2  // 0.2 (easy) to 0.4 (hard)
    
    for attempt := 0; attempt < 100; attempt++ {
        // Phase 1: Generate candidate
        baseGrid := generateLatinSquare(size)
        mask := generateValidMask(size, blackRatio)
        puzzle := constructPuzzle(baseGrid, mask)
        
        // Phase 2: Verify uniqueness
        solutionCount := countSolutions(puzzle, size, 2)
        
        if solutionCount == 1 {
            // Convert mask to 'provided' format (true = given, false = player filled)
            provided := make([][]bool, size)
            for y := range size {
                provided[y] = make([]bool, size)
                for x := range size {
                    provided[y][x] = true  // All numbers are given
                }
            }
            return puzzle, provided, nil
        }
        // else: discard and retry
    }
    
    return nil, nil, errors.New("failed to generate puzzle after max attempts")
}
```

---

## Tuning Parameters

| Parameter | Range | Effect |
|-----------|-------|--------|
| `blackRatio` | 0.15-0.35 | Higher = more black cells = harder to deduce |
| `maxAttempts` | 100-1000 | Retry limit for generation loop |
| `size` | 4-12 | Puzzle dimensions |

---

## Common Issues & Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| Generator hangs | Mask creates impossible connectivity | Increase `maxAttempts`, lower `blackRatio` |
| Few solutions found | Too many black cells | Reduce `blackRatio` |
| Generation too slow | Connectivity check O(n²) per cell | Use Union-Find instead of BFS |

### Optimization: Union-Find for Connectivity

```go
type UnionFind struct {
    parent []int
    size   []int
}

func (uf *UnionFind) find(x int) int {
    if uf.parent[x] != x {
        uf.parent[x] = uf.find(uf.parent[x])
    }
    return uf.parent[x]
}

func (uf *UnionFind) union(x, y int) {
    px, py := uf.find(x), uf.find(y)
    if px == py {
        return
    }
    if uf.size[px] < uf.size[py] {
        px, py = py, px
    }
    uf.parent[py] = px
    uf.size[px] += uf.size[py]
}
```

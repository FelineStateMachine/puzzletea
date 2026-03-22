package netwalk

import (
	"fmt"
	"math/bits"
)

type directionMask uint8

const (
	north directionMask = 1 << iota
	east
	south
	west
)

const allDirections = north | east | south | west

type cellKind uint8

const (
	emptyCell cellKind = iota
	nodeCell
	serverCell
)

type point struct {
	X int
	Y int
}

type tile struct {
	BaseMask        directionMask
	Rotation        uint8
	InitialRotation uint8
	Locked          bool
	Kind            cellKind
}

type Puzzle struct {
	Size  int
	Root  point
	Tiles [][]tile
}

type boardState struct {
	nonEmpty        int
	connected       int
	dangling        int
	locked          int
	solved          bool
	allMatched      bool
	connectedToRoot [][]bool
	tileHasDangling [][]bool
	rotatedMasks    [][]directionMask
}

type directionSpec struct {
	dx  int
	dy  int
	bit directionMask
	opp directionMask
}

var directions = []directionSpec{
	{dx: 0, dy: -1, bit: north, opp: south},
	{dx: 1, dy: 0, bit: east, opp: west},
	{dx: 0, dy: 1, bit: south, opp: north},
	{dx: -1, dy: 0, bit: west, opp: east},
}

func newPuzzle(size int) Puzzle {
	tiles := make([][]tile, size)
	for y := range size {
		tiles[y] = make([]tile, size)
	}
	return Puzzle{Size: size, Tiles: tiles}
}

func (p Puzzle) inBounds(x, y int) bool {
	return x >= 0 && x < p.Size && y >= 0 && y < p.Size
}

func (p Puzzle) activeAt(x, y int) bool {
	return p.inBounds(x, y) && p.Tiles[y][x].Kind != emptyCell
}

func (p Puzzle) firstActive() point {
	if p.activeAt(p.Root.X, p.Root.Y) {
		return p.Root
	}
	for y := range p.Size {
		for x := range p.Size {
			if p.activeAt(x, y) {
				return point{X: x, Y: y}
			}
		}
	}
	return point{}
}

func isActive(t tile) bool {
	return t.Kind != emptyCell
}

func rotateMask(mask directionMask, rotation uint8) directionMask {
	shift := rotation % 4
	if shift == 0 {
		return mask
	}
	value := uint8(mask & allDirections)
	return directionMask(((value << shift) | (value >> (4 - shift))) & uint8(allDirections))
}

func degree(mask directionMask) int {
	return bits.OnesCount8(uint8(mask))
}

func uniqueRotations(mask directionMask) []uint8 {
	seen := make(map[directionMask]struct{}, 4)
	rotations := make([]uint8, 0, 4)
	for rot := uint8(0); rot < 4; rot++ {
		rotated := rotateMask(mask, rot)
		if _, ok := seen[rotated]; ok {
			continue
		}
		seen[rotated] = struct{}{}
		rotations = append(rotations, rot)
	}
	return rotations
}

func maskGlyph(mask directionMask) string {
	switch mask {
	case 0:
		return " "
	case north:
		return "╵"
	case east:
		return "╶"
	case south:
		return "╷"
	case west:
		return "╴"
	case north | south:
		return "│"
	case east | west:
		return "─"
	case north | east:
		return "└"
	case east | south:
		return "┌"
	case south | west:
		return "┐"
	case west | north:
		return "┘"
	case north | east | south:
		return "├"
	case east | south | west:
		return "┬"
	case south | west | north:
		return "┤"
	case west | north | east:
		return "┴"
	case north | east | south | west:
		return "┼"
	default:
		return "?"
	}
}

func encodeMaskRows(tiles [][]tile) string {
	return encodeRows(len(tiles), func(x, y int) byte {
		return nibbleHex(uint8(tiles[y][x].BaseMask))
	})
}

func encodeRotationRows(tiles [][]tile, initial bool) string {
	return encodeRows(len(tiles), func(x, y int) byte {
		value := tiles[y][x].Rotation
		if initial {
			value = tiles[y][x].InitialRotation
		}
		return byte('0' + value%4)
	})
}

func encodeKindRows(tiles [][]tile) string {
	return encodeRows(len(tiles), func(x, y int) byte {
		switch tiles[y][x].Kind {
		case serverCell:
			return 'S'
		case nodeCell:
			return '#'
		default:
			return '.'
		}
	})
}

func encodeLockRows(tiles [][]tile) string {
	return encodeRows(len(tiles), func(x, y int) byte {
		if tiles[y][x].Locked {
			return '#'
		}
		return '.'
	})
}

func decodePuzzle(size int, masks, rotations, initial, kinds, locks string) (Puzzle, error) {
	if size <= 0 {
		return Puzzle{}, fmt.Errorf("invalid netwalk size %d", size)
	}
	maskRows, err := decodeRows(size, masks)
	if err != nil {
		return Puzzle{}, fmt.Errorf("decode masks: %w", err)
	}
	rotationRows, err := decodeRows(size, rotations)
	if err != nil {
		return Puzzle{}, fmt.Errorf("decode rotations: %w", err)
	}
	initialRows, err := decodeRows(size, initial)
	if err != nil {
		return Puzzle{}, fmt.Errorf("decode initial rotations: %w", err)
	}
	kindRows, err := decodeRows(size, kinds)
	if err != nil {
		return Puzzle{}, fmt.Errorf("decode kinds: %w", err)
	}
	lockRows, err := decodeRows(size, locks)
	if err != nil {
		return Puzzle{}, fmt.Errorf("decode locks: %w", err)
	}

	puzzle := newPuzzle(size)
	rootCount := 0
	for y := range size {
		for x := range size {
			maskValue, ok := parseNibble(maskRows[y][x])
			if !ok {
				return Puzzle{}, fmt.Errorf("invalid mask value %q at (%d,%d)", maskRows[y][x], x, y)
			}
			rotationValue := rotationRows[y][x]
			initialValue := initialRows[y][x]
			if rotationValue < '0' || rotationValue > '3' || initialValue < '0' || initialValue > '3' {
				return Puzzle{}, fmt.Errorf("invalid rotation at (%d,%d)", x, y)
			}

			t := tile{
				BaseMask:        directionMask(maskValue),
				Rotation:        uint8(rotationValue - '0'),
				InitialRotation: uint8(initialValue - '0'),
			}

			switch kindRows[y][x] {
			case 'S':
				t.Kind = serverCell
				puzzle.Root = point{X: x, Y: y}
				rootCount++
			case '#':
				t.Kind = nodeCell
			case '.':
				t.Kind = emptyCell
			default:
				return Puzzle{}, fmt.Errorf("invalid cell kind %q at (%d,%d)", kindRows[y][x], x, y)
			}

			if t.Kind == emptyCell && t.BaseMask != 0 {
				return Puzzle{}, fmt.Errorf("empty tile at (%d,%d) has mask", x, y)
			}
			if t.Kind != emptyCell && t.BaseMask == 0 {
				return Puzzle{}, fmt.Errorf("active tile at (%d,%d) has empty mask", x, y)
			}
			if lockRows[y][x] == '#' {
				t.Locked = true
			}
			puzzle.Tiles[y][x] = t
		}
	}

	if rootCount != 1 {
		return Puzzle{}, fmt.Errorf("expected exactly one root, got %d", rootCount)
	}

	return puzzle, nil
}

func encodeRows(size int, valueAt func(x, y int) byte) string {
	buf := make([]byte, 0, size*size+max(size-1, 0))
	for y := range size {
		for x := range size {
			buf = append(buf, valueAt(x, y))
		}
		if y < size-1 {
			buf = append(buf, '\n')
		}
	}
	return string(buf)
}

func decodeRows(size int, raw string) ([][]byte, error) {
	rows := make([][]byte, size)
	currentRow := 0
	currentCol := 0
	rows[currentRow] = make([]byte, size)

	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch == '\r' {
			continue
		}
		if ch == '\n' {
			if currentCol != size {
				return nil, fmt.Errorf("row %d has width %d, want %d", currentRow, currentCol, size)
			}
			currentRow++
			if currentRow >= size {
				return nil, fmt.Errorf("too many rows")
			}
			rows[currentRow] = make([]byte, size)
			currentCol = 0
			continue
		}
		if currentCol >= size {
			return nil, fmt.Errorf("row %d exceeds width %d", currentRow, size)
		}
		rows[currentRow][currentCol] = ch
		currentCol++
	}

	if currentRow != size-1 || currentCol != size {
		return nil, fmt.Errorf("incomplete grid")
	}

	return rows, nil
}

func nibbleHex(value uint8) byte {
	if value < 10 {
		return '0' + value
	}
	return 'a' + (value - 10)
}

func parseNibble(value byte) (uint8, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return 10 + value - 'a', true
	case value >= 'A' && value <= 'F':
		return 10 + value - 'A', true
	default:
		return 0, false
	}
}

func analyzePuzzle(p Puzzle) boardState {
	state := boardState{
		connectedToRoot: make([][]bool, p.Size),
		tileHasDangling: make([][]bool, p.Size),
		rotatedMasks:    make([][]directionMask, p.Size),
	}
	for y := range p.Size {
		state.connectedToRoot[y] = make([]bool, p.Size)
		state.tileHasDangling[y] = make([]bool, p.Size)
		state.rotatedMasks[y] = make([]directionMask, p.Size)
		for x := range p.Size {
			t := p.Tiles[y][x]
			if !isActive(t) {
				continue
			}
			state.nonEmpty++
			if t.Locked {
				state.locked++
			}
			state.rotatedMasks[y][x] = rotateMask(t.BaseMask, t.Rotation)
		}
	}

	allMatched := true
	halfEdges := 0
	for y := range p.Size {
		for x := range p.Size {
			t := p.Tiles[y][x]
			if !isActive(t) {
				continue
			}
			mask := state.rotatedMasks[y][x]
			for _, dir := range directions {
				if mask&dir.bit == 0 {
					continue
				}
				halfEdges++
				nx := x + dir.dx
				ny := y + dir.dy
				if !p.activeAt(nx, ny) {
					state.tileHasDangling[y][x] = true
					state.dangling++
					allMatched = false
					continue
				}
				neighborMask := state.rotatedMasks[ny][nx]
				if neighborMask&dir.opp == 0 {
					state.tileHasDangling[y][x] = true
					state.dangling++
					allMatched = false
				}
			}
		}
	}

	state.allMatched = allMatched
	if !p.activeAt(p.Root.X, p.Root.Y) {
		return state
	}

	queue := []point{p.Root}
	state.connectedToRoot[p.Root.Y][p.Root.X] = true
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		state.connected++
		mask := state.rotatedMasks[cur.Y][cur.X]
		for _, dir := range directions {
			if mask&dir.bit == 0 {
				continue
			}
			nx := cur.X + dir.dx
			ny := cur.Y + dir.dy
			if !p.activeAt(nx, ny) || state.connectedToRoot[ny][nx] {
				continue
			}
			neighborMask := state.rotatedMasks[ny][nx]
			if neighborMask&dir.opp == 0 {
				continue
			}
			state.connectedToRoot[ny][nx] = true
			queue = append(queue, point{X: nx, Y: ny})
		}
	}

	matchedEdges := halfEdges / 2
	state.solved = state.nonEmpty > 0 &&
		allMatched &&
		state.connected == state.nonEmpty &&
		matchedEdges == state.nonEmpty-1

	return state
}

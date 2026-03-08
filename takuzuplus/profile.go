package takuzuplus

import (
	"image"
	"math"
	"math/rand/v2"
)

const (
	endpointZeroProvided = iota
	endpointOneProvided
	endpointTwoProvided
	endpointClassCount
)

type ratioBand struct {
	Min float64
	Max float64
}

type relationProfile struct {
	MinRelations int
	MaxRelations int

	EndpointBands   [endpointClassCount]ratioBand
	EndpointWeights [endpointClassCount]int

	MinSpacing int
	MinRegions int
}

type relationMetrics struct {
	Total           int
	EndpointCounts  [endpointClassCount]int
	SameCount       int
	DifferentCount  int
	HorizontalCount int
	VerticalCount   int
	RegionMask      uint8
	OccupiedRegions int
	Positions       []image.Point
	MinExistingGap  int
	HasAnyRelations bool
}

func (p relationProfile) targetCount(rng *rand.Rand) int {
	if p.MaxRelations <= p.MinRelations {
		return p.MinRelations
	}
	return p.MinRelations + rng.IntN(p.MaxRelations-p.MinRelations+1)
}

func relationProfiles() []relationProfile {
	return []relationProfile{
		{
			MinRelations: 2,
			MaxRelations: 4,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.00, Max: 0.05},
				{Min: 0.70, Max: 0.85},
				{Min: 0.15, Max: 0.30},
			},
			EndpointWeights: [endpointClassCount]int{0, 10, 5},
			MinSpacing:      2,
		},
		{
			MinRelations: 3,
			MaxRelations: 5,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.00, Max: 0.05},
				{Min: 0.65, Max: 0.80},
				{Min: 0.20, Max: 0.35},
			},
			EndpointWeights: [endpointClassCount]int{0, 9, 5},
			MinSpacing:      2,
		},
		{
			MinRelations: 4,
			MaxRelations: 6,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.05, Max: 0.15},
				{Min: 0.60, Max: 0.75},
				{Min: 0.15, Max: 0.30},
			},
			EndpointWeights: [endpointClassCount]int{3, 9, 4},
			MinRegions:      3,
		},
		{
			MinRelations: 6,
			MaxRelations: 8,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.10, Max: 0.15},
				{Min: 0.60, Max: 0.70},
				{Min: 0.15, Max: 0.25},
			},
			EndpointWeights: [endpointClassCount]int{4, 8, 3},
			MinRegions:      3,
		},
		{
			MinRelations: 8,
			MaxRelations: 11,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.10, Max: 0.20},
				{Min: 0.60, Max: 0.70},
				{Min: 0.10, Max: 0.25},
			},
			EndpointWeights: [endpointClassCount]int{5, 8, 2},
			MinRegions:      3,
		},
		{
			MinRelations: 10,
			MaxRelations: 14,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.15, Max: 0.20},
				{Min: 0.55, Max: 0.65},
				{Min: 0.15, Max: 0.30},
			},
			EndpointWeights: [endpointClassCount]int{5, 7, 3},
			MinRegions:      3,
		},
		{
			MinRelations: 13,
			MaxRelations: 18,
			EndpointBands: [endpointClassCount]ratioBand{
				{Min: 0.15, Max: 0.20},
				{Min: 0.55, Max: 0.65},
				{Min: 0.15, Max: 0.30},
			},
			EndpointWeights: [endpointClassCount]int{5, 7, 3},
			MinRegions:      3,
		},
	}
}

func relationPenalty(profile relationProfile, size int, rels relations, provided [][]bool) int {
	metrics := analyzeRelations(rels, provided, size)
	penalty := 0

	switch {
	case metrics.Total < profile.MinRelations:
		penalty += 1000 * (profile.MinRelations - metrics.Total)
	case metrics.Total > profile.MaxRelations:
		penalty += 1000 * (metrics.Total - profile.MaxRelations)
	}

	if metrics.Total == 0 {
		return penalty
	}

	for class, band := range profile.EndpointBands {
		if !ratioWithinBand(metrics.EndpointCounts[class], metrics.Total, band) {
			penalty += bandPenalty(metrics.EndpointCounts[class], metrics.Total, band)
		}
	}

	if metrics.Total >= 5 {
		balanceBand := ratioBand{Min: 0.40, Max: 0.60}
		if !ratioWithinBand(metrics.SameCount, metrics.Total, balanceBand) {
			penalty += bandPenalty(metrics.SameCount, metrics.Total, balanceBand)
		}
		if !ratioWithinBand(metrics.HorizontalCount, metrics.Total, balanceBand) {
			penalty += bandPenalty(metrics.HorizontalCount, metrics.Total, balanceBand)
		}
	}

	if profile.MinSpacing > 0 && metrics.Total > 1 && metrics.MinExistingGap < profile.MinSpacing {
		penalty += 250 * (profile.MinSpacing - metrics.MinExistingGap)
	}

	if profile.MinRegions > 0 && metrics.OccupiedRegions < profile.MinRegions {
		penalty += 400 * (profile.MinRegions - metrics.OccupiedRegions)
	}

	return penalty
}

func analyzeRelations(rels relations, provided [][]bool, size int) relationMetrics {
	metrics := relationMetrics{
		MinExistingGap: math.MaxInt,
	}

	appendPosition := func(pos image.Point) {
		for _, prior := range metrics.Positions {
			dist := abs(pos.X-prior.X) + abs(pos.Y-prior.Y)
			if dist < metrics.MinExistingGap {
				metrics.MinExistingGap = dist
			}
		}
		metrics.Positions = append(metrics.Positions, pos)
		metrics.RegionMask |= relationRegionBit(size, pos)
	}

	for y, row := range rels.horizontal {
		for x, rel := range row {
			if rel == relationNone {
				continue
			}
			metrics.Total++
			metrics.HasAnyRelations = true
			metrics.HorizontalCount++
			if rel == relationSame {
				metrics.SameCount++
			} else {
				metrics.DifferentCount++
			}
			class := endpointClass(boolInt(provided[y][x]) + boolInt(provided[y][x+1]))
			metrics.EndpointCounts[class]++
			appendPosition(horizontalRelationPosition(x, y))
		}
	}

	for y, row := range rels.vertical {
		for x, rel := range row {
			if rel == relationNone {
				continue
			}
			metrics.Total++
			metrics.HasAnyRelations = true
			metrics.VerticalCount++
			if rel == relationSame {
				metrics.SameCount++
			} else {
				metrics.DifferentCount++
			}
			class := endpointClass(boolInt(provided[y][x]) + boolInt(provided[y+1][x]))
			metrics.EndpointCounts[class]++
			appendPosition(verticalRelationPosition(x, y))
		}
	}

	if metrics.MinExistingGap == math.MaxInt {
		metrics.MinExistingGap = profileDistanceUnset()
	}
	metrics.OccupiedRegions = bitsSet(metrics.RegionMask)

	return metrics
}

func ratioWithinBand(count, total int, band ratioBand) bool {
	if total == 0 {
		return band.Min <= 0
	}
	actual := float64(count) / float64(total)
	slack := 1.0 / float64(total)
	return actual >= maxFloat(0, band.Min-slack) && actual <= minFloat(1, band.Max+slack)
}

func bandPenalty(count, total int, band ratioBand) int {
	if total == 0 {
		return 0
	}
	actual := float64(count) / float64(total)
	switch {
	case actual < band.Min:
		return int(math.Ceil((band.Min - actual) * 1000))
	case actual > band.Max:
		return int(math.Ceil((actual - band.Max) * 1000))
	default:
		return 0
	}
}

func relationRegionBit(size int, pos image.Point) uint8 {
	if size <= 0 {
		return 0
	}
	cut := size
	regionX := 0
	regionY := 0
	if pos.X >= cut {
		regionX = 1
	}
	if pos.Y >= cut {
		regionY = 1
	}
	return 1 << uint(regionY*2+regionX)
}

func horizontalRelationPosition(x, y int) image.Point {
	return image.Pt(2*x+1, 2*y)
}

func verticalRelationPosition(x, y int) image.Point {
	return image.Pt(2*x, 2*y+1)
}

func endpointClass(providedCount int) int {
	switch providedCount {
	case 0:
		return endpointZeroProvided
	case 1:
		return endpointOneProvided
	default:
		return endpointTwoProvided
	}
}

func bitsSet(mask uint8) int {
	count := 0
	for mask > 0 {
		count += int(mask & 1)
		mask >>= 1
	}
	return count
}

func profileDistanceUnset() int {
	return math.MaxInt / 2
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

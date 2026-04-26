// Package difficulty defines PuzzleTea's shared Elo difficulty contract.
package difficulty

import "fmt"

type Elo int

const (
	MinElo     Elo = 0
	SoftCapElo Elo = 3000
)

type Confidence string

const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)

type Metrics map[string]float64

type Report struct {
	TargetElo  Elo
	ActualElo  Elo
	Confidence Confidence
	Metrics    Metrics
}

func ValidateElo(elo Elo) error {
	if elo < MinElo || elo > SoftCapElo {
		return fmt.Errorf("difficulty elo %d outside supported range %d..%d", elo, MinElo, SoftCapElo)
	}
	return nil
}

func ClampElo(elo Elo) Elo {
	if elo < MinElo {
		return MinElo
	}
	if elo > SoftCapElo {
		return SoftCapElo
	}
	return elo
}

func Score01(elo Elo) float64 {
	elo = ClampElo(elo)
	return float64(elo-MinElo) / float64(SoftCapElo-MinElo)
}

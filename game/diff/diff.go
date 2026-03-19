package diff

import (
	"github.com/lex00/sector-zero/game/scope"
)

// Result holds the outcome of comparing a player's trace to the target.
type Result struct {
	Score    float64 // 0.0-1.0
	Category string  // "empty"|"wrong_types"|"wrong_order"|"wrong_termination"|"near"|"exact"
	HintKey  string  // key for dialogue lookup
}

// algorithmicTypes are the pulse types that represent actual algorithmic
// operations. Decorative/visualisation pulses (pin, signal, bounds) are
// excluded so that templates which omit them can still score exact.
var algorithmicTypes = map[string]bool{
	"init":      true,
	"compare":   true,
	"swap":      true,
	"access":    true,
	"found":     true,
	"not_found": true,
	"done":      true,
	"split":     true,
	"merge":     true,
}

func filterAlgorithmic(pulses []scope.Pulse) []scope.Pulse {
	out := make([]scope.Pulse, 0, len(pulses))
	for _, p := range pulses {
		if algorithmicTypes[p.Type] {
			out = append(out, p)
		}
	}
	return out
}

// Diff compares a player trace to the target trace and returns a Result.
func Diff(player []scope.Pulse, target []scope.Pulse) Result {
	// Strip decorative pulses before scoring.
	player = filterAlgorithmic(player)
	target = filterAlgorithmic(target)

	if len(player) == 0 {
		return Result{
			Score:    0,
			Category: "empty",
			HintKey:  "empty",
		}
	}

	// Check type sequence match.
	typeScore := typeSequenceScore(player, target)

	// Check positional match (i, j, pos).
	posScore := positionalScore(player, target)

	// Check termination (does player have a "done" pulse?).
	hasDone := hasDonePulse(player)
	targetHasDone := hasDonePulse(target)

	// Weighted score.
	score := typeScore*0.6 + posScore*0.4

	var category string

	switch {
	case score == 1.0 && hasDone == targetHasDone:
		category = "exact"
	case score >= 0.90:
		if targetHasDone && !hasDone {
			category = "wrong_termination"
		} else {
			category = "near"
		}
	case score >= 0.60:
		category = "wrong_order"
	case score >= 0.30:
		category = "wrong_types"
	default:
		category = "wrong_types"
	}

	return Result{
		Score:    score,
		Category: category,
		HintKey:  category,
	}
}

// typeSequenceScore compares the sequence of pulse types.
func typeSequenceScore(player, target []scope.Pulse) float64 {
	if len(target) == 0 {
		return 0
	}

	// LCS of type sequences.
	pTypes := make([]string, len(player))
	tTypes := make([]string, len(target))
	for i, p := range player {
		pTypes[i] = p.Type
	}
	for i, p := range target {
		tTypes[i] = p.Type
	}

	lcsLen := lcs(pTypes, tTypes)
	return float64(lcsLen) / float64(len(target))
}

// positionalScore compares i/j/pos values where types match.
func positionalScore(player, target []scope.Pulse) float64 {
	if len(target) == 0 {
		return 0
	}

	matchCount := 0
	total := 0

	// Walk both sequences in lock-step where types match.
	pi := 0
	for ti := 0; ti < len(target) && pi < len(player); ti++ {
		// Find next matching type in player.
		for pi < len(player) && player[pi].Type != target[ti].Type {
			pi++
		}
		if pi >= len(player) {
			break
		}

		tp := target[ti]
		pp := player[pi]

		if tp.Type == pp.Type {
			total++
			if pulsePosEqual(pp, tp) {
				matchCount++
			}
			pi++
		}
	}

	if total == 0 {
		return 0
	}
	return float64(matchCount) / float64(total)
}

// pulsePosEqual returns true if the positional fields of two pulses match.
func pulsePosEqual(a, b scope.Pulse) bool {
	switch a.Type {
	case "compare", "swap":
		return a.I == b.I && a.J == b.J
	case "pin", "access", "found":
		return a.Pos == b.Pos
	case "signal":
		return a.Name == b.Name && intSliceEqual(a.Positions, b.Positions)
	case "bounds":
		return a.Low == b.Low && a.High == b.High
	case "split", "merge":
		return a.Left == b.Left && a.Mid == b.Mid && a.Right == b.Right
	default:
		return true
	}
}

// hasDonePulse returns true if the slice contains a "done" pulse.
func hasDonePulse(pulses []scope.Pulse) bool {
	for _, p := range pulses {
		if p.Type == "done" {
			return true
		}
	}
	return false
}

// lcs computes the length of the longest common subsequence of two string slices.
func lcs(a, b []string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	// Use two-row DP to save memory.
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else if prev[j] > curr[j-1] {
				curr[j] = prev[j]
			} else {
				curr[j] = curr[j-1]
			}
		}
		prev, curr = curr, prev
		for j := range curr {
			curr[j] = 0
		}
	}
	return prev[len(b)]
}

// intSliceEqual compares two int slices for equality.
func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

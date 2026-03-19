package puzzles

import (
	"fmt"
	"sort"
)

var registered []Puzzle

// Register adds a Puzzle to the registry. Called from each puzzle file's init().
// Panics if the puzzle fails validation — fail-fast at startup.
func Register(p Puzzle) {
	if err := p.Validate(); err != nil {
		panic(fmt.Sprintf("puzzles.Register: %v", err))
	}
	registered = append(registered, p)
}

// Get returns the Puzzle with the given 1-based ID, or registered[0].
func Get(id int) Puzzle {
	for _, p := range registered {
		if p.ID == id {
			return p
		}
	}
	if len(registered) > 0 {
		return registered[0]
	}
	return Puzzle{}
}

// all returns all registered puzzles sorted by ID.
func all() []Puzzle {
	out := make([]Puzzle, len(registered))
	copy(out, registered)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

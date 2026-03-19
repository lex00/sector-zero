package puzzles

import "embed"

//go:embed data
var puzzleData embed.FS

// GetPuzzle returns the puzzle with the given 1-based ID.
// Returns the first puzzle if id is out of range.
func GetPuzzle(id int) Puzzle {
	return Get(id)
}

// All returns all puzzles sorted by ID.
func All() []Puzzle {
	return all()
}

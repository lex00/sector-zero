package puzzles

import (
	"embed"
	"fmt"

	"github.com/lex00/sector-zero/game/scope"
)

//go:embed data
var puzzleData embed.FS

// Puzzle holds the metadata and content for a single game puzzle.
type Puzzle struct {
	ID           int
	Title        string
	TraceFile    string                       // path relative to puzzles/data/
	Dialogue     map[string]map[string]string // help_level → hint_key → text
	ScaffoldCode map[string]string            // lang → scaffold code for OPEN level
	Challenge    ChallengeSpec
	Script       []LessonStep
}

// GetPuzzle returns the puzzle with the given 1-based ID.
// Returns the first puzzle if id is out of range.
func GetPuzzle(id int) Puzzle {
	return specToPuzzle(GetSpec(id))
}

// All returns all puzzles.
func All() []Puzzle {
	specs := AllSpecs()
	out := make([]Puzzle, len(specs))
	for i, s := range specs {
		out[i] = specToPuzzle(s)
	}
	return out
}

// LoadTrace reads and parses the embedded trace file for the given puzzle.
func LoadTrace(pz Puzzle) ([]scope.Pulse, error) {
	data, err := puzzleData.ReadFile(pz.TraceFile)
	if err != nil {
		return nil, fmt.Errorf("load trace %q: %w", pz.TraceFile, err)
	}
	return scope.ParseTrace(data)
}

// GetDialogue returns the dialogue text for a given help level and hint key.
// Falls back through levels: OPEN → SIGNAL → STATIC → BLACKOUT.
func GetDialogue(pz Puzzle, helpLevel, hintKey string) string {
	if texts, ok := pz.Dialogue[helpLevel]; ok {
		if s, ok := texts[hintKey]; ok {
			return s
		}
	}
	// Fallback chain.
	fallbacks := []string{"SIGNAL", "STATIC", "BLACKOUT"}
	for _, lvl := range fallbacks {
		if lvl == helpLevel {
			continue
		}
		if texts, ok := pz.Dialogue[lvl]; ok {
			if s, ok := texts[hintKey]; ok {
				return s
			}
		}
	}
	return "*...*"
}

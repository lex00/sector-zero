package puzzles

import "sort"

var allSpecs []PuzzleSpec

// Register adds a PuzzleSpec to the registry. Called from each puzzle file's init().
func Register(s PuzzleSpec) {
	allSpecs = append(allSpecs, s)
}

// GetSpec returns the PuzzleSpec with the given 1-based ID, or allSpecs[0].
func GetSpec(id int) PuzzleSpec {
	for _, s := range allSpecs {
		if s.ID == id {
			return s
		}
	}
	if len(allSpecs) > 0 {
		return allSpecs[0]
	}
	return PuzzleSpec{}
}

// AllSpecs returns all registered specs sorted by ID.
func AllSpecs() []PuzzleSpec {
	out := make([]PuzzleSpec, len(allSpecs))
	copy(out, allSpecs)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// specToPuzzle converts a PuzzleSpec to the legacy Puzzle type used by the scene layer.
func specToPuzzle(s PuzzleSpec) Puzzle {
	dialogue := make(map[string]map[string]string)
	add := func(level string, h HintSet) {
		if !h.isEmpty() {
			dialogue[level] = h.toMap()
		}
	}
	add("BLACKOUT", s.Dialogue.Blackout)
	add("STATIC", s.Dialogue.Static)
	add("SIGNAL", s.Dialogue.Signal)
	add("OPEN", s.Dialogue.Open)

	// Build ScaffoldCode from the template (kept raw with {N} markers for the choice panel;
	// also usable as a read-only reference in non-OPEN modes).
	scaffold := make(map[string]string)
	for lang, tmpl := range s.Challenge.Template {
		scaffold[lang] = tmpl
	}

	return Puzzle{
		ID:           s.ID,
		Title:        s.Title,
		TraceFile:    s.TraceFile,
		Dialogue:     dialogue,
		ScaffoldCode: scaffold,
		Challenge:    s.Challenge,
		Script:       s.Script,
	}
}

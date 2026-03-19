package puzzles

import (
	"fmt"

	"github.com/lex00/sector-zero/game/scope"
)

// Puzzle is the strongly-typed plugin definition for one puzzle.
// All fields are required — missing fields won't compile.
type Puzzle struct {
	ID        int
	Title     string
	TraceFile string // relative to puzzles/data/
	Dialogue  DialogueSpec
	Challenge ChallengeSpec
	Script    []LessonStep
}

// Validate checks that all required fields are populated.
// Called by Register — panics at startup if a puzzle is malformed.
func (p Puzzle) Validate() error {
	if p.ID <= 0 {
		return fmt.Errorf("puzzle ID must be > 0, got %d", p.ID)
	}
	if p.Title == "" {
		return fmt.Errorf("puzzle %d: Title is empty", p.ID)
	}
	if p.TraceFile == "" {
		return fmt.Errorf("puzzle %d: TraceFile is empty", p.ID)
	}
	if len(p.Challenge.Template) == 0 {
		return fmt.Errorf("puzzle %d: Challenge.Template is empty", p.ID)
	}
	if len(p.Challenge.Blanks) == 0 {
		return fmt.Errorf("puzzle %d: Challenge.Blanks is empty", p.ID)
	}
	return nil
}

// LoadTrace reads and parses the embedded trace file for this puzzle.
func (p Puzzle) LoadTrace() ([]scope.Pulse, error) {
	data, err := puzzleData.ReadFile(p.TraceFile)
	if err != nil {
		return nil, fmt.Errorf("load trace %q: %w", p.TraceFile, err)
	}
	return scope.ParseTrace(data)
}

// GetDialogue returns the dialogue text for a given help level and hint key.
// Falls back through levels: try exact level first, then SIGNAL→STATIC→BLACKOUT
// (OPEN is not included in the fallback chain).
func (p Puzzle) GetDialogue(helpLevel, hintKey string) string {
	levels := map[string]HintSet{
		"BLACKOUT": p.Dialogue.Blackout,
		"STATIC":   p.Dialogue.Static,
		"SIGNAL":   p.Dialogue.Signal,
		"OPEN":     p.Dialogue.Open,
	}
	if h, ok := levels[helpLevel]; ok {
		if s := pick(h, hintKey); s != "" {
			return s
		}
	}
	for _, lvl := range []string{"SIGNAL", "STATIC", "BLACKOUT"} {
		if lvl == helpLevel {
			continue
		}
		if s := pick(levels[lvl], hintKey); s != "" {
			return s
		}
	}
	return "*...*"
}

// pick looks up a hint key in a HintSet and returns the value if non-empty.
func pick(h HintSet, key string) string {
	return h.toMap()[key]
}

// DialogueSpec holds hints at all four difficulty levels.
type DialogueSpec struct {
	Blackout HintSet
	Static   HintSet
	Signal   HintSet
	Open     HintSet
}

// HintSet is the set of feedback messages for one difficulty level.
type HintSet struct {
	Empty            string
	WrongTypes       string
	WrongOrder       string
	WrongTermination string
	Near             string
	Exact            string
}

// isEmpty returns true when every field is the zero string.
func (h HintSet) isEmpty() bool {
	return h.Empty == "" && h.WrongTypes == "" && h.WrongOrder == "" &&
		h.WrongTermination == "" && h.Near == "" && h.Exact == ""
}

// toMap converts a HintSet to a map[string]string for key-based lookup.
func (h HintSet) toMap() map[string]string {
	return map[string]string{
		"empty":             h.Empty,
		"wrong_types":       h.WrongTypes,
		"wrong_order":       h.WrongOrder,
		"wrong_termination": h.WrongTermination,
		"near":              h.Near,
		"exact":             h.Exact,
	}
}

// ChallengeSpec defines the fill-in-the-blank coding challenge.
type ChallengeSpec struct {
	Blanks   []Blank
	Template map[string]string // lang → template string with {0}, {1}... markers
}

// LessonStep is one scripted moment in a puzzle's guided experience.
//
// Matching: for each event, the step with the highest After that is still
// <= runAttempts wins. This lets later steps override earlier ones as the
// player makes more attempts.
//
// Trigger values (can combine with "+" e.g. "scope_ref+guide_next"):
//
//	guide_next    — reveal the next incorrect blank and switch to OPEN mode
//	guide_all     — reveal all blanks
//	scope_ref     — switch scope back to the reference trace
//	scope_player  — switch scope to the player's last trace
//	scope_pause   — pause the scope animation
//	scope_play    — resume the scope animation
//	mode_open     — switch help level to OPEN
type LessonStep struct {
	On      string // event: "load" | "run" | "error"
	Result  string // diff result filter, empty = any: "empty"|"wrong_types"|"wrong_order"|"wrong_termination"|"near"|"exact"
	After   int    // activates only when total run attempts >= After (0 = always)
	Message string // popup message; empty = use DialogueSpec fallback
	Style   string // popup style override: "run"|"guide"|"error"|"solved"
	Trigger string // side-effect (see above); multiple joined with "+"
}

// Blank is one fill-in-the-blank slot with multiple choice options.
type Blank struct {
	Label       string   // shown in the choice picker
	Choices     []string // display text for each option
	Correct     int      // index of the correct answer
	Explanation string   // shown in dialogue when guided solve reveals this blank
}

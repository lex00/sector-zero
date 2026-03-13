package opossum

// State represents the opossum's current emotional/physical state.
type State int

const (
	Idle       State = iota
	Curious          // heat 30-60%
	Uneasy           // heat 60-85%
	Afraid           // heat 85-99%
	Dead             // BOOM
	Recovering       // after reboot
	Awed             // on solve
)

// Model is the opossum state machine.
type Model struct {
	State State
}

// New creates a new opossum starting in Idle state.
func New() Model {
	return Model{State: Idle}
}

// UpdateFromHeat transitions the opossum state based on heat level (0.0-1.0).
func (m *Model) UpdateFromHeat(heat float64) {
	// Don't override special event states with heat transitions.
	switch m.State {
	case Dead, Recovering, Awed:
		return
	}

	switch {
	case heat >= 1.0:
		m.State = Dead
	case heat >= 0.85:
		m.State = Afraid
	case heat >= 0.60:
		m.State = Uneasy
	case heat >= 0.30:
		m.State = Curious
	default:
		m.State = Idle
	}
}

// StateFromSolve transitions the opossum to the Awed state (puzzle solved).
func (m *Model) StateFromSolve() {
	m.State = Awed
}

// StateFromBoom transitions the opossum to the Dead state (thermal overload).
func (m *Model) StateFromBoom() {
	m.State = Dead
}

// StateFromReboot transitions the opossum to the Recovering state.
func (m *Model) StateFromReboot() {
	m.State = Recovering
}

// Reaction returns the stage-direction text for the current state.
func (m Model) Reaction() string {
	switch m.State {
	case Idle:
		return ""
	case Curious:
		return "*ears perk*"
	case Uneasy:
		return "*shifts weight from foot to foot*"
	case Afraid:
		return "*backs toward the tunnel entrance*"
	case Dead:
		return "*plays dead*"
	case Recovering:
		return "*slowly rights himself*"
	case Awed:
		return "*stares into the light*"
	default:
		return ""
	}
}

// Name returns a display label for the current state.
func (m Model) Name() string {
	switch m.State {
	case Idle:
		return "Idle"
	case Curious:
		return "Curious"
	case Uneasy:
		return "Uneasy"
	case Afraid:
		return "Afraid"
	case Dead:
		return "Dead"
	case Recovering:
		return "Recovering"
	case Awed:
		return "Awed"
	default:
		return "Unknown"
	}
}

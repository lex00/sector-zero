package scope

import (
	"bufio"
	"bytes"
	"encoding/json"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Pulse represents a single event in an algorithm trace.
type Pulse struct {
	V         int    `json:"v"`
	Type      string `json:"type"`
	Net       string `json:"net"`
	Values    []int  `json:"values,omitempty"`
	I         int    `json:"i,omitempty"`
	J         int    `json:"j,omitempty"`
	Name      string `json:"name,omitempty"`
	Pos       int    `json:"pos,omitempty"`
	Positions []int  `json:"positions,omitempty"`
	Low       int    `json:"low,omitempty"`
	High      int    `json:"high,omitempty"`
	Mid       int    `json:"mid,omitempty"`
	Left      int    `json:"left,omitempty"`
	Right     int    `json:"right,omitempty"`
	Value     int    `json:"value,omitempty"`
}

// ParseTrace parses an NDJSON trace file into a slice of Pulses.
func ParseTrace(data []byte) ([]Pulse, error) {
	var pulses []Pulse
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var p Pulse
		if err := json.Unmarshal(line, &p); err != nil {
			continue
		}
		pulses = append(pulses, p)
	}
	return pulses, scanner.Err()
}

// NetState holds the current values for a named array net.
type NetState struct {
	Values     []int
	Normalised []float64
}

// TickMsg is sent by the animation ticker.
type TickMsg time.Time

// VictoryTickMsg is sent by the faster victory animation ticker.
type VictoryTickMsg time.Time

// VictoryTick returns a fast tick command for the victory animation.
func VictoryTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return VictoryTickMsg(t)
	})
}

// Model manages animated trace playback in the scope panel.
type Model struct {
	Trace       []Pulse
	TargetTrace []Pulse

	CurrentPulse int
	Paused       bool

	Width  int
	Height int

	netState map[string]*NetState
	pins     map[string]map[string]int
	signals  map[string]map[string][]int
	bounds   map[string][2]int // low, high per net
	accesses map[string]int    // net → flashed position (-1 if none)

	HeatCorrupt  float64 // 0-1, causes border corruption
	BrailleNoise float64 // noise bleeds in at hot stage

	LoopCount int // incremented each time the trace restarts from the beginning

	Victory      bool // true after puzzle is solved — plays victory animation
	VictoryFrame int  // current frame of the victory animation
}

// New returns an initialised scope Model.
func New(width, height int) Model {
	return Model{
		Width:    width,
		Height:   height,
		netState: make(map[string]*NetState),
		pins:     make(map[string]map[string]int),
		signals:  make(map[string]map[string][]int),
		bounds:   make(map[string][2]int),
		accesses: make(map[string]int),
	}
}

// SetTrace loads a new trace and resets playback state.
func (m *Model) SetTrace(pulses []Pulse) {
	m.Trace = pulses
	m.CurrentPulse = 0
	m.netState = make(map[string]*NetState)
	m.pins = make(map[string]map[string]int)
	m.signals = make(map[string]map[string][]int)
	m.bounds = make(map[string][2]int)
	m.accesses = make(map[string]int)
}

// Tick advances the animation by one pulse.
func Tick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init returns the initial command (start ticker).
func (m Model) Init() (Model, tea.Cmd) {
	return m, Tick()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if !m.Paused && len(m.Trace) > 0 {
			if m.CurrentPulse >= len(m.Trace) {
				// Loop: clear state, restart, and count the loop.
				m.CurrentPulse = 0
				m.LoopCount++
				m.netState = make(map[string]*NetState)
				m.pins = make(map[string]map[string]int)
				m.signals = make(map[string]map[string][]int)
				m.bounds = make(map[string][2]int)
				m.accesses = make(map[string]int)
			}
			m.applyPulse(m.Trace[m.CurrentPulse])
			m.CurrentPulse++
		}
		return m, Tick()

	case VictoryTickMsg:
		m.VictoryFrame++
		return m, VictoryTick()

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m *Model) applyPulse(p Pulse) {
	net := p.Net
	switch p.Type {
	case "init":
		ns := &NetState{Values: make([]int, len(p.Values))}
		copy(ns.Values, p.Values)
		ns.Normalised = normalise(ns.Values)
		m.netState[net] = ns
		// Clear per-net ancillary state.
		delete(m.pins, net)
		delete(m.signals, net)

	case "compare":
		if m.signals[net] == nil {
			m.signals[net] = make(map[string][]int)
		}
		// Clear any lingering swap highlight, set compare.
		delete(m.signals[net], "swap")
		m.signals[net]["compare"] = []int{p.I, p.J}

	case "swap":
		if ns, ok := m.netState[net]; ok {
			if p.I < len(ns.Values) && p.J < len(ns.Values) {
				ns.Values[p.I], ns.Values[p.J] = ns.Values[p.J], ns.Values[p.I]
				ns.Normalised = normalise(ns.Values)
			}
		}
		if m.signals[net] == nil {
			m.signals[net] = make(map[string][]int)
		}
		// Replace compare highlight with swap highlight so bars show orange for one tick.
		delete(m.signals[net], "compare")
		m.signals[net]["swap"] = []int{p.I, p.J}

	case "pin":
		if m.pins[net] == nil {
			m.pins[net] = make(map[string]int)
		}
		m.pins[net][p.Name] = p.Pos

	case "signal":
		if m.signals[net] == nil {
			m.signals[net] = make(map[string][]int)
		}
		m.signals[net][p.Name] = p.Positions

	case "access":
		m.accesses[net] = p.Pos

	case "found":
		if m.signals[net] == nil {
			m.signals[net] = make(map[string][]int)
		}
		m.signals[net]["found"] = []int{p.Pos}

	case "not_found":
		if m.signals[net] != nil {
			delete(m.signals[net], "found")
		}

	case "bounds":
		m.bounds[net] = [2]int{p.Low, p.High}

	case "split", "merge":
		if m.signals[net] == nil {
			m.signals[net] = make(map[string][]int)
		}
		m.signals[net]["left"] = []int{p.Left}
		m.signals[net]["mid"] = []int{p.Mid}
		m.signals[net]["right"] = []int{p.Right}

	case "write":
		if ns, ok := m.netState[net]; ok {
			if p.Pos >= 0 && p.Pos < len(ns.Values) {
				ns.Values[p.Pos] = p.Value
				ns.Normalised = normalise(ns.Values)
			}
		}
		m.accesses[net] = p.Pos

	case "done":
		if m.signals[net] != nil {
			m.signals[net] = make(map[string][]int)
		}
		if m.pins[net] != nil {
			m.pins[net] = make(map[string]int)
		}
	}
}

// normalise converts int values to float64 in [0, 1].
func normalise(values []int) []float64 {
	if len(values) == 0 {
		return nil
	}
	maxV := values[0]
	for _, v := range values {
		if v > maxV {
			maxV = v
		}
	}
	if maxV == 0 {
		maxV = 1
	}
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = float64(v) / float64(maxV)
	}
	return out
}

package scope

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

// ParseTrace parses an NDJSON trace file into a slice of Pulses.
func ParseTrace(data []byte) ([]Pulse, error) {
	var pulses []Pulse
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var p Pulse
		if err := json.Unmarshal(line, &p); err != nil {
			return nil, fmt.Errorf("parse pulse: %w", err)
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
		if !m.Paused && m.CurrentPulse < len(m.Trace) {
			m.applyPulse(m.Trace[m.CurrentPulse])
			m.CurrentPulse++
		}
		return m, Tick()

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
		if ns, ok := m.netState[net]; ok {
			_ = ns
		}
		if m.signals[net] == nil {
			m.signals[net] = make(map[string][]int)
		}
		m.signals[net]["compare"] = []int{p.I, p.J}

	case "swap":
		if ns, ok := m.netState[net]; ok {
			if p.I < len(ns.Values) && p.J < len(ns.Values) {
				ns.Values[p.I], ns.Values[p.J] = ns.Values[p.J], ns.Values[p.I]
				ns.Normalised = normalise(ns.Values)
			}
		}
		// Clear compare highlight after swap.
		if m.signals[net] != nil {
			delete(m.signals[net], "compare")
		}

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

	case "done":
		if m.signals[net] != nil {
			m.signals[net] = make(map[string][]int)
		}
		if m.pins[net] != nil {
			m.pins[net] = make(map[string]int)
		}
	}
}

// View renders the scope panel as a styled string.
func (m Model) View() string {
	panelW := m.Width
	panelH := m.Height
	if panelW <= 0 {
		panelW = 40
	}
	if panelH <= 0 {
		panelH = 20
	}

	innerW := panelW - 2 // subtract border chars
	innerH := panelH - 2
	if innerW <= 0 {
		innerW = 1
	}
	if innerH <= 0 {
		innerH = 1
	}

	var body strings.Builder

	// Collect the first net's state for rendering.
	var netName string
	var ns *NetState
	for k, v := range m.netState {
		netName = k
		ns = v
		break
	}

	var brailleRows []string
	if ns != nil {
		highlights := make(map[int]string)
		if sigs, ok := m.signals[netName]; ok {
			for sigName, positions := range sigs {
				for _, pos := range positions {
					highlights[pos] = sigName
				}
			}
		}
		brailleRows = RenderBars(ns.Normalised, highlights, innerW, innerH-2)
	} else {
		// Empty state - show blank braille.
		brailleRows = make([]string, innerH-2)
		for i := range brailleRows {
			brailleRows[i] = strings.Repeat(string(brailleChar(0)), innerW)
		}
	}

	// Add braille noise at high heat.
	if m.BrailleNoise > 0 {
		brailleRows = addNoise(brailleRows, m.BrailleNoise)
	}

	// Render each braille row.
	for _, row := range brailleRows {
		body.WriteString(row)
		body.WriteRune('\n')
	}

	// Pin labels row.
	pinRow := renderPins(m.pins[netName], ns, innerW)
	body.WriteString(pinRow)
	body.WriteRune('\n')

	// Pulse counter.
	total := len(m.Trace)
	counter := fmt.Sprintf(" pulse %d / %d ", m.CurrentPulse, total)
	if len(counter) > innerW {
		counter = counter[:innerW]
	}
	body.WriteString(counter)

	// Border style - corrupts at high heat.
	borderStyle := lipgloss.RoundedBorder()
	borderColour := lipgloss.Color("10") // green
	switch {
	case m.HeatCorrupt >= 0.85:
		borderColour = lipgloss.Color("9") // red
	case m.HeatCorrupt >= 0.60:
		borderColour = lipgloss.Color("208") // orange
	case m.HeatCorrupt >= 0.30:
		borderColour = lipgloss.Color("11") // yellow
	}

	boxStyle := lipgloss.NewStyle().
		Border(borderStyle).
		BorderForeground(borderColour).
		Width(innerW).
		Height(innerH)

	content := body.String()

	// Corrupt border chars at high heat.
	rendered := boxStyle.Render(content)
	if m.HeatCorrupt >= 0.60 {
		rendered = corruptBorder(rendered, m.HeatCorrupt)
	}

	return rendered
}

// renderPins builds a pin-label row beneath the braille chart.
func renderPins(pins map[string]int, ns *NetState, width int) string {
	if ns == nil || len(pins) == 0 {
		return strings.Repeat(" ", width)
	}
	row := make([]rune, width)
	for i := range row {
		row[i] = ' '
	}
	barWidth := 1
	if len(ns.Values) > 0 {
		dotCols := width * 2
		barWidth = dotCols / len(ns.Values)
		if barWidth < 2 {
			barWidth = 2
		}
	}
	for name, pos := range pins {
		label := []rune(name)
		cellCol := pos * barWidth / 2
		for i, ch := range label {
			if cellCol+i < width {
				row[cellCol+i] = ch
			}
		}
	}
	return string(row)
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

// addNoise randomly replaces some braille cells with noise characters.
func addNoise(rows []string, level float64) []string {
	noisy := make([]string, len(rows))
	for i, row := range rows {
		runes := []rune(row)
		for j, ch := range runes {
			if rand.Float64() < level*0.15 {
				// Replace with a random braille char.
				mask := byte(rand.Intn(256))
				runes[j] = brailleChar(mask)
			} else {
				runes[j] = ch
			}
		}
		noisy[i] = string(runes)
	}
	return noisy
}

// corruptBorder replaces some border characters with ~ and other noise.
func corruptBorder(s string, heat float64) string {
	// Simple line-by-line corruption on first and last lines.
	lines := strings.Split(s, "\n")
	noiseChars := []rune{'~', '!', '?', '%', '&'}
	corrupt := func(line string) string {
		runes := []rune(line)
		for i, ch := range runes {
			if ch == '─' || ch == '━' {
				if rand.Float64() < (heat-0.60)*2.0 {
					runes[i] = noiseChars[rand.Intn(len(noiseChars))]
				}
			}
		}
		return string(runes)
	}
	if len(lines) > 0 {
		lines[0] = corrupt(lines[0])
	}
	if len(lines) > 1 {
		lines[len(lines)-1] = corrupt(lines[len(lines)-1])
	}
	return strings.Join(lines, "\n")
}

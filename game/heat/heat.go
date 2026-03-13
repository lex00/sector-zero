package heat

import (
	"fmt"
	"strings"
	"time"
)

// Model tracks thermal state and fuse count.
type Model struct {
	level        float64
	fuses        int
	lastActivity time.Time
}

const (
	fusesMax        = 3
	dissipatePerSec = 0.01
)

// New returns an initialised Model with full fuses and zero heat.
func New() Model {
	return Model{
		level:        0,
		fuses:        fusesMax,
		lastActivity: time.Now(),
	}
}

// NewWithState creates a Model from saved values.
func NewWithState(level float64, fuses int) Model {
	return Model{
		level:        clamp(level),
		fuses:        fuses,
		lastActivity: time.Now(),
	}
}

// Level returns the current heat level in [0, 1].
func (m Model) Level() float64 { return m.level }

// Stage returns a human-readable heat stage string.
func (m Model) Stage() string {
	switch {
	case m.level >= 1.0:
		return "boom"
	case m.level >= 0.85:
		return "critical"
	case m.level >= 0.60:
		return "hot"
	case m.level >= 0.30:
		return "warm"
	default:
		return "cool"
	}
}

// Add increases the heat level by delta (clamped to [0, 1]).
func (m *Model) Add(delta float64) {
	m.level = clamp(m.level + delta)
	m.lastActivity = time.Now()
}

// Set sets the heat level directly (clamped to [0, 1]).
func (m *Model) Set(level float64) {
	m.level = clamp(level)
}

// Tick dissipates heat by one tick unit (~1%/sec when called every second).
func (m *Model) Tick() {
	if m.level > 0 {
		m.level = clamp(m.level - dissipatePerSec)
	}
}

// Boom should be called when heat hits 100%. Burns one fuse and resets heat to 40%.
// Returns true if there are no more fuses remaining after burning.
func (m *Model) Boom() bool {
	if m.fuses > 0 {
		m.fuses--
	}
	m.level = 0.40
	return m.fuses == 0
}

// HasFuses reports whether at least one fuse remains.
func (m Model) HasFuses() bool { return m.fuses > 0 }

// FuseCount returns the number of remaining fuses.
func (m Model) FuseCount() int { return m.fuses }

// FusesMax returns the maximum number of fuses.
func (m Model) FusesMax() int { return fusesMax }

// HeatBar renders a width-character wide ASCII heat bar.
// Example: ░░░░░▓▓▓▓▓░░
func (m Model) HeatBar(width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(float64(width) * m.level)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)

	// Colour based on stage.
	var colour string
	switch m.Stage() {
	case "warm":
		colour = "yellow"
	case "hot":
		colour = "208" // orange
	case "critical":
		colour = "red"
	case "boom":
		colour = "red"
	default:
		colour = "green"
	}
	_ = colour // lipgloss colouring is applied by the caller scene
	return bar
}

// FuseDisplay renders a fuse indicator string like "⊡ ⊡ ⊠".
// ⊡ = intact fuse, ⊠ = burned fuse.
func (m Model) FuseDisplay() string {
	parts := make([]string, fusesMax)
	for i := 0; i < fusesMax; i++ {
		if i < m.fuses {
			parts[i] = "⊡"
		} else {
			parts[i] = "⊠"
		}
	}
	return fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
}

// HeatDeltaForCategory returns the appropriate heat gain for a diff category.
func HeatDeltaForCategory(category string) float64 {
	switch category {
	case "empty":
		return 0.20
	case "wrong_types":
		return 0.15
	case "wrong_order":
		return 0.10
	case "wrong_termination":
		return 0.08
	case "near":
		return 0.05
	case "exact":
		return 0 // no heat gain on exact match
	default:
		return 0.10
	}
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

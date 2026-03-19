package scope

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
	if m.Victory {
		brailleRows = m.renderVictory(innerW, innerH-2)
	} else if ns != nil {
		highlights := make(map[int]string)
		if sigs, ok := m.signals[netName]; ok {
			for sigName, positions := range sigs {
				for _, pos := range positions {
					highlights[pos] = sigName
				}
			}
		}
		// Access pulse: lower priority than compare/swap/found — only set if not already highlighted.
		if pos, ok := m.accesses[netName]; ok {
			if _, exists := highlights[pos]; !exists {
				highlights[pos] = "access"
			}
		}
		brailleRows = RenderBars(ns.Normalised, ns.Values, highlights, innerW, innerH-2)
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

// victoryNoise returns a pseudo-random braille mask for cell (r, c) at frame f.
// Uses a simple integer hash — fast and deterministic, good enough for animation.
func victoryNoise(r, c, f int) byte {
	x := uint32(r*0x9E3779B9) ^ uint32(c*0x6C62272E) ^ uint32(f*0x517CC1B7)
	x ^= x >> 16
	x *= 0x45D9F3B
	x ^= x >> 16
	return byte(x)
}

// renderVictory generates braille rows for the post-solve victory animation.
// Three bright signal sweeps race horizontally across a field of random noise.
func (m Model) renderVictory(width, height int) []string {
	f := m.VictoryFrame

	// Three horizontal signal sweeps at different row positions and speeds.
	type sweep struct{ row, speed int }
	sweeps := []sweep{
		{height * 1 / 5, 3},
		{height * 2 / 5, 2},
		{height * 3 / 5, 4},
	}

	dimStyle    := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // dark green
	brightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan

	rows := make([]string, height)
	for r := 0; r < height; r++ {
		type cell struct {
			ch     rune
			bright bool
		}
		cells := make([]cell, width)

		// Background: random braille noise, slowly shifting each frame.
		for c := 0; c < width; c++ {
			cells[c] = cell{ch: brailleChar(victoryNoise(r, c, f/3))}
		}

		// Overlay signal sweeps.
		for _, sw := range sweeps {
			// Column of the leading edge; wraps with a short off-screen pause.
			col := (f*sw.speed)%(width+20) - 10
			for dr := -1; dr <= 1; dr++ {
				if r != sw.row+dr {
					continue
				}
				// 5-cell wide pulse: full-dot cap at leading edge, fading behind.
				widths := []struct {
					dc   int
					full bool
				}{
					{0, true}, {-1, true}, {1, false}, {-2, false}, {-3, false},
				}
				for _, w := range widths {
					c := col + w.dc
					if c >= 0 && c < width {
						var mask byte
						if w.full {
							mask = 0xFF
						} else {
							mask = victoryNoise(r, c, f)&0xAA | 0x55
						}
						cells[c] = cell{ch: brailleChar(mask), bright: true}
					}
				}
			}
		}

		// Render grouped runs of same-brightness cells.
		var sb strings.Builder
		i := 0
		for i < width {
			bright := cells[i].bright
			j := i
			for j < width && cells[j].bright == bright {
				j++
			}
			var seg strings.Builder
			for k := i; k < j; k++ {
				seg.WriteRune(cells[k].ch)
			}
			if bright {
				sb.WriteString(brightStyle.Render(seg.String()))
			} else {
				sb.WriteString(dimStyle.Render(seg.String()))
			}
			i = j
		}
		rows[r] = sb.String()
	}
	return rows
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

// addNoise randomly replaces some braille cells with noise characters.
// It skips ANSI escape sequences so it doesn't corrupt color codes.
func addNoise(rows []string, level float64) []string {
	noisy := make([]string, len(rows))
	for i, row := range rows {
		runes := []rune(row)
		out := make([]rune, 0, len(runes))
		inEscape := false
		for _, ch := range runes {
			if ch == 0x1B { // ESC — start of ANSI sequence
				inEscape = true
				out = append(out, ch)
				continue
			}
			if inEscape {
				out = append(out, ch)
				if ch == 'm' {
					inEscape = false
				}
				continue
			}
			if rand.Float64() < level*0.15 {
				mask := byte(rand.Intn(256))
				out = append(out, brailleChar(mask))
			} else {
				out = append(out, ch)
			}
		}
		noisy[i] = string(out)
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

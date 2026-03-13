package scenes

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RebootTickMsg advances the reboot animation.
type RebootTickMsg time.Time

// RebootDoneMsg signals that the reboot sequence is complete.
type RebootDoneMsg struct{}

// Reboot is the post-BOOM reboot scene.
type Reboot struct {
	Progress float64 // 0.0 → 1.0
	Width    int
	Height   int
	Done     bool
	phase    int // 0=filling, 1=opossum, 2=transitioning
	phaseAge int
}

// NewReboot creates an initialised Reboot scene.
func NewReboot(width, height int) Reboot {
	return Reboot{Width: width, Height: height}
}

func rebootTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return RebootTickMsg(t)
	})
}

// Init starts the reboot animation.
func (m Reboot) Init() (Reboot, tea.Cmd) {
	return m, rebootTick()
}

// Update handles reboot animation ticks.
func (m Reboot) Update(msg tea.Msg) (Reboot, tea.Cmd) {
	switch msg := msg.(type) {
	case RebootTickMsg:
		_ = msg
		switch m.phase {
		case 0: // filling progress bar
			m.Progress += 0.015 // ~3 seconds to fill at 50ms tick
			if m.Progress >= 1.0 {
				m.Progress = 1.0
				m.phase = 1
				m.phaseAge = 0
			}
		case 1: // opossum message
			m.phaseAge++
			if m.phaseAge > 40 { // ~2 seconds
				m.phase = 2
				m.phaseAge = 0
			}
		case 2: // transitioning back to puzzle
			m.phaseAge++
			if m.phaseAge > 20 {
				m.Done = true
				return m, func() tea.Msg { return RebootDoneMsg{} }
			}
		}
		return m, rebootTick()

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// View renders the reboot sequence.
func (m Reboot) View() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}
	h := m.Height
	if h <= 0 {
		h = 24
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	brightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	italicStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Italic(true)

	var lines []string

	// Fill screen line by line based on progress.
	totalLines := h - 4 // reserve space for the header and progress bar
	filledLines := int(m.Progress * float64(totalLines))

	header := brightStyle.Render("[ REBOOTING... ]")
	lines = append(lines, header)
	lines = append(lines, "")

	// Progress bar.
	barWidth := w - 4
	if barWidth < 10 {
		barWidth = 10
	}
	filledBar := int(m.Progress * float64(barWidth))
	bar := "[" + strings.Repeat("▓", filledBar) + strings.Repeat("░", barWidth-filledBar) + "]"
	pct := fmt.Sprintf(" %3.0f%%", m.Progress*100)
	lines = append(lines, brightStyle.Render(bar+pct))
	lines = append(lines, "")

	// Fill lines.
	fillLine := strings.Repeat("·", w-2)
	for i := 0; i < totalLines; i++ {
		if i < filledLines {
			lines = append(lines, dimStyle.Render(fillLine))
		} else {
			lines = append(lines, strings.Repeat(" ", w-2))
		}
	}

	// Opossum message overlay (shown in phase 1+).
	if m.phase >= 1 {
		// Replace middle lines with opossum messages.
		mid := len(lines) / 2
		opLines := []string{
			"",
			italicStyle.Render("*slowly rights himself*"),
			"",
			brightStyle.Render("systems nominal"),
			dimStyle.Render("heat: 40%  fuses: -1"),
			"",
		}
		for i, ol := range opLines {
			idx := mid - len(opLines)/2 + i
			if idx >= 0 && idx < len(lines) {
				lines[idx] = ol
			}
		}
	}

	if m.phase >= 2 {
		mid := len(lines)/2 + 3
		if mid < len(lines) {
			lines[mid] = brightStyle.Render("[ resuming... ]")
		}
	}

	return strings.Join(lines, "\n")
}

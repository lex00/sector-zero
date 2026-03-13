package scenes

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BoomTickMsg advances the explosion animation.
type BoomTickMsg time.Time

// RebootMsg signals transition to the reboot scene.
type RebootMsg struct{}

// Boom is the thermal overload explosion scene.
type Boom struct {
	Width   int
	Height  int
	Tick    int
	Done    bool
	fading  bool
	fadeVal float64
}

// NewBoom creates an initialised Boom scene.
func NewBoom(width, height int) Boom {
	return Boom{Width: width, Height: height}
}

func boomTick() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(t time.Time) tea.Msg {
		return BoomTickMsg(t)
	})
}

// Init rings the terminal bell and starts the animation ticker.
func (m Boom) Init() (Boom, tea.Cmd) {
	// Terminal bell is printed as a side-effect string in the view.
	return m, boomTick()
}

// Update handles boom animation ticks.
func (m Boom) Update(msg tea.Msg) (Boom, tea.Cmd) {
	switch msg := msg.(type) {
	case BoomTickMsg:
		_ = msg
		m.Tick++
		maxRadius := int(math.Sqrt(float64(m.Width*m.Width+m.Height*m.Height))) / 2
		if m.fading {
			m.fadeVal += 0.05
			if m.fadeVal >= 1.0 {
				m.Done = true
				return m, func() tea.Msg { return RebootMsg{} }
			}
		} else if m.Tick > maxRadius+10 {
			m.fading = true
		}
		return m, boomTick()

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// View renders the explosion animation.
func (m Boom) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	// Add bell on first tick.
	bell := ""
	if m.Tick == 1 {
		bell = "\a"
	}

	cx := m.Width / 2
	cy := m.Height / 2

	// Expansion chars by ring mod.
	shockChars := []rune{'*', '#', '!', '@', '%', '&'}

	// Fading: dim everything to black.
	if m.fading {
		fadeChars := [][]rune{
			{'*', '#', '!'},
			{':', '.', ' '},
			{' ', ' ', ' '},
		}
		fadeIdx := int(m.fadeVal * float64(len(fadeChars)))
		if fadeIdx >= len(fadeChars) {
			fadeIdx = len(fadeChars) - 1
		}
		chars := fadeChars[fadeIdx]

		rows := make([]string, m.Height)
		for y := 0; y < m.Height; y++ {
			row := make([]rune, m.Width)
			for x := 0; x < m.Width; x++ {
				dx := float64(x - cx)
				dy := float64(y-cy) * 2 // compensate for terminal cell aspect ratio
				dist := math.Sqrt(dx*dx + dy*dy)
				_ = dist
				idx := (x + y) % len(chars)
				row[x] = chars[idx]
			}
			rows[y] = string(row)
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		return bell + style.Render(strings.Join(rows, "\n"))
	}

	radius := m.Tick

	rows := make([]string, m.Height)
	for y := 0; y < m.Height; y++ {
		row := make([]rune, m.Width)
		for x := 0; x < m.Width; x++ {
			dx := float64(x - cx)
			dy := float64(y-cy) * 2
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= float64(radius) {
				ring := int(float64(radius) - dist)
				row[x] = shockChars[ring%len(shockChars)]
			} else {
				row[x] = ' '
			}
		}
		rows[y] = string(row)
	}

	// Colour: yellow at edges, red at centre.
	var coloured []string
	for y, row := range rows {
		dy := math.Abs(float64(y - cy))
		var colour lipgloss.Color
		switch {
		case dy < float64(cy)*0.2:
			colour = lipgloss.Color("9") // bright red
		case dy < float64(cy)*0.5:
			colour = lipgloss.Color("208") // orange
		default:
			colour = lipgloss.Color("11") // yellow
		}
		coloured = append(coloured, lipgloss.NewStyle().Foreground(colour).Render(row))
	}

	return bell + strings.Join(coloured, "\n")
}

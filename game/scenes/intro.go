package scenes

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// introLines is the full narrative shown on the intro screen.
var introLines = []string{
	"[ SECTOR ZERO ]",
	"",
	"Somewhere over central Florida, something fell.",
	"",
	"The heat was wrong. The sound was wrong.",
	"The crater smelled like copper and static.",
	"",
	"An opossum found it first.",
	"",
	"She had found things before. A Dorito. Half a tennis ball.",
	"This was not those things.",
	"",
	"She pushed the button with her nose.",
	"",
	"Three hundred miles away, a decommissioned relay post",
	"lit up for the first time in eleven years.",
	"",
	"That was six weeks ago.",
	"",
	"You are here now.",
	"",
	"[ press any key ]",
}

// SceneID identifies a named scene in the top-level router.
type SceneID uint8

const (
	SceneIntro  SceneID = iota
	ScenePuzzle
	SceneBoom
	SceneReboot
)

// IntroTickMsg advances the line reveal.
type IntroTickMsg time.Time

// TransitionMsg asks the top-level model to switch scenes.
type TransitionMsg struct{ Scene SceneID }

// Intro is the intro scene model.
type Intro struct {
	Lines    []string
	Revealed int
	Done     bool
	Width    int
	Height   int
}

// NewIntro creates an initialised Intro model.
func NewIntro(width, height int) Intro {
	return Intro{
		Lines:  introLines,
		Width:  width,
		Height: height,
	}
}

func introTick() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
		return IntroTickMsg(t)
	})
}

// Init starts the line-reveal ticker.
func (m Intro) Init() (Intro, tea.Cmd) {
	return m, introTick()
}

// Update handles intro messages.
func (m Intro) Update(msg tea.Msg) (Intro, tea.Cmd) {
	switch msg := msg.(type) {
	case IntroTickMsg:
		_ = msg
		if m.Revealed < len(m.Lines) {
			m.Revealed++
			if m.Revealed < len(m.Lines) {
				return m, introTick()
			}
			// All lines revealed — mark done, wait for key.
			m.Done = true
		}

	case tea.KeyMsg:
		_ = msg
		if m.Done {
			return m, func() tea.Msg { return TransitionMsg{Scene: ScenePuzzle} }
		}
		// Pressing a key before all lines are revealed shows them all at once.
		m.Revealed = len(m.Lines)
		m.Done = true

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// View renders the intro screen.
func (m Intro) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // bright green
		MarginBottom(1)

	italicStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("8")) // dark grey

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // yellow
		Bold(true)

	containerStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Width(m.Width).
		Height(m.Height)

	var lines []string
	for i, line := range m.Lines {
		if i >= m.Revealed {
			break
		}
		var rendered string
		switch {
		case line == "[ SECTOR ZERO ]":
			rendered = titleStyle.Render(line)
		case line == "[ press any key ]":
			rendered = promptStyle.Render(line)
		case len(line) > 0 && line[0] == '*' && line[len(line)-1] == '*':
			rendered = italicStyle.Render(line)
		default:
			rendered = line
		}
		lines = append(lines, rendered)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return containerStyle.Render(body)
}

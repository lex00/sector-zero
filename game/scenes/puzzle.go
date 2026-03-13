package scenes

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lex00/sector-zero/game/diff"
	"github.com/lex00/sector-zero/game/heat"
	"github.com/lex00/sector-zero/game/opossum"
	"github.com/lex00/sector-zero/game/puzzles"
	"github.com/lex00/sector-zero/game/runner"
	"github.com/lex00/sector-zero/game/scope"
)

// Focus targets.
const (
	FocusScope = iota
	FocusCode
	FocusDialogue
)

// Available languages in cycle order.
var languages = []string{"python", "go", "js", "rust", "java", "ruby"}

// RunCodeMsg is sent when Ctrl+R is pressed.
type RunCodeMsg struct{}

// RunResultMsg carries the result of running the player's code.
type RunResultMsg struct {
	Pulses []scope.Pulse
	Err    error
}

// HeatTickMsg is the 1-second heat dissipation tick.
type HeatTickMsg time.Time

// SolvedMsg signals that a puzzle was solved exactly.
type SolvedMsg struct{ PuzzleID int }

// PuzzleQuitMsg signals the player wants to quit.
type PuzzleQuitMsg struct{}

// Puzzle is the main four-panel game scene.
type Puzzle struct {
	Width  int
	Height int

	focus int

	ScopeModel   scope.Model
	CodeArea     textarea.Model
	HeatModel    heat.Model
	OpossumModel opossum.Model
	PuzzleData   puzzles.Puzzle

	HelpLevel  int // 0=BLACKOUT,1=STATIC,2=SIGNAL,3=OPEN
	Dialogue   []string
	SyntaxErr  string
	LangIndex  int
	AvailLangs map[string]bool

	FullscreenScope bool
	LastKeystroke   time.Time
	running         bool
}

var helpLevels = []string{"BLACKOUT", "STATIC", "SIGNAL", "OPEN"}

// NewPuzzle creates an initialised Puzzle scene.
func NewPuzzle(width, height int, pz puzzles.Puzzle, hlevel string, heatLevel float64, fuses int) Puzzle {
	// Textarea for code input.
	ta := textarea.New()
	ta.SetWidth(40)
	ta.SetHeight(20)
	ta.Placeholder = "// write your solution here"
	_ = ta.Focus() // returns tea.Cmd in newer bubbles; discard here
	ta.ShowLineNumbers = true
	ta.CharLimit = 0

	scopeM := scope.New(width/2, height-10)

	// Load target trace.
	pulses, err := puzzles.LoadTrace(pz)
	if err == nil {
		scopeM.SetTrace(pulses)
	}

	// Help level index.
	hli := 2 // default SIGNAL
	for i, lvl := range helpLevels {
		if lvl == hlevel {
			hli = i
			break
		}
	}

	// Set scaffold code for OPEN level.
	code := ""
	if hli == 3 {
		if sc, ok := pz.ScaffoldCode["python"]; ok {
			code = sc
		}
	}
	ta.SetValue(code)

	return Puzzle{
		Width:        width,
		Height:       height,
		focus:        FocusCode,
		ScopeModel:   scopeM,
		CodeArea:     ta,
		HeatModel:    heat.NewWithState(heatLevel, fuses),
		OpossumModel: opossum.New(),
		PuzzleData:   pz,
		HelpLevel:    hli,
		Dialogue:     []string{puzzles.GetDialogue(pz, helpLevels[hli], "empty")},
		AvailLangs:   runner.CheckRuntimes(),
		LastKeystroke: time.Now(),
	}
}

// Init starts the scope animation and heat ticker.
func (m Puzzle) Init() (Puzzle, tea.Cmd) {
	scopeM, scopeCmd := m.ScopeModel.Init()
	m.ScopeModel = scopeM
	return m, tea.Batch(scopeCmd, heatTick())
}

func heatTick() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return HeatTickMsg(t)
	})
}

// Update handles all puzzle input and messages.
func (m Puzzle) Update(msg tea.Msg) (Puzzle, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.ScopeModel.Width = m.scopeWidth()
		m.ScopeModel.Height = m.scopeHeight()

	case HeatTickMsg:
		_ = msg
		m.HeatModel.Tick()
		// Update scope corruption based on heat.
		m.ScopeModel.HeatCorrupt = m.HeatModel.Level()
		m.ScopeModel.BrailleNoise = 0
		if m.HeatModel.Stage() == "hot" || m.HeatModel.Stage() == "critical" {
			m.ScopeModel.BrailleNoise = m.HeatModel.Level()
		}
		// Check for boom.
		if m.HeatModel.Level() >= 1.0 {
			m.OpossumModel.StateFromBoom()
			return m, func() tea.Msg { return TransitionMsg{Scene: "boom"} }
		}
		m.OpossumModel.UpdateFromHeat(m.HeatModel.Level())
		cmds = append(cmds, heatTick())

	case RunCodeMsg:
		_ = msg
		if !m.running {
			m.running = true
			lang := languages[m.LangIndex]
			code := m.CodeArea.Value()
			cmds = append(cmds, func() tea.Msg {
				pulses, err := runner.Run(lang, code, 0)
				return RunResultMsg{Pulses: pulses, Err: err}
			})
		}

	case RunResultMsg:
		m.running = false
		if msg.Err != nil {
			m.SyntaxErr = msg.Err.Error()
			m.pushDialogue("[ error ] " + shortErr(msg.Err))
			m.HeatModel.Add(0.10)
		} else {
			// Load player trace into scope for side-by-side feel.
			targetPulses, _ := puzzles.LoadTrace(m.PuzzleData)
			result := diff.Diff(msg.Pulses, targetPulses)

			// Show player trace in scope.
			m.ScopeModel.SetTrace(msg.Pulses)
			m.ScopeModel.CurrentPulse = 0

			hint := puzzles.GetDialogue(m.PuzzleData, helpLevels[m.HelpLevel], result.HintKey)
			m.pushDialogue(hint)

			if result.Category == "exact" {
				m.OpossumModel.StateFromSolve()
				m.HeatModel.Set(0)
				cmds = append(cmds, func() tea.Msg { return SolvedMsg{PuzzleID: m.PuzzleData.ID} })
			} else {
				m.HeatModel.Add(result.HeatDelta)
				m.OpossumModel.UpdateFromHeat(m.HeatModel.Level())
			}
		}

	case scope.TickMsg:
		scopeM, cmd := m.ScopeModel.Update(msg)
		m.ScopeModel = scopeM
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		m.LastKeystroke = time.Now()
		switch msg.String() {
		case "ctrl+c", "q":
			return m, func() tea.Msg { return PuzzleQuitMsg{} }

		case "tab":
			m.focus = (m.focus + 1) % 3
			if m.focus == FocusCode {
				focusCmd := m.CodeArea.Focus()
				if focusCmd != nil {
					cmds = append(cmds, focusCmd)
				}
			} else {
				m.CodeArea.Blur()
			}

		case "ctrl+r":
			cmds = append(cmds, func() tea.Msg { return RunCodeMsg{} })

		case "ctrl+x":
			m.CodeArea.SetValue("")

		case "ctrl+f":
			m.FullscreenScope = !m.FullscreenScope

		case "ctrl+l":
			m.LangIndex = (m.LangIndex + 1) % len(languages)
			// If OPEN level, update scaffold.
			if m.HelpLevel == 3 {
				lang := languages[m.LangIndex]
				if sc, ok := m.PuzzleData.ScaffoldCode[lang]; ok {
					m.CodeArea.SetValue(sc)
				} else {
					m.CodeArea.SetValue("")
				}
			}

		case "ctrl+h":
			m.HelpLevel = (m.HelpLevel + 1) % len(helpLevels)

		case "ctrl+p":
			m.ScopeModel.Paused = !m.ScopeModel.Paused

		default:
			if m.focus == FocusCode {
				var cmd tea.Cmd
				m.CodeArea, cmd = m.CodeArea.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the four-panel layout.
func (m Puzzle) View() string {
	if m.FullscreenScope {
		return m.renderFullscreenScope()
	}

	hud := m.renderHUD()
	scopePanel := m.renderScope()
	codePanel := m.renderCode()
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, scopePanel, codePanel)
	dialoguePanel := m.renderDialogue()

	return lipgloss.JoinVertical(lipgloss.Left, hud, topRow, dialoguePanel)
}

// renderHUD renders the top status bar.
func (m Puzzle) renderHUD() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}

	hudStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("15")).
		Width(w).
		Padding(0, 1)

	// Puzzle info.
	puzzleInfo := fmt.Sprintf("PUZZLE %d: %s", m.PuzzleData.ID, m.PuzzleData.Title)

	// Help level.
	helpColour := map[string]string{
		"BLACKOUT": "8",
		"STATIC":   "7",
		"SIGNAL":   "11",
		"OPEN":     "10",
	}
	hlvl := helpLevels[m.HelpLevel]
	helpDisplay := lipgloss.NewStyle().
		Foreground(lipgloss.Color(helpColour[hlvl])).
		Bold(true).
		Render(hlvl)

	// Heat bar.
	barWidth := 20
	heatBar := m.HeatModel.HeatBar(barWidth)
	heatColour := heatBarColour(m.HeatModel.Stage())
	heatDisplay := lipgloss.NewStyle().Foreground(lipgloss.Color(heatColour)).Render("▐" + heatBar + "▌")

	// Fuses.
	fuseDisplay := m.HeatModel.FuseDisplay()

	// Language.
	lang := strings.ToUpper(languages[m.LangIndex])
	if m.running {
		lang = "[ running... ]"
	}

	// Layout: puzzle | help | heat | fuses | lang
	parts := []string{
		puzzleInfo,
		"  ",
		helpDisplay,
		"  ",
		heatDisplay,
		"  ",
		fuseDisplay,
		"  ",
		lang,
	}

	return hudStyle.Render(strings.Join(parts, ""))
}

// renderScope renders the left SCOPE panel.
func (m Puzzle) renderScope() string {
	sw := m.scopeWidth()
	sh := m.scopeHeight()

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	label := labelStyle.Render("SCOPE")
	if m.ScopeModel.Paused {
		label += lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(" [paused]")
	}

	m.ScopeModel.Width = sw
	m.ScopeModel.Height = sh - 1

	scopeView := m.ScopeModel.View()

	return lipgloss.NewStyle().Width(sw).Render(
		lipgloss.JoinVertical(lipgloss.Left, label, scopeView),
	)
}

// renderCode renders the right CODE panel.
func (m Puzzle) renderCode() string {
	cw := m.codeWidth()
	ch := m.codeHeight()

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)

	label := labelStyle.Render(fmt.Sprintf("CODE [%s]", strings.ToUpper(languages[m.LangIndex])))

	m.CodeArea.SetWidth(cw - 2)
	m.CodeArea.SetHeight(ch - 3)

	codeView := m.CodeArea.View()

	borderColour := lipgloss.Color("4")
	if m.focus == FocusCode {
		borderColour = lipgloss.Color("12")
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColour).
		Width(cw - 2).
		Height(ch - 1)

	errorLine := ""
	if m.SyntaxErr != "" {
		errorLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("✗ " + truncate(m.SyntaxErr, cw-4))
	}

	return lipgloss.NewStyle().Width(cw).Render(
		lipgloss.JoinVertical(lipgloss.Left, label, boxStyle.Render(codeView), errorLine),
	)
}

// renderDialogue renders the bottom DIALOGUE panel.
func (m Puzzle) renderDialogue() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}
	dh := m.dialogueHeight()

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")).
		Bold(true)

	label := labelStyle.Render("DIALOGUE")

	// Opossum reaction.
	reaction := m.OpossumModel.Reaction()
	opStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	// Last few dialogue lines.
	lines := m.Dialogue
	if len(lines) > 3 {
		lines = lines[len(lines)-3:]
	}

	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	var renderedLines []string
	for _, l := range lines {
		renderedLines = append(renderedLines, textStyle.Render(l))
	}
	if reaction != "" {
		renderedLines = append(renderedLines, opStyle.Render(reaction))
	}

	keybinds := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("^R run  ^X clear  ^F scope  ^L lang  ^H help  ^P pause  tab focus  q quit")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Width(w - 2).
		Height(dh - 1)

	inner := lipgloss.JoinVertical(lipgloss.Left,
		append(renderedLines, "", keybinds)...,
	)

	return lipgloss.JoinVertical(lipgloss.Left, label, boxStyle.Render(inner))
}

// renderFullscreenScope renders the scope panel filling the entire terminal.
func (m Puzzle) renderFullscreenScope() string {
	m.ScopeModel.Width = m.Width
	m.ScopeModel.Height = m.Height - 1
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		Render("SCOPE  [full]  ctrl+f to exit")
	return lipgloss.JoinVertical(lipgloss.Left, header, m.ScopeModel.View())
}

// pushDialogue appends text to the dialogue history.
func (m *Puzzle) pushDialogue(s string) {
	m.Dialogue = append(m.Dialogue, s)
	if len(m.Dialogue) > 20 {
		m.Dialogue = m.Dialogue[len(m.Dialogue)-20:]
	}
}

// Layout helpers.

func (m Puzzle) hudHeight() int    { return 1 }
func (m Puzzle) dialogueHeight() int {
	h := 7
	if m.Height > 40 {
		h = 9
	}
	return h
}
func (m Puzzle) topRowHeight() int {
	return m.Height - m.hudHeight() - m.dialogueHeight()
}
func (m Puzzle) scopeWidth() int {
	w := m.Width * 40 / 100
	if w < 20 {
		w = 20
	}
	return w
}
func (m Puzzle) codeWidth() int {
	return m.Width - m.scopeWidth()
}
func (m Puzzle) scopeHeight() int { return m.topRowHeight() }
func (m Puzzle) codeHeight() int  { return m.topRowHeight() }

func heatBarColour(stage string) string {
	switch stage {
	case "warm":
		return "11"
	case "hot":
		return "208"
	case "critical", "boom":
		return "9"
	default:
		return "10"
	}
}

func shortErr(err error) string {
	s := err.Error()
	lines := strings.SplitN(s, "\n", 3)
	if len(lines) > 0 {
		return lines[0]
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

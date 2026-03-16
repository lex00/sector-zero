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
	choicePanel  ChoicePanel

	HelpLevel   int    // 0=BLACKOUT,1=STATIC,2=SIGNAL,3=OPEN
	popup       *string // non-nil = show modal overlay; dismissed by any key
	popupStyle  string  // "run", "solved", "error", "guide"
	LangIndex  int
	AvailLangs map[string]bool

	FullscreenScope  bool
	LastKeystroke    time.Time
	running          bool
	solved           bool
	solvedLoopTarget int  // LoopCount value that triggers the "advance" popup
	gated            bool // true = ^R blocked until animation loops once
	runAttempts      int
	playerTrace      []scope.Pulse // last trace from player's run
	scopeShowsPlayer bool          // true = show player trace, false = show reference
}

var helpLevels = []string{"BLACKOUT", "STATIC", "SIGNAL", "OPEN"}

// NewPuzzle creates an initialised Puzzle scene.
func NewPuzzle(width, height int, pz puzzles.Puzzle, hlevel string, heatLevel float64, fuses int) Puzzle {
	// Textarea for code input.
	ta := textarea.New()
	ta.SetWidth(40)
	ta.SetHeight(20)
	ta.Placeholder = "// write your solution here"
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

	cp := NewChoicePanel(pz.Challenge, "python", width, height)

	return Puzzle{
		Width:           width,
		Height:          height,
		focus:           FocusCode,
		ScopeModel:      scopeM,
		CodeArea:        ta,
		HeatModel:       heat.NewWithState(heatLevel, fuses),
		OpossumModel:    opossum.New(),
		PuzzleData:      pz,
		choicePanel:     cp,
		HelpLevel:       hli,
		AvailLangs:      runner.CheckRuntimes(),
		LastKeystroke:   time.Now(),
		FullscreenScope: false,
	}
}

// matchStep returns the highest-threshold matching LessonStep for the given event.
// Iterating in reverse means later steps (higher After) take precedence over
// earlier ones when multiple steps share the same On+Result — so "After: 2"
// overrides "After: 0" once the player has made enough attempts.
func (m *Puzzle) matchStep(on, result string) *puzzles.LessonStep {
	var best *puzzles.LessonStep
	for i := range m.PuzzleData.Script {
		s := &m.PuzzleData.Script[i]
		if s.On != on {
			continue
		}
		if s.Result != "" && s.Result != result {
			continue
		}
		if m.runAttempts < s.After {
			continue
		}
		// Keep the step with the highest After threshold (most specific).
		if best == nil || s.After >= best.After {
			best = s
		}
	}
	return best
}

// applyTrigger executes one or more "+" separated trigger tokens from a LessonStep.
func (m *Puzzle) applyTrigger(trigger string) {
	for _, tok := range strings.Split(trigger, "+") {
		switch strings.TrimSpace(tok) {
		case "guide_next":
			if m.HelpLevel != 3 {
				m.HelpLevel = 3
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}
			for i, blank := range m.PuzzleData.Challenge.Blanks {
				if m.choicePanel.Selections[i] != blank.Correct {
					m.choicePanel.Selections[i] = blank.Correct
					m.choicePanel.FocusedBlank = i
					break
				}
			}
		case "guide_all":
			if m.HelpLevel != 3 {
				m.HelpLevel = 3
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}
			for i, blank := range m.PuzzleData.Challenge.Blanks {
				m.choicePanel.Selections[i] = blank.Correct
			}
		case "scope_ref":
			ref, err := puzzles.LoadTrace(m.PuzzleData)
			if err == nil {
				m.ScopeModel.SetTrace(ref)
				m.scopeShowsPlayer = false
			}
		case "scope_player":
			if len(m.playerTrace) > 0 {
				m.ScopeModel.SetTrace(m.playerTrace)
				m.scopeShowsPlayer = true
			}
		case "scope_pause":
			m.ScopeModel.Paused = true
		case "scope_play":
			m.ScopeModel.Paused = false
		case "mode_open":
			if m.HelpLevel != 3 {
				m.HelpLevel = 3
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}
		case "gate":
			m.gated = true
		}
	}
}

// Init starts the scope animation, heat ticker, and focuses the code area.
func (m Puzzle) Init() (Puzzle, tea.Cmd) {
	scopeM, scopeCmd := m.ScopeModel.Init()
	m.ScopeModel = scopeM
	focusCmd := m.CodeArea.Focus()
	// Fire the load step if the puzzle defines one.
	if step := m.matchStep("load", ""); step != nil {
		msg := step.Message
		style := step.Style
		if style == "" {
			style = "guide"
		}
		m.popup = &msg
		m.popupStyle = style
	}
	clearCmd := func() tea.Msg { return tea.ClearScreen() }
	return m, tea.Batch(clearCmd, scopeCmd, heatTick(), focusCmd)
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
		m.CodeArea.SetWidth(m.codeWidth() - 2)
		m.CodeArea.SetHeight(m.codeHeight() - 4)
		m.choicePanel.Width = m.codeWidth()
		m.choicePanel.Height = m.codeHeight()

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
			var code string
			if m.HelpLevel == 3 {
				code = m.choicePanel.ResolvedCode()
			} else {
				code = m.CodeArea.Value()
			}
			cmds = append(cmds, func() tea.Msg {
				pulses, err := runner.Run(lang, code, 0)
				return RunResultMsg{Pulses: pulses, Err: err}
			})
		}

	case RunResultMsg:
		m.running = false
		m.runAttempts++
		if msg.Err != nil {
			popMsg := "✗  " + shortErr(msg.Err)
			if step := m.matchStep("error", ""); step != nil && step.Message != "" {
				popMsg = step.Message
			}
			m.setPopup(popMsg, "error")
			m.HeatModel.Add(0.10)
		} else {
			targetPulses, _ := puzzles.LoadTrace(m.PuzzleData)
			result := diff.Diff(msg.Pulses, targetPulses)

			m.playerTrace = msg.Pulses
			m.scopeShowsPlayer = true
			m.ScopeModel.SetTrace(msg.Pulses)

			// Determine message and style from script or DialogueSpec fallback.
			popMsg := puzzles.GetDialogue(m.PuzzleData, helpLevels[m.HelpLevel], result.HintKey)
			popStyle := "run"
			if result.Category == "exact" {
				popStyle = "solved"
			}
			if step := m.matchStep("run", result.Category); step != nil {
				if step.Message != "" {
					popMsg = step.Message
				}
				if step.Style != "" {
					popStyle = step.Style
				}
				if step.Trigger != "" {
					m.applyTrigger(step.Trigger)
				}
			}
			if result.Category == "exact" {
				m.OpossumModel.StateFromSolve()
				m.HeatModel.Set(0)
				m.solved = true
				// Go fullscreen immediately and wait for one full loop before advancing.
				m.FullscreenScope = true
				m.ScopeModel.Paused = false
				m.solvedLoopTarget = m.ScopeModel.LoopCount + 1
			} else {
				// Append score so the player can see proximity.
				popMsg += fmt.Sprintf("\n\nscore: %d%%", int(result.Score*100))
				m.setPopup(popMsg, popStyle)
				m.HeatModel.Add(result.HeatDelta)
				m.OpossumModel.UpdateFromHeat(m.HeatModel.Level())
			}
		}

	case scope.TickMsg:
		prevLoops := m.ScopeModel.LoopCount
		scopeM, cmd := m.ScopeModel.Update(msg)
		m.ScopeModel = scopeM
		cmds = append(cmds, cmd)
		// Clear gate when the reference animation completes its first full loop.
		if m.gated && !m.scopeShowsPlayer && m.ScopeModel.LoopCount > prevLoops {
			m.gated = false
		}
		// After solving: show "advancing" popup once the animation completes one loop.
		if m.solved && m.popup == nil && m.ScopeModel.LoopCount >= m.solvedLoopTarget {
			msg := "puzzle solved.\n\nadvancing to next puzzle..."
			m.popup = &msg
			m.popupStyle = "solved"
		}
		// Sync the choice panel gutter with the current animation pulse.
		if cp := m.ScopeModel.CurrentPulse; cp > 0 && cp <= len(m.ScopeModel.Trace) {
			m.choicePanel.ActivePulseType = m.ScopeModel.Trace[cp-1].Type
		}

	case tea.KeyMsg:
		m.LastKeystroke = time.Now()
		key := msg.String()

		// Quit is always handled first.
		if key == "ctrl+c" || key == "q" {
			return m, func() tea.Msg { return PuzzleQuitMsg{} }
		}

		// Popup dismissal.
		if m.popup != nil {
			m.popup = nil
			// When solved the "advancing" popup is the final gate — any key advances.
			if m.solved {
				id := m.PuzzleData.ID
				return m, func() tea.Msg { return SolvedMsg{PuzzleID: id} }
			}
			// For normal popups let action keys fall through so ^G ^R etc work immediately.
			switch key {
			case "ctrl+g", "ctrl+r", "ctrl+v", "ctrl+p", "ctrl+x", "ctrl+h":
				// fall through to action handling below
			default:
				return m, nil
			}
		}
		switch key {
		case "ctrl+r":
			if m.gated {
				m.setPopup("Watch the reference animation first.\n\n^R unlocks when it completes one full loop.", "guide")
			} else {
				cmds = append(cmds, func() tea.Msg { return RunCodeMsg{} })
			}

		case "ctrl+x":
			if m.HelpLevel == 3 {
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			} else {
				m.CodeArea.SetValue("")
			}

		case "ctrl+f":
			m.FullscreenScope = !m.FullscreenScope
			if !m.FullscreenScope {
				m.focus = FocusCode
				if focusCmd := m.CodeArea.Focus(); focusCmd != nil {
					cmds = append(cmds, focusCmd)
				}
			}

		case "ctrl+l":
			// Cycle to the next available runtime, skipping unavailable ones.
			for range languages {
				m.LangIndex = (m.LangIndex + 1) % len(languages)
				if m.AvailLangs[languages[m.LangIndex]] {
					break
				}
			}
			m.choicePanel.Lang = languages[m.LangIndex]
			// If non-OPEN level, update scaffold.
			if m.HelpLevel != 3 {
				lang := languages[m.LangIndex]
				if sc, ok := m.PuzzleData.ScaffoldCode[lang]; ok {
					m.CodeArea.SetValue(sc)
				} else {
					m.CodeArea.SetValue("")
				}
			}

		case "ctrl+h":
			m.HelpLevel = (m.HelpLevel + 1) % len(helpLevels)
			if m.HelpLevel == 3 {
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}

		case "ctrl+g":
			// Switch to OPEN mode if not already there.
			if m.HelpLevel != 3 {
				m.HelpLevel = 3
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}
			// Reveal the next incorrect blank, one at a time.
			blanks := m.PuzzleData.Challenge.Blanks
			total := len(blanks)
			revealed := 0
			for i, blank := range blanks {
				if m.choicePanel.Selections[i] == blank.Correct {
					revealed++
				}
			}
			if revealed == total {
				m.setPopup("All blanks correct — press ^R to run.", "guide")
			} else {
				for i, blank := range blanks {
					if m.choicePanel.Selections[i] != blank.Correct {
						m.choicePanel.Selections[i] = blank.Correct
						m.choicePanel.FocusedBlank = i
						m.setPopup(fmt.Sprintf("blank %d / %d  —  %s\n\n%s", revealed+1, total, blank.Label, blank.Explanation), "guide")
						break
					}
				}
			}

		case "ctrl+v":
			if len(m.playerTrace) > 0 {
				m.scopeShowsPlayer = !m.scopeShowsPlayer
				if m.scopeShowsPlayer {
					m.ScopeModel.SetTrace(m.playerTrace)
				} else {
					ref, _ := puzzles.LoadTrace(m.PuzzleData)
					m.ScopeModel.SetTrace(ref)
				}
				m.ScopeModel.CurrentPulse = 0
			}

		case "ctrl+p":
			m.ScopeModel.Paused = !m.ScopeModel.Paused

		default:
			if m.HelpLevel == 3 {
				var cmd tea.Cmd
				m.choicePanel, cmd = m.choicePanel.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				var cmd tea.Cmd
				m.CodeArea, cmd = m.CodeArea.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the layout.
func (m Puzzle) View() string {
	w := m.Width
	h := m.Height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	if m.FullscreenScope {
		return lipgloss.NewStyle().Width(w).Height(h).Render(m.renderFullscreenScope())
	}

	hud := m.renderHUD()
	controls := m.renderControls()

	var middle string
	if m.popup != nil {
		middle = m.renderPopup()
	} else {
		scopePanel := m.renderScope()
		codePanel := m.renderCode()
		middle = lipgloss.JoinHorizontal(lipgloss.Top, scopePanel, codePanel)
	}

	// One blank row at the top matches the intro's Padding(2,4) offset so that
	// terminal chrome (tab bars etc.) hiding row 1 doesn't eat the HUD.
	content := lipgloss.JoinVertical(lipgloss.Left, "", hud, middle, controls)
	// Wrap in a fixed-size container so BubbleTea always sees exactly h rows —
	// the same guarantee the intro scene provides via its containerStyle.
	return lipgloss.NewStyle().Width(w).Height(h).Render(content)
}

// renderHUD renders the top status bar.
func (m Puzzle) renderHUD() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}

	// Solved state: replace the whole HUD with a green banner.
	if m.solved {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("2")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Width(w).
			Padding(0, 1).
			Render(fmt.Sprintf("★  PUZZLE %d: %s — SOLVED   press any key to continue",
				m.PuzzleData.ID, m.PuzzleData.Title))
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

	// Language / gate indicator.
	lang := strings.ToUpper(languages[m.LangIndex])
	if m.running {
		lang = "[ running... ]"
	} else if m.gated {
		lang = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("[ watch animation ]")
	}

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

	scopeTag := "REFERENCE"
	if m.scopeShowsPlayer {
		scopeTag = "YOUR RUN"
	}
	label := labelStyle.Render("SCOPE") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(" ["+scopeTag+"]")
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

	var codeView string
	if m.HelpLevel == 3 {
		m.choicePanel.Width = cw
		m.choicePanel.Height = ch
		codeView = m.choicePanel.View()
	} else {
		m.CodeArea.SetWidth(cw - 2)
		m.CodeArea.SetHeight(ch - 4)
		borderColour := lipgloss.Color("4")
		if m.focus == FocusCode {
			borderColour = lipgloss.Color("12")
		}
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColour).
			Width(cw - 2).
			Height(ch - 4)
		codeView = boxStyle.Render(m.CodeArea.View())
	}

	return lipgloss.NewStyle().Width(cw).MaxHeight(ch).Render(
		lipgloss.JoinVertical(lipgloss.Left, label, codeView),
	)
}

// renderControls renders the fixed bottom keybinds bar.
// Always exactly dialogueHeight() = 4 lines: label + bordered 1-line content.
func (m Puzzle) renderControls() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Render("CONTROLS")

	var keys string
	if m.HelpLevel == 3 {
		keys = "^R run  Tab blank  ↑↓ pick  ^G guide  ^X reset  ^V toggle  ^F scope  ^H hint  q quit"
	} else {
		keys = "^R run  ^X clear  ^F scope  ^L lang  ^H hint  ^P pause  q quit"
	}
	keybinds := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(keys)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Width(w - 2).
		Height(1).
		Render(keybinds)

	return lipgloss.JoinVertical(lipgloss.Left, label, box)
}

// renderPopup renders a centered modal overlay in the top-row area.
func (m Puzzle) renderPopup() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}
	trh := m.topRowHeight()

	msg := ""
	if m.popup != nil {
		msg = *m.popup
	}

	// Style parameters based on popup type.
	var (
		borderColor lipgloss.Color
		header      string
		msgColor    lipgloss.Color
	)
	switch m.popupStyle {
	case "solved":
		borderColor = lipgloss.Color("10")  // bright green
		header = "★  SOLVED  ★"
		msgColor = lipgloss.Color("10")
	case "error":
		borderColor = lipgloss.Color("9")   // red
		header = "ERROR"
		msgColor = lipgloss.Color("9")
	case "guide":
		borderColor = lipgloss.Color("14")  // cyan
		header = "GUIDE"
		msgColor = lipgloss.Color("15")
	default: // "run"
		borderColor = lipgloss.Color("13")  // magenta
		header = "RESULT"
		msgColor = lipgloss.Color("15")
	}

	headerStyle := lipgloss.NewStyle().Foreground(borderColor).Bold(true)
	msgStyle := lipgloss.NewStyle().Foreground(msgColor)
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	opStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)

	reaction := m.OpossumModel.Reaction()

	parts := []string{
		headerStyle.Render(header),
		"",
		msgStyle.Render(msg),
	}
	if reaction != "" {
		parts = append(parts, "", opStyle.Render(reaction))
	}
	parts = append(parts, "", promptStyle.Render("— press any key —"))

	inner := lipgloss.JoinVertical(lipgloss.Center, parts...)

	popW := min(w-12, 60)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 3).
		Width(popW).
		Render(inner)

	placed := lipgloss.Place(w, trh,
		lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("235")),
	)
	// Clip to trh rows in case the popup box is taller than the available space.
	return lipgloss.NewStyle().MaxHeight(trh).Render(placed)
}

// renderFullscreenScope renders the scope panel filling the entire terminal.
func (m Puzzle) renderFullscreenScope() string {
	m.ScopeModel.Width = m.Width
	m.ScopeModel.Height = m.Height - 1

	var header string
	if m.solved {
		header = lipgloss.NewStyle().
			Background(lipgloss.Color("2")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Width(m.Width).
			Padding(0, 1).
			Render(fmt.Sprintf("★  SOLVED — PUZZLE %d: %s   watch the trace",
				m.PuzzleData.ID, m.PuzzleData.Title))
	} else {
		header = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Render(fmt.Sprintf("SCOPE  PUZZLE %d: %s  — ctrl+f to return", m.PuzzleData.ID, m.PuzzleData.Title))
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, m.ScopeModel.View())
}

// setPopup sets the modal popup message with an optional style tag.
func (m *Puzzle) setPopup(s, style string) {
	m.popup = &s
	m.popupStyle = style
}

// Layout helpers.

func (m Puzzle) hudHeight() int      { return 1 }
func (m Puzzle) dialogueHeight() int { return 4 } // label(1) + border(2) + content(1)
func (m Puzzle) topRowHeight() int {
	// Subtract 2: 1 for the top spacer row, 1 to stay within the Height(h) container.
	return m.Height - m.hudHeight() - m.dialogueHeight() - 2
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

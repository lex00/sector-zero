package scenes

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lex00/sector-zero/game/heat"
	"github.com/lex00/sector-zero/game/opossum"
	"github.com/lex00/sector-zero/game/puzzles"
	"github.com/lex00/sector-zero/game/runner"
	"github.com/lex00/sector-zero/game/save"
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
	heatModel    heat.Model
	OpossumModel opossum.Model
	PuzzleData   puzzles.Puzzle
	choicePanel  ChoicePanel

	HelpLevel   save.HelpLevel // 0=BLACKOUT,1=STATIC,2=SIGNAL,3=OPEN
	popup       *string        // non-nil = show modal overlay; dismissed by any key
	popupStyle  string         // "run", "solved", "error", "guide"
	LangIndex   int
	AvailLangs  map[string]bool

	FullscreenScope  bool
	LastKeystroke    time.Time
	running          bool
	solved           bool
	gated            bool // true = ^R blocked until animation loops once
	runAttempts      int
	playerTrace      []scope.Pulse // last trace from player's run
	scopeShowsPlayer bool          // true = show player trace, false = show reference
}

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
	pulses, err := pz.LoadTrace()
	if err == nil {
		scopeM.SetTrace(pulses)
	}

	// Help level index.
	hli := save.HelpLevelFromString(hlevel)

	// Set scaffold code for OPEN level.
	code := ""
	if hli == save.HelpOpen {
		if sc, ok := pz.Challenge.Template["python"]; ok {
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
		heatModel:       heat.NewWithState(heatLevel, fuses),
		OpossumModel:    opossum.New(),
		PuzzleData:      pz,
		choicePanel:     cp,
		HelpLevel:       hli,
		AvailLangs:      runner.CheckRuntimes(),
		LastKeystroke:   time.Now(),
		FullscreenScope: false,
	}
}

// HeatLevel returns the current heat level (0–1).
func (m Puzzle) HeatLevel() float64 { return m.heatModel.Level() }

// FuseCount returns the number of remaining fuses.
func (m Puzzle) FuseCount() int { return m.heatModel.FuseCount() }

// TriggerBoom burns one fuse and resets heat to 40%.
func (m *Puzzle) TriggerBoom() { m.heatModel.Boom() }

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
		m.heatModel.Tick()
		// Update scope corruption based on heat.
		m.ScopeModel.HeatCorrupt = m.heatModel.Level()
		m.ScopeModel.BrailleNoise = 0
		if m.heatModel.Stage() == "hot" || m.heatModel.Stage() == "critical" {
			m.ScopeModel.BrailleNoise = m.heatModel.Level()
		}
		// Check for boom.
		if m.heatModel.Level() >= 1.0 {
			m.OpossumModel.StateFromBoom()
			return m, func() tea.Msg { return TransitionMsg{Scene: SceneBoom} }
		}
		m.OpossumModel.UpdateFromHeat(m.heatModel.Level())
		cmds = append(cmds, heatTick())

	case RunCodeMsg:
		var cmd tea.Cmd
		m, cmd = m.handleRunCode(msg)
		cmds = append(cmds, cmd)

	case RunResultMsg:
		var cmd tea.Cmd
		m, cmd = m.handleRunResult(msg)
		cmds = append(cmds, cmd)

	case scope.VictoryTickMsg:
		scopeM, cmd := m.ScopeModel.Update(msg)
		m.ScopeModel = scopeM
		cmds = append(cmds, cmd)
		// Show the congratulations popup after ~2 seconds of animation (25 frames × 80ms).
		if m.solved && m.popup == nil && m.ScopeModel.VictoryFrame >= 25 {
			s := "congratulations.\n\npress any key to continue."
			m.popup = &s
			m.popupStyle = "solved"
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
		// Sync the choice panel gutter with the current animation pulse.
		if cp := m.ScopeModel.CurrentPulse; cp > 0 && cp <= len(m.ScopeModel.Trace) {
			m.choicePanel.ActivePulseType = m.ScopeModel.Trace[cp-1].Type
		}

	case tea.KeyMsg:
		var cmd tea.Cmd
		m, cmd = m.handleKeyMsg(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

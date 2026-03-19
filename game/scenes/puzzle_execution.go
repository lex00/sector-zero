package scenes

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lex00/sector-zero/game/diff"
	"github.com/lex00/sector-zero/game/heat"
	"github.com/lex00/sector-zero/game/runner"
	"github.com/lex00/sector-zero/game/save"
	"github.com/lex00/sector-zero/game/scope"
)

// runEval holds the pure outcome of evaluating a player's run against the target.
type runEval struct {
	popupMsg   string
	popupStyle string
	heatDelta  float64
	triggers   []string
	solved     bool
	score      float64
}

// evaluateRun compares playerPulses to the target trace and returns what should
// happen next. It is a pure method: it reads model state but does not mutate it.
func (m Puzzle) evaluateRun(playerPulses []scope.Pulse) runEval {
	targetPulses, _ := m.PuzzleData.LoadTrace()
	result := diff.Diff(playerPulses, targetPulses)

	popMsg := m.PuzzleData.GetDialogue(m.HelpLevel.String(), result.HintKey)
	popStyle := "run"
	var triggers []string
	solved := false

	if result.Category == "exact" {
		popStyle = "solved"
		solved = true
	}

	if step := m.matchStep("run", result.Category); step != nil {
		if step.Message != "" {
			popMsg = step.Message
		}
		if step.Style != "" {
			popStyle = step.Style
		}
		if step.Trigger != "" {
			triggers = append(triggers, step.Trigger)
		}
	}

	return runEval{
		popupMsg:   popMsg,
		popupStyle: popStyle,
		heatDelta:  heat.HeatDeltaForCategory(result.Category),
		triggers:   triggers,
		solved:     solved,
		score:      result.Score,
	}
}

// handleRunCode starts execution of the player's code.
func (m Puzzle) handleRunCode(_ RunCodeMsg) (Puzzle, tea.Cmd) {
	if m.running {
		return m, nil
	}
	m.running = true
	lang := languages[m.LangIndex]
	var code string
	if m.HelpLevel == save.HelpOpen {
		code = m.choicePanel.ResolvedCode()
	} else {
		code = m.CodeArea.Value()
	}
	cmd := func() tea.Msg {
		pulses, err := runner.Run(lang, code, 0)
		return RunResultMsg{Pulses: pulses, Err: err}
	}
	return m, cmd
}

// handleRunResult processes the result of a code execution.
func (m Puzzle) handleRunResult(msg RunResultMsg) (Puzzle, tea.Cmd) {
	var cmds []tea.Cmd
	m.running = false
	m.runAttempts++

	if msg.Err != nil {
		popMsg := "✗  " + shortErr(msg.Err)
		if step := m.matchStep("error", ""); step != nil && step.Message != "" {
			popMsg = step.Message
		}
		m.setPopup(popMsg, "error")
		m.heatModel.Add(heat.HeatDeltaForCategory("error"))
		return m, nil
	}

	eval := m.evaluateRun(msg.Pulses)

	m.playerTrace = msg.Pulses
	m.scopeShowsPlayer = true
	m.ScopeModel.SetTrace(msg.Pulses)

	for _, trigger := range eval.triggers {
		m.applyTrigger(trigger)
	}

	if eval.solved {
		m.OpossumModel.StateFromSolve()
		m.heatModel.Set(0)
		m.solved = true
		m.ScopeModel.Victory = true
		cmds = append(cmds, scope.VictoryTick())
	} else {
		popMsg := eval.popupMsg + fmt.Sprintf("\n\nscore: %d%%", int(eval.score*100))
		m.setPopup(popMsg, eval.popupStyle)
		m.heatModel.Add(eval.heatDelta)
		m.OpossumModel.UpdateFromHeat(m.heatModel.Level())
	}

	return m, tea.Batch(cmds...)
}

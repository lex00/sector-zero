package scenes

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lex00/sector-zero/game/puzzles"
	"github.com/lex00/sector-zero/game/save"
)

// matchStep returns the highest-threshold matching LessonStep for the given event.
// Iterating in reverse means later steps (higher After) take precedence over
// earlier ones when multiple steps share the same On+Result — so "After: 2"
// overrides "After: 0" once the player has made enough attempts.
func (m Puzzle) matchStep(on, result string) *puzzles.LessonStep {
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
			if m.HelpLevel != save.HelpOpen {
				m.HelpLevel = save.HelpOpen
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
			if m.HelpLevel != save.HelpOpen {
				m.HelpLevel = save.HelpOpen
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}
			for i, blank := range m.PuzzleData.Challenge.Blanks {
				m.choicePanel.Selections[i] = blank.Correct
			}
		case "scope_ref":
			ref, err := m.PuzzleData.LoadTrace()
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
			if m.HelpLevel != save.HelpOpen {
				m.HelpLevel = save.HelpOpen
				m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
			}
		case "gate":
			m.gated = true
		}
	}
}

// handleKeyMsg processes keyboard input and returns the updated model and command.
func (m Puzzle) handleKeyMsg(msg tea.KeyMsg) (Puzzle, tea.Cmd) {
	var cmds []tea.Cmd
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
		if m.HelpLevel == save.HelpOpen {
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
		if m.HelpLevel != save.HelpOpen {
			lang := languages[m.LangIndex]
			if sc, ok := m.PuzzleData.Challenge.Template[lang]; ok {
				m.CodeArea.SetValue(sc)
			} else {
				m.CodeArea.SetValue("")
			}
		}

	case "ctrl+h":
		m.HelpLevel = (m.HelpLevel + 1) % 4
		if m.HelpLevel == save.HelpOpen {
			m.choicePanel = NewChoicePanel(m.PuzzleData.Challenge, languages[m.LangIndex], m.codeWidth(), m.codeHeight())
		}

	case "ctrl+g":
		// Switch to OPEN mode if not already there.
		if m.HelpLevel != save.HelpOpen {
			m.HelpLevel = save.HelpOpen
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
				ref, _ := m.PuzzleData.LoadTrace()
				m.ScopeModel.SetTrace(ref)
			}
			m.ScopeModel.CurrentPulse = 0
		}

	case "ctrl+p":
		m.ScopeModel.Paused = !m.ScopeModel.Paused

	default:
		if m.HelpLevel == save.HelpOpen {
			var cmd tea.Cmd
			m.choicePanel, cmd = m.choicePanel.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			var cmd tea.Cmd
			m.CodeArea, cmd = m.CodeArea.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

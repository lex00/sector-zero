package scenes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lex00/sector-zero/game/puzzles"
)

var blankRe = regexp.MustCompile(`\{(\d+)\}`)

// ChoicePanel is the fill-in-the-blank code editor used in OPEN mode.
type ChoicePanel struct {
	Spec            puzzles.ChallengeSpec
	Lang            string
	Selections      []int // current selection index per blank
	FocusedBlank    int   // index of the focused blank; -1 if none
	ActivePulseType string // set by the scene each animation tick
	Width           int
	Height          int
}

// NewChoicePanel returns an initialised ChoicePanel for the given spec and language.
// All blanks start at -1 (the "SELECT" sentinel) so the player must make a choice
// before running — substituting -1 produces the literal "SELECT" which fails to parse.
func NewChoicePanel(spec puzzles.ChallengeSpec, lang string, width, height int) ChoicePanel {
	sels := make([]int, len(spec.Blanks))
	for i := range sels {
		sels[i] = -1
	}
	focused := 0
	if len(spec.Blanks) == 0 {
		focused = -1
	}
	return ChoicePanel{
		Spec:         spec,
		Lang:         lang,
		Selections:   sels,
		FocusedBlank: focused,
		Width:        width,
		Height:       height,
	}
}

// ResolvedCode returns the template with all blanks substituted by the current selections.
func (cp ChoicePanel) ResolvedCode() string {
	tmpl, ok := cp.Spec.Template[cp.Lang]
	if !ok {
		tmpl = cp.Spec.Template["python"]
	}
	return blankRe.ReplaceAllStringFunc(tmpl, func(match string) string {
		idxStr := match[1 : len(match)-1]
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx >= len(cp.Spec.Blanks) {
			return match
		}
		blank := cp.Spec.Blanks[idx]
		sel := cp.Selections[idx]
		if sel == -1 {
			return "SELECT"
		}
		if sel < len(blank.Choices) {
			return blank.Choices[sel]
		}
		return match
	})
}

// Update handles Tab, Shift+Tab, Up/Down arrow, and number keys.
func (cp ChoicePanel) Update(msg tea.Msg) (ChoicePanel, tea.Cmd) {
	if len(cp.Spec.Blanks) == 0 {
		return cp, nil
	}
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return cp, nil
	}
	switch keyMsg.String() {
	case "tab":
		cp.FocusedBlank = (cp.FocusedBlank + 1) % len(cp.Spec.Blanks)
	case "shift+tab":
		cp.FocusedBlank = (cp.FocusedBlank - 1 + len(cp.Spec.Blanks)) % len(cp.Spec.Blanks)
	case "up", "k":
		if cp.FocusedBlank >= 0 {
			n := len(cp.Spec.Blanks[cp.FocusedBlank].Choices)
			cur := cp.Selections[cp.FocusedBlank]
			if cur == -1 {
				cp.Selections[cp.FocusedBlank] = n - 1 // wrap from SELECT to last
			} else {
				cp.Selections[cp.FocusedBlank] = (cur - 1 + n) % n
			}
		}
	case "down", "j":
		if cp.FocusedBlank >= 0 {
			n := len(cp.Spec.Blanks[cp.FocusedBlank].Choices)
			cur := cp.Selections[cp.FocusedBlank]
			if cur == -1 {
				cp.Selections[cp.FocusedBlank] = 0 // SELECT → first choice
			} else {
				cp.Selections[cp.FocusedBlank] = (cur + 1) % n
			}
		}
	default:
		// Number keys 1–9 for direct selection.
		if len(keyMsg.Runes) == 1 {
			d := int(keyMsg.Runes[0] - '1')
			if cp.FocusedBlank >= 0 && d >= 0 && d < len(cp.Spec.Blanks[cp.FocusedBlank].Choices) {
				cp.Selections[cp.FocusedBlank] = d
			}
		}
	}
	return cp, nil
}

// View renders the template with inline choice widgets and a dropdown for the focused blank.
func (cp ChoicePanel) View() string {
	if cp.Width <= 0 {
		cp.Width = 60
	}
	if cp.Height <= 0 {
		cp.Height = 16
	}

	innerW := cp.Width - 2
	if innerW < 1 {
		innerW = 1
	}

	unfocusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)
	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("11")).
		Bold(true)
	selectStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	selectFocusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("1")).
		Italic(true)
	activeGutter := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("▶ ")
	inactiveGutter := "  "

	// Render the template line-by-line, substituting {N} markers and adding
	// a gutter indicator on the line that matches the current animation pulse.
	activePattern := probeCallForPulse(cp.ActivePulseType)
	rawLines := strings.Split(cp.currentTemplate(), "\n")
	renderedLines := make([]string, len(rawLines))
	for i, rawLine := range rawLines {
		// Substitute {N} markers on this line.
		substituted := blankRe.ReplaceAllStringFunc(rawLine, func(match string) string {
			idxStr := match[1 : len(match)-1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil || idx >= len(cp.Spec.Blanks) {
				return match
			}
			blank := cp.Spec.Blanks[idx]
			sel := cp.Selections[idx]
			if sel == -1 {
				if idx == cp.FocusedBlank {
					return selectFocusedStyle.Render("[SELECT]")
				}
				return selectStyle.Render("[SELECT]")
			}
			choice := "?"
			if sel < len(blank.Choices) {
				choice = blank.Choices[sel]
			}
			if idx == cp.FocusedBlank {
				return focusedStyle.Render("[" + choice + "]")
			}
			return unfocusedStyle.Render("[" + choice + "]")
		})
		// Add gutter: highlight the line whose raw text contains the active probe call.
		if activePattern != "" && strings.Contains(rawLine, activePattern) {
			renderedLines[i] = activeGutter + substituted
		} else {
			renderedLines[i] = inactiveGutter + substituted
		}
	}
	rendered := strings.Join(renderedLines, "\n")

	// Dropdown for the focused blank.
	dropdown := cp.renderDropdown()
	dropdownLines := strings.Count(dropdown, "\n") + 1
	if dropdown == "" {
		dropdownLines = 0
	}

	codeH := cp.Height - 4 - dropdownLines
	if codeH < 1 {
		codeH = 1
	}

	codeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Width(innerW).
		Height(codeH).
		Render(rendered)

	if dropdown == "" {
		return codeBox
	}

	dropBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("11")).
		Width(innerW).
		Render(dropdown)

	return lipgloss.JoinVertical(lipgloss.Left, codeBox, dropBox)
}

func (cp ChoicePanel) currentTemplate() string {
	tmpl, ok := cp.Spec.Template[cp.Lang]
	if !ok {
		tmpl = cp.Spec.Template["python"]
	}
	return tmpl
}

// probeCallForPulse maps a pulse type to the probe method name used in templates.
func probeCallForPulse(pulseType string) string {
	switch pulseType {
	case "init":
		return "p.init"
	case "compare":
		return "p.compare"
	case "swap":
		return "p.swap"
	case "done":
		return "p.done"
	case "access":
		return "p.access"
	case "found":
		return "p.found"
	case "not_found":
		return "p.not_found"
	case "bounds":
		return "p.bounds"
	case "pin":
		return "p.pin"
	case "signal":
		return "p.signal"
	}
	return ""
}

func (cp ChoicePanel) renderDropdown() string {
	if cp.FocusedBlank < 0 || cp.FocusedBlank >= len(cp.Spec.Blanks) {
		return ""
	}
	blank := cp.Spec.Blanks[cp.FocusedBlank]

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	selectLineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	sel := cp.Selections[cp.FocusedBlank]

	var sb strings.Builder
	sb.WriteString(labelStyle.Render(fmt.Sprintf("blank %d — %s", cp.FocusedBlank+1, blank.Label)))
	sb.WriteRune('\n')
	// Show the sentinel "select" line when no choice has been made yet.
	if sel == -1 {
		sb.WriteString(selectLineStyle.Render("▶ — select —"))
		sb.WriteRune('\n')
	}
	for i, c := range blank.Choices {
		if sel != -1 && i == sel {
			sb.WriteString(selectedStyle.Render("▶ " + c))
		} else {
			sb.WriteString(normalStyle.Render("  " + c))
		}
		if i < len(blank.Choices)-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

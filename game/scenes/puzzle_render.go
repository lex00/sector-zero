package scenes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/lex00/sector-zero/game/save"
)

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
	hlvl := m.HelpLevel.String()
	helpDisplay := lipgloss.NewStyle().
		Foreground(lipgloss.Color(helpColour[hlvl])).
		Bold(true).
		Render(hlvl)

	// Heat bar.
	barWidth := 20
	heatBar := m.heatModel.HeatBar(barWidth)
	heatColour := heatBarColour(m.heatModel.Stage())
	heatDisplay := lipgloss.NewStyle().Foreground(lipgloss.Color(heatColour)).Render("▐" + heatBar + "▌")

	// Fuses.
	fuseDisplay := m.heatModel.FuseDisplay()

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
	if m.HelpLevel == save.HelpOpen {
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
	if m.HelpLevel == save.HelpOpen {
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
		borderColor = lipgloss.Color("10") // bright green
		header = "★  SOLVED  ★"
		msgColor = lipgloss.Color("10")
	case "error":
		borderColor = lipgloss.Color("9") // red
		header = "ERROR"
		msgColor = lipgloss.Color("9")
	case "guide":
		borderColor = lipgloss.Color("14") // cyan
		header = "GUIDE"
		msgColor = lipgloss.Color("15")
	default: // "run"
		borderColor = lipgloss.Color("13") // magenta
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

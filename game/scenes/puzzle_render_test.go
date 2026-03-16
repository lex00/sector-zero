package scenes

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lex00/sector-zero/game/puzzles"
)

func newTestPuzzleSize(w, h int) Puzzle {
	pz := puzzles.GetPuzzle(1)
	p := NewPuzzle(w, h, pz, "SIGNAL", 0, 3)
	p, _ = p.Init()
	// Simulate WindowSizeMsg so stored CodeArea gets proper dimensions.
	p, _ = p.Update(tea.WindowSizeMsg{Width: w, Height: h})
	// Dismiss any load popup so keypresses go to the textarea.
	if p.popup != nil {
		p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	}
	return p
}

func TestCodePanelVisibleInSplitView(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 50}} {
		w, h := size[0], size[1]
		p := newTestPuzzleSize(w, h)

		// Type some text.
		for _, r := range "hello" {
			p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}

		view := p.View()
		plain := trimANSI(view)

		if !containsRaw(plain, "hello") {
			t.Errorf("terminal %dx%d: 'hello' not visible in View output", w, h)
			// Show the code panel section.
			lines := strings.Split(plain, "\n")
			codeStart := -1
			for i, l := range lines {
				if strings.Contains(l, "CODE") {
					codeStart = i
					break
				}
			}
			if codeStart >= 0 {
				end := codeStart + 15
				if end > len(lines) {
					end = len(lines)
				}
				t.Logf("CODE panel section:\n%s", strings.Join(lines[codeStart:end], "\n"))
			}
		} else {
			t.Logf("terminal %dx%d: OK - 'hello' visible", w, h)
		}
	}
}

func TestCodePanelDimensions(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		w, h := size[0], size[1]
		p := newTestPuzzleSize(w, h)

		cw := p.codeWidth()
		ch := p.codeHeight()
		taH := ch - 4
		taW := cw - 2

		t.Logf("terminal %dx%d: codeWidth=%d codeHeight=%d textareaW=%d textareaH=%d",
			w, h, cw, ch, taW, taH)

		if taH <= 0 {
			t.Errorf("terminal %dx%d: textarea height %d <= 0", w, h, taH)
		}
		if taW <= 0 {
			t.Errorf("terminal %dx%d: textarea width %d <= 0", w, h, taW)
		}
	}
}

package scenes

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lex00/sector-zero/game/puzzles"
)

// sendKey simulates a printable key press through the puzzle Update.
func sendKey(m Puzzle, r rune) Puzzle {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	m, _ = m.Update(msg)
	return m
}

func newTestPuzzle() Puzzle {
	pz := puzzles.GetPuzzle(1)
	p := NewPuzzle(120, 40, pz, "SIGNAL", 0, 3)
	p, _ = p.Init()
	// Dismiss any load popup so keypresses go to the textarea.
	if p.popup != nil {
		p = sendKey(p, ' ')
	}
	return p
}

func TestFocusOnInit(t *testing.T) {
	p := newTestPuzzle()
	if !p.CodeArea.Focused() {
		t.Fatal("textarea should be focused after Init, but it is not")
	}
}

func TestTypingInsertsCharacters(t *testing.T) {
	p := newTestPuzzle()
	p = sendKey(p, 'a')
	p = sendKey(p, 'b')
	p = sendKey(p, 'c')

	got := p.CodeArea.Value()
	if got != "abc" {
		t.Fatalf("expected value %q after typing 'abc', got %q", "abc", got)
	}
}

func TestTypingVisibleInView(t *testing.T) {
	p := newTestPuzzle()
	p = sendKey(p, 'a')
	p = sendKey(p, 'b')
	p = sendKey(p, 'c')

	view := p.View()
	if !contains(view, "abc") {
		t.Fatalf("expected 'abc' to appear in puzzle View() output after typing, but it did not.\nView output (trimmed):\n%s", trimANSI(view))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRaw(s, sub))
}

func containsRaw(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// trimANSI strips ANSI escape codes for readable test output.
func trimANSI(s string) string {
	var out []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
			continue
		}
		out = append(out, s[i])
		i++
	}
	return string(out)
}

func TestFocusRouting(t *testing.T) {
	p := newTestPuzzle()

	if p.focus != FocusCode {
		t.Fatalf("expected initial focus %d (FocusCode), got %d", FocusCode, p.focus)
	}

	// Tab away and back.
	tab := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\t'}}
	tab.Type = tea.KeyTab
	for i := 0; i < 3; i++ {
		p, _ = p.Update(tab)
	}

	if p.focus != FocusCode {
		t.Fatalf("expected focus back at FocusCode after 3 tabs, got %d", p.focus)
	}
	if !p.CodeArea.Focused() {
		t.Fatal("textarea should be focused after tabbing back to FocusCode")
	}
}

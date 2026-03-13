package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lex00/sector-zero/game/puzzles"
	"github.com/lex00/sector-zero/game/save"
	"github.com/lex00/sector-zero/game/scenes"
)

// topModel is the root Bubble Tea model that owns scene routing.
type topModel struct {
	scene  string // "intro"|"puzzle"|"boom"|"reboot"
	width  int
	height int

	sv     save.Save
	puzzle scenes.Puzzle
	intro  scenes.Intro
	boom   scenes.Boom
	reboot scenes.Reboot
}

func initialModel() topModel {
	sv, err := save.Load()
	if err != nil {
		sv = save.Default()
	}

	return topModel{
		scene: "intro",
		sv:    sv,
		intro: scenes.NewIntro(0, 0), // dimensions updated on first WindowSizeMsg
	}
}

func (m topModel) Init() tea.Cmd {
	_, cmd := m.intro.Init()
	return cmd
}

func (m topModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Pass resize to active sub-model.
		switch m.scene {
		case "intro":
			m.intro, _ = m.intro.Update(msg)
		case "puzzle":
			m.puzzle, _ = m.puzzle.Update(msg)
		case "boom":
			m.boom, _ = m.boom.Update(msg)
		case "reboot":
			m.reboot, _ = m.reboot.Update(msg)
		}
		return m, nil

	// ─── Scene transitions ─────────────────────────────────────────────────

	case scenes.TransitionMsg:
		return m.transitionTo(msg.Scene)

	case scenes.RebootMsg:
		_ = msg
		return m.transitionTo("reboot")

	case scenes.RebootDoneMsg:
		_ = msg
		// After reboot: restore puzzle with reduced heat and one fewer fuse.
		return m.transitionTo("puzzle")

	case scenes.SolvedMsg:
		// Mark puzzle complete in save, advance to next puzzle.
		m.sv.Completed = appendUnique(m.sv.Completed, msg.PuzzleID)
		next := msg.PuzzleID + 1
		if next > len(puzzles.All()) {
			next = 1 // loop back for now
		}
		m.sv.CurrentPuzzle = next
		_ = m.sv.Write()
		return m.transitionTo("puzzle")

	case scenes.PuzzleQuitMsg:
		_ = msg
		// Save state and quit.
		m.sv.Heat = m.puzzle.HeatModel.Level()
		m.sv.FusesRemaining = m.puzzle.HeatModel.FuseCount()
		m.sv.CurrentPuzzle = m.puzzle.PuzzleData.ID
		m.sv.HelpLevel = helpLevelString(m.puzzle.HelpLevel)
		_ = m.sv.Write()
		return m, tea.Quit
	}

	// ─── Route to active sub-model ─────────────────────────────────────────

	var cmd tea.Cmd
	switch m.scene {
	case "intro":
		m.intro, cmd = m.intro.Update(msg)
	case "puzzle":
		m.puzzle, cmd = m.puzzle.Update(msg)
	case "boom":
		m.boom, cmd = m.boom.Update(msg)
	case "reboot":
		m.reboot, cmd = m.reboot.Update(msg)
	}
	return m, cmd
}

func (m topModel) View() string {
	switch m.scene {
	case "intro":
		return m.intro.View()
	case "puzzle":
		return m.puzzle.View()
	case "boom":
		return m.boom.View()
	case "reboot":
		return m.reboot.View()
	default:
		return "loading..."
	}
}

// transitionTo switches to a new scene and returns the initialised model + cmd.
func (m topModel) transitionTo(scene string) (tea.Model, tea.Cmd) {
	m.scene = scene
	switch scene {
	case "intro":
		intro := scenes.NewIntro(m.width, m.height)
		var cmd tea.Cmd
		m.intro, cmd = intro.Init()
		return m, cmd

	case "puzzle":
		pz := puzzles.GetPuzzle(m.sv.CurrentPuzzle)
		fuses := m.sv.FusesRemaining
		if fuses <= 0 {
			fuses = 3
		}
		pzScene := scenes.NewPuzzle(
			m.width, m.height,
			pz,
			m.sv.HelpLevel,
			m.sv.Heat,
			fuses,
		)
		var cmd tea.Cmd
		m.puzzle, cmd = pzScene.Init()
		return m, cmd

	case "boom":
		boom := scenes.NewBoom(m.width, m.height)
		var cmd tea.Cmd
		m.boom, cmd = boom.Init()
		return m, cmd

	case "reboot":
		// Burn a fuse.
		m.puzzle.HeatModel.Boom()
		m.sv.FusesRemaining = m.puzzle.HeatModel.FuseCount()
		m.sv.Heat = m.puzzle.HeatModel.Level()
		_ = m.sv.Write()

		reboot := scenes.NewReboot(m.width, m.height)
		var cmd tea.Cmd
		m.reboot, cmd = reboot.Init()
		return m, cmd
	}

	return m, nil
}

func helpLevelString(idx int) string {
	levels := []string{"BLACKOUT", "STATIC", "SIGNAL", "OPEN"}
	if idx >= 0 && idx < len(levels) {
		return levels[idx]
	}
	return "SIGNAL"
}

func appendUnique(slice []int, val int) []int {
	for _, v := range slice {
		if v == val {
			return slice
		}
	}
	return append(slice, val)
}

func main() {
	m := initialModel()

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

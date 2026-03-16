package scope

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Braille unicode block starts at U+2800.
// Cell layout (2 columns × 4 rows of dots):
//
//	col0  col1
//	dot1  dot4   (bit 0x01, 0x08)
//	dot2  dot5   (bit 0x02, 0x10)
//	dot3  dot6   (bit 0x04, 0x20)
//	dot7  dot8   (bit 0x40, 0x80)
const brailleBase = 0x2800

// dot bit positions: [row][col]
var dotBit = [4][2]byte{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

// brailleChar returns the braille character for a given 8-bit dot mask.
func brailleChar(mask byte) rune {
	return rune(brailleBase + int(mask))
}

// tagColor maps a highlight tag to a terminal color.
func tagColor(tag string) lipgloss.Color {
	switch tag {
	case "compare":
		return lipgloss.Color("11") // yellow — inspecting
	case "swap":
		return lipgloss.Color("208") // orange — moving
	case "pivot":
		return lipgloss.Color("13") // magenta — pivot marker
	case "found":
		return lipgloss.Color("10") // green — match
	case "access":
		return lipgloss.Color("14") // cyan — read
	case "left", "right", "split":
		return lipgloss.Color("12") // blue — partition boundary
	case "mid", "merge":
		return lipgloss.Color("13") // magenta — midpoint / merge marker
	case "write":
		return lipgloss.Color("14") // cyan — value being written back
	default:
		return lipgloss.Color("") // no override
	}
}

// RenderBars renders a bar chart as braille characters with per-tag coloring.
//
// values: normalised bar heights in [0.0, 1.0].
// labels: raw integer values shown as text in the bottom row (nil = no labels).
// highlights: index → highlight tag ("compare", "swap", "pivot", "found", "access").
// width, height: panel size in terminal character cells.
//
// Returns a slice of strings (with embedded ANSI color), one per terminal row.
func RenderBars(values []float64, labels []int, highlights map[int]string, width, height int) []string {
	if len(values) == 0 || width <= 0 || height <= 0 {
		empty := make([]string, height)
		for i := range empty {
			empty[i] = strings.Repeat(" ", width)
		}
		return empty
	}

	// Each braille cell is 2 dot-columns wide and 4 dot-rows tall.
	dotCols := width * 2
	dotRows := height * 4

	// Assign dot-columns to each bar.
	barWidth := dotCols / len(values)
	if barWidth < 2 {
		barWidth = 2
	}

	// Build a dot grid [dotRow][dotCol] = filled (true/false).
	grid := make([][]bool, dotRows)
	for r := range grid {
		grid[r] = make([]bool, dotCols)
	}

	for idx, v := range values {
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		filledDots := int(v * float64(dotRows))
		startCol := idx * barWidth

		for dc := startCol; dc < startCol+barWidth && dc < dotCols; dc++ {
			for dr := dotRows - filledDots; dr < dotRows; dr++ {
				grid[dr][dc] = true
			}
		}
	}

	// Render grid into braille characters (plain, no color yet).
	rows := make([]string, height)
	for cellRow := 0; cellRow < height; cellRow++ {
		var sb strings.Builder
		for cellCol := 0; cellCol < width; cellCol++ {
			var mask byte
			for dr := 0; dr < 4; dr++ {
				for dc := 0; dc < 2; dc++ {
					dotR := cellRow*4 + dr
					dotC := cellCol*2 + dc
					if dotR < dotRows && dotC < dotCols && grid[dotR][dotC] {
						mask |= dotBit[dr][dc]
					}
				}
			}
			sb.WriteRune(brailleChar(mask))
		}
		rows[cellRow] = sb.String()
	}

	// Overlay integer labels on the bottom row (plain text, before ANSI coloring).
	if len(labels) == len(values) && height > 0 {
		labelRunes := make([]rune, width)
		for i := range labelRunes {
			labelRunes[i] = ' '
		}
		for idx, v := range labels {
			startCell := (idx * barWidth) / 2
			endCell := ((idx + 1) * barWidth) / 2
			cellW := endCell - startCell
			if cellW <= 0 || startCell >= width {
				continue
			}
			s := strconv.Itoa(v)
			if len(s) > cellW {
				continue
			}
			col := startCell + (cellW-len(s))/2
			for j, ch := range s {
				if col+j < width {
					labelRunes[col+j] = ch
				}
			}
		}
		rows[height-1] = string(labelRunes)
	}

	// Build a per-cell-column color map from highlights.
	// Also mark the top cap of each highlighted bar with a full braille char.
	cellColor := make(map[int]lipgloss.Color)
	if highlights != nil {
		for idx, tag := range highlights {
			if idx < 0 || idx >= len(values) {
				continue
			}
			color := tagColor(tag)
			barDotColStart := idx * barWidth
			startCell := barDotColStart / 2
			endCell := (barDotColStart + barWidth) / 2
			for c := startCell; c < endCell && c < width; c++ {
				cellColor[c] = color
			}

			// Replace the top cell of the bar with a full-dot cap.
			v := values[idx]
			if v < 0 {
				v = 0
			}
			if v > 1 {
				v = 1
			}
			filledDots := int(v * float64(dotRows))
			topDotRow := dotRows - filledDots
			if topDotRow >= dotRows {
				continue
			}
			topCellRow := topDotRow / 4
			cellCol := barDotColStart / 2
			if topCellRow >= 0 && topCellRow < height && cellCol >= 0 && cellCol < width {
				runes := []rune(rows[topCellRow])
				if cellCol < len(runes) {
					runes[cellCol] = brailleChar(0xFF)
					rows[topCellRow] = string(runes)
				}
			}
		}
	}

	// Apply color to each row by grouping consecutive same-colored cells.
	if len(cellColor) > 0 {
		for r, row := range rows {
			runes := []rune(row)
			var out strings.Builder
			var group strings.Builder
			var lastColor lipgloss.Color

			flush := func() {
				if group.Len() == 0 {
					return
				}
				seg := group.String()
				group.Reset()
				if lastColor != "" {
					out.WriteString(lipgloss.NewStyle().Foreground(lastColor).Render(seg))
				} else {
					out.WriteString(seg)
				}
			}

			for cellCol, ch := range runes {
				c := cellColor[cellCol] // "" if not highlighted
				if c != lastColor {
					flush()
					lastColor = c
				}
				group.WriteRune(ch)
			}
			flush()
			rows[r] = out.String()
		}
	}

	return rows
}

// RenderBinary renders a binary search visualisation.
// low, high, mid: current search bounds/midpoint indices.
// target: value being searched for.
// size: array size.
// width, height: panel dimensions in terminal cells.
func RenderBinary(low, high, mid, target, size, width, height int) []string {
	if size <= 0 || width <= 0 || height <= 0 {
		empty := make([]string, height)
		for i := range empty {
			empty[i] = strings.Repeat(" ", width)
		}
		return empty
	}

	// Build normalised values: all at 0.5 except mid at 1.0.
	values := make([]float64, size)
	highlights := make(map[int]string)
	for i := range values {
		switch {
		case i == mid:
			values[i] = 1.0
			highlights[i] = "pivot"
		case i >= low && i <= high:
			values[i] = 0.6
		default:
			values[i] = 0.2
		}
	}
	if target >= 0 && target < size {
		highlights[target] = "found"
	}

	return RenderBars(values, nil, highlights, width, height)
}

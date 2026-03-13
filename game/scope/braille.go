package scope

import (
	"strings"
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

// RenderBars renders a bar chart as braille characters.
//
// values: normalised bar heights in [0.0, 1.0].
// highlights: index → highlight tag ("compare", "swap", "pivot", "found", "access").
// width, height: panel size in terminal character cells.
//
// Returns a slice of strings, one per terminal row.
func RenderBars(values []float64, highlights map[int]string, width, height int) []string {
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

	// Render grid into braille characters.
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

	// Overlay highlight caps on top of highlighted bars.
	if highlights != nil {
		for idx, tag := range highlights {
			if idx < 0 || idx >= len(values) {
				continue
			}
			_ = tag
			// Mark the top cell of the highlighted bar with a dense braille char.
			// Find the topmost filled dot row for this bar.
			barDotColStart := idx * barWidth
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
				// Replace that cell with a full braille char (all dots on).
				runes := []rune(rows[topCellRow])
				if cellCol < len(runes) {
					runes[cellCol] = brailleChar(0xFF)
					rows[topCellRow] = string(runes)
				}
			}
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

	return RenderBars(values, highlights, width, height)
}

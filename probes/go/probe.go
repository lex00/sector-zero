// Package probe is the Go probe library for the Sector Zero puzzle game.
// It instruments algorithms and emits NDJSON pulse events to os.Stdout.
//
// Usage:
//
//	p := probe.New()
//	p.Init("arr", []int{64, 34, 25, 12, 22, 11, 90})
//	p.Compare("arr", 0, 1)
//	p.Swap("arr", 0, 1)
//	p.Done("arr")
package probe

import (
	"encoding/json"
	"fmt"
	"os"
)

// Probe tracks array state and emits NDJSON pulses to os.Stdout.
type Probe struct {
	state map[string][]int
}

// New creates a new Probe instance.
func New() *Probe {
	return &Probe{state: make(map[string][]int)}
}

func (p *Probe) emit(fields map[string]any) {
	// Ensure version field is present and first by building the JSON manually.
	type pulse struct {
		V   int    `json:"v"`
		Type string `json:"type"`
	}
	// We marshal the full fields map but need "v" first. Use ordered approach.
	fields["v"] = 1
	b, err := json.Marshal(fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "probe: marshal error: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stdout, "%s\n", b)
}

// Init declares a named array with initial values.
func (p *Probe) Init(net string, values []int) {
	cp := make([]int, len(values))
	copy(cp, values)
	p.state[net] = cp
	p.emit(map[string]any{
		"type":   "init",
		"net":    net,
		"values": values,
	})
}

// Compare signals that positions i and j are being compared. Does not mutate state.
func (p *Probe) Compare(net string, i, j int) {
	p.emit(map[string]any{
		"type": "compare",
		"net":  net,
		"i":    i,
		"j":    j,
	})
}

// Swap signals that positions i and j are being swapped and updates internal state.
func (p *Probe) Swap(net string, i, j int) {
	if arr, ok := p.state[net]; ok {
		arr[i], arr[j] = arr[j], arr[i]
	}
	p.emit(map[string]any{
		"type": "swap",
		"net":  net,
		"i":    i,
		"j":    j,
	})
}

// Pin attaches a named cursor to a position.
func (p *Probe) Pin(net, name string, pos int) {
	p.emit(map[string]any{
		"type": "pin",
		"net":  net,
		"name": name,
		"pos":  pos,
	})
}

// Signal emits a named highlight over a set of positions.
func (p *Probe) Signal(net, name string, positions []int) {
	p.emit(map[string]any{
		"type":      "signal",
		"net":       net,
		"name":      name,
		"positions": positions,
	})
}

// Access signals a single read at pos.
func (p *Probe) Access(net string, pos int) {
	p.emit(map[string]any{
		"type": "access",
		"net":  net,
		"pos":  pos,
	})
}

// Found signals that the search target was found at pos.
func (p *Probe) Found(net string, pos int) {
	p.emit(map[string]any{
		"type": "found",
		"net":  net,
		"pos":  pos,
	})
}

// NotFound signals that the search target was not found.
func (p *Probe) NotFound(net string) {
	p.emit(map[string]any{
		"type": "not_found",
		"net":  net,
	})
}

// Bounds signals the current search window [low, high].
func (p *Probe) Bounds(net string, low, high int) {
	p.emit(map[string]any{
		"type": "bounds",
		"net":  net,
		"low":  low,
		"high": high,
	})
}

// Split signals a divide step: subarray [left, right) split at mid.
func (p *Probe) Split(net string, left, mid, right int) {
	p.emit(map[string]any{
		"type":  "split",
		"net":   net,
		"left":  left,
		"mid":   mid,
		"right": right,
	})
}

// Merge signals a merge step: two subarrays being merged into [left, right).
func (p *Probe) Merge(net string, left, mid, right int) {
	p.emit(map[string]any{
		"type":  "merge",
		"net":   net,
		"left":  left,
		"mid":   mid,
		"right": right,
	})
}

// Done signals that all operations on this net are complete.
func (p *Probe) Done(net string) {
	p.emit(map[string]any{
		"type": "done",
		"net":  net,
	})
}

// State returns a copy of the current tracked values for net (useful for debugging).
func (p *Probe) State(net string) []int {
	arr, ok := p.state[net]
	if !ok {
		return nil
	}
	cp := make([]int, len(arr))
	copy(cp, arr)
	return cp
}

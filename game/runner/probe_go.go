// +build ignore

// Minimal Go probe — bundled with game binary and extracted at runtime.
// Players import this as "probe" from their code directory.
package main

import (
	"encoding/json"
	"fmt"
)

// Probe tracks algorithm execution and emits NDJSON pulses to stdout.
type Probe struct {
	state map[string][]int
}

// NewProbe creates a new Probe.
func NewProbe() *Probe {
	return &Probe{state: make(map[string][]int)}
}

func emit(v interface{}) {
	b, _ := json.Marshal(v)
	fmt.Println(string(b))
}

// Init records the initial array state.
func (p *Probe) Init(net string, values []int) {
	cp := make([]int, len(values))
	copy(cp, values)
	p.state[net] = cp
	emit(map[string]interface{}{"v": 1, "type": "init", "net": net, "values": values})
}

// Compare records a comparison between positions i and j.
func (p *Probe) Compare(net string, i, j int) {
	emit(map[string]interface{}{"v": 1, "type": "compare", "net": net, "i": i, "j": j})
}

// Swap swaps positions i and j and records the event.
func (p *Probe) Swap(net string, i, j int) {
	if s, ok := p.state[net]; ok {
		if i < len(s) && j < len(s) {
			s[i], s[j] = s[j], s[i]
		}
	}
	emit(map[string]interface{}{"v": 1, "type": "swap", "net": net, "i": i, "j": j})
}

// Pin marks a named pointer at a position.
func (p *Probe) Pin(net, name string, pos int) {
	emit(map[string]interface{}{"v": 1, "type": "pin", "net": net, "name": name, "pos": pos})
}

// Signal records a named signal over a set of positions.
func (p *Probe) Signal(net, name string, positions []int) {
	emit(map[string]interface{}{"v": 1, "type": "signal", "net": net, "name": name, "positions": positions})
}

// Access records a single element access.
func (p *Probe) Access(net string, pos int) {
	emit(map[string]interface{}{"v": 1, "type": "access", "net": net, "pos": pos})
}

// Found records a successful search result.
func (p *Probe) Found(net string, pos int) {
	emit(map[string]interface{}{"v": 1, "type": "found", "net": net, "pos": pos})
}

// NotFound records a failed search.
func (p *Probe) NotFound(net string) {
	emit(map[string]interface{}{"v": 1, "type": "not_found", "net": net})
}

// Bounds records search bounds.
func (p *Probe) Bounds(net string, low, high int) {
	emit(map[string]interface{}{"v": 1, "type": "bounds", "net": net, "low": low, "high": high})
}

// Done marks the end of algorithm execution on a net.
func (p *Probe) Done(net string) {
	emit(map[string]interface{}{"v": 1, "type": "done", "net": net})
}

// Split records a divide-and-conquer split.
func (p *Probe) Split(net string, left, mid, right int) {
	emit(map[string]interface{}{"v": 1, "type": "split", "net": net, "left": left, "mid": mid, "right": right})
}

// Merge records a divide-and-conquer merge.
func (p *Probe) Merge(net string, left, mid, right int) {
	emit(map[string]interface{}{"v": 1, "type": "merge", "net": net, "left": left, "mid": mid, "right": right})
}

// Write records a single element write-back.
func (p *Probe) Write(net string, pos, value int) {
	if s, ok := p.state[net]; ok {
		if pos >= 0 && pos < len(s) {
			s[pos] = value
		}
	}
	emit(map[string]interface{}{"v": 1, "type": "write", "net": net, "pos": pos, "value": value})
}

// Lower-case aliases so Python-style template choices compile in Go.
func (p *Probe) not_found(net string)             { p.NotFound(net) }
func (p *Probe) found(net string, pos int)        { p.Found(net, pos) }
func (p *Probe) split(net string, l, m, r int)    { p.Split(net, l, m, r) }
func (p *Probe) merge(net string, l, m, r int)    { p.Merge(net, l, m, r) }
func (p *Probe) write(net string, pos, value int) { p.Write(net, pos, value) }

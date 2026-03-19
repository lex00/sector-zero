# Sector Zero

> *Something is still running in the substrate. You don't know what it is. You're going to figure it out.*

Sector Zero is a gamified algorithm puzzle game for your terminal. You are stationed at a decommissioned relay post with a piece of alien hardware that is still drawing power. It runs processes you don't recognize. Your job is to study its behavior — pulse by pulse, frame by frame — and reproduce it exactly.

An opossum has been living inside the rack unit. She was here before you arrived and she has opinions about how the work is going. Watch her. When the heat climbs she gets nervous. When you blow a fuse she plays dead. When you crack a puzzle she stares into the light like she always knew you would.

No algorithm names. No hints about what you're looking at. Just the signal — and the opossum.

---

## Installation

**Homebrew (macOS / Linux):**
```sh
brew install lex00/tap/sector-zero
```

**Go install:**
```sh
go install github.com/lex00/sector-zero/game@latest
```

**Binary download:**

Grab the latest release for your platform from the [releases page](https://github.com/lex00/sector-zero/releases). Drop the binary somewhere on your `$PATH` and you're done.

---

## Quick Start

```sh
sector-zero
```

That's it. The game opens in your terminal, loads your save (or creates a fresh one), and drops you at the artifact interface. Use the arrow keys to navigate, `Enter` to select, and `?` to open the help overlay.

Your save state lives at `~/.sector-zero/save.json`. Delete it to start over.

---

## How to Play

### The Setup

Each puzzle presents an **artifact** — a black-box system executing some internal process. The artifact's behavior is recorded as a **Trace**: a stream of structured pulses that encode everything observable about its execution.

You have two jobs:

1. **Study the artifact.** Watch its Trace replay on the Scope (the visualization surface). You'll see an array rendered as Braille bar columns, named pins floating above positions, colored signals lighting up regions. Watch until you have a hypothesis about what the artifact is doing.

2. **Match the artifact.** Write an implementation in your language of choice, instrument it using the Probe library, and submit your Trace. The game compares your Trace against the reference pulse by pulse. Exact match = puzzle cleared.

### The Probe / Trace / Scope Model

```
Your code  →  Probe library  →  Trace (NDJSON)  →  Scope (visualization)
```

**Probe:** A small library (one file, no dependencies) that you import into your solution. You call its methods as your algorithm runs. The Probe writes a Trace to stdout.

**Trace:** An NDJSON stream — one JSON object per line — where each line is one observable moment (a Pulse) in the algorithm's execution. Pulses cover comparisons, swaps, named index positions, highlighted regions, and search outcomes.

**Scope:** The terminal UI. It reads a Trace and renders it as an animated visualization. You can pause, step forward, step backward, and adjust playback speed. In compare mode, your Trace plays alongside the reference and divergences are flagged in real time.

The key insight: **the Trace is the ground truth.** You don't just need to produce the right answer — you need to produce the right answer *the right way*. Same comparisons, same swaps, same structural operations, in the same order.

---

## Probe Library

The Probe API is identical across all supported languages. Write your algorithm in the built-in editor, instrument it with Probe calls, and submit. Supported languages: Python, Go, JavaScript, Ruby, Rust, Java.

The game runs your code directly from the built-in editor — no external tools needed. Select your language, write your solution, press `Enter` to submit. The game executes your code, captures the Trace, and compares it against the reference.

The Probe is automatically available in your solution; you do not need to import or install it separately. The game injects it at runtime.

### Python

```python
def solve(arr):
    p = Probe()
    p.init("arr", arr)

    # ... your algorithm here ...
    # call p.compare(), p.swap(), p.pin(), p.signal(), etc.

    p.done("arr")

solve([5, 3, 1, 4, 2])
```

### Go

```go
func main() {
    p := NewProbe()
    arr := []int{5, 3, 1, 4, 2}
    p.Init("arr", arr)
    // ...
    p.Done("arr")
}
```

### JavaScript

```js
const p = new Probe();

function solve(arr) {
    p.init("arr", arr);
    // ...
    p.done("arr");
}

solve([5, 3, 1, 4, 2]);
```

### Ruby

```ruby
p = Probe.new

def solve(arr, p)
  p.init("arr", arr)
  # ...
  p.done("arr")
end

solve([5, 3, 1, 4, 2], p)
```

### Rust

```rust
fn main() {
    let p = Probe::new();
    let mut arr = vec![5, 3, 1, 4, 2];
    p.init("arr", &arr);
    // ...
    p.done("arr");
}
```

### Java

```java
public class Solution {
    public static void main(String[] args) throws Exception {
        Probe p = new Probe();
        int[] arr = {5, 3, 1, 4, 2};
        p.init("arr", arr);
        // ...
        p.done("arr");
    }
}
```

### Probe API Reference

| Method | Description |
|--------|-------------|
| `init(net, values)` | Declare a named array with initial values |
| `compare(net, i, j)` | Signal a comparison between positions i and j |
| `swap(net, i, j)` | Signal and record a swap of positions i and j |
| `pin(net, name, pos)` | Attach a named marker to a position |
| `signal(net, name, positions)` | Highlight a named set of positions |
| `access(net, pos)` | Signal a read at a single position |
| `found(net, pos)` | Signal that a search target was found |
| `not_found(net)` | Signal that a search target was not found |
| `bounds(net, low, high)` | Signal the current search window |
| `split(net, left, mid, right)` | Signal a divide step |
| `merge(net, left, mid, right)` | Signal a merge step |
| `write(net, pos, value)` | Write a value to a single position (mutates state) |
| `done(net)` | Signal completion of all operations |

---

## Help Levels

Sector Zero lets you tune how much the game reveals. Change your help level in the settings menu at any time.

| Level | What you see |
|-------|-------------|
| `BLACKOUT` | Reference Trace plays once at full speed. No labels. No replay. You're on your own. |
| `STATIC` | Reference Trace available on demand, but no pin labels or signal annotations in the hint view. |
| `SIGNAL` | Full pin labels and signal annotations visible in the reference Trace. |
| `OPEN` | Full annotations, pause/step controls on reference Trace, heat-free hint mode. *(Default)* |

These are named after diagnostic instrument settings, not difficulty tiers. Increasing the level gives you more signal, not a simpler problem.

---

## Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate menus |
| `Enter` | Select / confirm |
| `Esc` | Back / cancel |
| `Space` | Pause / resume playback |
| `←` / `→` | Step backward / forward one pulse |
| `[` / `]` | Decrease / increase playback speed |
| `r` | Restart playback from the beginning |
| `s` | Switch between reference and your trace (compare mode) |
| `h` | Toggle help overlay |
| `?` | Open key bindings reference |
| `q` | Quit to menu |

---

## Heat

Incorrect submissions raise your heat. High heat depletes fuses (you start with three per session). Cooling happens automatically while you're studying — the artifact rewards patience.

The heat display is in the top-right corner of the Scope. If you find yourself submitting blind and repeatedly, slow down and watch the reference Trace again.

---

## Contributing

### Development Setup

**Prerequisites:** Go 1.21+, Python 3.9+ (for reference trace generation)

```sh
git clone https://github.com/lex00/sector-zero.git
cd sector-zero

# Install Go dependencies
make tidy

# Generate puzzle trace files from reference implementations
make traces

# Build and run
make run
```

### Project Structure

```
game/               Go source for the binary (bubbletea TUI)
  diff/             Trace comparison logic
  heat/             Heat mechanic
  opossum/          The opossum
  puzzles/
    data/           Compiled .trace files (generated by make traces)
  runner/           Code execution engine (all six languages)
  save/             Persistent save state
  scope/            Braille renderer and Trace parser
probes/             Per-language probe libraries (standalone packages)
  go/               probe.go
  java/             Probe.java (Maven)
  js/               probe.js (npm)
  python/           probe.py
  ruby/             probe.rb
  rust/             src/lib.rs (Cargo)
ref/                Reference implementations (Python) — source of truth for traces
spec/               trace.schema.json (JSON Schema draft-07)
docs/               Hugo site source
```

### Adding a Puzzle

1. Write a reference implementation in `ref/puzzle_name.py` using `probe.py`.
2. Run `make traces` to generate the `.trace` file.
3. Register the puzzle in `game/puzzles/puzzles.go`.
4. Add a test in `game/puzzles/puzzle_name_test.go`.

### Adding a Probe Language

1. Create `probes/<language>/` with a single-file probe implementing the full API.
2. The probe must emit valid NDJSON conforming to `spec/trace.schema.json`.
3. Add a test that validates sample output against the schema.
4. Document installation in this README.

### Running Tests

```sh
make test
```

---

## License

MIT. See [LICENSE](LICENSE).

---

*The opossum was here before you. She'll be here after.*

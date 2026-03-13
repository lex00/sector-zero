# sector-zero-probe (Python)

A probe library for the [Sector Zero](https://github.com/lex00/sector-zero) puzzle game. Instruments your algorithm and emits an NDJSON trace to stdout that the game engine replays.

## Installation

```bash
pip install .
# or just copy probe.py next to your script
```

## Usage

```python
from probe import Probe

p = Probe()
p.init("arr", [64, 34, 25, 12, 22, 11, 90])

arr = [64, 34, 25, 12, 22, 11, 90]
n = len(arr)
for i in range(n):
    p.pin("arr", "i", i)
    for j in range(n - i - 1):
        p.pin("arr", "j", j)
        p.compare("arr", j, j + 1)
        if arr[j] > arr[j + 1]:
            p.swap("arr", j, j + 1)
            arr[j], arr[j + 1] = arr[j + 1], arr[j]

p.done("arr")
```

Run your script and pipe / redirect stdout to capture the trace:

```bash
python my_sort.py > my_sort.trace
```

## API

| Method | Description |
|--------|-------------|
| `init(net, values)` | Declare array `net` with initial values. Always call first. |
| `compare(net, i, j)` | Observe a comparison between indices `i` and `j`. No state change. |
| `swap(net, i, j)` | Record a swap of indices `i` and `j`. Updates internal state. |
| `pin(net, name, pos)` | Attach a named cursor (e.g. `"i"`, `"mid"`) to position `pos`. |
| `signal(net, name, positions)` | Highlight a set of positions under a named label. |
| `access(net, pos)` | Record a single read at `pos`. |
| `found(net, pos)` | Record that the target was found at `pos`. |
| `not_found(net)` | Record that the target was not found. |
| `bounds(net, low, high)` | Record the current search window. |
| `split(net, left, mid, right)` | Record a divide step. |
| `merge(net, left, mid, right)` | Record a merge step. |
| `done(net)` | Signal completion. |

All output goes to `stdout`. Redirect to a file to capture a `.trace` for the game.

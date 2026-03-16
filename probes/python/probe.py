# sector-zero-probe — Python probe library
# Usage: from probe import Probe; p = Probe(); p.init("arr", [1,2,3]); p.compare("arr",0,1); ...
import json
import sys


class Probe:
    """Emits NDJSON pulse events to stdout for the Sector Zero puzzle game."""

    def __init__(self, out=None):
        self._state = {}
        self._out = out or sys.stdout

    def _emit(self, obj):
        obj["v"] = 1
        # Move "v" to front for readability
        ordered = {"v": obj.pop("v")}
        ordered.update(obj)
        print(json.dumps(ordered), file=self._out, flush=True)

    def init(self, net, values):
        """Declare a named array and set its initial values."""
        self._state[net] = list(values)
        self._emit({"type": "init", "net": net, "values": list(values)})

    def compare(self, net, i, j):
        """Signal that positions i and j are being compared. Does not mutate state."""
        self._emit({"type": "compare", "net": net, "i": i, "j": j})

    def swap(self, net, i, j):
        """Signal that positions i and j are being swapped. Updates internal state."""
        arr = self._state.get(net)
        if arr is not None:
            arr[i], arr[j] = arr[j], arr[i]
        self._emit({"type": "swap", "net": net, "i": i, "j": j})

    def pin(self, net, name, pos):
        """Attach a named marker (e.g. 'i', 'j', 'mid') to a position."""
        self._emit({"type": "pin", "net": net, "name": name, "pos": pos})

    def signal(self, net, name, positions):
        """Emit a named signal highlighting a set of positions."""
        self._emit({"type": "signal", "net": net, "name": name, "positions": list(positions)})

    def access(self, net, pos):
        """Signal a single-element read at pos."""
        self._emit({"type": "access", "net": net, "pos": pos})

    def found(self, net, pos):
        """Signal that the search target was found at pos."""
        self._emit({"type": "found", "net": net, "pos": pos})

    def not_found(self, net):
        """Signal that the search target was not found."""
        self._emit({"type": "not_found", "net": net})

    def bounds(self, net, low, high):
        """Signal the current search window [low, high]."""
        self._emit({"type": "bounds", "net": net, "low": low, "high": high})

    def split(self, net, left, mid, right):
        """Signal that a subarray [left, right) is being split at mid."""
        self._emit({"type": "split", "net": net, "left": left, "mid": mid, "right": right})

    def merge(self, net, left, mid, right):
        """Signal that two subarrays are being merged into [left, right)."""
        self._emit({"type": "merge", "net": net, "left": left, "mid": mid, "right": right})

    def write(self, net, pos, value):
        """Update a single element at pos and emit a write pulse."""
        arr = self._state.get(net)
        if arr is not None and 0 <= pos < len(arr):
            arr[pos] = value
        self._emit({"type": "write", "net": net, "pos": pos, "value": value})

    def done(self, net):
        """Signal that all operations on this net are complete."""
        self._emit({"type": "done", "net": net})

    def state(self, net):
        """Return a copy of the current internal state for net (for debugging)."""
        return list(self._state.get(net, []))

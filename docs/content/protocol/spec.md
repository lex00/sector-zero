---
title: "Trace Format Specification"
date: 2024-01-01
draft: false
weight: 1
---

# Trace Format Specification

Version: **1**

## Overview

A Trace is a sequence of Pulse objects encoding the complete observable behavior of an algorithm execution. Traces are serialized as **NDJSON** (Newline-Delimited JSON): one JSON object per line, UTF-8 encoded, with a Unix newline (`\n`) as the record separator.

The Trace format is the contract between the Probe (player-side instrumentation) and the Scope (game-side visualization and evaluation). The game evaluates correctness by comparing a submitted Trace against a reference Trace stored in the puzzle's data file.

---

## Versioning

Every Pulse carries a version field `"v"` as its first key. The current version is `1`.

Future versions may add new pulse types or new optional fields on existing types. The game will reject Traces where `"v"` is absent or contains an unrecognized version number.

Backward compatibility policy: new optional fields may be added to existing pulse types in minor revisions without a version bump. Removing fields or changing the semantics of existing fields requires a version bump.

---

## Encoding Rules

- One JSON object per line.
- No trailing commas. No comments. Standard JSON only.
- The `"v"` key must appear first in each object (for readability and fast rejection of unknown versions).
- The `"type"` key must appear second.
- All string values are UTF-8.
- Integer fields (`i`, `j`, `pos`, `low`, `high`, `left`, `mid`, `right`) must be non-negative integers unless otherwise noted.
- Array fields (`values`, `positions`) must be JSON arrays.

---

## Pulse Types

### `init`

Declares a named Net and sets its initial values. Must be the first Pulse referencing a given Net name. Subsequent Pulses on the same Net operate on this initial state as mutated by any intervening `swap` Pulses.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. Always `1`. |
| `type` | string | yes | Always `"init"`. |
| `net` | string | yes | The name of the Net being declared (e.g. `"arr"`). |
| `values` | array of numbers | yes | The initial values of the Net, ordered by position. |

**Example:**
```json
{"v":1,"type":"init","net":"arr","values":[5,3,1,4,2]}
```

---

### `compare`

Signals that two positions are being compared. Does not mutate the Net's state. Used to visualize the comparison step in sort algorithms or the probe step in search algorithms.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"compare"`. |
| `net` | string | yes | The Net on which the comparison occurs. |
| `i` | integer | yes | Index of the first element being compared. |
| `j` | integer | yes | Index of the second element being compared. |

**Example:**
```json
{"v":1,"type":"compare","net":"arr","i":0,"j":1}
```

---

### `swap`

Signals that two positions are being swapped. Mutates the Net's logical state: after a `swap` pulse, positions `i` and `j` hold each other's values. The Scope updates its internal representation accordingly.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"swap"`. |
| `net` | string | yes | The Net on which the swap occurs. |
| `i` | integer | yes | Index of the first element. |
| `j` | integer | yes | Index of the second element. |

**Example:**
```json
{"v":1,"type":"swap","net":"arr","i":0,"j":1}
```

---

### `pin`

Attaches a named marker to a specific position on a Net. Pins are used to track named indices as they move through the algorithm (`i`, `j`, `lo`, `hi`, `mid`, `pivot`, etc.). A subsequent `pin` pulse with the same `name` moves the pin to the new position.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"pin"`. |
| `net` | string | yes | The Net the pin is attached to. |
| `name` | string | yes | The pin's label (e.g. `"i"`, `"lo"`, `"pivot"`). |
| `pos` | integer | yes | The position (0-based index) the pin points to. |

**Example:**
```json
{"v":1,"type":"pin","net":"arr","name":"i","pos":2}
```

---

### `signal`

Emits a named highlight covering a set of positions on a Net. Signals are used to mark regions of interest: the active partition, the current search window, the subarray being merged. A subsequent `signal` pulse with the same `name` replaces the previous highlight.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"signal"`. |
| `net` | string | yes | The Net the signal is on. |
| `name` | string | yes | The signal's label (e.g. `"window"`, `"partition"`). |
| `positions` | array of integers | yes | The positions (0-based indices) covered by the signal. |

**Example:**
```json
{"v":1,"type":"signal","net":"arr","name":"window","positions":[1,2,3,4]}
```

---

### `access`

Signals a single-element read at a position. Used in search algorithms to indicate that a value at a specific index is being inspected. Does not mutate state.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"access"`. |
| `net` | string | yes | The Net being accessed. |
| `pos` | integer | yes | The position being read. |

**Example:**
```json
{"v":1,"type":"access","net":"arr","pos":5}
```

---

### `found`

Signals that a search target was located at the given position. Terminal event for successful searches.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"found"`. |
| `net` | string | yes | The Net that was searched. |
| `pos` | integer | yes | The position where the target was found. |

**Example:**
```json
{"v":1,"type":"found","net":"arr","pos":3}
```

---

### `not_found`

Signals that a search target was not present in the Net. Terminal event for unsuccessful searches.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"not_found"`. |
| `net` | string | yes | The Net that was searched. |

**Example:**
```json
{"v":1,"type":"not_found","net":"arr"}
```

---

### `bounds`

Signals the current active search window as a `[low, high]` range. Used in search algorithms that narrow a range iteratively.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"bounds"`. |
| `net` | string | yes | The Net the bounds apply to. |
| `low` | integer | yes | The lower bound of the current window (inclusive). |
| `high` | integer | yes | The upper bound of the current window (inclusive). |

**Example:**
```json
{"v":1,"type":"bounds","net":"arr","low":0,"high":9}
```

---

### `split`

Signals that a subarray `[left, right)` is being divided at `mid`. Used by divide-and-conquer algorithms to indicate the start of a recursive split. `mid` is the index at which the split occurs; the two resulting sub-ranges are `[left, mid)` and `[mid, right)`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"split"`. |
| `net` | string | yes | The Net being split. |
| `left` | integer | yes | Start index of the subarray (inclusive). |
| `mid` | integer | yes | The split point. |
| `right` | integer | yes | End index of the subarray (exclusive). |

**Example:**
```json
{"v":1,"type":"split","net":"arr","left":0,"mid":2,"right":5}
```

---

### `merge`

Signals that two adjacent subarrays are being merged into `[left, right)`. The two source sub-ranges are `[left, mid)` and `[mid, right)`. This event marks the start of a merge operation; `swap` events that follow perform the actual reordering.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"merge"`. |
| `net` | string | yes | The Net being merged into. |
| `left` | integer | yes | Start index of the left subarray (inclusive). |
| `mid` | integer | yes | Boundary between left and right subarrays. |
| `right` | integer | yes | End index of the right subarray (exclusive). |

**Example:**
```json
{"v":1,"type":"merge","net":"arr","left":0,"mid":2,"right":5}
```

---

### `write`

Writes a value to a single position on a Net. Mutates the Net's logical state: after a `write` pulse, position `pos` holds `value`. Used by algorithms that write to a scratch or output array directly (e.g. merge sort's merge step).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"write"`. |
| `net` | string | yes | The Net being written to. |
| `pos` | integer | yes | The position being written. |
| `value` | number | yes | The value being written to that position. |

**Example:**
```json
{"v":1,"type":"write","net":"arr","pos":2,"value":7}
```

---

### `done`

Signals that all operations on a Net are complete. The final state of the Net after all preceding `swap` events is the algorithm's output. Every Trace must end with a `done` pulse for each Net that was initialized.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `v` | integer | yes | Protocol version. |
| `type` | string | yes | Always `"done"`. |
| `net` | string | yes | The Net that is now complete. |

**Example:**
```json
{"v":1,"type":"done","net":"arr"}
```

---

## Example Trace: Bubble Sort on a Five-Element Array

The following is a complete, valid Trace for bubble sort on `[5, 3, 1, 4, 2]`. Pin `i` tracks the outer loop; pin `j` tracks the inner comparison cursor.

```ndjson
{"v":1,"type":"init","net":"arr","values":[5,3,1,4,2]}
{"v":1,"type":"pin","net":"arr","name":"i","pos":0}
{"v":1,"type":"pin","net":"arr","name":"j","pos":0}
{"v":1,"type":"compare","net":"arr","i":0,"j":1}
{"v":1,"type":"swap","net":"arr","i":0,"j":1}
{"v":1,"type":"pin","net":"arr","name":"j","pos":1}
{"v":1,"type":"compare","net":"arr","i":1,"j":2}
{"v":1,"type":"swap","net":"arr","i":1,"j":2}
{"v":1,"type":"pin","net":"arr","name":"j","pos":2}
{"v":1,"type":"compare","net":"arr","i":2,"j":3}
{"v":1,"type":"swap","net":"arr","i":2,"j":3}
{"v":1,"type":"pin","net":"arr","name":"j","pos":3}
{"v":1,"type":"compare","net":"arr","i":3,"j":4}
{"v":1,"type":"swap","net":"arr","i":3,"j":4}
{"v":1,"type":"pin","net":"arr","name":"i","pos":1}
{"v":1,"type":"pin","net":"arr","name":"j","pos":0}
{"v":1,"type":"compare","net":"arr","i":0,"j":1}
{"v":1,"type":"compare","net":"arr","i":1,"j":2}
{"v":1,"type":"swap","net":"arr","i":1,"j":2}
{"v":1,"type":"pin","net":"arr","name":"j","pos":1}
{"v":1,"type":"compare","net":"arr","i":2,"j":3}
{"v":1,"type":"swap","net":"arr","i":2,"j":3}
{"v":1,"type":"pin","net":"arr","name":"j","pos":2}
{"v":1,"type":"compare","net":"arr","i":3,"j":4}
{"v":1,"type":"pin","net":"arr","name":"i","pos":2}
{"v":1,"type":"pin","net":"arr","name":"j","pos":0}
{"v":1,"type":"compare","net":"arr","i":0,"j":1}
{"v":1,"type":"compare","net":"arr","i":1,"j":2}
{"v":1,"type":"compare","net":"arr","i":2,"j":3}
{"v":1,"type":"swap","net":"arr","i":2,"j":3}
{"v":1,"type":"pin","net":"arr","name":"j","pos":2}
{"v":1,"type":"compare","net":"arr","i":3,"j":4}
{"v":1,"type":"pin","net":"arr","name":"i","pos":3}
{"v":1,"type":"pin","net":"arr","name":"j","pos":0}
{"v":1,"type":"compare","net":"arr","i":0,"j":1}
{"v":1,"type":"compare","net":"arr","i":1,"j":2}
{"v":1,"type":"compare","net":"arr","i":2,"j":3}
{"v":1,"type":"pin","net":"arr","name":"i","pos":4}
{"v":1,"type":"done","net":"arr"}
```

After executing all `swap` events in the above Trace, the Net `"arr"` holds `[1, 2, 3, 4, 5]`.

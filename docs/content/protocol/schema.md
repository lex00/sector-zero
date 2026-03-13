---
title: "Schema Reference"
date: 2024-01-01
draft: false
weight: 2
---

# Trace Schema Reference

The canonical machine-readable schema for a single Pulse object lives at [`spec/trace.schema.json`](https://github.com/lex00/sector-zero/blob/main/spec/trace.schema.json) in the repository root.

This page is an annotated reference for that schema. It explains each field, documents valid values, and notes which fields are required vs. optional.

---

## Schema Overview

The schema is written in **JSON Schema draft-07**. It describes a single Pulse object using `oneOf` with discrimination on the `"type"` field. Each variant (each pulse type) has its own subschema with exactly the fields it requires.

Every valid Pulse must satisfy exactly one of the `oneOf` variants. A Pulse that matches zero variants or more than one variant is invalid.

---

## Common Fields

These fields appear on every Pulse variant.

### `v`

- **Type:** `integer`
- **Required:** yes
- **Valid values:** `1` (the only current version)
- **Description:** Protocol version number. Must be the first key in the object. Enables fast rejection of Pulses from future or unknown protocol versions without parsing the rest of the object.

### `type`

- **Type:** `string`
- **Required:** yes
- **Valid values:** `"init"`, `"compare"`, `"swap"`, `"pin"`, `"signal"`, `"access"`, `"found"`, `"not_found"`, `"bounds"`, `"split"`, `"merge"`, `"done"`
- **Description:** The pulse type. Determines which variant subschema applies and which additional fields are required. Must be the second key in the object (by convention; not enforced by the schema).

### `net`

- **Type:** `string`
- **Required:** yes (on all variants)
- **Valid values:** Any non-empty string. Conventionally short identifiers like `"arr"`, `"left"`, `"scratch"`.
- **Description:** The name of the Net this Pulse applies to. Must match the `net` value used in the preceding `init` pulse for that Net.

---

## Variant Fields

### `init` — `values`

- **Type:** `array` of `number`
- **Required:** yes
- **Valid values:** Any JSON array of numbers. May include negative numbers and decimals; the Scope normalizes for display.
- **Description:** The initial ordered contents of the Net. Position 0 is the leftmost element in the Scope visualization.

### `compare` — `i`, `j`

- **Type:** `integer` (both)
- **Required:** yes (both)
- **Valid values:** Non-negative integers less than the length of the Net.
- **Description:** The two positions being compared. `i` and `j` may be equal (a self-comparison), though this is unusual. The Scope highlights both positions during playback of this Pulse.

### `swap` — `i`, `j`

- **Type:** `integer` (both)
- **Required:** yes (both)
- **Valid values:** Non-negative integers less than the length of the Net. May be equal (a no-op swap).
- **Description:** The two positions whose values are exchanged. The Scope's internal state is updated: after this Pulse, position `i` holds the value formerly at `j` and vice versa.

### `pin` — `name`, `pos`

- **Type:** `name` is `string`; `pos` is `integer`
- **Required:** yes (both)
- **Valid values:** `name` — any non-empty string, typically a single letter or short label. `pos` — non-negative integer less than the length of the Net.
- **Description:** Attaches (or moves) the named pin to `pos`. If a pin with this `name` already exists on this Net, the Scope moves it. If no pin with this name exists, the Scope creates it.

### `signal` — `name`, `positions`

- **Type:** `name` is `string`; `positions` is `array` of `integer`
- **Required:** yes (both)
- **Valid values:** `name` — any non-empty string. `positions` — array of non-negative integers, each less than the length of the Net. May be empty (which clears the signal).
- **Description:** Sets the positions covered by the named signal. An empty `positions` array removes the signal from the display. Each signal name is rendered in a distinct color on the Scope.

### `access` — `pos`

- **Type:** `integer`
- **Required:** yes
- **Valid values:** Non-negative integer less than the length of the Net.
- **Description:** The position being read. The Scope briefly highlights this position during playback.

### `found` — `pos`

- **Type:** `integer`
- **Required:** yes
- **Valid values:** Non-negative integer less than the length of the Net.
- **Description:** The position at which the search target was found. This is a terminal event: no further Pulses on this Net are expected (other than `done`).

### `not_found` — (no additional fields)

- **Description:** No variant-specific fields. The absence of a result is itself the signal. This is a terminal event: no further Pulses on this Net are expected (other than `done`).

### `bounds` — `low`, `high`

- **Type:** `integer` (both)
- **Required:** yes (both)
- **Valid values:** Non-negative integers. `low` must be less than or equal to `high`. Both must be less than the length of the Net.
- **Description:** The current active search window. The Scope renders `[low, high]` as a highlighted range. Consecutive `bounds` Pulses show the window narrowing.

### `split` — `left`, `mid`, `right`

- **Type:** `integer` (all three)
- **Required:** yes (all three)
- **Valid values:** Non-negative integers. `left` <= `mid` < `right`. All values less than or equal to the length of the Net.
- **Description:** Marks the beginning of a divide step. The Scope visualizes the two resulting sub-ranges: `[left, mid)` and `[mid, right)`.

### `merge` — `left`, `mid`, `right`

- **Type:** `integer` (all three)
- **Required:** yes (all three)
- **Valid values:** Non-negative integers. `left` <= `mid` < `right`. All values less than or equal to the length of the Net.
- **Description:** Marks the beginning of a merge step. The Scope visualizes the two source sub-ranges being merged into `[left, right)`.

### `done` — (no additional fields)

- **Description:** No variant-specific fields. Signals the end of all operations on this Net. Every Net that receives an `init` Pulse must eventually receive a `done` Pulse. The Scope uses this to know when playback is complete and to finalize comparison results.

---

## Validation

To validate a Trace against the schema, use any JSON Schema draft-07 validator. Example using `ajv-cli`:

```sh
# Install once
npm install -g ajv-cli

# Validate each pulse in a trace file
while IFS= read -r line; do
  echo "$line" | ajv validate -s spec/trace.schema.json -d /dev/stdin
done < my_trace.ndjson
```

The game itself validates incoming Traces at load time and will display a parse error with the offending line number if validation fails.

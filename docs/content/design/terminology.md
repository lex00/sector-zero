---
title: "Terminology"
date: 2024-01-01
draft: false
weight: 2
---

# Sector Zero — Terminology

## Why Chip-Fab Language?

The artifact is not a program in any conventional sense. It does not run on a known instruction set. It does not have a filesystem, a heap, or a call stack that you can inspect with familiar tools. Whatever it is, it predates the software abstractions we use to talk about computation.

When you attach your equipment to it, you are doing what any reverse engineer does with unknown hardware: probing contacts, reading signal lines, watching for patterns on a scope. The language of chip fabrication and electronics diagnostics is the honest vocabulary for this activity. It describes what you are *actually doing* — observing electrical behavior — without importing assumptions from the software world that may not apply here.

Using terms like "array" and "function" and "variable" would feel wrong. Not because they are inaccurate (the artifact does, in some sense, operate on sequences of values), but because they presuppose a familiarity with the artifact's internals that you do not have. You are looking at it from the outside. You are reading the pins.

The chip-fab vocabulary also reinforces the game's central fiction: the artifact is advanced circuitry of unknown origin, and you are a technician reverse-engineering it. The terminology keeps you in that headspace. It makes the familiar feel alien enough to be interesting again.

---

## Glossary

### Trace

**Concept:** Event stream

**In-world meaning:** The artifact's output signal — the continuous stream of structured pulses that the artifact emits as it operates. When you attach your Scope to the artifact, you are reading its Trace. When you run your Probe, it generates a Trace that the game compares against the artifact's reference Trace.

The Trace is the ground truth of the game. Matching it exactly is the win condition.

---

### Pulse

**Concept:** Single step / single event

**In-world meaning:** One clock cycle of the artifact's mind. A Pulse is the smallest observable unit of the Trace — a single typed event representing one moment of the artifact's operation. Examples: a comparison between two positions, a swap, the movement of a pin.

The artifact's behavior is made legible one Pulse at a time. The Scope advances through Pulses frame by frame, and the player's understanding of the artifact is built by watching the sequence.

---

### Probe

**Concept:** Per-language instrumentation library

**In-world meaning:** The equipment you use to interface with the artifact — or rather, to make *your* implementation emit a signal in the same format as the artifact's. A Probe is a small library (one file, no dependencies) that you import into your solution. You call its methods as your algorithm executes. The Probe writes a Trace to stdout in the artifact's signal format.

There is a Probe for each supported language. The Probe API is identical across languages: `init`, `compare`, `swap`, `pin`, `signal`, `access`, `found`, `not_found`, `bounds`, `split`, `merge`, `done`.

---

### Scope

**Concept:** Renderer / visualization surface

**In-world meaning:** The artifact's face — the display surface that renders the Trace as an animated terminal visualization. The Scope is what you stare at. It shows the artifact's internal state evolving pulse by pulse: Braille-encoded bars for array values, floating pin labels, colored signal highlights.

The Scope has two modes: *study mode*, where you watch the reference Trace at your own pace, and *compare mode*, where your submitted Trace plays alongside the reference and divergences are flagged in real time.

---

### Net

**Concept:** Named data structure (typically an array)

**In-world meaning:** A named channel inside the artifact — an internal signal bus carrying a sequence of values. Most puzzles operate on a single Net (conventionally named `"arr"`), but more complex puzzles may involve multiple Nets (e.g., a source Net and a scratch Net for merge operations).

When you call `init` on your Probe, you are declaring a Net and setting its initial values. All subsequent Pulses reference a Net by name to specify which channel they apply to.

---

### Signal

**Concept:** Annotation / highlight

**In-world meaning:** A named energy pattern on a Net — a highlighted range of positions that carries semantic meaning in the context of the algorithm. Signals are used to mark regions of interest: the active partition in a sort, the current search window in a search, the subarray being merged.

A Signal has a name (a short string) and a list of positions. The Scope renders Signals as colored overlays on the affected positions. Unlike Pins, Signals do not point to a single position — they illuminate a region.

---

### Pin

**Concept:** Named pointer / index marker

**In-world meaning:** A contact point on a Net — a named marker attached to a single position. Pins are used to track named indices as they move through the algorithm: `i`, `j`, `lo`, `hi`, `mid`, `pivot`.

The Scope renders Pins as small floating labels above the bar at the indicated position. As a Pin moves from pulse to pulse, the label migrates across the display. Watching Pin movement is often the first step toward understanding what the artifact is doing.

---

### Run

**Concept:** Full session / submission attempt

**In-world meaning:** One attempt to match the artifact's behavior — a complete execution of your implementation from `init` to `done`. A Run produces a Trace that the game evaluates against the reference Trace.

The Heat mechanic is scoped to a Run: heat accumulates as you make incorrect submissions within a session, and cools between Runs. Fuses are consumed when heat reaches critical levels.

---

## Quick Reference Table

| Term | Concept | In-world meaning |
|------|---------|-----------------|
| Trace | Event stream | The artifact's output signal |
| Pulse | Single step/event | One clock cycle of the artifact's mind |
| Probe | Per-language lib | What the player uses to interface with the artifact |
| Scope | Renderer | The artifact's face — the visualization surface |
| Net | Named data structure | A named channel inside the artifact |
| Signal | Annotation/highlight | A named energy pattern on a net |
| Pin | Named pointer | A contact point on a net |
| Run | Full session | One attempt to match the artifact's behavior |

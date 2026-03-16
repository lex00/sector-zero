---
title: "Design Overview"
date: 2024-01-01
draft: false
weight: 1
---

# Sector Zero — Design Overview

## Concept

Sector Zero is a gamified algorithm puzzle game that runs entirely in your terminal. You are not a student. You are a technician stationed at a decommissioned relay post somewhere in the Florida panhandle, and you have found something buried in the substrate of an old rack unit — an artifact of unknown origin, still drawing power, still running.

The artifact does not respond to conventional interfaces. It speaks in a signal language of its own: pulses of activity across named internal channels, patterns that repeat with eerie consistency. Your job is to study those patterns and reproduce them — not to name them, not to understand them in any academic sense, but to *match* them.

The game's aesthetic is deliberate: green phosphor on black, monospace geometry, the low-frequency hum of hardware that has been running since before the internet. The 80s DOS terminal look is not nostalgia for its own sake — it is the correct visual register for a game about reading machine behavior at close range. The opossum that occasionally appears in the corner of your scope? She was here before you. She's not going anywhere.

## The Artifact as Visualization Surface

The artifact is the game's central fiction and its core UX metaphor at once.

Every puzzle presents you with an artifact — a black-box system executing some internal process. You cannot read its source code. You cannot ask it what it is doing. What you *can* do is attach a Scope to its output channel and watch the signal.

The Scope is the visualization surface: a full-terminal rendering of the artifact's internal state as it evolves pulse by pulse. Arrays appear as columns of Braille-encoded bars. Named pins float above positions like diagnostic probes. Signals light up ranges of the array in distinct colors. The display updates as each pulse fires.

This is not a debugger. It is closer to an oscilloscope: you observe waveforms, identify recurring signatures, and eventually reconstruct the logic that produces them — not by reading documentation, but by watching long enough.

The artifact metaphor earns its weight because it answers the question every player will ask: "Why can't I just look up the algorithm?" The artifact predates human documentation. There is no lookup. There is only the signal.

## Three-Layer Architecture

Sector Zero is built around a strict three-layer model:

```
Probe  →  Trace  →  Scope
```

### Layer 1: Probe (Player Interface)

The Probe is a small library available in the player's language of choice — Python, Go, JavaScript, Ruby, Rust, and Java. The player writes an algorithm using the Probe API: calling `init`, `compare`, `swap`, `pin`, `signal`, and related methods as their code runs. The Probe emits a stream of structured events to stdout.

The Probe API is intentionally minimal. It does not enforce any particular algorithm. It does not know what puzzle you are solving. It simply instruments your code and produces a signal.

### Layer 2: Trace (Signal Format)

The Trace is the lingua franca of the system: an NDJSON stream of typed pulse objects. Each line is one pulse — one observable moment in the artifact's execution. Pulses carry enough information to reconstruct the full internal state of the data structure at any point in time.

The Trace is the ground truth. The game does not evaluate correctness by checking your final output array. It evaluates correctness by comparing your Trace, pulse by pulse, against the reference Trace stored in the puzzle's data file. To pass a puzzle, you must not only sort the array correctly — you must sort it *the same way the artifact sorts it*.

### Layer 3: Scope (Visualization)

The Scope is the terminal UI rendered by the Go binary. It reads a Trace — either the reference Trace (for study) or your submitted Trace (for comparison) — and renders it as an animated visualization.

The Scope presents the artifact's behavior as a live display: array bars, pin labels, signal highlights, frame-by-frame playback controls. It is the window into the artifact's mind, and it is the primary feedback loop for the player.

## Core Design Philosophy: "The Visualization IS the Instruction"

The single most important design rule in Sector Zero: **algorithm names never appear in the game.**

No puzzle is titled "Bubble Sort." No hint says "try dividing the array in half." No help text mentions "pivot." The Scope shows you what the artifact does. The Scope is your only instruction manual.

This is not an artificial constraint imposed for difficulty. It reflects a genuine belief about how algorithmic intuition is built. Reading that something is a "divide and conquer" algorithm gives you a label. Watching an array split, recurse, and merge — watching it happen ten times at different speeds until the pattern is in your hands — gives you *understanding*. The label can come later, outside the game, when you recognize the pattern in the wild and finally have a name to put to it.

The mystery is the point. The artifact does not owe you an explanation. Neither does the game.

## Target Audience

Sector Zero is for developers who already know how to write code and want to build deeper algorithmic intuition — not by grinding LeetCode, but by *watching* and *imitating*.

The ideal player has written a few sorting functions, knows roughly what O(n log n) means, and has never quite internalized *why* some approaches are faster than others. They are comfortable in a terminal. They appreciate atmosphere. They do not need a hint button — but they might use the help level system when they're genuinely stuck, because the help level system is designed to feel like adjusting a diagnostic instrument, not admitting defeat.

Secondary audience: educators who want a novel, lore-driven way to introduce algorithm concepts in courses or workshops.

## Platform

Sector Zero is a single Go binary with no runtime dependencies. Players download it (or install via Homebrew) and run it in any modern terminal emulator. There is no server, no account, no cloud sync. Save state is a single JSON file at `~/.sector-zero/save.json`.

The probe libraries are separate, language-specific packages. The Python probe is a single-file library with no dependencies beyond the standard library. This keeps the barrier to entry as low as possible: download the binary, grab the probe file, start writing.

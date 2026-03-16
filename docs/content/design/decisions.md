---
title: "Design Decisions"
date: 2024-01-01
draft: false
weight: 3
---

# Sector Zero — Key Design Decisions

This document records the reasoning behind the game's most significant design choices. These are not post-hoc rationalizations — they are the constraints that shaped everything else.

---

## 1. No Algorithm Names

**Decision:** Algorithm names never appear in the game. No puzzle title, hint, or piece of UI will say "Bubble Sort," "Merge Sort," "Binary Search," or any other conventional name.

**Why:** Names short-circuit the learning process. The moment a player reads "Bubble Sort," their brain reaches for a cached description — probably something about adjacent swaps and multiple passes — and they start pattern-matching against that description rather than against the Trace in front of them.

Sector Zero is designed to build *intuition*, not to reinforce *labels*. Intuition is built by observation and imitation: watching the artifact operate, forming hypotheses about what it might be doing, testing those hypotheses by running your own implementation and comparing Traces.

The mystery is also generative. A player who has never heard of merge sort will watch the artifact split an array in half, recurse into each half, and merge the results — and they will feel genuine wonder at the first moment of recognition. That moment of recognition is the game's core reward. Naming the algorithm in advance trades that reward for a smaller one: the satisfaction of recognizing a term you already know.

Algorithm names belong to the world outside the game. Players who complete a puzzle and want a name to put to what they've learned can find one in thirty seconds. The game does not need to provide it.

---

## 2. Terminal-Only

**Decision:** Sector Zero runs in a terminal emulator. There is no web version, no GUI app, no Electron wrapper.

**Why:** Three reasons, in order of importance.

*Focus.* A terminal window is a minimal, distraction-free environment. There is no browser chrome, no notification dot, no temptation to open a new tab. The entire screen is the game. This is especially important for a game that asks the player to observe carefully — the terminal frame is a focusing device.

*Accessibility.* A single Go binary with no runtime dependencies installs everywhere. Linux, macOS, Windows WSL. No JVM, no Node, no browser. The barrier to entry is as low as it can be. The game should be a `brew install` or a `go install` away from any developer's workflow.

*Aesthetic honesty.* The 80s DOS look is not decoration — it is the game's visual register. Green phosphor on black, Braille-encoded bar graphs, monospace geometry: these are not stylistic choices bolted onto a game that could have looked like anything. They are what the game *is*. The terminal is the artifact's natural habitat. Running it in a browser would be like watching a film on your phone: technically possible, fundamentally wrong.

---

## 3. Braille Rendering

**Decision:** Array bars are rendered using Unicode Braille characters rather than ASCII art (e.g., `█`, `▄`, or `|`).

**Why:** Density and aesthetics, in equal measure.

A Braille character occupies a single terminal cell but encodes an 8-bit pattern (a 2×4 grid of dots). This gives 8× the vertical resolution of a single-cell block character. A Braille bar graph can render meaningful value differences in an array of 20+ elements in a reasonable terminal width without resorting to horizontal scrolling.

The visual result also looks genuinely alien. A column of Braille characters does not look like a conventional bar chart. It looks like something a machine produced — which is exactly right. The artifact is not making a bar chart for your benefit. It is displaying its internal state in its own visual language, and you are learning to read it.

The Braille rendering requires a terminal and font that support Unicode, which is a non-issue for any modern development environment. On systems where Braille renders poorly, the Scope can fall back to block characters — but the default is Braille, and it is the intended experience.

---

## 4. Help Level System

**Decision:** Players can choose from four help levels: BLACKOUT, STATIC, SIGNAL, and OPEN.

**Why:** Not all players have the same tolerance for mystery, and not all puzzles have the same ambient difficulty. The help level system lets players tune the experience to their current state without changing the game's fundamental structure.

- **BLACKOUT**: No hints of any kind. The reference Trace plays once at full speed and you are on your own. For players who want the full alien-artifact experience.
- **STATIC**: The Scope shows the reference Trace on demand, but no structural hints (no pin labels in the hint view, no annotated regions).
- **SIGNAL**: Pin labels are visible in the reference Trace. Signals are labeled. You can see *what* is being tracked, even if you don't know *why*.
- **OPEN**: Full annotations, pause/step controls on the reference Trace, and a heat-free hint mode. For players who are genuinely stuck or using the game in a learning context. This is the default.

The help level is a game setting, not a puzzle property. Players can change it at any time. Lowering the help level mid-puzzle awards no bonus — it is simply a tool for the player's own use.

The naming (BLACKOUT → OPEN) is intentional: it maps to the metaphor of adjusting the gain on a diagnostic instrument. You are not "getting hints." You are increasing the sensitivity of your scope.

---

## 5. Heat Mechanic

**Decision:** Each incorrect submission raises the player's heat. High heat depletes fuses. Cooling requires time between submissions.

**Why:** Without a cost to incorrect submissions, the optimal strategy is to thrash: make a random change, submit, observe the diff, adjust. This is effective but shallow — it teaches you to binary-search your way to a passing Trace without ever understanding the algorithm.

Heat creates urgency. After two or three wrong submissions in quick succession, the heat indicator starts climbing and the player is incentivized to *stop and think* before submitting again. The cooling mechanic — heat drops while the player is in study mode or has closed the submit dialog — directly rewards thoughtful observation over mechanical iteration.

Fuses are the hard limiter: a finite resource (three per session by default) that is consumed when heat reaches critical levels. Running out of fuses does not end the game, but it does lock out the submit function for the current Run, forcing the player to start a new Run with a fresh array.

The mechanic is designed to feel like hardware, not penalty: you are not being punished for being wrong, you are managing the thermal load on your diagnostic equipment. The framing matters. Players do not feel judged by heat; they feel constrained by physics.

---

## 6. Trace-as-Ground-Truth

**Decision:** Correctness is evaluated by comparing the player's full Trace against the reference Trace, pulse by pulse. A correct final array is not sufficient to pass a puzzle.

**Why:** Many different algorithms can sort an array correctly. The game is not asking "can you sort an array?" It is asking "can you *behave like the artifact*?"

This constraint is what makes the game a puzzle rather than a coding challenge. It forces the player to understand not just the output but the *process* — the specific sequence of comparisons, swaps, and structural operations that the artifact uses. Getting the right answer the wrong way is, in this game, the wrong answer.

Trace comparison also makes the feedback loop precise. When your Trace diverges from the reference, the Scope highlights the exact pulse where divergence begins and shows what you emitted versus what was expected. This is far more actionable than "your output array is wrong."

The practical implication: if a puzzle is based on a specific variant of an algorithm (e.g., a particular pivot selection strategy in quicksort), the player must match that variant exactly. This is intentional. The artifact has a specific mind. Your goal is to reproduce it.

---

## 7. Multi-Language Support

**Decision:** The Probe library is available in multiple languages: Python, Go, JavaScript, Ruby, Rust, and Java.

**Why:** The game is for developers, and developers have a primary language. Forcing every player to write Python creates a friction that has nothing to do with algorithmic understanding — it is just a language tax. A player who thinks in Go should be able to think in Go. A player who teaches JavaScript should be able to demo the game in JavaScript.

The Probe API is deliberately simple and language-agnostic: a dozen methods on a stateless object, all emitting to stdout. Implementing a new Probe is a few hours of work and requires no knowledge of the game's internals. Community contributions are welcome.

The Trace format is the contract. As long as a Probe emits valid NDJSON in the trace schema format, the game will accept it.

---

## 8. No External LSP

**Decision:** The game does not integrate with any Language Server Protocol implementation. There is no in-editor autocomplete, type checking, or documentation hover for the Probe API.

**Why:** The Probe API is small enough that it does not need it.

The full API surface is twelve methods. Each method takes two to four arguments, all of which are either strings or integers. A player can learn the entire API by reading the probe file — which is a single, short, well-commented source file in their language.

LSP integration would require installing a language server, configuring an editor plugin, and maintaining compatibility across multiple editors and language server versions. This is significant complexity for a marginal benefit. The game-integrated documentation (available in the Scope's help overlay) is sufficient for the API surface in question.

The probe file is intentionally readable. The docstrings are the documentation. If a player wants richer editor support, they can write a wrapper with their own type annotations — but the game will not require it.

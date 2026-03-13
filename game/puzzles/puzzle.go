package puzzles

import (
	"embed"
	"fmt"

	"github.com/lex00/sector-zero/game/scope"
)

//go:embed data
var puzzleData embed.FS

// Puzzle holds the metadata and content for a single game puzzle.
type Puzzle struct {
	ID           int
	Title        string
	TraceFile    string                       // path relative to puzzles/data/
	Dialogue     map[string]map[string]string // help_level → hint_key → text
	ScaffoldCode map[string]string            // lang → scaffold code for OPEN level
}

// allPuzzles is the full puzzle list.
var allPuzzles = []Puzzle{
	{
		ID:        1,
		Title:     "Bubble Sort",
		TraceFile: "data/bubble_sort.trace",
		Dialogue: map[string]map[string]string{
			"BLACKOUT": {
				"empty":             "...",
				"wrong_types":       "...",
				"wrong_order":       "...",
				"wrong_termination": "...",
				"near":              "...",
				"exact":             "...",
			},
			"STATIC": {
				"empty":             "*static*",
				"wrong_types":       "*something moves — but not like that*",
				"wrong_order":       "*closer — but the rhythm is wrong*",
				"wrong_termination": "*it flickers and dies too soon*",
				"near":              "*almost. so close.*",
				"exact":             "*the artifact glows. layer 2 unsealed.*",
			},
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*something moves — but not in the right way*",
				"wrong_order":       "*the artifact stirs — you're closer*",
				"wrong_termination": "*it runs too long. or not long enough*",
				"near":              "*the opossum's eyes widen*",
				"exact":             "*the artifact glows. layer 2 unsealed.*",
			},
			"OPEN": {
				"empty":             "The artifact waits. It needs the compare-and-swap pattern.",
				"wrong_types":       "Check your pulse types: init, compare, swap, done.",
				"wrong_order":       "The sequence is right but the indices are off. Watch the j loop.",
				"wrong_termination": "You need a 'done' pulse at the end.",
				"near":              "Almost! Check your loop bounds — n-i-1 for the inner loop.",
				"exact":             "Layer 2 unsealed. The opossum is impressed.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def sort(arr):
    p.init("arr", arr)
    n = len(arr)
    for i in range(n):
        for j in range(___):  # fill this in
            p.pin("arr", "j", j)
            p.signal("arr", "compare", [j, j+1])
            p.compare("arr", j, j+1)
            if arr[j] ___ arr[j+1]:  # fill this in
                p.swap("arr", j, j+1)
                arr[j], arr[j+1] = arr[j+1], arr[j]
    p.done("arr")

sort([64, 34, 25, 12, 22, 11, 90])
`,
			"go": `package main

// p := NewProbe()
// call p.Init, p.Compare, p.Swap, p.Done

func main() {
    p := NewProbe()
    arr := []int{64, 34, 25, 12, 22, 11, 90}
    p.Init("arr", arr)
    n := len(arr)
    for i := 0; i < n; i++ {
        for j := 0; j < ___; j++ { // fill this in
            p.Compare("arr", j, j+1)
            if arr[j] ___ arr[j+1] { // fill this in
                p.Swap("arr", j, j+1)
                arr[j], arr[j+1] = arr[j+1], arr[j]
            }
        }
    }
    p.Done("arr")
}
`,
			"js": `const { Probe } = require('./probe_js.js');
const p = new Probe();

function sort(arr) {
    p.init("arr", arr);
    const n = arr.length;
    for (let i = 0; i < n; i++) {
        for (let j = 0; j < ___; j++) { // fill this in
            p.compare("arr", j, j+1);
            if (arr[j] ___ arr[j+1]) { // fill this in
                p.swap("arr", j, j+1);
                [arr[j], arr[j+1]] = [arr[j+1], arr[j]];
            }
        }
    }
    p.done("arr");
}

sort([64, 34, 25, 12, 22, 11, 90]);
`,
		},
	},
	{
		ID:        2,
		Title:     "Linear Search",
		TraceFile: "data/02_linear_search.trace",
		Dialogue: map[string]map[string]string{
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*it scans — but not like that*",
				"wrong_order":       "*scanning — but something is off*",
				"wrong_termination": "*the scan ends too soon*",
				"near":              "*it hums. almost.*",
				"exact":             "*target located. layer 3 unsealed.*",
			},
			"OPEN": {
				"empty":             "Walk the array and emit access pulses. Emit found when the target is found.",
				"wrong_types":       "You need: init, access for each element, found or not_found, done.",
				"wrong_order":       "Access element by element left-to-right.",
				"wrong_termination": "Stop when found. Emit not_found if the loop finishes.",
				"near":              "Almost there — check your found/not_found logic.",
				"exact":             "Layer 3 unsealed.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def search(arr, target):
    p.init("arr", arr)
    for i in range(len(arr)):
        p.pin("arr", "i", i)
        p.access("arr", i)
        if arr[i] == target:
            p.found("arr", i)
            p.done("arr")
            return i
    p.not_found("arr")
    p.done("arr")
    return -1

search([3, 7, 1, 9, 4, 6, 2], 9)
`,
		},
	},
	{
		ID:        3,
		Title:     "Binary Search",
		TraceFile: "data/binary_search.trace",
		Dialogue: map[string]map[string]string{
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*the search pattern is wrong*",
				"wrong_order":       "*it narrows — but not quite right*",
				"wrong_termination": "*the search terminates incorrectly*",
				"near":              "*the artifact pulses — almost*",
				"exact":             "*binary lock disengaged. layer 4 unsealed.*",
			},
			"OPEN": {
				"empty":             "Halve the search space with each step. Emit bounds and mid pulses.",
				"wrong_types":       "You need: init, bounds, access (at mid), found or not_found, done.",
				"wrong_order":       "Update low or high based on the midpoint comparison.",
				"wrong_termination": "Loop while low <= high.",
				"near":              "Nearly right — double-check your mid calculation.",
				"exact":             "Layer 4 unsealed.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def binary_search(arr, target):
    p.init("arr", arr)
    low, high = 0, len(arr) - 1
    while low <= high:
        p.bounds("arr", low, high)
        mid = (low + high) // 2
        p.pin("arr", "mid", mid)
        p.access("arr", mid)
        if arr[mid] == target:
            p.found("arr", mid)
            p.done("arr")
            return mid
        elif arr[mid] < target:
            low = mid + 1
        else:
            high = mid - 1
    p.not_found("arr")
    p.done("arr")
    return -1

binary_search([2, 5, 8, 12, 16, 23, 38, 56, 72, 91], 23)
`,
		},
	},
	{
		ID:        4,
		Title:     "Selection Sort",
		TraceFile: "data/04_selection_sort.trace",
		Dialogue: map[string]map[string]string{
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*the selection pattern is wrong*",
				"wrong_order":       "*it selects — but not in the right order*",
				"wrong_termination": "*the selection ends too soon*",
				"near":              "*the opossum cocks his head*",
				"exact":             "*minimum found. layer 5 unsealed.*",
			},
			"OPEN": {
				"empty":             "Find the minimum each pass, swap it to the front.",
				"wrong_types":       "Track the min index. Emit compare for each check, swap when done.",
				"wrong_order":       "The min pointer should update whenever a smaller element is found.",
				"wrong_termination": "Outer loop runs n-1 times.",
				"near":              "Almost — check the swap condition (only swap if min != i).",
				"exact":             "Layer 5 unsealed.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def selection_sort(arr):
    p.init("arr", arr)
    n = len(arr)
    for i in range(n - 1):
        p.pin("arr", "i", i)
        min_idx = i
        p.pin("arr", "min", min_idx)
        for j in range(i + 1, n):
            p.compare("arr", min_idx, j)
            if arr[j] < arr[min_idx]:
                min_idx = j
                p.pin("arr", "min", min_idx)
        if min_idx != i:
            p.swap("arr", i, min_idx)
            arr[i], arr[min_idx] = arr[min_idx], arr[i]
    p.done("arr")

selection_sort([29, 10, 14, 37, 13])
`,
		},
	},
	{
		ID:        5,
		Title:     "Insertion Sort",
		TraceFile: "data/05_insertion_sort.trace",
		Dialogue: map[string]map[string]string{
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*the insertion pattern is wrong*",
				"wrong_order":       "*it inserts — but backwards*",
				"wrong_termination": "*the insertion loop is cut short*",
				"near":              "*so close. one detail off.*",
				"exact":             "*card inserted. layer 6 unsealed.*",
			},
			"OPEN": {
				"empty":             "Pick each element and shift larger elements right.",
				"wrong_types":       "You need compare pulses while shifting, swap for each shift.",
				"wrong_order":       "Walk backwards from i while arr[j-1] > arr[j].",
				"wrong_termination": "The outer loop starts at index 1.",
				"near":              "Nearly there.",
				"exact":             "Layer 6 unsealed.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def insertion_sort(arr):
    p.init("arr", arr)
    n = len(arr)
    for i in range(1, n):
        p.pin("arr", "i", i)
        j = i
        while j > 0:
            p.compare("arr", j, j-1)
            if arr[j] < arr[j-1]:
                p.swap("arr", j, j-1)
                arr[j], arr[j-1] = arr[j-1], arr[j]
                j -= 1
            else:
                break
    p.done("arr")

insertion_sort([5, 2, 4, 6, 1, 3])
`,
		},
	},
	{
		ID:        6,
		Title:     "Merge Sort",
		TraceFile: "data/06_merge_sort.trace",
		Dialogue: map[string]map[string]string{
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*the divide pattern is wrong*",
				"wrong_order":       "*it divides — but the merge is off*",
				"wrong_termination": "*the merge is incomplete*",
				"near":              "*the halves align...*",
				"exact":             "*divided and conquered. layer 7 unsealed.*",
			},
			"OPEN": {
				"empty":             "Divide array in half recursively, merge sorted halves.",
				"wrong_types":       "Emit split on divide, merge on combine, compare during merge.",
				"wrong_order":       "Recurse left then right, then merge.",
				"wrong_termination": "Base case: subarrays of length 1 need no split.",
				"near":              "Almost — watch the merge indices.",
				"exact":             "Layer 7 unsealed.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def merge_sort(arr, left, right):
    if right - left <= 1:
        return
    mid = (left + right) // 2
    p.signal("arr", "split", [left, mid, right])
    merge_sort(arr, left, mid)
    merge_sort(arr, mid, right)
    p.signal("arr", "merge", [left, mid, right])
    # merge step
    tmp = []
    i, j = left, mid
    while i < mid and j < right:
        p.compare("arr", i, j)
        if arr[i] <= arr[j]:
            tmp.append(arr[i]); i += 1
        else:
            tmp.append(arr[j]); j += 1
    tmp.extend(arr[i:mid])
    tmp.extend(arr[j:right])
    arr[left:right] = tmp

arr = [38, 27, 43, 3, 9, 82, 10]
p.init("arr", arr)
merge_sort(arr, 0, len(arr))
p.done("arr")
`,
		},
	},
	{
		ID:        7,
		Title:     "Quick Sort",
		TraceFile: "data/07_quick_sort.trace",
		Dialogue: map[string]map[string]string{
			"SIGNAL": {
				"empty":             "*the artifact receives nothing*",
				"wrong_types":       "*the partition pattern is wrong*",
				"wrong_order":       "*it partitions — but around the wrong pivot*",
				"wrong_termination": "*the recursion doesn't bottom out*",
				"near":              "*the pivot is almost right*",
				"exact":             "*sector zero unlocked.*",
			},
			"OPEN": {
				"empty":             "Partition around a pivot, recurse on each half.",
				"wrong_types":       "Pin the pivot, compare each element, swap when needed.",
				"wrong_order":       "Lomuto or Hoare — pick one and stick to it.",
				"wrong_termination": "Base case: partition of size <= 1.",
				"near":              "Almost — check your pivot placement after partition.",
				"exact":             "SECTOR ZERO UNLOCKED.",
			},
		},
		ScaffoldCode: map[string]string{
			"python": `from probe_python import Probe
p = Probe()

def partition(arr, low, high):
    pivot = arr[high]
    p.pin("arr", "pivot", high)
    i = low - 1
    for j in range(low, high):
        p.compare("arr", j, high)
        if arr[j] <= pivot:
            i += 1
            if i != j:
                p.swap("arr", i, j)
                arr[i], arr[j] = arr[j], arr[i]
    p.swap("arr", i+1, high)
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1

def quicksort(arr, low, high):
    if low < high:
        pi = partition(arr, low, high)
        quicksort(arr, low, pi - 1)
        quicksort(arr, pi + 1, high)

arr = [10, 80, 30, 90, 40, 50, 70]
p.init("arr", arr)
quicksort(arr, 0, len(arr) - 1)
p.done("arr")
`,
		},
	},
}

// GetPuzzle returns the puzzle with the given 1-based ID.
// Returns the first puzzle if id is out of range.
func GetPuzzle(id int) Puzzle {
	for _, pz := range allPuzzles {
		if pz.ID == id {
			return pz
		}
	}
	return allPuzzles[0]
}

// All returns all puzzles.
func All() []Puzzle {
	return allPuzzles
}

// LoadTrace reads and parses the embedded trace file for the given puzzle.
func LoadTrace(pz Puzzle) ([]scope.Pulse, error) {
	data, err := puzzleData.ReadFile(pz.TraceFile)
	if err != nil {
		return nil, fmt.Errorf("load trace %q: %w", pz.TraceFile, err)
	}
	return scope.ParseTrace(data)
}

// GetDialogue returns the dialogue text for a given help level and hint key.
// Falls back through levels: OPEN → SIGNAL → STATIC → BLACKOUT.
func GetDialogue(pz Puzzle, helpLevel, hintKey string) string {
	if texts, ok := pz.Dialogue[helpLevel]; ok {
		if s, ok := texts[hintKey]; ok {
			return s
		}
	}
	// Fallback chain.
	fallbacks := []string{"SIGNAL", "STATIC", "BLACKOUT"}
	for _, lvl := range fallbacks {
		if lvl == helpLevel {
			continue
		}
		if texts, ok := pz.Dialogue[lvl]; ok {
			if s, ok := texts[hintKey]; ok {
				return s
			}
		}
	}
	return "*...*"
}

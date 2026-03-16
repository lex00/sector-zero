package puzzles

func init() {
	Register(PuzzleSpec{
		ID:        3,
		Title:     "Binary Search",
		TraceFile: "data/03_binary_search.trace",
		Dialogue: DialogueSpec{
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*the search pattern is wrong*",
				WrongOrder:       "*it narrows — but not quite right*",
				WrongTermination: "*the search terminates incorrectly*",
				Near:             "*the artifact pulses — almost*",
				Exact:            "*binary lock disengaged. layer 4 unsealed.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "You need: init, bounds, access (at mid), found or not_found, done.",
				WrongOrder:       "Update low or high based on the midpoint comparison.",
				WrongTermination: "Loop while low <= high.",
				Near:             "Nearly right — double-check your mid calculation.",
				Exact:            "Layer 4 unsealed.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Binary Search: eliminate half the array each step.\n\nCompare the middle element to the target:\n• Equal → found\n• Target is larger → search the right half (low = mid+1)\n• Target is smaller → search the left half (high = mid-1)\n\nRequires a sorted array.", Style: "guide"},
			{On: "run", Result: "empty", Message: "Select your choices and press ^R."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Pulse sequence: init → bounds → access at mid → found or not_found → done.\nbounds fires every iteration to show the current search window."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Let me set the loop condition for you.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Sequence is off.\nEach loop: emit bounds with current low/high, compute mid, access arr[mid], then compare."},
			{On: "run", Result: "wrong_termination", Message: "Loop condition: while low <= high.\nWhen low passes high the window is empty — the target isn't in the array."},
			{On: "run", Result: "near", Message: "Almost — check the mid formula: (low + high) // 2\nInteger floor division keeps mid as a valid index."},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "loop condition",
					Choices:     []string{"low <= high", "low < high", "low != high", "low < len(arr)"},
					Correct:     0,
					Explanation: "low <= high: the search window is valid as long as low hasn't passed high — when they cross, the target isn't present",
				},
				{
					Label:       "mid calculation",
					Choices:     []string{"(low + high) / 2", "low + high", "(low + high) * 2", "high - low"},
					Correct:     0,
					Explanation: "(low + high) / 2: integer division splits the window at the midpoint, always landing on a valid index",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def binary_search(arr, target):
    p.init("arr", arr)
    low, high = 0, len(arr) - 1
    while {0}:
        p.bounds("arr", low, high)
        mid = int({1})
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
				"go": `package main

func main() {
    p := NewProbe()
    arr := []int{2, 5, 8, 12, 16, 23, 38, 56, 72, 91}
    target := 23
    p.Init("arr", arr)
    low, high := 0, len(arr)-1
    for {0} {
        p.Bounds("arr", low, high)
        mid := {1}
        p.Pin("arr", "mid", mid)
        p.Access("arr", mid)
        if arr[mid] == target {
            p.Found("arr", mid)
            p.Done("arr")
            return
        } else if arr[mid] < target {
            low = mid + 1
        } else {
            high = mid - 1
        }
    }
    p.NotFound("arr")
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function binarySearch(arr, target) {
    p.init("arr", arr);
    let low = 0, high = arr.length - 1;
    while ({0}) {
        p.bounds("arr", low, high);
        const mid = Math.floor({1});
        p.pin("arr", "mid", mid);
        p.access("arr", mid);
        if (arr[mid] === target) {
            p.found("arr", mid);
            p.done("arr");
            return mid;
        } else if (arr[mid] < target) {
            low = mid + 1;
        } else {
            high = mid - 1;
        }
    }
    p.notFound("arr");
    p.done("arr");
    return -1;
}

binarySearch([2, 5, 8, 12, 16, 23, 38, 56, 72, 91], 23);
`,
			"ruby": `p = Probe.new

def binary_search(p, arr, target)
  p.init("arr", arr)
  low, high = 0, arr.length - 1
  while {0}
    p.bounds("arr", low, high)
    mid = {1}
    p.pin("arr", "mid", mid)
    p.access("arr", mid)
    if arr[mid] == target
      p.found("arr", mid)
      p.done("arr")
      return mid
    elsif arr[mid] < target
      low = mid + 1
    else
      high = mid - 1
    end
  end
  p.not_found("arr")
  p.done("arr")
  -1
end

binary_search(p, [2, 5, 8, 12, 16, 23, 38, 56, 72, 91], 23)
`,
			"java": `public class Solution {
    public static void main(String[] args) {
        Probe p = new Probe();
        int[] arr = {2, 5, 8, 12, 16, 23, 38, 56, 72, 91};
        int target = 23;
        p.init("arr", arr);
        int low = 0, high = arr.length - 1;
        while ({0}) {
            p.bounds("arr", low, high);
            int mid = {1};
            p.pin("arr", "mid", mid);
            p.access("arr", mid);
            if (arr[mid] == target) {
                p.found("arr", mid);
                p.done("arr");
                return;
            } else if (arr[mid] < target) {
                low = mid + 1;
            } else {
                high = mid - 1;
            }
        }
        p.notFound("arr");
        p.done("arr");
    }
}
`,
			"rust": `fn main() {
    let p = Probe::new();
    let arr = vec![2i64, 5, 8, 12, 16, 23, 38, 56, 72, 91];
    let target = 23i64;
    p.init("arr", &arr);
    let mut low: isize = 0;
    let mut high: isize = arr.len() as isize - 1;
    while {0} {
        p.bounds("arr", low as usize, high as usize);
        let mid = ({1}) as usize;
        p.pin("arr", "mid", mid);
        p.access("arr", mid);
        if arr[mid] == target {
            p.found("arr", mid);
            p.done("arr");
            return;
        } else if arr[mid] < target {
            low = mid as isize + 1;
        } else {
            high = mid as isize - 1;
        }
    }
    p.not_found("arr");
    p.done("arr");
}
`,
			},
		},
	})
}

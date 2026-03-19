package puzzles

func init() {
	Register(Puzzle{
		ID:        2,
		Title:     "Linear Search",
		TraceFile: "data/02_linear_search.trace",
		Dialogue: DialogueSpec{
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*it scans — but not like that*",
				WrongOrder:       "*scanning — but something is off*",
				WrongTermination: "*the scan ends too soon*",
				Near:             "*it hums. almost.*",
				Exact:            "*target located. layer 3 unsealed.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "You need: init, access for each element, found or not_found, done.",
				WrongOrder:       "Access element by element left-to-right.",
				WrongTermination: "Stop when found. Emit not_found if the loop finishes.",
				Near:             "Almost there — check your found/not_found logic.",
				Exact:            "Layer 3 unsealed.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Linear Search: scan every element left to right.\n\nFor each element: access it, compare it to the target.\nIf it matches — emit found with the index.\nIf the loop ends without a match — emit not_found.", Style: "guide"},
			{On: "run", Result: "empty", Message: "Select your choices and press ^R."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Pulse sequence: init → access each element → found (if match) or not_found (at end) → done."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Let me set blank 1 — the comparison condition.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Pulse order is off.\naccess comes before the comparison; found exits early, not_found comes after the loop."},
			{On: "run", Result: "wrong_termination", Message: "The loop finished without a match.\nBlank 2 needs to call p.not_found() before p.done()."},
			{On: "run", Result: "near", Message: "Nearly there — check blank 2: p.not_found(\"arr\") goes outside and after the for loop."},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "found condition",
					Choices:     []string{"arr[i] == target", "arr[i] != target", "i == target", "arr[i] > target"},
					Correct:     0,
					Explanation: "arr[i] == target: check each element against what we're looking for — equality means we found it",
				},
				{
					Label:       "end of loop result",
					Choices:     []string{"p.not_found(\"arr\")", "p.found(\"arr\", -1)", "pass", "return -1"},
					Correct:     0,
					Explanation: "p.not_found: if the loop finishes without a match the target isn't in the array — signal that explicitly",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def search(arr, target):
    p.init("arr", arr)
    for i in range(len(arr)):
        p.pin("arr", "i", i)
        p.access("arr", i)
        if {0}:
            p.found("arr", i)
            p.done("arr")
            return i
    {1}
    p.done("arr")
    return -1

search([3, 7, 1, 9, 4, 6, 2], 9)
`,
				"go": `package main

func main() {
    p := NewProbe()
    arr := []int{3, 7, 1, 9, 4, 6, 2}
    target := 9
    p.Init("arr", arr)
    for i := 0; i < len(arr); i++ {
        p.Pin("arr", "i", i)
        p.Access("arr", i)
        if {0} {
            p.Found("arr", i)
            p.Done("arr")
            return
        }
    }
    {1}
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function search(arr, target) {
    p.init("arr", arr);
    for (let i = 0; i < arr.length; i++) {
        p.pin("arr", "i", i);
        p.access("arr", i);
        if ({0}) {
            p.found("arr", i);
            p.done("arr");
            return i;
        }
    }
    {1}
    p.done("arr");
    return -1;
}

search([3, 7, 1, 9, 4, 6, 2], 9);
`,
			"ruby": `p = Probe.new

def search(p, arr, target)
  p.init("arr", arr)
  arr.length.times do |i|
    p.pin("arr", "i", i)
    p.access("arr", i)
    if {0}
      p.found("arr", i)
      p.done("arr")
      return i
    end
  end
  {1}
  p.done("arr")
  -1
end

search(p, [3, 7, 1, 9, 4, 6, 2], 9)
`,
			"java": `public class Solution {
    public static void main(String[] args) {
        Probe p = new Probe();
        int[] arr = {3, 7, 1, 9, 4, 6, 2};
        int target = 9;
        p.init("arr", arr);
        for (int i = 0; i < arr.length; i++) {
            p.pin("arr", "i", i);
            p.access("arr", i);
            if ({0}) {
                p.found("arr", i);
                p.done("arr");
                return;
            }
        }
        {1};
        p.done("arr");
    }
}
`,
			"rust": `fn main() {
    let p = Probe::new();
    let arr = vec![3i64, 7, 1, 9, 4, 6, 2];
    let target = 9i64;
    p.init("arr", &arr);
    for i in 0..arr.len() {
        p.pin("arr", "i", i);
        p.access("arr", i);
        if {0} {
            p.found("arr", i);
            p.done("arr");
            return;
        }
    }
    {1};
    p.done("arr");
}
`,
			},
		},
	})
}

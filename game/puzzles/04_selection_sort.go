package puzzles

func init() {
	Register(Puzzle{
		ID:        4,
		Title:     "Selection Sort",
		TraceFile: "data/04_selection_sort.trace",
		Dialogue: DialogueSpec{
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*the selection pattern is wrong*",
				WrongOrder:       "*it selects — but not in the right order*",
				WrongTermination: "*the selection ends too soon*",
				Near:             "*the opossum cocks his head*",
				Exact:            "*minimum found. layer 5 unsealed.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "Track the min index. Emit compare for each check, swap when done.",
				WrongOrder:       "The min pointer should update whenever a smaller element is found.",
				WrongTermination: "Outer loop runs n-1 times.",
				Near:             "Almost — check the swap condition (only swap if min != i).",
				Exact:            "Layer 5 unsealed.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Selection Sort: find the minimum element in the unsorted portion, then swap it into position.\n\nOuter loop: position to fill (i)\nInner loop: scan i+1..end to find the minimum index\nSwap arr[i] with arr[min_idx]", Style: "guide"},
			{On: "run", Result: "empty", Message: "Select your choices and press ^R."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Pulse sequence: init → compare to track minimum → swap minimum to front → done."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Let me reveal the first blank for you.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Right pulses, wrong positions.\nInner loop starts at i+1 (not 0) — elements before i are already sorted."},
			{On: "run", Result: "near", Message: "Close — check both loop bounds and the swap: only swap when min_idx != i to avoid redundant moves."},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "outer loop range",
					Choices:     []string{"n - 1", "n", "n + 1", "n - 2"},
					Correct:     0,
					Explanation: "n-1 passes: after placing n-1 minimums, the last element is already in the right position automatically",
				},
				{
					Label:       "update min condition",
					Choices:     []string{"arr[j] < arr[min_idx]", "arr[j] > arr[min_idx]", "j < min_idx", "arr[j] == arr[min_idx]"},
					Correct:     0,
					Explanation: "arr[j] < arr[min_idx]: update the min pointer whenever we find something smaller — we want the smallest in the unsorted region",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def selection_sort(arr):
    p.init("arr", arr)
    n = len(arr)
    for i in range({0}):
        p.pin("arr", "i", i)
        min_idx = i
        p.pin("arr", "min", min_idx)
        for j in range(i + 1, n):
            p.compare("arr", min_idx, j)
            if {1}:
                min_idx = j
                p.pin("arr", "min", min_idx)
        if min_idx != i:
            p.swap("arr", i, min_idx)
            arr[i], arr[min_idx] = arr[min_idx], arr[i]
    p.done("arr")

selection_sort([29, 10, 14, 37, 13])
`,
				"go": `package main

func main() {
    p := NewProbe()
    arr := []int{29, 10, 14, 37, 13}
    p.Init("arr", arr)
    n := len(arr)
    for i := 0; i < {0}; i++ {
        p.Pin("arr", "i", i)
        min_idx := i
        p.Pin("arr", "min", min_idx)
        for j := i + 1; j < n; j++ {
            p.Compare("arr", min_idx, j)
            if {1} {
                min_idx = j
                p.Pin("arr", "min", min_idx)
            }
        }
        if min_idx != i {
            p.Swap("arr", i, min_idx)
            arr[i], arr[min_idx] = arr[min_idx], arr[i]
        }
    }
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function selectionSort(arr) {
    p.init("arr", arr);
    const n = arr.length;
    for (let i = 0; i < {0}; i++) {
        p.pin("arr", "i", i);
        let min_idx = i;
        p.pin("arr", "min", min_idx);
        for (let j = i + 1; j < n; j++) {
            p.compare("arr", min_idx, j);
            if ({1}) {
                min_idx = j;
                p.pin("arr", "min", min_idx);
            }
        }
        if (min_idx !== i) {
            p.swap("arr", i, min_idx);
            [arr[i], arr[min_idx]] = [arr[min_idx], arr[i]];
        }
    }
    p.done("arr");
}

selectionSort([29, 10, 14, 37, 13]);
`,
			"ruby": `p = Probe.new

def selection_sort(p, arr)
  p.init("arr", arr)
  n = arr.length
  ({0}).times do |i|
    p.pin("arr", "i", i)
    min_idx = i
    p.pin("arr", "min", min_idx)
    (i + 1...n).each do |j|
      p.compare("arr", min_idx, j)
      if {1}
        min_idx = j
        p.pin("arr", "min", min_idx)
      end
    end
    if min_idx != i
      p.swap("arr", i, min_idx)
      arr[i], arr[min_idx] = arr[min_idx], arr[i]
    end
  end
  p.done("arr")
end

selection_sort(p, [29, 10, 14, 37, 13])
`,
			"java": `public class Solution {
    public static void main(String[] args) {
        Probe p = new Probe();
        int[] arr = {29, 10, 14, 37, 13};
        p.init("arr", arr);
        int n = arr.length;
        for (int i = 0; i < {0}; i++) {
            p.pin("arr", "i", i);
            int min_idx = i;
            p.pin("arr", "min", min_idx);
            for (int j = i + 1; j < n; j++) {
                p.compare("arr", min_idx, j);
                if ({1}) {
                    min_idx = j;
                    p.pin("arr", "min", min_idx);
                }
            }
            if (min_idx != i) {
                p.swap("arr", i, min_idx);
                int tmp = arr[i]; arr[i] = arr[min_idx]; arr[min_idx] = tmp;
            }
        }
        p.done("arr");
    }
}
`,
			"rust": `fn main() {
    let p = Probe::new();
    let mut arr = vec![29i64, 10, 14, 37, 13];
    p.init("arr", &arr);
    let n = arr.len();
    for i in 0..{0} {
        p.pin("arr", "i", i);
        let mut min_idx = i;
        p.pin("arr", "min", min_idx);
        for j in (i + 1)..n {
            p.compare("arr", min_idx, j);
            if {1} {
                min_idx = j;
                p.pin("arr", "min", min_idx);
            }
        }
        if min_idx != i {
            p.swap("arr", i, min_idx);
            arr.swap(i, min_idx);
        }
    }
    p.done("arr");
}
`,
			},
		},
	})
}

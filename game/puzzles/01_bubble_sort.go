package puzzles

func init() {
	Register(PuzzleSpec{
		ID:        1,
		Title:     "Bubble Sort",
		TraceFile: "data/bubble_sort.trace",
		Dialogue: DialogueSpec{
			Blackout: HintSet{
				Empty: "...", WrongTypes: "...", WrongOrder: "...",
				WrongTermination: "...", Near: "...", Exact: "...",
			},
			Static: HintSet{
				Empty:            "*static*",
				WrongTypes:       "*something moves — but not like that*",
				WrongOrder:       "*closer — but the rhythm is wrong*",
				WrongTermination: "*it flickers and dies too soon*",
				Near:             "*almost. so close.*",
				Exact:            "*the artifact glows. layer 2 unsealed.*",
			},
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*something moves — but not in the right way*",
				WrongOrder:       "*the artifact stirs — you're closer*",
				WrongTermination: "*it runs too long. or not long enough*",
				Near:             "*the opossum's eyes widen*",
				Exact:            "*the artifact glows. layer 2 unsealed.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "Check your pulse types: init, compare, swap, done.",
				WrongOrder:       "The sequence is right but the indices are off. Watch the j loop.",
				WrongTermination: "You need a 'done' pulse at the end.",
				Near:             "Almost! Check your loop bounds — n-i-1 for the inner loop.",
				Exact:            "Layer 2 unsealed. The opossum is impressed.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Bubble Sort: compare adjacent pairs and swap if out of order.\n\nEach full pass moves the largest unsorted element to its final position at the right end.\n\nWatch the scope — yellow = comparing, orange = swapping.\n\n^R unlocks after one loop.", Style: "guide", Trigger: "gate"},
			{On: "run", Result: "empty", Message: "Select choices for each blank then press ^R to run."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Watch the reference scope:\ncompare fires on every adjacent pair check,\nswap fires only when elements are out of order."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Still wrong types — press ^G and I'll set the first blank for you.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Right pulses, wrong indices.\n\nInner loop: j goes from 0 to n-i-1.\nThe last i elements are already in place — no need to compare them."},
			{On: "run", Result: "wrong_termination", Message: "Add p.done(\"arr\") after the outer loop finishes — it signals the sort is complete."},
			{On: "run", Result: "near", Message: "Very close.\nBlank 0: inner bound is n-i-1\nBlank 1: swap condition is >"},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "inner loop bound",
					Choices:     []string{"n-i-1", "n-i", "n-1", "n"},
					Correct:     0,
					Explanation: "n-i-1: after each outer pass the largest i elements are already in place at the end — comparing them again wastes work",
				},
				{
					Label:       "swap condition",
					Choices:     []string{">", "<", ">=", "<="},
					Correct:     0,
					Explanation: "Use > so that when the left element is bigger we swap — this bubbles the maximum rightward on every pass",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def sort(arr):
    p.init("arr", arr)
    n = len(arr)
    for i in range(n):
        for j in range({0}):
            p.pin("arr", "j", j)
            p.signal("arr", "compare", [j, j+1])
            p.compare("arr", j, j+1)
            if arr[j] {1} arr[j+1]:
                p.swap("arr", j, j+1)
                arr[j], arr[j+1] = arr[j+1], arr[j]
    p.done("arr")

sort([64, 34, 25, 12, 22, 11, 90])
`,
				"go": `package main

func main() {
    p := NewProbe()
    arr := []int{64, 34, 25, 12, 22, 11, 90}
    p.Init("arr", arr)
    n := len(arr)
    for i := 0; i < n; i++ {
        for j := 0; j < {0}; j++ {
            p.Compare("arr", j, j+1)
            if arr[j] {1} arr[j+1] {
                p.Swap("arr", j, j+1)
                arr[j], arr[j+1] = arr[j+1], arr[j]
            }
        }
    }
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function sort(arr) {
    p.init("arr", arr);
    const n = arr.length;
    for (let i = 0; i < n; i++) {
        for (let j = 0; j < {0}; j++) {
            p.compare("arr", j, j+1);
            if (arr[j] {1} arr[j+1]) {
                p.swap("arr", j, j+1);
                [arr[j], arr[j+1]] = [arr[j+1], arr[j]];
            }
        }
    }
    p.done("arr");
}

sort([64, 34, 25, 12, 22, 11, 90]);
`,
			"ruby": `p = Probe.new

def sort(p, arr)
  p.init("arr", arr)
  n = arr.length
  n.times do |i|
    ({0}).times do |j|
      p.compare("arr", j, j+1)
      if arr[j] {1} arr[j+1]
        p.swap("arr", j, j+1)
        arr[j], arr[j+1] = arr[j+1], arr[j]
      end
    end
  end
  p.done("arr")
end

sort(p, [64, 34, 25, 12, 22, 11, 90])
`,
			"java": `public class Solution {
    public static void main(String[] args) {
        Probe p = new Probe();
        int[] arr = {64, 34, 25, 12, 22, 11, 90};
        p.init("arr", arr);
        int n = arr.length;
        for (int i = 0; i < n; i++) {
            for (int j = 0; j < {0}; j++) {
                p.compare("arr", j, j+1);
                if (arr[j] {1} arr[j+1]) {
                    p.swap("arr", j, j+1);
                    int tmp = arr[j]; arr[j] = arr[j+1]; arr[j+1] = tmp;
                }
            }
        }
        p.done("arr");
    }
}
`,
			"rust": `fn main() {
    let p = Probe::new();
    let mut arr = vec![64i64, 34, 25, 12, 22, 11, 90];
    p.init("arr", &arr);
    let n = arr.len();
    for i in 0..n {
        for j in 0..{0} {
            p.compare("arr", j, j+1);
            if arr[j] {1} arr[j+1] {
                p.swap("arr", j, j+1);
                arr.swap(j, j+1);
            }
        }
    }
    p.done("arr");
}
`,
			},
		},
	})
}

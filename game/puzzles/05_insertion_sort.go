package puzzles

func init() {
	Register(Puzzle{
		ID:        5,
		Title:     "Insertion Sort",
		TraceFile: "data/05_insertion_sort.trace",
		Dialogue: DialogueSpec{
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*the insertion pattern is wrong*",
				WrongOrder:       "*it inserts — but backwards*",
				WrongTermination: "*the insertion loop is cut short*",
				Near:             "*so close. one detail off.*",
				Exact:            "*card inserted. layer 6 unsealed.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "You need compare pulses while shifting, swap for each shift.",
				WrongOrder:       "Walk backwards from i while arr[j-1] > arr[j].",
				WrongTermination: "The outer loop starts at index 1.",
				Near:             "Nearly there.",
				Exact:            "Layer 6 unsealed.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Insertion Sort: grow a sorted region one element at a time.\n\nFor each new element, slide it left past any larger neighbours until it's in the right spot.\n\nThink: sorting a hand of cards by inserting each new card into the correct position.", Style: "guide"},
			{On: "run", Result: "empty", Message: "Select your choices and press ^R."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Pulse sequence: init → compare current with previous → swap leftward while out of order → done."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Let me reveal the shift condition.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Direction is off.\nThe inner loop walks left (j decrements), comparing arr[j] with arr[j-1]."},
			{On: "run", Result: "near", Message: "Very close — inner loop: for j in range(i, 0, -1), stop when arr[j] >= arr[j-1]."},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "outer loop start",
					Choices:     []string{"1", "0", "2", "n-1"},
					Correct:     0,
					Explanation: "Start at index 1: a single element at index 0 is trivially sorted — pick each subsequent element and insert it into the sorted left portion",
				},
				{
					Label:       "shift condition",
					Choices:     []string{"arr[j] < arr[j-1]", "arr[j] > arr[j-1]", "arr[j] == arr[j-1]", "arr[j] <= arr[j-1]"},
					Correct:     0,
					Explanation: "arr[j] < arr[j-1]: keep shifting left while the current element is smaller than its left neighbor — this inserts it in sorted order",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def insertion_sort(arr):
    p.init("arr", arr)
    n = len(arr)
    for i in range({0}, n):
        p.pin("arr", "i", i)
        j = i
        while j > 0:
            p.compare("arr", j, j-1)
            if {1}:
                p.swap("arr", j, j-1)
                arr[j], arr[j-1] = arr[j-1], arr[j]
                j -= 1
            else:
                break
    p.done("arr")

insertion_sort([5, 2, 4, 6, 1, 3])
`,
				"go": `package main

func main() {
    p := NewProbe()
    arr := []int{5, 2, 4, 6, 1, 3}
    p.Init("arr", arr)
    n := len(arr)
    for i := {0}; i < n; i++ {
        p.Pin("arr", "i", i)
        j := i
        for j > 0 {
            p.Compare("arr", j, j-1)
            if {1} {
                p.Swap("arr", j, j-1)
                arr[j], arr[j-1] = arr[j-1], arr[j]
                j--
            } else {
                break
            }
        }
    }
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function insertionSort(arr) {
    p.init("arr", arr);
    const n = arr.length;
    for (let i = {0}; i < n; i++) {
        p.pin("arr", "i", i);
        let j = i;
        while (j > 0) {
            p.compare("arr", j, j-1);
            if ({1}) {
                p.swap("arr", j, j-1);
                [arr[j], arr[j-1]] = [arr[j-1], arr[j]];
                j--;
            } else {
                break;
            }
        }
    }
    p.done("arr");
}

insertionSort([5, 2, 4, 6, 1, 3]);
`,
			"ruby": `p = Probe.new

def insertion_sort(p, arr)
  p.init("arr", arr)
  n = arr.length
  ({0}...n).each do |i|
    p.pin("arr", "i", i)
    j = i
    while j > 0
      p.compare("arr", j, j-1)
      if {1}
        p.swap("arr", j, j-1)
        arr[j], arr[j-1] = arr[j-1], arr[j]
        j -= 1
      else
        break
      end
    end
  end
  p.done("arr")
end

insertion_sort(p, [5, 2, 4, 6, 1, 3])
`,
			"java": `public class Solution {
    public static void main(String[] args) {
        Probe p = new Probe();
        int[] arr = {5, 2, 4, 6, 1, 3};
        p.init("arr", arr);
        int n = arr.length;
        for (int i = {0}; i < n; i++) {
            p.pin("arr", "i", i);
            int j = i;
            while (j > 0) {
                p.compare("arr", j, j-1);
                if ({1}) {
                    p.swap("arr", j, j-1);
                    int tmp = arr[j]; arr[j] = arr[j-1]; arr[j-1] = tmp;
                    j--;
                } else {
                    break;
                }
            }
        }
        p.done("arr");
    }
}
`,
			"rust": `fn main() {
    let p = Probe::new();
    let mut arr = vec![5i64, 2, 4, 6, 1, 3];
    p.init("arr", &arr);
    let n = arr.len();
    for i in {0}..n {
        p.pin("arr", "i", i);
        let mut j = i;
        while j > 0 {
            p.compare("arr", j, j-1);
            if {1} {
                p.swap("arr", j, j-1);
                arr.swap(j, j-1);
                j -= 1;
            } else {
                break;
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

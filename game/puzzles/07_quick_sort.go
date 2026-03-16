package puzzles

func init() {
	Register(PuzzleSpec{
		ID:        7,
		Title:     "Quick Sort",
		TraceFile: "data/07_quick_sort.trace",
		Dialogue: DialogueSpec{
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*the partition pattern is wrong*",
				WrongOrder:       "*it partitions — but around the wrong pivot*",
				WrongTermination: "*the recursion doesn't bottom out*",
				Near:             "*the pivot is almost right*",
				Exact:            "*sector zero unlocked.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "Pin the pivot, compare each element, swap when needed.",
				WrongOrder:       "Lomuto or Hoare — pick one and stick to it.",
				WrongTermination: "Base case: partition of size <= 1.",
				Near:             "Almost — check your pivot placement after partition.",
				Exact:            "SECTOR ZERO UNLOCKED.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Quick Sort (Lomuto partition):\n\n1. Choose the last element as pivot (emit pin)\n2. Walk the array — if element <= pivot, extend the left partition (emit swap)\n3. Place pivot at its final position\n4. Recurse on left and right of pivot\n\nThe pivot ends up in its exact sorted position.", Style: "guide"},
			{On: "run", Result: "empty", Message: "Select your choices and press ^R."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Pulse sequence: pin pivot → compare each element → swap if ≤ pivot → swap pivot to final position → recurse."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Let me set the base case for you.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Partition logic is off.\ni starts at low-1 and only advances when a swap happens, keeping left partition growing."},
			{On: "run", Result: "wrong_termination", Message: "Base case: only recurse when low < high — a single element needs no sorting."},
			{On: "run", Result: "near", Message: "Close — blank 2 swap condition should be arr[j] <= pivot so equal elements go left of pivot."},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "recursion base case",
					Choices:     []string{"low < high", "low <= high", "low != high", "high - low > 1"},
					Correct:     0,
					Explanation: "low < high: only recurse when the partition has more than one element — a single element is already in its correct position",
				},
				{
					Label:       "swap condition",
					Choices:     []string{"arr[j] <= pivot", "arr[j] >= pivot", "arr[j] < pivot", "arr[j] > pivot"},
					Correct:     0,
					Explanation: "arr[j] <= pivot: move elements smaller than or equal to the pivot to the left partition — this places the pivot correctly after the loop",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def partition(arr, low, high):
    pivot = arr[high]
    p.pin("arr", "pivot", high)
    i = low - 1
    for j in range(low, high):
        p.compare("arr", j, high)
        if {1}:
            i += 1
            p.swap("arr", i, j)
            arr[i], arr[j] = arr[j], arr[i]
    p.swap("arr", i+1, high)
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1

def quicksort(arr, low, high):
    if {0}:
        pi = partition(arr, low, high)
        quicksort(arr, low, pi - 1)
        quicksort(arr, pi + 1, high)

arr = [64, 34, 25, 12, 22, 11, 90]
p.init("arr", arr)
quicksort(arr, 0, len(arr) - 1)
p.done("arr")
`,
				"go": `package main

func partition(p *Probe, arr []int, low, high int) int {
    pivot := arr[high]
    p.Pin("arr", "pivot", high)
    i := low - 1
    for j := low; j < high; j++ {
        p.Compare("arr", j, high)
        if {1} {
            i++
            p.Swap("arr", i, j)
            arr[i], arr[j] = arr[j], arr[i]
        }
    }
    p.Swap("arr", i+1, high)
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1
}

func quicksort(p *Probe, arr []int, low, high int) {
    if {0} {
        pi := partition(p, arr, low, high)
        quicksort(p, arr, low, pi-1)
        quicksort(p, arr, pi+1, high)
    }
}

func main() {
    p := NewProbe()
    arr := []int{64, 34, 25, 12, 22, 11, 90}
    p.Init("arr", arr)
    quicksort(p, arr, 0, len(arr)-1)
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function partition(arr, low, high) {
    const pivot = arr[high];
    p.pin("arr", "pivot", high);
    let i = low - 1;
    for (let j = low; j < high; j++) {
        p.compare("arr", j, high);
        if ({1}) {
            i++;
            p.swap("arr", i, j);
            [arr[i], arr[j]] = [arr[j], arr[i]];
        }
    }
    p.swap("arr", i+1, high);
    [arr[i+1], arr[high]] = [arr[high], arr[i+1]];
    return i + 1;
}

function quicksort(arr, low, high) {
    if ({0}) {
        const pi = partition(arr, low, high);
        quicksort(arr, low, pi - 1);
        quicksort(arr, pi + 1, high);
    }
}

const arr = [64, 34, 25, 12, 22, 11, 90];
p.init("arr", arr);
quicksort(arr, 0, arr.length - 1);
p.done("arr");
`,
			"ruby": `p = Probe.new

def partition(p, arr, low, high)
  pivot = arr[high]
  p.pin("arr", "pivot", high)
  i = low - 1
  (low...high).each do |j|
    p.compare("arr", j, high)
    if {1}
      i += 1
      p.swap("arr", i, j)
      arr[i], arr[j] = arr[j], arr[i]
    end
  end
  p.swap("arr", i+1, high)
  arr[i+1], arr[high] = arr[high], arr[i+1]
  i + 1
end

def quicksort(p, arr, low, high)
  if {0}
    pi = partition(p, arr, low, high)
    quicksort(p, arr, low, pi - 1)
    quicksort(p, arr, pi + 1, high)
  end
end

arr = [64, 34, 25, 12, 22, 11, 90]
p.init("arr", arr)
quicksort(p, arr, 0, arr.length - 1)
p.done("arr")
`,
			"java": `public class Solution {
    static Probe p = new Probe();

    static int partition(int[] arr, int low, int high) {
        int pivot = arr[high];
        p.pin("arr", "pivot", high);
        int i = low - 1;
        for (int j = low; j < high; j++) {
            p.compare("arr", j, high);
            if ({1}) {
                i++;
                p.swap("arr", i, j);
                int tmp = arr[i]; arr[i] = arr[j]; arr[j] = tmp;
            }
        }
        p.swap("arr", i+1, high);
        int tmp = arr[i+1]; arr[i+1] = arr[high]; arr[high] = tmp;
        return i + 1;
    }

    static void quicksort(int[] arr, int low, int high) {
        if ({0}) {
            int pi = partition(arr, low, high);
            quicksort(arr, low, pi - 1);
            quicksort(arr, pi + 1, high);
        }
    }

    public static void main(String[] args) {
        int[] arr = {64, 34, 25, 12, 22, 11, 90};
        p.init("arr", arr);
        quicksort(arr, 0, arr.length - 1);
        p.done("arr");
    }
}
`,
			"rust": `fn partition(p: &Probe, arr: &mut Vec<i64>, low: usize, high: usize) -> usize {
    let pivot = arr[high];
    p.pin("arr", "pivot", high);
    let mut i: isize = low as isize - 1;
    for j in low..high {
        p.compare("arr", j, high);
        if {1} {
            i += 1;
            p.swap("arr", i as usize, j);
            arr.swap(i as usize, j);
        }
    }
    let pos = (i + 1) as usize;
    p.swap("arr", pos, high);
    arr.swap(pos, high);
    pos
}

fn quicksort(p: &Probe, arr: &mut Vec<i64>, low: usize, high: usize) {
    if {0} {
        let pi = partition(p, arr, low, high);
        if pi > 0 { quicksort(p, arr, low, pi - 1); }
        quicksort(p, arr, pi + 1, high);
    }
}

fn main() {
    let p = Probe::new();
    let mut arr = vec![64i64, 34, 25, 12, 22, 11, 90];
    p.init("arr", &arr);
    let n = arr.len();
    quicksort(&p, &mut arr, 0, n - 1);
    p.done("arr");
}
`,
			},
		},
	})
}

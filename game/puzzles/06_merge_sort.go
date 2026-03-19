package puzzles

func init() {
	Register(Puzzle{
		ID:        6,
		Title:     "Merge Sort",
		TraceFile: "data/06_merge_sort.trace",
		Dialogue: DialogueSpec{
			Signal: HintSet{
				Empty:            "*the artifact receives nothing*",
				WrongTypes:       "*the divide pattern is wrong*",
				WrongOrder:       "*it divides — but the merge is off*",
				WrongTermination: "*the merge is incomplete*",
				Near:             "*the halves align...*",
				Exact:            "*divided and conquered. layer 7 unsealed.*",
			},
			Open: HintSet{
				Empty:            "No trace received. Pick your choices and press ^R to run.",
				WrongTypes:       "Emit split on divide, merge on combine, compare during merge.",
				WrongOrder:       "Recurse left then right, then merge.",
				WrongTermination: "Base case: subarrays of length 1 need no split.",
				Near:             "Almost — watch the merge indices.",
				Exact:            "Layer 7 unsealed.",
			},
		},
		Script: []LessonStep{
			{On: "load", Message: "Merge Sort: divide and conquer.\n\n1. Split the array at the midpoint (emit split)\n2. Recursively sort each half\n3. Merge the sorted halves back together (emit merge, then compare each pair)\n\nThe split/merge pulses mark each level of the recursion.", Style: "guide"},
			{On: "run", Result: "empty", Message: "Select your choices and press ^R."},
			{On: "run", Result: "wrong_types", After: 0, Message: "Pulse sequence: init → split → (recurse) → merge → compare during merge → done."},
			{On: "run", Result: "wrong_types", After: 2, Message: "Let me set the base case condition for you.", Style: "guide", Trigger: "guide_next"},
			{On: "run", Result: "wrong_order", Message: "Order is off — emit split before recursing, then merge after both recursive calls return."},
			{On: "run", Result: "wrong_termination", Message: "Base case: right - left <= 1 means a subarray of 0 or 1 elements — already sorted, return early."},
			{On: "run", Result: "near", Message: "Almost — check blank 2: mid = (left + right) // 2 for integer floor division."},
		},
		Challenge: ChallengeSpec{
			Blanks: []Blank{
				{
					Label:       "base case condition",
					Choices:     []string{"right - left <= 1", "right - left == 0", "left >= right", "right <= left"},
					Correct:     0,
					Explanation: "right - left <= 1: a subarray of 0 or 1 elements is already sorted — stop recursing and return",
				},
				{
					Label:       "mid calculation",
					Choices:     []string{"(left + right) / 2", "left + right", "(left + right) * 2", "right - left"},
					Correct:     0,
					Explanation: "(left + right) / 2: integer division splits the subarray at its midpoint so both halves are roughly equal in size",
				},
			},
			Template: map[string]string{
				"python": `from probe_python import Probe
p = Probe()

def merge_sort(arr, left, right):
    if {0}:
        return
    mid = int({1})
    p.split("arr", left, mid, right)
    merge_sort(arr, left, mid)
    merge_sort(arr, mid, right)
    p.merge("arr", left, mid, right)
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
    for k, v in enumerate(tmp):
        arr[left + k] = v
        p.write("arr", left + k, v)

arr = [64, 34, 25, 12, 22, 11, 90]
p.init("arr", arr)
merge_sort(arr, 0, len(arr))
p.done("arr")
`,
				"go": `package main

func mergeSort(p *Probe, arr []int, left, right int) {
    if {0} {
        return
    }
    mid := {1}
    p.Split("arr", left, mid, right)
    mergeSort(p, arr, left, mid)
    mergeSort(p, arr, mid, right)
    p.Merge("arr", left, mid, right)
    tmp := make([]int, 0, right-left)
    i, j := left, mid
    for i < mid && j < right {
        p.Compare("arr", i, j)
        if arr[i] <= arr[j] {
            tmp = append(tmp, arr[i]); i++
        } else {
            tmp = append(tmp, arr[j]); j++
        }
    }
    tmp = append(tmp, arr[i:mid]...)
    tmp = append(tmp, arr[j:right]...)
    for k, v := range tmp {
        arr[left+k] = v
        p.Write("arr", left+k, v)
    }
}

func main() {
    p := NewProbe()
    arr := []int{64, 34, 25, 12, 22, 11, 90}
    p.Init("arr", arr)
    mergeSort(p, arr, 0, len(arr))
    p.Done("arr")
}
`,
				"js": `const p = new Probe();

function mergeSort(arr, left, right) {
    if ({0}) return;
    const mid = Math.floor({1});
    p.split("arr", left, mid, right);
    mergeSort(arr, left, mid);
    mergeSort(arr, mid, right);
    p.merge("arr", left, mid, right);
    const tmp = [];
    let i = left, j = mid;
    while (i < mid && j < right) {
        p.compare("arr", i, j);
        if (arr[i] <= arr[j]) { tmp.push(arr[i]); i++; }
        else { tmp.push(arr[j]); j++; }
    }
    while (i < mid) { tmp.push(arr[i]); i++; }
    while (j < right) { tmp.push(arr[j]); j++; }
    for (let k = 0; k < tmp.length; k++) {
        arr[left + k] = tmp[k];
        p.write("arr", left + k, tmp[k]);
    }
}

const arr = [64, 34, 25, 12, 22, 11, 90];
p.init("arr", arr);
mergeSort(arr, 0, arr.length);
p.done("arr");
`,
			"ruby": `p = Probe.new

def merge_sort(p, arr, left, right)
  return if {0}
  mid = {1}
  p.split("arr", left, mid, right)
  merge_sort(p, arr, left, mid)
  merge_sort(p, arr, mid, right)
  p.merge("arr", left, mid, right)
  tmp = []
  i, j = left, mid
  while i < mid && j < right
    p.compare("arr", i, j)
    if arr[i] <= arr[j]
      tmp << arr[i]; i += 1
    else
      tmp << arr[j]; j += 1
    end
  end
  tmp.concat(arr[i...mid])
  tmp.concat(arr[j...right])
  tmp.each_with_index do |v, k|
    arr[left + k] = v
    p.write("arr", left + k, v)
  end
end

arr = [64, 34, 25, 12, 22, 11, 90]
p.init("arr", arr)
merge_sort(p, arr, 0, arr.length)
p.done("arr")
`,
			"java": `public class Solution {
    static Probe p = new Probe();

    static void mergeSort(int[] arr, int left, int right) {
        if ({0}) return;
        int mid = {1};
        p.split("arr", left, mid, right);
        mergeSort(arr, left, mid);
        mergeSort(arr, mid, right);
        p.merge("arr", left, mid, right);
        int[] tmp = new int[right - left];
        int i = left, j = mid, k = 0;
        while (i < mid && j < right) {
            p.compare("arr", i, j);
            if (arr[i] <= arr[j]) { tmp[k++] = arr[i++]; }
            else { tmp[k++] = arr[j++]; }
        }
        while (i < mid) tmp[k++] = arr[i++];
        while (j < right) tmp[k++] = arr[j++];
        for (int x = 0; x < tmp.length; x++) {
            arr[left + x] = tmp[x];
            p.write("arr", left + x, tmp[x]);
        }
    }

    public static void main(String[] args) {
        int[] arr = {64, 34, 25, 12, 22, 11, 90};
        p.init("arr", arr);
        mergeSort(arr, 0, arr.length);
        p.done("arr");
    }
}
`,
			"rust": `fn merge_sort(p: &Probe, arr: &mut Vec<i64>, left: usize, right: usize) {
    if {0} { return; }
    let mid = {1};
    p.split("arr", left, mid, right);
    merge_sort(p, arr, left, mid);
    merge_sort(p, arr, mid, right);
    p.merge("arr", left, mid, right);
    let mut tmp = Vec::new();
    let (mut i, mut j) = (left, mid);
    while i < mid && j < right {
        p.compare("arr", i, j);
        if arr[i] <= arr[j] { tmp.push(arr[i]); i += 1; }
        else { tmp.push(arr[j]); j += 1; }
    }
    while i < mid { tmp.push(arr[i]); i += 1; }
    while j < right { tmp.push(arr[j]); j += 1; }
    for (k, &v) in tmp.iter().enumerate() {
        arr[left + k] = v;
        p.write("arr", left + k, v);
    }
}

fn main() {
    let p = Probe::new();
    let mut arr = vec![64i64, 34, 25, 12, 22, 11, 90];
    p.init("arr", &arr);
    let n = arr.len();
    merge_sort(&p, &mut arr, 0, n);
    p.done("arr");
}
`,
			},
		},
	})
}

#!/usr/bin/env python3
# Reference: Binary Search — generates puzzle 2 trace
import sys
sys.path.insert(0, '../probes/python')
from probe import Probe

def run():
    arr = [11, 12, 22, 25, 34, 64, 90]  # sorted
    target = 25
    p = Probe()
    p.init("arr", arr)
    low, high = 0, len(arr) - 1
    while low <= high:
        mid = (low + high) // 2
        p.bounds("arr", low, high)
        p.pin("arr", "mid", mid)
        p.access("arr", mid)
        p.compare("arr", mid, -1)  # -1 signals "compare to target"
        if arr[mid] == target:
            p.found("arr", mid)
            break
        elif arr[mid] < target:
            low = mid + 1
        else:
            high = mid - 1
    else:
        p.not_found("arr")
    p.done("arr")

run()

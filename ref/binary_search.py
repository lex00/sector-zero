#!/usr/bin/env python3
# Reference: Binary Search — generates puzzle 2 trace
import sys
sys.path.insert(0, '../probes/python')
from probe import Probe

def run():
    arr = [2, 5, 8, 12, 16, 23, 38, 56, 72, 91]  # sorted
    target = 23
    p = Probe()
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
            return
        elif arr[mid] < target:
            low = mid + 1
        else:
            high = mid - 1
    p.not_found("arr")
    p.done("arr")

run()

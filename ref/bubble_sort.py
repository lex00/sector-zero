#!/usr/bin/env python3
# Reference: Bubble Sort — generates puzzle 1 trace
import sys
sys.path.insert(0, '../probes/python')
from probe import Probe

def run():
    arr = [64, 34, 25, 12, 22, 11, 90]
    p = Probe()
    p.init("arr", arr)
    n = len(arr)
    for i in range(n):
        p.pin("arr", "i", i)
        for j in range(n - i - 1):
            p.pin("arr", "j", j)
            p.signal("arr", "compare", [j, j+1])
            p.compare("arr", j, j+1)
            if arr[j] > arr[j+1]:
                p.swap("arr", j, j+1)
                arr[j], arr[j+1] = arr[j+1], arr[j]
    p.done("arr")

run()

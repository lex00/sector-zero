#!/usr/bin/env python3
# Reference: Quicksort — generates puzzle 4 trace
import sys
sys.path.insert(0, '../probes/python')
from probe import Probe

p = Probe()

def partition(arr, low, high):
    pivot = arr[high]
    p.pin("arr", "pivot", high)
    p.signal("arr", "pivot_zone", [high])
    i = low - 1
    for j in range(low, high):
        p.pin("arr", "j", j)
        p.compare("arr", j, high)
        if arr[j] <= pivot:
            i += 1
            p.swap("arr", i, j)
            arr[i], arr[j] = arr[j], arr[i]
    p.swap("arr", i+1, high)
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1

def quick_sort(arr, low, high):
    if low < high:
        pi = partition(arr, low, high)
        quick_sort(arr, low, pi - 1)
        quick_sort(arr, pi + 1, high)

def run():
    arr = [64, 34, 25, 12, 22, 11, 90]
    p.init("arr", arr)
    quick_sort(arr, 0, len(arr) - 1)
    p.done("arr")

run()

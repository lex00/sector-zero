#!/usr/bin/env python3
# Reference: Merge Sort — generates puzzle 3 trace
import sys
sys.path.insert(0, '../probes/python')
from probe import Probe

p = Probe()

def merge_sort(arr, left, right):
    if right - left <= 1:
        return
    mid = (left + right) // 2
    p.split("arr", left, mid, right)
    merge_sort(arr, left, mid)
    merge_sort(arr, mid, right)
    merge(arr, left, mid, right)

def merge(arr, left, mid, right):
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

def run():
    arr = [64, 34, 25, 12, 22, 11, 90]
    p.init("arr", arr)
    merge_sort(arr, 0, len(arr))
    p.done("arr")

run()

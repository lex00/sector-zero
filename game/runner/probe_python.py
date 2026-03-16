# Minimal probe bundled with game binary
import json, sys

class Probe:
    def __init__(self):
        self._state = {}

    def init(self, net, values):
        self._state[net] = list(values)
        print(json.dumps({"v":1,"type":"init","net":net,"values":list(values)}), flush=True)

    def compare(self, net, i, j):
        print(json.dumps({"v":1,"type":"compare","net":net,"i":i,"j":j}), flush=True)

    def swap(self, net, i, j):
        s = self._state.get(net, [])
        if i < len(s) and j < len(s):
            s[i], s[j] = s[j], s[i]
        print(json.dumps({"v":1,"type":"swap","net":net,"i":i,"j":j}), flush=True)

    def pin(self, net, name, pos):
        print(json.dumps({"v":1,"type":"pin","net":net,"name":name,"pos":pos}), flush=True)

    def signal(self, net, name, positions):
        print(json.dumps({"v":1,"type":"signal","net":net,"name":name,"positions":positions}), flush=True)

    def access(self, net, pos):
        print(json.dumps({"v":1,"type":"access","net":net,"pos":pos}), flush=True)

    def found(self, net, pos):
        print(json.dumps({"v":1,"type":"found","net":net,"pos":pos}), flush=True)

    def not_found(self, net):
        print(json.dumps({"v":1,"type":"not_found","net":net}), flush=True)

    def bounds(self, net, low, high):
        print(json.dumps({"v":1,"type":"bounds","net":net,"low":low,"high":high}), flush=True)

    def split(self, net, left, mid, right):
        print(json.dumps({"v":1,"type":"split","net":net,"left":left,"mid":mid,"right":right}), flush=True)

    def merge(self, net, left, mid, right):
        print(json.dumps({"v":1,"type":"merge","net":net,"left":left,"mid":mid,"right":right}), flush=True)

    def write(self, net, pos, value):
        s = self._state.get(net, [])
        if 0 <= pos < len(s):
            s[pos] = value
        print(json.dumps({"v":1,"type":"write","net":net,"pos":pos,"value":value}), flush=True)

    def done(self, net):
        print(json.dumps({"v":1,"type":"done","net":net}), flush=True)

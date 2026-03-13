'use strict';
// sector-zero-probe — JavaScript probe library (CommonJS, stdlib only)
// Usage:
//   const { Probe } = require('./probe');
//   const p = new Probe();
//   p.init('arr', [64, 34, 25, 12, 22, 11, 90]);
//   p.compare('arr', 0, 1);
//   p.swap('arr', 0, 1);
//   p.done('arr');

class Probe {
  constructor(out) {
    // out: optional writable stream; defaults to process.stdout
    this._state = {};
    this._out = out || process.stdout;
  }

  _emit(obj) {
    obj.v = 1;
    // Reorder so "v" appears first
    const ordered = { v: obj.v };
    for (const key of Object.keys(obj)) {
      if (key !== 'v') ordered[key] = obj[key];
    }
    this._out.write(JSON.stringify(ordered) + '\n');
  }

  /** Declare a named array with initial values. Always call first. */
  init(net, values) {
    this._state[net] = values.slice();
    this._emit({ type: 'init', net, values: values.slice() });
  }

  /** Signal a comparison between indices i and j. Does not mutate state. */
  compare(net, i, j) {
    this._emit({ type: 'compare', net, i, j });
  }

  /** Signal a swap of indices i and j. Updates internal state. */
  swap(net, i, j) {
    const arr = this._state[net];
    if (arr) {
      const tmp = arr[i];
      arr[i] = arr[j];
      arr[j] = tmp;
    }
    this._emit({ type: 'swap', net, i, j });
  }

  /** Attach a named cursor (e.g. 'i', 'mid') to position pos. */
  pin(net, name, pos) {
    this._emit({ type: 'pin', net, name, pos });
  }

  /** Emit a named signal highlighting a set of positions. */
  signal(net, name, positions) {
    this._emit({ type: 'signal', net, name, positions: positions.slice() });
  }

  /** Record a single read at pos. */
  access(net, pos) {
    this._emit({ type: 'access', net, pos });
  }

  /** Record that the target was found at pos. */
  found(net, pos) {
    this._emit({ type: 'found', net, pos });
  }

  /** Record that the target was not found. */
  notFound(net) {
    this._emit({ type: 'not_found', net });
  }

  /** Record the current search window [low, high]. */
  bounds(net, low, high) {
    this._emit({ type: 'bounds', net, low, high });
  }

  /** Record a divide step: subarray [left, right) split at mid. */
  split(net, left, mid, right) {
    this._emit({ type: 'split', net, left, mid, right });
  }

  /** Record a merge step: merging subarrays into [left, right). */
  merge(net, left, mid, right) {
    this._emit({ type: 'merge', net, left, mid, right });
  }

  /** Signal that all operations on this net are complete. */
  done(net) {
    this._emit({ type: 'done', net });
  }

  /** Return a copy of the current tracked values for net (for debugging). */
  state(net) {
    const arr = this._state[net];
    return arr ? arr.slice() : null;
  }
}

module.exports = { Probe };

# sector-zero-probe — Ruby probe library (stdlib only)
# Usage:
#   require_relative '../probes/ruby/probe'
#   p = Probe.new
#   p.init('arr', [64, 34, 25, 12, 22, 11, 90])
#   p.compare('arr', 0, 1)
#   p.swap('arr', 0, 1)
#   p.done('arr')

require 'json'

class Probe
  # Creates a new Probe. Optionally pass an IO object to direct output (default: $stdout).
  def initialize(out = $stdout)
    @state = {}
    @out = out
  end

  # Declare a named array with initial values. Always call first.
  def init(net, values)
    @state[net] = values.dup
    emit(type: 'init', net: net, values: values.dup)
  end

  # Signal a comparison between indices i and j. Does not mutate state.
  def compare(net, i, j)
    emit(type: 'compare', net: net, i: i, j: j)
  end

  # Signal a swap of indices i and j. Updates internal state.
  def swap(net, i, j)
    arr = @state[net]
    arr[i], arr[j] = arr[j], arr[i] if arr
    emit(type: 'swap', net: net, i: i, j: j)
  end

  # Attach a named cursor (e.g. 'i', 'mid') to position pos.
  def pin(net, name, pos)
    emit(type: 'pin', net: net, name: name, pos: pos)
  end

  # Emit a named signal highlighting a set of positions.
  def signal(net, name, positions)
    emit(type: 'signal', net: net, name: name, positions: positions.dup)
  end

  # Record a single read at pos.
  def access(net, pos)
    emit(type: 'access', net: net, pos: pos)
  end

  # Record that the target was found at pos.
  def found(net, pos)
    emit(type: 'found', net: net, pos: pos)
  end

  # Record that the target was not found.
  def not_found(net)
    emit(type: 'not_found', net: net)
  end

  # Record the current search window [low, high].
  def bounds(net, low, high)
    emit(type: 'bounds', net: net, low: low, high: high)
  end

  # Record a divide step: subarray [left, right) split at mid.
  def split(net, left, mid, right)
    emit(type: 'split', net: net, left: left, mid: mid, right: right)
  end

  # Record a merge step: merging subarrays into [left, right).
  def merge(net, left, mid, right)
    emit(type: 'merge', net: net, left: left, mid: mid, right: right)
  end

  # Signal that all operations on this net are complete.
  def done(net)
    emit(type: 'done', net: net)
  end

  # Return a copy of the current tracked values for net (for debugging).
  def state(net)
    @state[net]&.dup
  end

  private

  def emit(fields)
    # Ensure v:1 appears first in the JSON output
    obj = { v: 1 }.merge(fields)
    @out.puts JSON.generate(obj)
    @out.flush
  end
end

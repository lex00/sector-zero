package sectorzeroprobe;

import java.io.PrintStream;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

/**
 * Probe — Java probe library for the Sector Zero puzzle game.
 *
 * <p>Instruments algorithms and emits NDJSON pulse events to stdout.
 *
 * <p>Usage:
 * <pre>{@code
 * Probe p = new Probe();
 * p.init("arr", new int[]{64, 34, 25, 12, 22, 11, 90});
 * p.compare("arr", 0, 1);
 * p.swap("arr", 0, 1);
 * p.done("arr");
 * }</pre>
 */
public class Probe {

    private final Map<String, int[]> state = new HashMap<>();
    private final PrintStream out;

    /** Create a Probe that writes to System.out. */
    public Probe() {
        this(System.out);
    }

    /** Create a Probe that writes to a custom PrintStream (useful for testing). */
    public Probe(PrintStream out) {
        this.out = out;
    }

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /** Declare a named array with initial values. Always call first. */
    public void init(String net, int[] values) {
        state.put(net, Arrays.copyOf(values, values.length));
        emit("\"type\":\"init\",\"net\":" + jsonStr(net) + ",\"values\":" + jsonIntArr(values));
    }

    /** Signal a comparison between indices i and j. Does not mutate state. */
    public void compare(String net, int i, int j) {
        emit("\"type\":\"compare\",\"net\":" + jsonStr(net) + ",\"i\":" + i + ",\"j\":" + j);
    }

    /** Signal a swap of indices i and j. Updates internal state. */
    public void swap(String net, int i, int j) {
        int[] arr = state.get(net);
        if (arr != null) {
            int tmp = arr[i];
            arr[i] = arr[j];
            arr[j] = tmp;
        }
        emit("\"type\":\"swap\",\"net\":" + jsonStr(net) + ",\"i\":" + i + ",\"j\":" + j);
    }

    /** Attach a named cursor (e.g. "i", "mid") to position pos. */
    public void pin(String net, String name, int pos) {
        emit("\"type\":\"pin\",\"net\":" + jsonStr(net) + ",\"name\":" + jsonStr(name) + ",\"pos\":" + pos);
    }

    /** Emit a named signal highlighting a set of positions. */
    public void signal(String net, String name, int[] positions) {
        emit("\"type\":\"signal\",\"net\":" + jsonStr(net) + ",\"name\":" + jsonStr(name) + ",\"positions\":" + jsonIntArr(positions));
    }

    /** Record a single read at pos. */
    public void access(String net, int pos) {
        emit("\"type\":\"access\",\"net\":" + jsonStr(net) + ",\"pos\":" + pos);
    }

    /** Record that the target was found at pos. */
    public void found(String net, int pos) {
        emit("\"type\":\"found\",\"net\":" + jsonStr(net) + ",\"pos\":" + pos);
    }

    /** Record that the target was not found. */
    public void notFound(String net) {
        emit("\"type\":\"not_found\",\"net\":" + jsonStr(net));
    }

    /** Record the current search window [low, high]. */
    public void bounds(String net, int low, int high) {
        emit("\"type\":\"bounds\",\"net\":" + jsonStr(net) + ",\"low\":" + low + ",\"high\":" + high);
    }

    /** Record a divide step: subarray [left, right) split at mid. */
    public void split(String net, int left, int mid, int right) {
        emit("\"type\":\"split\",\"net\":" + jsonStr(net) + ",\"left\":" + left + ",\"mid\":" + mid + ",\"right\":" + right);
    }

    /** Record a merge step: merging subarrays into [left, right). */
    public void merge(String net, int left, int mid, int right) {
        emit("\"type\":\"merge\",\"net\":" + jsonStr(net) + ",\"left\":" + left + ",\"mid\":" + mid + ",\"right\":" + right);
    }

    /** Signal that all operations on this net are complete. */
    public void done(String net) {
        emit("\"type\":\"done\",\"net\":" + jsonStr(net));
    }

    /**
     * Return a copy of the current tracked values for net (for debugging).
     * Returns null if net has not been initialised.
     */
    public int[] getState(String net) {
        int[] arr = state.get(net);
        return arr == null ? null : Arrays.copyOf(arr, arr.length);
    }

    // -------------------------------------------------------------------------
    // Internal helpers
    // -------------------------------------------------------------------------

    private void emit(String fields) {
        out.println("{\"v\":1," + fields + "}");
        out.flush();
    }

    private static String jsonStr(String s) {
        // Minimal escaping — sufficient for well-formed net/name identifiers.
        return "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"") + "\"";
    }

    private static String jsonIntArr(int[] arr) {
        StringBuilder sb = new StringBuilder("[");
        for (int k = 0; k < arr.length; k++) {
            if (k > 0) sb.append(',');
            sb.append(arr[k]);
        }
        sb.append(']');
        return sb.toString();
    }
}

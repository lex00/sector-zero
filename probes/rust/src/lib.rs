//! sector-zero-probe — Rust probe library
//!
//! Instruments algorithms and emits NDJSON pulse events to stdout for the
//! Sector Zero puzzle game.
//!
//! # Example
//! ```no_run
//! use sector_zero_probe::Probe;
//!
//! let mut p = Probe::new();
//! p.init("arr", &[64, 34, 25, 12, 22, 11, 90]);
//! p.compare("arr", 0, 1);
//! p.swap("arr", 0, 1);
//! p.done("arr");
//! ```

use std::collections::HashMap;
use std::io::{self, Write};

/// Probe tracks named array state and emits NDJSON pulses to stdout.
pub struct Probe {
    state: HashMap<String, Vec<i64>>,
}

impl Probe {
    /// Create a new Probe instance.
    pub fn new() -> Self {
        Probe {
            state: HashMap::new(),
        }
    }

    fn emit(&self, fields: &str) {
        // fields is a pre-built JSON fragment (without the leading '{')
        // We prepend "v":1 to satisfy the pulse format.
        let line = format!("{{\"v\":1,{}}}", fields);
        let stdout = io::stdout();
        let mut handle = stdout.lock();
        writeln!(handle, "{}", line).expect("probe: write to stdout failed");
    }

    fn json_arr(values: &[i64]) -> String {
        let inner: Vec<String> = values.iter().map(|v| v.to_string()).collect();
        format!("[{}]", inner.join(","))
    }

    fn json_pos_arr(positions: &[usize]) -> String {
        let inner: Vec<String> = positions.iter().map(|v| v.to_string()).collect();
        format!("[{}]", inner.join(","))
    }

    /// Declare a named array with initial values. Always call first.
    pub fn init(&mut self, net: &str, values: &[i64]) {
        self.state.insert(net.to_string(), values.to_vec());
        self.emit(&format!(
            "\"type\":\"init\",\"net\":{},\"values\":{}",
            Self::json_str(net),
            Self::json_arr(values)
        ));
    }

    /// Signal a comparison between indices i and j. Does not mutate state.
    pub fn compare(&self, net: &str, i: usize, j: usize) {
        self.emit(&format!(
            "\"type\":\"compare\",\"net\":{},\"i\":{},\"j\":{}",
            Self::json_str(net),
            i,
            j
        ));
    }

    /// Signal a swap of indices i and j. Updates internal state.
    pub fn swap(&mut self, net: &str, i: usize, j: usize) {
        if let Some(arr) = self.state.get_mut(net) {
            arr.swap(i, j);
        }
        self.emit(&format!(
            "\"type\":\"swap\",\"net\":{},\"i\":{},\"j\":{}",
            Self::json_str(net),
            i,
            j
        ));
    }

    /// Attach a named cursor (e.g. "i", "mid") to position pos.
    pub fn pin(&self, net: &str, name: &str, pos: usize) {
        self.emit(&format!(
            "\"type\":\"pin\",\"net\":{},\"name\":{},\"pos\":{}",
            Self::json_str(net),
            Self::json_str(name),
            pos
        ));
    }

    /// Emit a named signal highlighting a set of positions.
    pub fn signal(&self, net: &str, name: &str, positions: &[usize]) {
        self.emit(&format!(
            "\"type\":\"signal\",\"net\":{},\"name\":{},\"positions\":{}",
            Self::json_str(net),
            Self::json_str(name),
            Self::json_pos_arr(positions)
        ));
    }

    /// Record a single read at pos.
    pub fn access(&self, net: &str, pos: usize) {
        self.emit(&format!(
            "\"type\":\"access\",\"net\":{},\"pos\":{}",
            Self::json_str(net),
            pos
        ));
    }

    /// Record that the target was found at pos.
    pub fn found(&self, net: &str, pos: usize) {
        self.emit(&format!(
            "\"type\":\"found\",\"net\":{},\"pos\":{}",
            Self::json_str(net),
            pos
        ));
    }

    /// Record that the target was not found.
    pub fn not_found(&self, net: &str) {
        self.emit(&format!(
            "\"type\":\"not_found\",\"net\":{}",
            Self::json_str(net)
        ));
    }

    /// Record the current search window [low, high].
    pub fn bounds(&self, net: &str, low: usize, high: usize) {
        self.emit(&format!(
            "\"type\":\"bounds\",\"net\":{},\"low\":{},\"high\":{}",
            Self::json_str(net),
            low,
            high
        ));
    }

    /// Record a divide step: subarray [left, right) split at mid.
    pub fn split(&self, net: &str, left: usize, mid: usize, right: usize) {
        self.emit(&format!(
            "\"type\":\"split\",\"net\":{},\"left\":{},\"mid\":{},\"right\":{}",
            Self::json_str(net),
            left,
            mid,
            right
        ));
    }

    /// Record a merge step: merging subarrays into [left, right).
    pub fn merge(&self, net: &str, left: usize, mid: usize, right: usize) {
        self.emit(&format!(
            "\"type\":\"merge\",\"net\":{},\"left\":{},\"mid\":{},\"right\":{}",
            Self::json_str(net),
            left,
            mid,
            right
        ));
    }

    /// Signal that all operations on this net are complete.
    pub fn done(&self, net: &str) {
        self.emit(&format!(
            "\"type\":\"done\",\"net\":{}",
            Self::json_str(net)
        ));
    }

    /// Return a copy of the current tracked values for net (for debugging).
    pub fn state(&self, net: &str) -> Option<Vec<i64>> {
        self.state.get(net).cloned()
    }

    // Minimal JSON string escaping (handles common cases; extend as needed).
    fn json_str(s: &str) -> String {
        format!("\"{}\"", s.replace('\\', "\\\\").replace('"', "\\\""))
    }
}

impl Default for Probe {
    fn default() -> Self {
        Self::new()
    }
}

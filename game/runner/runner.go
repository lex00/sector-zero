package runner

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lex00/sector-zero/game/scope"
)

//go:embed probe_python.py probe_js.js probe_go.go
var ProbeFiles embed.FS

const (
	defaultTimeout    = 5 * time.Second
	compiledTimeout   = 30 * time.Second
)

// CheckRuntimes checks which language runtimes are available on PATH.
func CheckRuntimes() map[string]bool {
	runtimes := map[string]string{
		"python": "python3",
		"go":     "go",
		"js":     "node",
		"rust":   "cargo",
		"java":   "javac",
		"ruby":   "ruby",
	}
	available := make(map[string]bool)
	for lang, cmd := range runtimes {
		_, err := exec.LookPath(cmd)
		available[lang] = err == nil
	}
	return available
}

// Run executes the player's code in the given language and returns the emitted
// Pulse trace. It enforces a timeout and cleans up temp files.
func Run(lang, code string, timeout time.Duration) ([]scope.Pulse, error) {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	tmpDir, err := os.MkdirTemp("", "sz_run_*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract probe files into tmpDir.
	if err := extractProbes(tmpDir); err != nil {
		return nil, fmt.Errorf("extract probes: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var output []byte

	switch lang {
	case "python":
		output, err = runPython(ctx, tmpDir, code)
	case "go":
		output, err = runGo(ctx, tmpDir, code)
	case "js":
		output, err = runJS(ctx, tmpDir, code)
	case "rust":
		timeout = compiledTimeout
		ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
		defer cancel2()
		output, err = runRust(ctx2, tmpDir, code)
	case "java":
		timeout = compiledTimeout
		ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
		defer cancel2()
		output, err = runJava(ctx2, tmpDir, code)
	case "ruby":
		output, err = runRuby(ctx, tmpDir, code)
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	if err != nil {
		return nil, err
	}

	return scope.ParseTrace(output)
}

// extractProbes writes the embedded probe files to dir.
func extractProbes(dir string) error {
	files := []string{"probe_python.py", "probe_js.js", "probe_go.go"}
	for _, name := range files {
		data, err := ProbeFiles.ReadFile(name)
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, name)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func runPython(ctx context.Context, dir, code string) ([]byte, error) {
	// Prepend probe import shim.
	shim := "import sys, os; sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))\n" +
		"from probe_python import Probe\n"
	full := shim + code

	srcFile := filepath.Join(dir, "solution.py")
	if err := os.WriteFile(srcFile, []byte(full), 0o644); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "python3", srcFile)
	cmd.Dir = dir
	return captureOutput(cmd)
}

func runGo(ctx context.Context, dir, code string) ([]byte, error) {
	// The probe_go.go file uses build tag ignore; we need to inline a thin wrapper.
	// For player Go code, we create a small module.
	modFile := `module szsolution

go 1.21
`
	// Wrap player code in a main package if not already present.
	if !strings.Contains(code, "package main") {
		code = "package main\n\n" + code
	}

	// Inject probe import alias if needed.
	if !strings.Contains(code, "probe.") {
		// Player doesn't use probe at all; run as-is.
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modFile), 0o644); err != nil {
		return nil, err
	}

	// Copy the Go probe as probe.go in the same package.
	probeData, err := ProbeFiles.ReadFile("probe_go.go")
	if err != nil {
		return nil, err
	}
	// Strip build tag for actual compilation.
	probeStr := strings.ReplaceAll(string(probeData), "// +build ignore\n\n", "")
	probeStr = strings.ReplaceAll(probeStr, "//go:build ignore\n\n", "")
	if err := os.WriteFile(filepath.Join(dir, "probe.go"), []byte(probeStr), 0o644); err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(dir, "solution.go"), []byte(code), 0o644); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = dir
	return captureOutput(cmd)
}

func runJS(ctx context.Context, dir, code string) ([]byte, error) {
	shim := "const { Probe } = require('./probe_js.js');\n"
	full := shim + code

	srcFile := filepath.Join(dir, "solution.js")
	if err := os.WriteFile(srcFile, []byte(full), 0o644); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "node", srcFile)
	cmd.Dir = dir
	return captureOutput(cmd)
}

func runRuby(ctx context.Context, dir, code string) ([]byte, error) {
	// Write a minimal Ruby probe.
	rubyProbe := `require 'json'

class Probe
  def initialize
    @state = {}
  end
  def init(net, values)
    @state[net] = values.dup
    puts JSON.generate({v:1,type:"init",net:net,values:values})
    $stdout.flush
  end
  def compare(net, i, j)
    puts JSON.generate({v:1,type:"compare",net:net,i:i,j:j})
    $stdout.flush
  end
  def swap(net, i, j)
    s = @state[net] || []
    s[i], s[j] = s[j], s[i]
    puts JSON.generate({v:1,type:"swap",net:net,i:i,j:j})
    $stdout.flush
  end
  def pin(net, name, pos)
    puts JSON.generate({v:1,type:"pin",net:net,name:name,pos:pos})
    $stdout.flush
  end
  def signal(net, name, positions)
    puts JSON.generate({v:1,type:"signal",net:net,name:name,positions:positions})
    $stdout.flush
  end
  def done(net)
    puts JSON.generate({v:1,type:"done",net:net})
    $stdout.flush
  end
end
`
	if err := os.WriteFile(filepath.Join(dir, "probe.rb"), []byte(rubyProbe), 0o644); err != nil {
		return nil, err
	}

	shim := "require_relative 'probe'\n"
	full := shim + code
	srcFile := filepath.Join(dir, "solution.rb")
	if err := os.WriteFile(srcFile, []byte(full), 0o644); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "ruby", srcFile)
	cmd.Dir = dir
	return captureOutput(cmd)
}

func runRust(ctx context.Context, dir, code string) ([]byte, error) {
	// Create a minimal Cargo project.
	cargoToml := `[package]
name = "szsolution"
version = "0.1.0"
edition = "2021"
`
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargoToml), 0o644); err != nil {
		return nil, err
	}

	// Prepend a minimal probe struct.
	probeRust := `use std::io::{self, Write};

struct Probe;
impl Probe {
    fn new() -> Self { Probe }
    fn init(&self, net: &str, values: &[i64]) {
        let s: Vec<String> = values.iter().map(|v| v.to_string()).collect();
        println!(r#"{{"v":1,"type":"init","net":"{}","values":[{}]}}"#, net, s.join(","));
        io::stdout().flush().unwrap();
    }
    fn compare(&self, net: &str, i: usize, j: usize) {
        println!(r#"{{"v":1,"type":"compare","net":"{}","i":{},"j":{}}}"#, net, i, j);
        io::stdout().flush().unwrap();
    }
    fn swap_vals(&self, net: &str, i: usize, j: usize) {
        println!(r#"{{"v":1,"type":"swap","net":"{}","i":{},"j":{}}}"#, net, i, j);
        io::stdout().flush().unwrap();
    }
    fn done(&self, net: &str) {
        println!(r#"{{"v":1,"type":"done","net":"{}"}}"#, net);
        io::stdout().flush().unwrap();
    }
}
`
	full := probeRust + "\n" + code
	if err := os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte(full), 0o644); err != nil {
		return nil, err
	}

	buildCmd := exec.CommandContext(ctx, "cargo", "build", "--release", "-q")
	buildCmd.Dir = dir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cargo build: %w\n%s", err, out)
	}

	runCmd := exec.CommandContext(ctx, filepath.Join(dir, "target", "release", "szsolution"))
	runCmd.Dir = dir
	return captureOutput(runCmd)
}

func runJava(ctx context.Context, dir, code string) ([]byte, error) {
	// Write a minimal Java probe.
	javaProbe := `import java.io.*;

public class Probe {
    public void init(String net, int[] values) throws IOException {
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < values.length; i++) {
            if (i > 0) sb.append(",");
            sb.append(values[i]);
        }
        sb.append("]");
        System.out.println("{\"v\":1,\"type\":\"init\",\"net\":\"" + net + "\",\"values\":" + sb + "}");
        System.out.flush();
    }
    public void compare(String net, int i, int j) {
        System.out.println("{\"v\":1,\"type\":\"compare\",\"net\":\"" + net + "\",\"i\":" + i + ",\"j\":" + j + "}");
        System.out.flush();
    }
    public void swap(String net, int i, int j) {
        System.out.println("{\"v\":1,\"type\":\"swap\",\"net\":\"" + net + "\",\"i\":" + i + ",\"j\":" + j + "}");
        System.out.flush();
    }
    public void done(String net) {
        System.out.println("{\"v\":1,\"type\":\"done\",\"net\":\"" + net + "\"}");
        System.out.flush();
    }
}
`
	if err := os.WriteFile(filepath.Join(dir, "Probe.java"), []byte(javaProbe), 0o644); err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(dir, "Solution.java"), []byte(code), 0o644); err != nil {
		return nil, err
	}

	compileCmd := exec.CommandContext(ctx, "javac", "Probe.java", "Solution.java")
	compileCmd.Dir = dir
	if out, err := compileCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("javac: %w\n%s", err, out)
	}

	runCmd := exec.CommandContext(ctx, "java", "Solution")
	runCmd.Dir = dir
	return captureOutput(runCmd)
}

// captureOutput runs cmd and returns its combined stdout. Stderr is captured
// separately and returned as part of the error on failure.
func captureOutput(cmd *exec.Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

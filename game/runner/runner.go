package runner

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lex00/sector-zero/game/scope"
)

//go:embed probe_python.py probe_js.js probe_go.go
var ProbeFiles embed.FS

const (
	defaultTimeout  = 5 * time.Second
	compiledTimeout = 30 * time.Second
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
		ctx2, cancel2 := context.WithTimeout(context.Background(), compiledTimeout)
		defer cancel2()
		output, err = runRust(ctx2, tmpDir, code)
	case "java":
		ctx2, cancel2 := context.WithTimeout(context.Background(), compiledTimeout)
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

package save

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Save holds persistent game state written to ~/.sector-zero/save.json.
type Save struct {
	Version        int     `json:"version"`
	HelpLevel      string  `json:"help_level"`      // "BLACKOUT"|"STATIC"|"SIGNAL"|"OPEN"
	CurrentPuzzle  int     `json:"current_puzzle"`
	Completed      []int   `json:"completed"`
	FusesRemaining int     `json:"fuses_remaining"`
	Heat           float64 `json:"heat"`
}

func saveDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sector-zero"), nil
}

func savePath() (string, error) {
	dir, err := saveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "save.json"), nil
}

// Default returns a fresh Save with sensible starting values.
func Default() Save {
	return Save{
		Version:        1,
		HelpLevel:      "OPEN",
		CurrentPuzzle:  1,
		Completed:      []int{},
		FusesRemaining: 3,
		Heat:           0,
	}
}

// Load reads ~/.sector-zero/save.json. Returns Default() if the file does not
// exist. Any other error (e.g. corrupt JSON) is returned to the caller.
func Load() (Save, error) {
	path, err := savePath()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), err
	}

	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return Default(), err
	}
	// Migrate pre-v1 saves: force OPEN level so new players get useful hints.
	if s.Version < 1 {
		s.HelpLevel = "OPEN"
		s.Version = 1
	}
	return s, nil
}

// Write serialises the Save and writes it to ~/.sector-zero/save.json,
// creating the directory if necessary.
func (s Save) Write() error {
	dir, err := saveDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(dir, "save.json")

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

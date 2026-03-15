package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerInit(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	if err := mgr.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	stateDir := filepath.Join(tmpDir, StateDirName)
	if _, err := os.Stat(stateDir); err != nil {
		t.Errorf("State directory not created: %v", err)
	}
}

func TestSetAndGetCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Set current
	if err := mgr.SetCurrent("auth/login", "/project/.feat.yml"); err != nil {
		t.Fatalf("SetCurrent failed: %v", err)
	}

	// Get current
	state, err := mgr.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	if state == nil {
		t.Fatal("Expected state, got nil")
	}

	if state.FeaturePath != "auth/login" {
		t.Errorf("FeaturePath = %q, want %q", state.FeaturePath, "auth/login")
	}

	if state.ManifestPath != "/project/.feat.yml" {
		t.Errorf("ManifestPath = %q, want %q", state.ManifestPath, "/project/.feat.yml")
	}

	if state.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestGetCurrentEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	state, err := mgr.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	if state != nil {
		t.Error("Expected nil state for empty directory")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Set then clear
	if err := mgr.SetCurrent("auth/login", "/project/.feat.yml"); err != nil {
		t.Fatalf("SetCurrent failed: %v", err)
	}

	if err := mgr.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	state, err := mgr.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	if state != nil {
		t.Error("Expected nil state after clear")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	if mgr.Exists() {
		t.Error("Expected Exists() = false before Init")
	}

	if err := mgr.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if !mgr.Exists() {
		t.Error("Expected Exists() = true after Init")
	}
}

func TestFormatState(t *testing.T) {
	tests := []struct {
		name     string
		state    *State
		contains []string
	}{
		{
			name:     "nil state",
			state:    nil,
			contains: []string{"No active feature"},
		},
		{
			name: "full state",
			state: &State{
				FeaturePath:  "auth/login",
				ManifestPath: "/project/.feat.yml",
			},
			contains: []string{"auth/login", "/project/.feat.yml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatState(tt.state)
			for _, s := range tt.contains {
				if !contains(output, s) {
					t.Errorf("FormatState() output missing %q: got %q", s, output)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

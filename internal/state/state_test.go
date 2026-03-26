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

	featuresDir := filepath.Join(stateDir, "features")
	if _, err := os.Stat(featuresDir); err != nil {
		t.Errorf("Features directory not created: %v", err)
	}
}

func TestSetAndGetCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Set current
	if err := mgr.SetCurrent("auth/login"); err != nil {
		t.Fatalf("SetCurrent failed: %v", err)
	}

	// Get current
	current, err := mgr.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	if current != "auth/login" {
		t.Errorf("GetCurrent() = %q, want %q", current, "auth/login")
	}
}

func TestGetCurrentEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Get current without setting
	current, err := mgr.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	if current != "" {
		t.Errorf("GetCurrent() = %q, want empty string", current)
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Set then clear
	if err := mgr.SetCurrent("auth/login"); err != nil {
		t.Fatalf("SetCurrent failed: %v", err)
	}

	if err := mgr.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	current, err := mgr.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	if current != "" {
		t.Error("Expected empty string after clear")
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

func TestSanitizeFeaturePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"auth/login", "auth-login"},
		{"feature-name", "feature-name"},
		{"deep/nested/path", "deep-nested-path"},
	}

	for _, tt := range tests {
		result := SanitizeFeaturePath(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFeaturePath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSetAndGetFeatureStep(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Set workflow for default step
	mgr.SetWorkflow([]string{"spec", "tests", "implement", "docs"})

	// Set step
	if err := mgr.SetFeatureStep("auth/login", "tests"); err != nil {
		t.Fatalf("SetFeatureStep failed: %v", err)
	}

	// Get step
	step, err := mgr.GetFeatureStep("auth/login")
	if err != nil {
		t.Fatalf("GetFeatureStep failed: %v", err)
	}

	if step != "tests" {
		t.Errorf("GetFeatureStep() = %q, want %q", step, "tests")
	}
}

func TestGetFeatureStepDefault(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Set workflow so we have a default
	mgr.SetWorkflow([]string{"spec", "tests", "implement", "docs"})

	// Get step without setting (should return first workflow step)
	step, err := mgr.GetFeatureStep("auth/login")
	if err != nil {
		t.Fatalf("GetFeatureStep failed: %v", err)
	}

	if step != "spec" {
		t.Errorf("GetFeatureStep() = %q, want %q (default)", step, "spec")
	}
}

func TestGetFeatureStepNoWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// No workflow set - should return empty string
	step, err := mgr.GetFeatureStep("auth/login")
	if err != nil {
		t.Fatalf("GetFeatureStep failed: %v", err)
	}

	if step != "" {
		t.Errorf("GetFeatureStep() = %q, want empty string", step)
	}
}

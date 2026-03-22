package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest with Tree structure
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test-project",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go", "auth/types.go"},
					Tests: []string{"auth/interface_test.go"},
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{"auth/login/handler.go", "auth/login/types.go"},
							Tests: []string{"auth/login/handler_test.go"},
						},
					},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "feat.yaml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create actual files
	authDir := filepath.Join(tmpDir, "auth", "login")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create implementation files
	files := []string{
		filepath.Join(authDir, "handler.go"),
		filepath.Join(authDir, "types.go"),
		filepath.Join(authDir, "handler_test.go"),
		filepath.Join(tmpDir, "auth", "interface.go"),
		filepath.Join(tmpDir, "auth", "types.go"),
		filepath.Join(tmpDir, "auth", "interface_test.go"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("package"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	// Load the feature
	l := New(m, manifestPath)
	result, err := l.Load("auth/login")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if result.FeaturePath != "auth/login" {
		t.Errorf("FeaturePath = %q, want %q", result.FeaturePath, "auth/login")
	}

	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	if len(result.Tests) != 1 {
		t.Errorf("Expected 1 test file, got %d", len(result.Tests))
	}

	// AncestorFiles should contain 2 implementation files (not tests)
	if len(result.AncestorFiles) != 2 {
		t.Errorf("Expected 2 ancestor files, got %d", len(result.AncestorFiles))
	}

	if len(result.MissingFiles) != 0 {
		t.Errorf("Expected 0 missing files, got %d: %v", len(result.MissingFiles), result.MissingFiles)
	}
}

func TestLoadMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/missing.go"},
					Tests: []string{"auth/missing_test.go"},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "feat.yaml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	l := New(m, manifestPath)
	result, err := l.Load("auth")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(result.MissingFiles) != 2 {
		t.Errorf("Expected 2 missing files, got %d", len(result.MissingFiles))
	}
}

func TestLoadNotLeaf(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go"},
					Children: map[string]manifest.Node{
						"login": {Files: []string{"auth/login.go"}},
					},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "feat.yaml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	l := New(m, manifestPath)
	_, err := l.Load("auth")
	if err == nil {
		t.Error("Expected error for boundary (non-feature)")
	}
}

func TestLoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     "test",
			Children: map[string]manifest.Node{},
		},
	}

	manifestPath := filepath.Join(tmpDir, "feat.yaml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	l := New(m, manifestPath)
	_, err := l.Load("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent feature")
	}
}

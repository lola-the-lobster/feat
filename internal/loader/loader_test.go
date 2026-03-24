package loader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest with Tree structure and higher max_files
	m := &manifest.Manifest{
		Config: manifest.Config{MaxFiles: 5},
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

func TestLoadMaxFilesExceeded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest with max_files = 2
	// Feature has 3 direct files, which exceeds max_files
	m := &manifest.Manifest{
		Config: manifest.Config{MaxFiles: 2},
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go"}, // 1 ancestor file
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{
								"auth/login/handler.go",
								"auth/login/types.go",
								"auth/login/validator.go", // 3 files, exceeds max_files=2
							},
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

	l := New(m, manifestPath)
	_, err := l.Load("auth/login")
	if err == nil {
		t.Error("Expected error when max_files exceeded")
	}

	var limitErr *ContextLimitExceededError
	if !errors.As(err, &limitErr) {
		t.Errorf("Expected ContextLimitExceededError, got %T: %v", err, err)
	} else {
		if limitErr.Total != 3 {
			t.Errorf("Expected Total = 3, got %d", limitErr.Total)
		}
		if limitErr.MaxFiles != 2 {
			t.Errorf("Expected MaxFiles = 2, got %d", limitErr.MaxFiles)
		}
	}
}

func TestLoadMaxFilesWithinLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest with max_files = 3
	m := &manifest.Manifest{
		Config: manifest.Config{MaxFiles: 3},
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go"}, // 1 ancestor file
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{"auth/login/handler.go"}, // 1 file = 2 total
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

	// Create the file
	authDir := filepath.Join(tmpDir, "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "interface.go"), []byte("package"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(authDir, "login"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "login", "handler.go"), []byte("package"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	l := New(m, manifestPath)
	result, err := l.Load("auth/login")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should succeed - 1 file <= 3 max_files
	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result.Files))
	}
	if len(result.AncestorFiles) != 1 {
		t.Errorf("Expected 1 ancestor file, got %d", len(result.AncestorFiles))
	}
}

func TestLoadMaxFilesDefault(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest without config (uses default = 3)
	// Feature has 4 direct files, which exceeds default max_files=3
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go"}, // 1 ancestor file
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{
								"auth/login/handler.go",
								"auth/login/types.go",
								"auth/login/validator.go",
								"auth/login/errors.go", // 4 files, exceeds default max_files=3
							},
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

	l := New(m, manifestPath)
	_, err := l.Load("auth/login")
	if err == nil {
		t.Error("Expected error when default max_files (3) exceeded")
	}

	var limitErr *ContextLimitExceededError
	if !errors.As(err, &limitErr) {
		t.Errorf("Expected ContextLimitExceededError, got %T: %v", err, err)
	} else {
		if limitErr.Total != 4 {
			t.Errorf("Expected Total = 4, got %d", limitErr.Total)
		}
		if limitErr.MaxFiles != 3 {
			t.Errorf("Expected MaxFiles = 3 (default), got %d", limitErr.MaxFiles)
		}
	}
}

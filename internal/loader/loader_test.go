package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest
	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{
			"auth": {
				Interface: "auth/interface.go",
				Children: map[string]manifest.Feature{
					"login": {
						Files: []string{"auth/login/handler.go", "auth/login/types.go"},
					},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, ".feat.yml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create actual files
	authDir := filepath.Join(tmpDir, "auth", "login")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "handler.go"), []byte("package login"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "types.go"), []byte("package login"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "auth", "interface.go"), []byte("package auth"), 0644); err != nil {
		t.Fatalf("Failed to create interface file: %v", err)
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

	if len(result.AncestorInterfaces) != 1 {
		t.Errorf("Expected 1 ancestor interface, got %d", len(result.AncestorInterfaces))
	}

	if len(result.MissingFiles) != 0 {
		t.Errorf("Expected 0 missing files, got %d: %v", len(result.MissingFiles), result.MissingFiles)
	}
}

func TestLoadMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{
			"auth": {
				Files: []string{"auth/missing.go"},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, ".feat.yml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	l := New(m, manifestPath)
	result, err := l.Load("auth")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(result.MissingFiles) != 1 {
		t.Errorf("Expected 1 missing file, got %d", len(result.MissingFiles))
	}
}

func TestLoadNotLeaf(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{
			"auth": {
				Interface: "auth/interface.go",
				Children: map[string]manifest.Feature{
					"login": {Files: []string{"auth/login.go"}},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, ".feat.yml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	l := New(m, manifestPath)
	_, err := l.Load("auth")
	if err == nil {
		t.Error("Expected error for non-leaf feature")
	}
}

func TestLoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{},
	}

	manifestPath := filepath.Join(tmpDir, ".feat.yml")
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	l := New(m, manifestPath)
	_, err := l.Load("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent feature")
	}
}

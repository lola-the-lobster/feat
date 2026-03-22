package split

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestSplit(t *testing.T) {
	tmpDir := t.TempDir()

	// Boundary node: has children, no files
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Children: map[string]manifest.Node{},
				},
			},
		},
	}

	opts := Options{
		ParentPath:  "auth",
		NewName:     "logout",
		CreateFiles: false,
		ManifestDir: tmpDir,
	}

	result, err := Split(m, opts)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if result.NewPath != "auth/logout" {
		t.Errorf("NewPath = %q, want %q", result.NewPath, "auth/logout")
	}

	// Verify node was added
	auth := m.Tree.Children["auth"]
	if _, ok := auth.Children["logout"]; !ok {
		t.Error("Expected 'logout' to be added to auth children")
	}
}

func TestSplitRootLevel(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     "test",
			Children: map[string]manifest.Node{},
		},
	}

	opts := Options{
		ParentPath:  "",
		NewName:     "payments",
		CreateFiles: false,
		ManifestDir: tmpDir,
	}

	result, err := Split(m, opts)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if result.NewPath != "payments" {
		t.Errorf("NewPath = %q, want %q", result.NewPath, "payments")
	}

	if _, ok := m.Tree.Children["payments"]; !ok {
		t.Error("Expected 'payments' to be added to root")
	}
}

func TestSplitDuplicate(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Children: map[string]manifest.Node{
						"login": {Files: []string{"auth/login.go"}},
					},
				},
			},
		},
	}

	opts := Options{
		ParentPath:  "auth",
		NewName:     "login",
		CreateFiles: false,
		ManifestDir: tmpDir,
	}

	_, err := Split(m, opts)
	if err == nil {
		t.Error("Expected error for duplicate node")
	}
}

func TestSplitFeatureParent(t *testing.T) {
	tmpDir := t.TempDir()

	// Feature node: has files, no children
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth.go"},
				},
			},
		},
	}

	opts := Options{
		ParentPath:  "auth",
		NewName:     "child",
		CreateFiles: false,
		ManifestDir: tmpDir,
	}

	_, err := Split(m, opts)
	if err == nil {
		t.Error("Expected error when splitting feature (leaf)")
	}
}

func TestSplitEmptyName(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     "test",
			Children: map[string]manifest.Node{},
		},
	}

	opts := Options{
		ParentPath:  "",
		NewName:     "",
		CreateFiles: false,
		ManifestDir: tmpDir,
	}

	_, err := Split(m, opts)
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

func TestSplitWithSlash(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     "test",
			Children: map[string]manifest.Node{},
		},
	}

	opts := Options{
		ParentPath:  "",
		NewName:     "foo/bar",
		CreateFiles: false,
		ManifestDir: tmpDir,
	}

	_, err := Split(m, opts)
	if err == nil {
		t.Error("Expected error for name with slash")
	}
}

func TestSplitCreateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"auth": {Children: map[string]manifest.Node{}},
			},
		},
	}

	opts := Options{
		ParentPath:  "auth",
		NewName:     "login",
		CreateFiles: true,
		ManifestDir: tmpDir,
	}

	result, err := Split(m, opts)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if len(result.FilesCreated) != 1 {
		t.Errorf("Expected 1 file created, got %d", len(result.FilesCreated))
	}

	// Verify file exists
	featFile := filepath.Join(tmpDir, "auth", "login", ".feat")
	if _, err := os.Stat(featFile); err != nil {
		t.Errorf("Expected file to exist: %s", featFile)
	}
}

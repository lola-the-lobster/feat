package split

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestSplit(t *testing.T) {
	tmpDir := t.TempDir()

	// Intermediate node: has children, no files
	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{
			"auth": {
				Children: map[string]manifest.Feature{},
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

	if result.NewFeaturePath != "auth/logout" {
		t.Errorf("NewFeaturePath = %q, want %q", result.NewFeaturePath, "auth/logout")
	}

	// Verify feature was added
	auth := m.Features["auth"]
	if _, ok := auth.Children["logout"]; !ok {
		t.Error("Expected 'logout' to be added to auth children")
	}
}

func TestSplitRootLevel(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{},
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

	if result.NewFeaturePath != "payments" {
		t.Errorf("NewFeaturePath = %q, want %q", result.NewFeaturePath, "payments")
	}

	if _, ok := m.Features["payments"]; !ok {
		t.Error("Expected 'payments' to be added to root")
	}
}

func TestSplitDuplicate(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{
			"auth": {
				Children: map[string]manifest.Feature{
					"login": {Files: []string{"auth/login.go"}},
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
		t.Error("Expected error for duplicate feature")
	}
}

func TestSplitLeafParent(t *testing.T) {
	tmpDir := t.TempDir()

	// Leaf node: has files, no children
	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{
			"auth": {
				Files: []string{"auth.go"},
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
		t.Error("Expected error when splitting leaf feature")
	}
}

func TestSplitEmptyName(t *testing.T) {
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Features: map[string]manifest.Feature{},
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
		Features: map[string]manifest.Feature{},
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
		Features: map[string]manifest.Feature{
			"auth": {Children: map[string]manifest.Feature{}},
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

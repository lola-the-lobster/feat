// Package loader handles resolving feature paths and loading file contents.
package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

// Result contains the resolved feature and its context.
type Result struct {
	// FeaturePath is the full path to the feature (e.g., "auth/password-reset").
	FeaturePath string

	// Node is the resolved node.
	Node *manifest.Node

	// Files are the absolute paths to the feature's implementation files.
	Files []string

	// Tests are the absolute paths to the feature's test files.
	Tests []string

	// AncestorFiles are the absolute paths to ancestor implementation files.
	AncestorFiles []string

	// MissingFiles are files declared in the manifest but not found on disk.
	MissingFiles []string
}

// Loader handles feature resolution and file loading.
type Loader struct {
	manifest     *manifest.Manifest
	manifestDir  string
	manifestPath string
}

// New creates a new loader for the given manifest.
func New(m *manifest.Manifest, manifestPath string) *Loader {
	return &Loader{
		manifest:     m,
		manifestDir:  filepath.Dir(manifestPath),
		manifestPath: manifestPath,
	}
}

// Load resolves a feature path and collects its files and ancestor files.
func (l *Loader) Load(featurePath string) (*Result, error) {
	node, ancestors, err := l.manifest.GetFeature(featurePath)
	if err != nil {
		return nil, fmt.Errorf("resolving feature: %w", err)
	}

	if node == nil {
		return nil, fmt.Errorf("feature not found: %s", featurePath)
	}

	if !node.IsFeature() {
		return nil, fmt.Errorf("'%s' is not a feature (it has children)", featurePath)
	}

	result := &Result{
		FeaturePath:   featurePath,
		Node:          node,
		Files:         make([]string, 0, len(node.Files)),
		Tests:         make([]string, 0, len(node.Tests)),
		AncestorFiles: make([]string, 0, len(ancestors)),
		MissingFiles:  []string{},
	}

	// Resolve implementation file paths relative to manifest directory
	for _, file := range node.Files {
		absPath := l.resolvePath(file)
		result.Files = append(result.Files, absPath)

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			result.MissingFiles = append(result.MissingFiles, absPath)
		}
	}

	// Resolve test file paths
	for _, file := range node.Tests {
		absPath := l.resolvePath(file)
		result.Tests = append(result.Tests, absPath)

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			result.MissingFiles = append(result.MissingFiles, absPath)
		}
	}

	// Resolve ancestor file paths (only implementation files, not tests)
	for _, file := range ancestors {
		absPath := l.resolvePath(file)
		result.AncestorFiles = append(result.AncestorFiles, absPath)
	}

	return result, nil
}

// resolvePath converts a manifest-relative path to an absolute path.
func (l *Loader) resolvePath(relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(l.manifestDir, relPath)
}

// ReadFile reads the contents of a file.
func (l *Loader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// FormatResult returns a human-readable string representation of the result.
func FormatResult(r *Result) string {
	var output string

	output += fmt.Sprintf("Feature: %s\n", r.FeaturePath)
	output += fmt.Sprintf("Files: %d\n", len(r.Files))
	for _, f := range r.Files {
		exists := ""
		if _, err := os.Stat(f); os.IsNotExist(err) {
			exists = " (missing)"
		}
		output += fmt.Sprintf("  - %s%s\n", f, exists)
	}

	if len(r.Tests) > 0 {
		output += fmt.Sprintf("Tests: %d\n", len(r.Tests))
		for _, f := range r.Tests {
			exists := ""
			if _, err := os.Stat(f); os.IsNotExist(err) {
				exists = " (missing)"
			}
			output += fmt.Sprintf("  - %s%s\n", f, exists)
		}
	}

	if len(r.AncestorFiles) > 0 {
		output += fmt.Sprintf("Ancestor Files: %d\n", len(r.AncestorFiles))
		for _, f := range r.AncestorFiles {
			output += fmt.Sprintf("  - %s\n", f)
		}
	}

	if len(r.MissingFiles) > 0 {
		output += fmt.Sprintf("\nWarning: %d file(s) declared but not found on disk\n", len(r.MissingFiles))
	}

	return output
}

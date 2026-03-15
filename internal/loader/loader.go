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

	// Feature is the resolved feature node.
	Feature *manifest.Feature

	// Files are the absolute paths to the feature's implementation files.
	Files []string

	// AncestorInterfaces are the absolute paths to ancestor interface files.
	AncestorInterfaces []string

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

// Load resolves a feature path and collects its files and ancestor interfaces.
func (l *Loader) Load(featurePath string) (*Result, error) {
	feature, ancestors, err := l.manifest.GetFeature(featurePath)
	if err != nil {
		return nil, fmt.Errorf("resolving feature: %w", err)
	}

	if feature == nil {
		return nil, fmt.Errorf("feature not found: %s", featurePath)
	}

	if !feature.IsLeaf() {
		return nil, fmt.Errorf("'%s' is not a leaf feature (it has no files)", featurePath)
	}

	result := &Result{
		FeaturePath:        featurePath,
		Feature:            feature,
		Files:              make([]string, 0, len(feature.Files)),
		AncestorInterfaces: make([]string, 0, len(ancestors)),
		MissingFiles:       []string{},
	}

	// Resolve file paths relative to manifest directory
	for _, file := range feature.Files {
		absPath := l.resolvePath(file)
		result.Files = append(result.Files, absPath)

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			result.MissingFiles = append(result.MissingFiles, absPath)
		}
	}

	// Resolve ancestor interface paths
	for _, iface := range ancestors {
		absPath := l.resolvePath(iface)
		result.AncestorInterfaces = append(result.AncestorInterfaces, absPath)
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

	if len(r.AncestorInterfaces) > 0 {
		output += fmt.Sprintf("Interfaces: %d\n", len(r.AncestorInterfaces))
		for _, iface := range r.AncestorInterfaces {
			output += fmt.Sprintf("  - %s\n", iface)
		}
	}

	if len(r.MissingFiles) > 0 {
		output += fmt.Sprintf("\nWarning: %d file(s) declared but not found on disk\n", len(r.MissingFiles))
	}

	return output
}

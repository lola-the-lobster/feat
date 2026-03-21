// Package split handles creating new features in the manifest.
package split

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

// Options configures the split operation.
type Options struct {
	// ParentPath is the path to the parent feature (e.g., "auth" or "auth/password-reset").
	// Use empty string for root-level features.
	ParentPath string

	// NewName is the name for the new feature (e.g., "confirmation").
	NewName string

	// CreateFiles if true, creates empty files on disk.
	CreateFiles bool

	// ManifestDir is the directory containing the manifest (for file creation).
	ManifestDir string
}

// Result contains the outcome of a split operation.
type Result struct {
	// NewFeaturePath is the full path to the created feature.
	NewFeaturePath string

	// FilesCreated are the files that were created on disk (if any).
	FilesCreated []string

	// ManifestUpdated is true if the manifest was successfully saved.
	ManifestUpdated bool
}

// Split creates a new feature under the given parent path.
func Split(m *manifest.Manifest, opts Options) (*Result, error) {
	if opts.NewName == "" {
		return nil, fmt.Errorf("new feature name cannot be empty")
	}
	if strings.Contains(opts.NewName, "/") {
		return nil, fmt.Errorf("new feature name cannot contain '/': %s", opts.NewName)
	}

	result := &Result{
		FilesCreated: []string{},
	}

	// Determine the full path and get the parent map
	var parentMap map[string]manifest.Feature
	var newFeaturePath string

	if opts.ParentPath == "" {
		// Root-level feature
		if m.Tree.Features == nil {
			m.Tree.Features = make(map[string]manifest.Feature)
		}
		parentMap = m.Tree.Features
		newFeaturePath = opts.NewName
	} else {
		// Nested feature
		parts := splitPath(opts.ParentPath)
		if len(parts) == 0 {
			return nil, fmt.Errorf("invalid parent path: %s", opts.ParentPath)
		}

		// Navigate to the parent, creating intermediate nodes if needed
		var err error
		parentMap, err = navigateToParent(m, parts)
		if err != nil {
			return nil, err
		}
		newFeaturePath = opts.ParentPath + "/" + opts.NewName
	}

	result.NewFeaturePath = newFeaturePath

	// Check if feature already exists
	if _, exists := parentMap[opts.NewName]; exists {
		return nil, fmt.Errorf("feature already exists: %s", newFeaturePath)
	}

	// Create the new feature (leaf with empty files list)
	newFeature := manifest.Feature{
		Files: []string{},
	}

	// Add to parent's children
	parentMap[opts.NewName] = newFeature

	// Create files on disk if requested
	if opts.CreateFiles {
		files, err := createFeatureFiles(opts.ManifestDir, newFeaturePath, opts.NewName)
		if err != nil {
			return nil, fmt.Errorf("creating feature files: %w", err)
		}
		result.FilesCreated = files
	}

	result.ManifestUpdated = true
	return result, nil
}

// navigateToParent navigates to the parent feature's Children map.
// Creates intermediate nodes if they don't exist.
func navigateToParent(m *manifest.Manifest, parts []string) (map[string]manifest.Feature, error) {
	if m.Tree.Features == nil {
		m.Tree.Features = make(map[string]manifest.Feature)
	}

	current := m.Tree.Features

	for i, part := range parts {
		f, exists := current[part]
		if !exists {
			// Create intermediate node
			f = manifest.Feature{
				Children: make(map[string]manifest.Feature),
			}
			current[part] = f
		}

		// If this is a leaf, we can't add children
		if f.IsLeaf() {
			return nil, fmt.Errorf("cannot add children to leaf feature: %s", strings.Join(parts[:i+1], "/"))
		}

		// If this is the last part, return its Children map
		if i == len(parts)-1 {
			if f.Children == nil {
				f.Children = make(map[string]manifest.Feature)
				current[part] = f
			}
			return f.Children, nil
		}

		// Descend into children
		if f.Children == nil {
			f.Children = make(map[string]manifest.Feature)
			current[part] = f
		}
		current = f.Children
	}

	return current, nil
}

// createFeatureFiles creates empty files for a new feature.
func createFeatureFiles(manifestDir, featurePath, featureName string) ([]string, error) {
	// Create directory structure
	dirPath := filepath.Join(manifestDir, featurePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("creating directory %s: %w", dirPath, err)
	}

	// Create a placeholder file
	placeholder := filepath.Join(dirPath, ".feat")
	if err := os.WriteFile(placeholder, []byte("# "+featureName+" feature\n"), 0644); err != nil {
		return nil, fmt.Errorf("creating placeholder file: %w", err)
	}

	return []string{placeholder}, nil
}

// splitPath splits a feature path into parts.
func splitPath(path string) []string {
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// FormatResult returns a human-readable string representation of the result.
func FormatResult(r *Result) string {
	var output string
	output += fmt.Sprintf("Created feature: %s\n", r.NewFeaturePath)

	if len(r.FilesCreated) > 0 {
		output += fmt.Sprintf("Files created: %d\n", len(r.FilesCreated))
		for _, f := range r.FilesCreated {
			output += fmt.Sprintf("  - %s\n", f)
		}
	}

	if r.ManifestUpdated {
		output += "Manifest updated.\n"
	}

	return output
}

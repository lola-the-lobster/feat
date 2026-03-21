// Package manifest handles parsing and serialization of feat.yaml files.
package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest represents the root feat.yaml file.
type Manifest struct {
	// Features is a map of feature names to Feature nodes.
	// The root level contains systems/subsystems (intermediate nodes)
	// and leaves are actual features.
	Features map[string]Feature `yaml:",inline"`
}

// Feature represents either an intermediate node (subsystem) or a leaf (actual feature).
type Feature struct {
	// Files are the implementation files for this feature.
	// Both intermediate nodes and leaves can have files.
	Files []string `yaml:"files,omitempty"`

	// Tests are the test files for this feature.
	// Kept separate from Files to distinguish implementation from tests.
	Tests []string `yaml:"tests,omitempty"`

	// Children are nested features (subsystems).
	// If Children is non-nil, this is an intermediate node.
	// If Children is nil, this is a leaf (if it has content) or placeholder (if empty).
	Children map[string]Feature `yaml:",inline"`
}

// IsLeaf returns true if this feature is a leaf node.
// A leaf has content (files/tests) and no children.
// Note: An empty feature (no files, no tests, no children) is a placeholder, not a leaf.
func (f Feature) IsLeaf() bool {
	if f.Children != nil {
		return false // Has children → intermediate
	}
	// No children - check if it has content
	hasContent := len(f.Files) > 0 || len(f.Tests) > 0
	return hasContent
}

// IsIntermediate returns true if this feature is an intermediate node.
// Intermediate nodes either:
//   - Have explicit children (non-nil Children map)
//   - Are empty placeholders (no files, no tests, no children defined yet)
func (f Feature) IsIntermediate() bool {
	if f.Children != nil {
		return true // Has explicit children
	}
	// No children - check if it's an empty placeholder
	hasContent := len(f.Files) > 0 || len(f.Tests) > 0
	return !hasContent // Empty = placeholder = intermediate
}

// HasContent returns true if this feature has any files or tests.
func (f Feature) HasContent() bool {
	return len(f.Files) > 0 || len(f.Tests) > 0
}

// AllFiles returns both implementation and test files.
func (f Feature) AllFiles() []string {
	return append(f.Files, f.Tests...)
}

// Load reads a manifest from the given path.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return &m, nil
}

// Save writes the manifest to the given path.
func (m *Manifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// GetFeature retrieves a feature by its path (e.g., "auth/password-reset").
// Returns nil if not found.
func (m *Manifest) GetFeature(path string) (*Feature, []string, error) {
	if path == "" {
		return nil, nil, fmt.Errorf("empty feature path")
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("invalid feature path: %s", path)
	}

	// Collect ancestor implementation files as we traverse
	// Note: ancestor test files are NOT included (tests are feature-specific)
	var ancestors []string

	current := m.Features
	var feature *Feature

	for i, part := range parts {
		f, ok := current[part]
		if !ok {
			return nil, nil, fmt.Errorf("feature not found: %s", path)
		}

		// If this is the last part, we've found our feature
		if i == len(parts)-1 {
			feature = &f
			break
		}

		// Otherwise, this is a parent node - collect its implementation files (not tests)
		ancestors = append(ancestors, f.Files...)

		// Descend into children
		if f.Children == nil {
			return nil, nil, fmt.Errorf("path continues but %s has no children", part)
		}
		current = f.Children
	}

	return feature, ancestors, nil
}

// splitPath splits a feature path like "auth/password-reset" into parts.
func splitPath(path string) []string {
	var parts []string
	var current string
	for _, r := range path {
		if r == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// Validate checks the manifest for common issues.
func (m *Manifest) Validate() []string {
	var issues []string

	if len(m.Features) == 0 {
		issues = append(issues, "manifest has no features defined")
	}

	for name, f := range m.Features {
		issues = append(issues, validateFeature(name, f)...)
	}

	return issues
}

func validateFeature(name string, f Feature) []string {
	var issues []string

	// Note: Intermediate nodes CAN have files/tests. They're not just boundaries.
	// An auth subsystem might have auth/interface.go AND auth/login/ children.

	// Validate children recursively
	if f.Children != nil {
		for childName, child := range f.Children {
			issues = append(issues, validateFeature(name+"/"+childName, child)...)
		}
	}

	return issues
}

// Init creates a minimal manifest at the given path.
func Init(path string) error {
	m := &Manifest{
		Features: make(map[string]Feature),
	}
	return m.Save(path)
}

// Package manifest handles parsing and serialization of .feat.yml files.
package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest represents the root .feat.yml file.
type Manifest struct {
	// Features is a map of feature names to Feature nodes.
	// The root level contains systems/subsystems (intermediate nodes)
	// and leaves are actual features.
	Features map[string]Feature `yaml:",inline"`
}

// Feature represents either an intermediate node (subsystem) or a leaf (actual feature).
type Feature struct {
	// Interface is the path to the interface file (for intermediate nodes).
	Interface string `yaml:"interface,omitempty"`

	// Deps are external dependencies (optional, at any level).
	Deps []string `yaml:"deps,omitempty"`

	// Files are the implementation files (for leaf features only).
	Files []string `yaml:"files,omitempty"`

	// Children are nested features.
	Children map[string]Feature `yaml:",inline"`
}

// IsLeaf returns true if this feature has files (it's a leaf node).
func (f Feature) IsLeaf() bool {
	return len(f.Files) > 0
}

// IsIntermediate returns true if this feature has an interface or children.
func (f Feature) IsIntermediate() bool {
	return f.Interface != "" || len(f.Children) > 0
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

	// Collect ancestor interfaces as we traverse
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

		// Otherwise, this is a parent node - if it has an interface, track it
		if f.Interface != "" {
			ancestors = append(ancestors, f.Interface)
		}

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

	// Check for invalid combinations
	if f.IsLeaf() && len(f.Children) > 0 {
		issues = append(issues, fmt.Sprintf("feature %s has both files and children", name))
	}

	// Warn about features with no interface and no children (orphaned intermediate)
	if !f.IsLeaf() && f.Interface == "" && len(f.Children) == 0 {
		issues = append(issues, fmt.Sprintf("feature %s has no interface and no children", name))
	}

	// Validate children recursively
	for childName, child := range f.Children {
		issues = append(issues, validateFeature(name+"/"+childName, child)...)
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

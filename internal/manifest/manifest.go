// Package manifest handles parsing and serialization of feat.yaml files.
package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest represents the root feat.yaml file.
type Manifest struct {
	// Tree is the root of the feature hierarchy.
	Tree Tree `yaml:"tree"`
}

// Tree represents the root node of the feature hierarchy.
type Tree struct {
	// Name is the project name.
	Name string `yaml:"name"`

	// Files are files at the root level (e.g., go.mod, README.md).
	Files []string `yaml:"files,omitempty"`

	// Children are the top-level nodes (boundaries or features).
	Children map[string]Node `yaml:"children,omitempty"`
}

// Node is the interface for all nodes in the tree.
// A Node is either a Boundary (intermediate, has children) or a Feature (leaf).
type Node struct {
	// Files are the implementation files for this node.
	// Both boundaries and features can have files.
	Files []string `yaml:"files,omitempty"`

	// Tests are the test files for this node.
	// Only features have tests (boundaries don't have direct tests).
	Tests []string `yaml:"tests,omitempty"`

	// Children are nested nodes.
	// If Children is non-nil, this node is a Boundary (intermediate).
	// If Children is nil, this node is a Feature (leaf).
	Children map[string]Node `yaml:",inline"`
}

// IsFeature returns true if this node is a Feature (leaf).
// A feature has content (files/tests) and no children.
// Note: An empty node (no files, no tests, no children) is a placeholder boundary.
func (n Node) IsFeature() bool {
	if n.Children != nil {
		return false // Has children → boundary
	}
	// No children - check if it has content
	hasContent := len(n.Files) > 0 || len(n.Tests) > 0
	return hasContent
}

// IsBoundary returns true if this node is a Boundary (intermediate).
// Boundaries either:
//   - Have explicit children (non-nil Children map)
//   - Are empty placeholders (no files, no tests, no children defined yet)
func (n Node) IsBoundary() bool {
	if n.Children != nil {
		return true // Has explicit children
	}
	// No children - check if it's an empty placeholder
	hasContent := len(n.Files) > 0 || len(n.Tests) > 0
	return !hasContent // Empty = placeholder = boundary
}

// HasContent returns true if this node has any files or tests.
func (n Node) HasContent() bool {
	return len(n.Files) > 0 || len(n.Tests) > 0
}

// AllFiles returns both implementation and test files.
func (n Node) AllFiles() []string {
	return append(n.Files, n.Tests...)
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

// GetNode retrieves a node by its path (e.g., "auth/password-reset").
// Returns nil if not found.
func (m *Manifest) GetNode(path string) (*Node, []string, error) {
	if path == "" {
		return nil, nil, fmt.Errorf("empty path")
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("invalid path: %s", path)
	}

	// Collect ancestor implementation files as we traverse
	// Note: ancestor test files are NOT included (tests are feature-specific)
	var ancestors []string

	current := m.Tree.Children
	var node *Node

	for i, part := range parts {
		n, ok := current[part]
		if !ok {
			return nil, nil, fmt.Errorf("node not found: %s", path)
		}

		// If this is the last part, we've found our node
		if i == len(parts)-1 {
			node = &n
			break
		}

		// Otherwise, this is a parent node - collect its implementation files (not tests)
		ancestors = append(ancestors, n.Files...)

		// Descend into children
		if n.Children == nil {
			return nil, nil, fmt.Errorf("path continues but %s has no children", part)
		}
		current = n.Children
	}

	return node, ancestors, nil
}

// GetFeature is an alias for GetNode for backward compatibility.
// It returns an error if the node is a boundary (not a feature).
func (m *Manifest) GetFeature(path string) (*Node, []string, error) {
	node, ancestors, err := m.GetNode(path)
	if err != nil {
		return nil, nil, err
	}
	if node != nil && node.IsBoundary() {
		return nil, nil, fmt.Errorf("%s is a boundary, not a feature", path)
	}
	return node, ancestors, nil
}

// splitPath splits a path like "auth/password-reset" into parts.
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

	if m.Tree.Name == "" {
		issues = append(issues, "manifest tree has no name")
	}

	if len(m.Tree.Children) == 0 {
		issues = append(issues, "manifest has no children defined")
	}

	for name, n := range m.Tree.Children {
		issues = append(issues, validateNode(name, n)...)
	}

	return issues
}

func validateNode(name string, n Node) []string {
	var issues []string

	// Note: Boundaries CAN have files. They're not just containers.
	// An auth subsystem might have auth/interface.go AND auth/login/ children.

	// Validate children recursively
	if n.Children != nil {
		for childName, child := range n.Children {
			issues = append(issues, validateNode(name+"/"+childName, child)...)
		}
	}

	return issues
}

// Init creates a minimal manifest at the given path.
func Init(path string, projectName string) error {
	m := &Manifest{
		Tree: Tree{
			Name:     projectName,
			Children: make(map[string]Node),
		},
	}
	return m.Save(path)
}

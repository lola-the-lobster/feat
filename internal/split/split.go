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
	// ParentPath is the path to the parent node (e.g., "auth" or "auth/password-reset").
	// Use empty string for root-level nodes.
	ParentPath string

	// NewName is the name for the new node (e.g., "confirmation").
	NewName string

	// CreateFiles if true, creates empty files on disk.
	CreateFiles bool

	// ManifestDir is the directory containing the manifest (for file creation).
	ManifestDir string
}

// Result contains the outcome of a split operation.
type Result struct {
	// NewPath is the full path to the created node.
	NewPath string

	// FilesCreated are the files that were created on disk (if any).
	FilesCreated []string

	// ManifestUpdated is true if the manifest was successfully saved.
	ManifestUpdated bool
}

// Split creates a new node under the given parent path.
func Split(m *manifest.Manifest, opts Options) (*Result, error) {
	if opts.NewName == "" {
		return nil, fmt.Errorf("new name cannot be empty")
	}
	if strings.Contains(opts.NewName, "/") {
		return nil, fmt.Errorf("new name cannot contain '/': %s", opts.NewName)
	}

	result := &Result{
		FilesCreated: []string{},
	}

	// Determine the full path and get the parent map
	var parentMap map[string]manifest.Node
	var newPath string

	if opts.ParentPath == "" {
		// Root-level node
		if m.Tree.Children == nil {
			m.Tree.Children = make(map[string]manifest.Node)
		}
		parentMap = m.Tree.Children
		newPath = opts.NewName
	} else {
		// Nested node
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
		newPath = opts.ParentPath + "/" + opts.NewName
	}

	result.NewPath = newPath

	// Check if node already exists
	if _, exists := parentMap[opts.NewName]; exists {
		return nil, fmt.Errorf("node already exists: %s", newPath)
	}

	// Create the new node (feature with empty files list)
	newNode := manifest.Node{
		Files: []string{},
	}

	// Add to parent's children
	parentMap[opts.NewName] = newNode

	// Create files on disk if requested
	if opts.CreateFiles {
		files, err := createNodeFiles(opts.ManifestDir, newPath, opts.NewName)
		if err != nil {
			return nil, fmt.Errorf("creating node files: %w", err)
		}
		result.FilesCreated = files
	}

	result.ManifestUpdated = true
	return result, nil
}

// navigateToParent navigates to the parent node's Children map.
// Creates intermediate nodes if they don't exist.
func navigateToParent(m *manifest.Manifest, parts []string) (map[string]manifest.Node, error) {
	if m.Tree.Children == nil {
		m.Tree.Children = make(map[string]manifest.Node)
	}

	current := m.Tree.Children

	for i, part := range parts {
		n, exists := current[part]
		if !exists {
			// Create intermediate boundary node
			n = manifest.Node{
				Children: make(map[string]manifest.Node),
			}
			current[part] = n
		}

		// If this is a feature (leaf), we can't add children
		if n.IsFeature() {
			return nil, fmt.Errorf("cannot add children to feature: %s", strings.Join(parts[:i+1], "/"))
		}

		// If this is the last part, return its Children map
		if i == len(parts)-1 {
			if n.Children == nil {
				n.Children = make(map[string]manifest.Node)
				current[part] = n
			}
			return n.Children, nil
		}

		// Descend into children
		if n.Children == nil {
			n.Children = make(map[string]manifest.Node)
			current[part] = n
		}
		current = n.Children
	}

	return current, nil
}

// createNodeFiles creates empty files for a new node.
func createNodeFiles(manifestDir, nodePath, nodeName string) ([]string, error) {
	// Create directory structure
	dirPath := filepath.Join(manifestDir, nodePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("creating directory %s: %w", dirPath, err)
	}

	// Create a placeholder file
	placeholder := filepath.Join(dirPath, ".feat")
	if err := os.WriteFile(placeholder, []byte("# "+nodeName+" feature\n"), 0644); err != nil {
		return nil, fmt.Errorf("creating placeholder file: %w", err)
	}

	return []string{placeholder}, nil
}

// splitPath splits a path into parts.
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
	output += fmt.Sprintf("Created node: %s\n", r.NewPath)

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

// Package formatter provides JSON and text formatting for feat CLI output.
package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lola-the-lobster/feat/internal/loader"
	"github.com/lola-the-lobster/feat/internal/manifest"
	"github.com/lola-the-lobster/feat/internal/state"
)

// FeatureJSON represents a feature in JSON output.
type FeatureJSON struct {
	Path   string   `json:"path"`
	Type   string   `json:"type"`
	Files  []string `json:"files,omitempty"`
	Tests  []string `json:"tests,omitempty"`
}

// ListJSON represents the JSON output for the list command.
type ListJSON struct {
	Project  string        `json:"project"`
	Features []FeatureJSON `json:"features"`
}

// StatusJSON represents the JSON output for the status command.
type StatusJSON struct {
	CurrentFeature string   `json:"current_feature,omitempty"`
	ManifestPath   string   `json:"manifest_path,omitempty"`
	Files          []string `json:"files,omitempty"`
	Tests          []string `json:"tests,omitempty"`
	Ancestors      []string `json:"ancestors,omitempty"`
	MissingFiles   []string `json:"missing_files,omitempty"`
	Timestamp      string   `json:"timestamp,omitempty"`
}

// ErrorJSON represents a JSON error response.
type ErrorJSON struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// FormatListJSON returns a JSON representation of the manifest tree.
func FormatListJSON(m *manifest.Manifest) ([]byte, error) {
	output := ListJSON{
		Project:  m.Tree.Name,
		Features: collectFeatures(m.Tree.Children, ""),
	}

	return json.MarshalIndent(output, "", "  ")
}

// collectFeatures recursively collects all features from the tree.
func collectFeatures(children map[string]manifest.Node, prefix string) []FeatureJSON {
	var features []FeatureJSON

	// Sort keys for consistent output
	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		node := children[name]
		var path string
		if prefix == "" {
			path = name
		} else {
			path = prefix + "/" + name
		}

		if node.IsFeature() {
			features = append(features, FeatureJSON{
				Path:  path,
				Type:  "feature",
				Files: node.Files,
				Tests: node.Tests,
			})
		} else {
			// Boundary node
			features = append(features, FeatureJSON{
				Path:  path,
				Type:  "boundary",
				Files: node.Files,
				Tests: node.Tests,
			})

			// Recurse into children
			if len(node.Children) > 0 {
				features = append(features, collectFeatures(node.Children, path)...)
			}
		}
	}

	return features
}

// FormatStatusJSON returns a JSON representation of the current status.
func FormatStatusJSON(s *state.State, result *loader.Result) ([]byte, error) {
	output := StatusJSON{}

	if s != nil {
		output.CurrentFeature = s.FeaturePath
		output.ManifestPath = s.ManifestPath
		if !s.Timestamp.IsZero() {
			output.Timestamp = s.Timestamp.Format(time.RFC3339)
		}
	}

	if result != nil {
		output.Files = result.Files
		output.Tests = result.Tests
		output.Ancestors = result.AncestorFiles
		output.MissingFiles = result.MissingFiles
	}

	return json.MarshalIndent(output, "", "  ")
}

// FormatErrorJSON returns a JSON error response.
func FormatErrorJSON(err error, code int) []byte {
	output := ErrorJSON{
		Error: err.Error(),
		Code:  code,
	}

	data, _ := json.Marshal(output)
	return data
}

// PrintErrorJSON prints a JSON error to stderr and exits with the given code.
func PrintErrorJSON(err error, code int) {
	data := FormatErrorJSON(err, code)
	fmt.Fprintln(os.Stderr, string(data))
	os.Exit(code)
}

// ResolvePaths makes relative paths absolute for JSON output.
func ResolvePaths(paths []string, baseDir string) []string {
	resolved := make([]string, len(paths))
	for i, p := range paths {
		if filepath.IsAbs(p) {
			resolved[i] = p
		} else {
			resolved[i] = filepath.Join(baseDir, p)
		}
	}
	return resolved
}

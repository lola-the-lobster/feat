// Package tree handles formatting and displaying the feature tree.
package tree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

// Printer handles formatting the tree for display.
type Printer struct {
	// IndentString is the string used for each level of indentation.
	// Default is two spaces.
	IndentString string
}

// NewPrinter creates a new tree printer with default settings.
func NewPrinter() *Printer {
	return &Printer{
		IndentString: "  ",
	}
}

// Print outputs the feature tree in a readable format.
// Format:
//   auth/
//     password-reset
//     email-verification
func (p *Printer) Print(m *manifest.Manifest) string {
	var b strings.Builder
	p.printFeatures(m.Features, 0, &b)
	return b.String()
}

// printFeatures recursively prints features in sorted order.
func (p *Printer) printFeatures(features map[string]manifest.Feature, depth int, b *strings.Builder) {
	// Sort keys for consistent output
	names := make([]string, 0, len(features))
	for name := range features {
		names = append(names, name)
	}
	sort.Strings(names)

	indent := strings.Repeat(p.IndentString, depth)

	for _, name := range names {
		feature := features[name]
		p.printFeature(name, feature, indent, depth, b)
	}
}

// printFeature prints a single feature node.
func (p *Printer) printFeature(name string, f manifest.Feature, indent string, depth int, b *strings.Builder) {
	if f.IsLeaf() {
		// Leaf feature: just print the name
		fmt.Fprintf(b, "%s%s\n", indent, name)
	} else {
		// Intermediate node: print name with trailing slash
		fmt.Fprintf(b, "%s%s/\n", indent, name)
		// Recurse into children
		if len(f.Children) > 0 {
			p.printFeatures(f.Children, depth+1, b)
		}
	}
}

// ListFormat returns a simple list of all feature paths (for machine-readable output).
func ListPaths(m *manifest.Manifest) []string {
	var paths []string
	collectPaths(m.Features, "", &paths)
	return paths
}

// collectPaths recursively collects all feature paths.
func collectPaths(features map[string]manifest.Feature, prefix string, paths *[]string) {
	names := make([]string, 0, len(features))
	for name := range features {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		feature := features[name]
		var path string
		if prefix == "" {
			path = name
		} else {
			path = prefix + "/" + name
		}

		if feature.IsLeaf() {
			*paths = append(*paths, path)
		} else {
			// Intermediate nodes with files can be valid targets
			if len(feature.Files) > 0 {
				*paths = append(*paths, path+"/")
			}
			// Recurse into children
			if len(feature.Children) > 0 {
				collectPaths(feature.Children, path, paths)
			}
		}
	}
}

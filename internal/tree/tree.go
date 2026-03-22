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

// Print outputs the tree in a readable format.
// Format:
//   auth/
//     password-reset (feature)
//     email-verification (feature)
func (p *Printer) Print(m *manifest.Manifest) string {
	var b strings.Builder
	p.printChildren(m.Tree.Children, 0, &b)
	return b.String()
}

// printChildren recursively prints nodes in sorted order.
func (p *Printer) printChildren(children map[string]manifest.Node, depth int, b *strings.Builder) {
	// Sort keys for consistent output
	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)

	indent := strings.Repeat(p.IndentString, depth)

	for _, name := range names {
		node := children[name]
		p.printNode(name, node, indent, depth, b)
	}
}

// printNode prints a single node.
func (p *Printer) printNode(name string, n manifest.Node, indent string, depth int, b *strings.Builder) {
	if n.IsFeature() {
		// Feature (leaf): print the name with (feature)
		fmt.Fprintf(b, "%s%s (feature)\n", indent, name)
	} else {
		// Boundary (intermediate): print name with trailing slash
		fmt.Fprintf(b, "%s%s/\n", indent, name)
		// Recurse into children
		if len(n.Children) > 0 {
			p.printChildren(n.Children, depth+1, b)
		}
	}
}

// ListPaths returns a simple list of all feature paths (for machine-readable output).
func ListPaths(m *manifest.Manifest) []string {
	var paths []string
	collectPaths(m.Tree.Children, "", &paths)
	return paths
}

// collectPaths recursively collects all feature paths.
func collectPaths(children map[string]manifest.Node, prefix string, paths *[]string) {
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
			*paths = append(*paths, path)
		} else {
			// Boundaries with files can be valid targets
			if len(node.Files) > 0 {
				*paths = append(*paths, path+"/")
			}
			// Recurse into children
			if len(node.Children) > 0 {
				collectPaths(node.Children, path, paths)
			}
		}
	}
}

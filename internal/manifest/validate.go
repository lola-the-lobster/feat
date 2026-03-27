// Package manifest handles parsing and serialization of feat.yaml files.
package manifest

import (
	"fmt"
)

// CircularError represents a circular reference in the manifest tree.
type CircularError struct {
	Path   string
	Node   string
	Cycle  string
}

func (e *CircularError) Error() string {
	return fmt.Sprintf("circular reference detected at %q: %s references ancestor %s", e.Path, e.Node, e.Cycle)
}

// ValidateCircular checks for circular parent-child relationships in the tree.
// It returns an error if a node has an ancestor with the same name, creating a cycle.
func (m *Manifest) ValidateCircular() error {
	for name, node := range m.Tree.Children {
		if err := validateNodeCircular(name, node, []string{name}); err != nil {
			return err
		}
	}
	return nil
}

// validateNodeCircular recursively checks for circular references.
// ancestors is the path from root to the current node (inclusive).
func validateNodeCircular(name string, n Node, ancestors []string) error {
	if n.Children == nil {
		return nil
	}

	for childName, child := range n.Children {
		// Check if this child name already exists in ancestors
		for _, ancestor := range ancestors {
			if childName == ancestor {
				return &CircularError{
					Path:  buildPath(ancestors),
					Node:  childName,
					Cycle: ancestor,
				}
			}
		}

		// Add this child to ancestors and recurse
		newAncestors := append(ancestors, childName)
		if err := validateNodeCircular(childName, child, newAncestors); err != nil {
			return err
		}
	}

	return nil
}

// buildPath joins ancestor names with "/" for error messages.
func buildPath(ancestors []string) string {
	if len(ancestors) == 0 {
		return ""
	}
	result := ancestors[0]
	for i := 1; i < len(ancestors); i++ {
		result += "/" + ancestors[i]
	}
	return result
}
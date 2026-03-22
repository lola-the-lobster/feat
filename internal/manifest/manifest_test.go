package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp manifest
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "feat.yaml")

	content := `
tree:
  name: my-project
  files:
    - root.go
  children:
    auth:
      files:
        - auth/interface.go
      tests:
        - auth/interface_test.go
      children:
        login:
          files:
            - auth/login/handler.go
            - auth/login/types.go
          tests:
            - auth/login/handler_test.go
`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test manifest: %v", err)
	}

	m, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if m.Tree.Name != "my-project" {
		t.Errorf("Expected name 'my-project', got %q", m.Tree.Name)
	}

	if len(m.Tree.Files) != 1 || m.Tree.Files[0] != "root.go" {
		t.Errorf("Expected root files ['root.go'], got %v", m.Tree.Files)
	}

	if len(m.Tree.Children) != 1 {
		t.Errorf("Expected 1 root child, got %d", len(m.Tree.Children))
	}

	auth, ok := m.Tree.Children["auth"]
	if !ok {
		t.Fatal("Expected 'auth' node")
	}

	if len(auth.Files) != 1 || auth.Files[0] != "auth/interface.go" {
		t.Errorf("Expected auth files ['auth/interface.go'], got %v", auth.Files)
	}

	if len(auth.Tests) != 1 || auth.Tests[0] != "auth/interface_test.go" {
		t.Errorf("Expected auth tests ['auth/interface_test.go'], got %v", auth.Tests)
	}

	if len(auth.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(auth.Children))
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/feat.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "feat.yaml")

	m := &Manifest{
		Tree: Tree{
			Name: "test-project",
			Children: map[string]Node{
				"auth": {
					Files: []string{"auth/interface.go"},
					Tests: []string{"auth/interface_test.go"},
					Children: map[string]Node{
						"login": {
							Files: []string{"auth/login/handler.go"},
							Tests: []string{"auth/login/handler_test.go"},
						},
					},
				},
			},
		},
	}

	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("Manifest file not created: %v", err)
	}

	// Load it back and verify
	m2, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if m2.Tree.Name != "test-project" {
		t.Errorf("Expected name 'test-project', got %q", m2.Tree.Name)
	}

	if len(m2.Tree.Children) != 1 {
		t.Errorf("Expected 1 child after reload, got %d", len(m2.Tree.Children))
	}

	auth := m2.Tree.Children["auth"]
	if len(auth.Tests) != 1 {
		t.Errorf("Expected 1 test after reload, got %d", len(auth.Tests))
	}
}

func TestGetNode(t *testing.T) {
	m := &Manifest{
		Tree: Tree{
			Name: "test",
			Children: map[string]Node{
				"auth": {
					Files: []string{"auth/interface.go"},
					Children: map[string]Node{
						"login": {
							Files: []string{"auth/login/handler.go"},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		path            string
		expectError     bool
		expectFeature   bool
		expectAncestors int
	}{
		{"auth/login", false, true, 1}, // 1 ancestor file: auth/interface.go
		{"auth", false, false, 0},      // 0 ancestors (root level, boundary)
		{"nonexistent", true, false, 0},
		{"auth/nonexistent", true, false, 0},
		{"", true, false, 0},
	}

	for _, tt := range tests {
		n, ancestors, err := m.GetNode(tt.path)
		if tt.expectError {
			if err == nil {
				t.Errorf("GetNode(%q): expected error, got nil", tt.path)
			}
			continue
		}
		if err != nil {
			t.Errorf("GetNode(%q): unexpected error: %v", tt.path, err)
			continue
		}
		if n == nil {
			t.Errorf("GetNode(%q): expected node, got nil", tt.path)
			continue
		}
		if n.IsFeature() != tt.expectFeature {
			t.Errorf("GetNode(%q): IsFeature() = %v, want %v", tt.path, n.IsFeature(), tt.expectFeature)
		}
		if len(ancestors) != tt.expectAncestors {
			t.Errorf("GetNode(%q): len(ancestors) = %d, want %d", tt.path, len(ancestors), tt.expectAncestors)
		}
	}
}

func TestNodeIsFeature(t *testing.T) {
	tests := []struct {
		n    Node
		want bool
	}{
		{Node{Files: []string{"a.go"}}, true},
		{Node{Tests: []string{"a_test.go"}}, true},
		{Node{Files: []string{"a.go"}, Tests: []string{"a_test.go"}}, true},
		{Node{Files: []string{}}, false}, // empty
		{Node{Children: map[string]Node{}}, false},
		{Node{Files: []string{"a.go"}, Children: map[string]Node{"child": {}}}, false},
	}

	for _, tt := range tests {
		got := tt.n.IsFeature()
		if got != tt.want {
			t.Errorf("IsFeature() = %v, want %v for %+v", got, tt.want, tt.n)
		}
	}
}

func TestNodeIsBoundary(t *testing.T) {
	tests := []struct {
		n    Node
		want bool
	}{
		// Has children → boundary
		{Node{Children: map[string]Node{"child": {}}}, true},
		// Has files and children → boundary
		{Node{Files: []string{"a.go"}, Children: map[string]Node{"child": {}}}, true},
		// Has files only → feature
		{Node{Files: []string{"a.go"}}, false},
		// Has tests only → feature
		{Node{Tests: []string{"a_test.go"}}, false},
		// Has both files and tests → feature
		{Node{Files: []string{"a.go"}, Tests: []string{"a_test.go"}}, false},
		// Empty boundary → boundary (placeholder)
		{Node{}, true},
		// Empty children map → boundary
		{Node{Children: map[string]Node{}}, true},
	}

	for _, tt := range tests {
		got := tt.n.IsBoundary()
		if got != tt.want {
			t.Errorf("IsBoundary() = %v, want %v for %+v", got, tt.want, tt.n)
		}
	}
}

func TestNodeAllFiles(t *testing.T) {
	n := Node{
		Files: []string{"a.go", "b.go"},
		Tests: []string{"a_test.go"},
	}

	all := n.AllFiles()
	if len(all) != 3 {
		t.Errorf("Expected 3 files, got %d: %v", len(all), all)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		m      Manifest
		issues int
	}{
		{
			name:   "empty manifest",
			m:      Manifest{Tree: Tree{Name: "", Children: map[string]Node{}}},
			issues: 2, // no name, no children
		},
		{
			name: "missing name",
			m: Manifest{
				Tree: Tree{
					Name:     "",
					Children: map[string]Node{"auth": {Files: []string{"auth.go"}}},
				},
			},
			issues: 1,
		},
		{
			name: "valid tree",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {
							Files: []string{"auth/interface.go"},
							Children: map[string]Node{
								"login": {Files: []string{"auth/login.go"}}},
						},
					},
				},
			},
			issues: 0,
		},
		{
			name: "valid feature with tests",
			m: Manifest{
				Tree: Tree{
					Name:     "my-project",
					Children: map[string]Node{"auth": {Files: []string{"auth.go"}, Tests: []string{"auth_test.go"}}},
				},
			},
			issues: 0,
		},
		{
			name: "valid empty boundary",
			m: Manifest{
				Tree: Tree{
					Name:     "my-project",
					Children: map[string]Node{"payments": {}},
				},
			},
			issues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.m.Validate()
			if len(issues) != tt.issues {
				t.Errorf("Validate() returned %d issues, want %d: %v", len(issues), tt.issues, issues)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "feat.yaml")

	if err := Init(manifestPath, "my-project"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	m, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if m.Tree.Name != "my-project" {
		t.Errorf("Expected name 'my-project', got %q", m.Tree.Name)
	}

	// Note: empty maps become nil when serialized through YAML, which is fine
}

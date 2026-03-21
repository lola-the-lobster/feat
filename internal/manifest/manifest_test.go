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
  features:
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

	if len(m.Tree.Features) != 1 {
		t.Errorf("Expected 1 root feature, got %d", len(m.Tree.Features))
	}

	auth, ok := m.Tree.Features["auth"]
	if !ok {
		t.Fatal("Expected 'auth' feature")
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
			Features: map[string]Feature{
				"auth": {
					Files: []string{"auth/interface.go"},
					Tests: []string{"auth/interface_test.go"},
					Children: map[string]Feature{
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

	if len(m2.Tree.Features) != 1 {
		t.Errorf("Expected 1 feature after reload, got %d", len(m2.Tree.Features))
	}

	auth := m2.Tree.Features["auth"]
	if len(auth.Tests) != 1 {
		t.Errorf("Expected 1 test after reload, got %d", len(auth.Tests))
	}
}

func TestGetFeature(t *testing.T) {
	m := &Manifest{
		Tree: Tree{
			Name: "test",
			Features: map[string]Feature{
				"auth": {
					Files: []string{"auth/interface.go"},
					Children: map[string]Feature{
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
		expectLeaf      bool
		expectAncestors int
	}{
		{"auth/login", false, true, 1}, // 1 ancestor file: auth/interface.go
		{"auth", false, false, 0},      // 0 ancestors (root level, not leaf)
		{"nonexistent", true, false, 0},
		{"auth/nonexistent", true, false, 0},
		{"", true, false, 0},
	}

	for _, tt := range tests {
		f, ancestors, err := m.GetFeature(tt.path)
		if tt.expectError {
			if err == nil {
				t.Errorf("GetFeature(%q): expected error, got nil", tt.path)
			}
			continue
		}
		if err != nil {
			t.Errorf("GetFeature(%q): unexpected error: %v", tt.path, err)
			continue
		}
		if f == nil {
			t.Errorf("GetFeature(%q): expected feature, got nil", tt.path)
			continue
		}
		if f.IsLeaf() != tt.expectLeaf {
			t.Errorf("GetFeature(%q): IsLeaf() = %v, want %v", tt.path, f.IsLeaf(), tt.expectLeaf)
		}
		if len(ancestors) != tt.expectAncestors {
			t.Errorf("GetFeature(%q): len(ancestors) = %d, want %d", tt.path, len(ancestors), tt.expectAncestors)
		}
	}
}

func TestFeatureIsLeaf(t *testing.T) {
	tests := []struct {
		f    Feature
		want bool
	}{
		{Feature{Files: []string{"a.go"}}, true},
		{Feature{Tests: []string{"a_test.go"}}, true},
		{Feature{Files: []string{"a.go"}, Tests: []string{"a_test.go"}}, true},
		{Feature{Files: []string{}}, false},
		{Feature{Children: map[string]Feature{}}, false},
		{Feature{Files: []string{"a.go"}, Children: map[string]Feature{"child": {}}}, false},
	}

	for _, tt := range tests {
		got := tt.f.IsLeaf()
		if got != tt.want {
			t.Errorf("IsLeaf() = %v, want %v for %+v", got, tt.want, tt.f)
		}
	}
}

func TestFeatureIsIntermediate(t *testing.T) {
	tests := []struct {
		f    Feature
		want bool
	}{
		// Has children → intermediate
		{Feature{Children: map[string]Feature{"child": {}}}, true},
		// Has files and children → intermediate
		{Feature{Files: []string{"a.go"}, Children: map[string]Feature{"child": {}}}, true},
		// Has files only → leaf
		{Feature{Files: []string{"a.go"}}, false},
		// Has tests only → leaf
		{Feature{Tests: []string{"a_test.go"}}, false},
		// Has both files and tests → leaf
		{Feature{Files: []string{"a.go"}, Tests: []string{"a_test.go"}}, false},
		// Empty subsystem → intermediate (boundary placeholder)
		{Feature{}, true},
		// Empty children map → intermediate
		{Feature{Children: map[string]Feature{}}, true},
	}

	for _, tt := range tests {
		got := tt.f.IsIntermediate()
		if got != tt.want {
			t.Errorf("IsIntermediate() = %v, want %v for %+v", got, tt.want, tt.f)
		}
	}
}

func TestFeatureAllFiles(t *testing.T) {
	f := Feature{
		Files: []string{"a.go", "b.go"},
		Tests: []string{"a_test.go"},
	}

	all := f.AllFiles()
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
			m:      Manifest{Tree: Tree{Name: "", Features: map[string]Feature{}}},
			issues: 2, // no name, no features
		},
		{
			name: "missing name",
			m: Manifest{
				Tree: Tree{
					Name:     "",
					Features: map[string]Feature{"auth": {Files: []string{"auth.go"}}},
				},
			},
			issues: 1,
		},
		{
			name: "valid feature tree",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Features: map[string]Feature{
						"auth": {
							Files: []string{"auth/interface.go"},
							Children: map[string]Feature{
								"login": {Files: []string{"auth/login.go"}}},
						},
					},
				},
			},
			issues: 0,
		},
		{
			name: "valid leaf with tests",
			m: Manifest{
				Tree: Tree{
					Name:     "my-project",
					Features: map[string]Feature{"auth": {Files: []string{"auth.go"}, Tests: []string{"auth_test.go"}}},
				},
			},
			issues: 0,
		},
		{
			name: "valid empty subsystem",
			m: Manifest{
				Tree: Tree{
					Name:     "my-project",
					Features: map[string]Feature{"payments": {}},
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

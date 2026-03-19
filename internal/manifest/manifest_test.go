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
auth:
  interface: auth/interface.go
  login:
    files:
      - auth/login/handler.go
      - auth/login/types.go
`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test manifest: %v", err)
	}

	m, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(m.Features) != 1 {
		t.Errorf("Expected 1 root feature, got %d", len(m.Features))
	}

	auth, ok := m.Features["auth"]
	if !ok {
		t.Fatal("Expected 'auth' feature")
	}

	if auth.Interface != "auth/interface.go" {
		t.Errorf("Expected interface 'auth/interface.go', got %s", auth.Interface)
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
		Features: map[string]Feature{
			"auth": {
				Interface: "auth/interface.go",
				Children: map[string]Feature{
					"login": {
						Files: []string{"auth/login/handler.go"},
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

	if len(m2.Features) != 1 {
		t.Errorf("Expected 1 feature after reload, got %d", len(m2.Features))
	}
}

func TestGetFeature(t *testing.T) {
	m := &Manifest{
		Features: map[string]Feature{
			"auth": {
				Interface: "auth/interface.go",
				Children: map[string]Feature{
					"login": {
						Files: []string{"auth/login/handler.go"},
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
		{"auth/login", false, true, 1}, // 1 ancestor: auth/interface.go
		{"auth", false, false, 0},    // 0 ancestors (root level)
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
		{Feature{Files: []string{}}, false},
		{Feature{Interface: "iface.go"}, false},
		{Feature{Children: map[string]Feature{}}, false},
	}

	for _, tt := range tests {
		got := tt.f.IsLeaf()
		if got != tt.want {
			t.Errorf("IsLeaf() = %v, want %v", got, tt.want)
		}
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
			m:      Manifest{Features: map[string]Feature{}},
			issues: 1,
		},
		{
			name: "valid feature",
			m: Manifest{
				Features: map[string]Feature{
					"auth": {
						Interface: "auth/interface.go",
						Children: map[string]Feature{
							"login": {Files: []string{"auth/login.go"}},
						},
					},
				},
			},
			issues: 0,
		},
		{
			name: "orphaned intermediate",
			m: Manifest{
				Features: map[string]Feature{
					"auth": {Children: map[string]Feature{}},
				},
			},
			issues: 1,
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

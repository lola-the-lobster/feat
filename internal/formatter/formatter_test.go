package formatter

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/lola-the-lobster/feat/internal/loader"
	"github.com/lola-the-lobster/feat/internal/manifest"
	"github.com/lola-the-lobster/feat/internal/state"
)

func TestFormatListJSON(t *testing.T) {
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test-project",
			Children: map[string]manifest.Node{
				"auth": {
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{"auth/login.go"},
							Tests: []string{"auth/login_test.go"},
						},
					},
					Files: []string{"auth/middleware.go"},
				},
				"api": {
					Files: []string{"api/routes.go"},
				},
			},
		},
	}

	data, err := FormatListJSON(m)
	if err != nil {
		t.Fatalf("FormatListJSON failed: %v", err)
	}

	var output ListJSON
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.Project != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", output.Project)
	}

	if len(output.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(output.Features))
	}
}

func TestFormatStatusJSON(t *testing.T) {
	s := &state.State{
		FeaturePath:  "auth/login",
		ManifestPath: "/path/to/feat.yaml",
		Timestamp:    time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
	}

	result := &loader.Result{
		FeaturePath:   "auth/login",
		Files:         []string{"/path/to/auth/login.go"},
		Tests:         []string{"/path/to/auth/login_test.go"},
		AncestorFiles: []string{"/path/to/auth/middleware.go"},
		MissingFiles:  []string{},
	}

	data, err := FormatStatusJSON(s, result)
	if err != nil {
		t.Fatalf("FormatStatusJSON failed: %v", err)
	}

	var output StatusJSON
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.CurrentFeature != "auth/login" {
		t.Errorf("Expected feature 'auth/login', got '%s'", output.CurrentFeature)
	}

	if len(output.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(output.Files))
	}

	if len(output.Ancestors) != 1 {
		t.Errorf("Expected 1 ancestor, got %d", len(output.Ancestors))
	}
}

func TestFormatStatusJSON_NilState(t *testing.T) {
	data, err := FormatStatusJSON(nil, nil)
	if err != nil {
		t.Fatalf("FormatStatusJSON failed: %v", err)
	}

	var output StatusJSON
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.CurrentFeature != "" {
		t.Errorf("Expected empty feature, got '%s'", output.CurrentFeature)
	}
}

func TestFormatErrorJSON(t *testing.T) {
	err := errors.New("feature not found: auth/bad")
	data := FormatErrorJSON(err, 4)

	var output ErrorJSON
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.Code != 4 {
		t.Errorf("Expected code 4, got %d", output.Code)
	}

	if output.Error == "" {
		t.Error("Expected non-empty error message")
	}
}

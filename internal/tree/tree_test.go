package tree

import (
	"strings"
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestPrint(t *testing.T) {
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "my-project",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go"},
					Tests: []string{"auth/interface_test.go"},
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{"auth/login.go"},
							Tests: []string{"auth/login_test.go"},
						},
						"logout": {
							Files: []string{"auth/logout.go"},
						},
					},
				},
				"payments": {
					Files: []string{"payments.go"},
					Tests: []string{"payments_test.go"},
				},
			},
		},
	}

	printer := NewPrinter()
	output := printer.Print(m)

	// Check that output contains expected nodes
	if !strings.Contains(output, "auth/") {
		t.Error("Expected output to contain 'auth/'")
	}
	if !strings.Contains(output, "login") {
		t.Error("Expected output to contain 'login'")
	}
	if !strings.Contains(output, "logout") {
		t.Error("Expected output to contain 'logout'")
	}
	if !strings.Contains(output, "payments") {
		t.Error("Expected output to contain 'payments'")
	}

	// Check ordering (should be alphabetical: auth, payments)
	authIdx := strings.Index(output, "auth")
	paymentsIdx := strings.Index(output, "payments")
	if authIdx == -1 || paymentsIdx == -1 {
		t.Error("Expected to find both auth and payments")
	} else if authIdx > paymentsIdx {
		t.Error("Expected auth to come before payments (alphabetical)")
	}
}

func TestPrintEmpty(t *testing.T) {
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     "my-project",
			Children: map[string]manifest.Node{},
		},
	}

	printer := NewPrinter()
	output := printer.Print(m)

	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestListPaths(t *testing.T) {
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "my-project",
			Children: map[string]manifest.Node{
				"auth": {
					Files: []string{"auth/interface.go"},
					Children: map[string]manifest.Node{
						"login": {Files: []string{"auth/login.go"}},
					},
				},
				"payments": {
					Files: []string{"payments.go"},
				},
			},
		},
	}

	paths := ListPaths(m)

	// Should include: auth/ (boundary with files), auth/login (feature), payments (feature)
	if len(paths) != 3 {
		t.Errorf("Expected 3 paths, got %d: %v", len(paths), paths)
	}

	// Check that paths are sorted
	for i := 1; i < len(paths); i++ {
		if paths[i] < paths[i-1] {
			t.Errorf("Paths not sorted: %v", paths)
		}
	}

	// Check specific paths exist
	hasAuth := false
	hasAuthLogin := false
	hasPayments := false
	for _, p := range paths {
		if p == "auth/" {
			hasAuth = true
		}
		if p == "auth/login" {
			hasAuthLogin = true
		}
		if p == "payments" {
			hasPayments = true
		}
	}
	if !hasAuth {
		t.Error("Expected paths to contain 'auth/'")
	}
	if !hasAuthLogin {
		t.Error("Expected paths to contain 'auth/login'")
	}
	if !hasPayments {
		t.Error("Expected paths to contain 'payments'")
	}
}

func TestListPathsEmpty(t *testing.T) {
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     "my-project",
			Children: map[string]manifest.Node{},
		},
	}

	paths := ListPaths(m)

	if len(paths) != 0 {
		t.Errorf("Expected 0 paths, got %d", len(paths))
	}
}

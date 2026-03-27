package manifest

import (
	"testing"
)

func TestValidateCircular(t *testing.T) {
	tests := []struct {
		name    string
		m       Manifest
		wantErr bool
		cycleAt string
	}{
		{
			name: "no circular reference",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {
							Files: []string{"auth/interface.go"},
							Children: map[string]Node{
								"login": {
									Files: []string{"auth/login/handler.go"},
									Children: map[string]Node{
										"password-reset": {
											Files: []string{"auth/login/password-reset.go"},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "circular reference at root level",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {
							Children: map[string]Node{
								"login": {
									Children: map[string]Node{
										"auth": { // circular - "auth" is ancestor
											Files: []string{"auth.go"},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			cycleAt: "auth",
		},
		{
			name: "circular reference nested deeper",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"api": {
							Children: map[string]Node{
								"v1": {
									Children: map[string]Node{
										"users": {
											Children: map[string]Node{
												"api": { // circular - "api" is ancestor
													Files: []string{"api.go"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			cycleAt: "api",
		},
		{
			name: "sibling nodes with same name - not circular",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {
							Children: map[string]Node{
								"login": {
									Files: []string{"auth/login.go"},
								},
							},
						},
						"payments": {
							Children: map[string]Node{
								"login": { // same name as auth/login, but different branch
									Files: []string{"payments/login.go"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "self-reference at root",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {
							Children: map[string]Node{
								"auth": { // immediate self-reference
									Files: []string{"auth.go"},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			cycleAt: "auth",
		},
		{
			name: "deep circular chain",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"a": {
							Children: map[string]Node{
								"b": {
									Children: map[string]Node{
										"c": {
											Children: map[string]Node{
												"d": {
													Children: map[string]Node{
														"a": { // cycles back to "a"
															Files: []string{"a.go"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			cycleAt: "a",
		},
		{
			name: "empty children - no circular",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple branches, one circular",
			m: Manifest{
				Tree: Tree{
					Name: "my-project",
					Children: map[string]Node{
						"auth": {
							Children: map[string]Node{
								"login": {
									Children: map[string]Node{
										"auth": { // circular
											Files: []string{"auth.go"},
										},
									},
								},
							},
						},
						"users": {
							Files: []string{"users.go"},
						},
					},
				},
			},
			wantErr: true,
			cycleAt: "auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.ValidateCircular()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCircular() expected error, got nil")
				} else {
					// Check that the error message contains the cycle point
					circularErr, ok := err.(*CircularError)
					if !ok {
						t.Errorf("expected *CircularError, got %T", err)
					} else if circularErr.Node != tt.cycleAt {
						t.Errorf("expected cycle at %q, got node %q", tt.cycleAt, circularErr.Node)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCircular() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCircularError(t *testing.T) {
	err := &CircularError{
		Path:  "auth/login",
		Node:  "auth",
		Cycle: "auth",
	}
	got := err.Error()
	want := `circular reference detected at "auth/login": auth references ancestor auth`
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestBuildPath(t *testing.T) {
	tests := []struct {
		ancestors []string
		want      string
	}{
		{[]string{}, ""},
		{[]string{"auth"}, "auth"},
		{[]string{"auth", "login"}, "auth/login"},
		{[]string{"a", "b", "c", "d"}, "a/b/c/d"},
	}

	for _, tt := range tests {
		got := buildPath(tt.ancestors)
		if got != tt.want {
			t.Errorf("buildPath(%v) = %q, want %q", tt.ancestors, got, tt.want)
		}
	}
}
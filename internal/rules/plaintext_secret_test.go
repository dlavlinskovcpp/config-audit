package rules

import "testing"

func TestPlaintextSecretRuleFlagsPlaintextSecrets(t *testing.T) {
	tests := []struct {
		name     string
		root     map[string]any
		wantPath string
	}{
		{
			name: "database password",
			root: map[string]any{
				"database": map[string]any{
					"password": "supersecret",
				},
			},
			wantPath: "database.password",
		},
		{
			name: "database pwd shortcut",
			root: map[string]any{
				"database": map[string]any{
					"pwd": "supersecret",
				},
			},
			wantPath: "database.pwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(PlaintextSecretRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Path != tt.wantPath {
				t.Fatalf("expected %q, got %q", tt.wantPath, problems[0].Path)
			}
		})
	}
}

func TestPlaintextSecretRuleIgnoresReferencedOrEmptyValues(t *testing.T) {
	tests := []struct {
		name string
		root map[string]any
	}{
		{
			name: "environment variable reference",
			root: map[string]any{"password": "${PASSWORD}"},
		},
		{
			name: "secret manager reference",
			root: map[string]any{"api_key": "vault:kv/app/api_key"},
		},
		{
			name: "blank password",
			root: map[string]any{"password": "   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(PlaintextSecretRule{}, tt.root)
			if len(problems) != 0 {
				t.Fatalf("expected 0 problems, got %d", len(problems))
			}
		})
	}
}

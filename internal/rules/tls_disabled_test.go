package rules

import "testing"

func TestTLSDisabledRuleFlagsDisabledTLSAndVerification(t *testing.T) {
	tests := []struct {
		name     string
		root     map[string]any
		wantPath string
	}{
		{
			name: "tls.enabled false",
			root: map[string]any{
				"tls": map[string]any{
					"enabled": false,
				},
			},
			wantPath: "tls.enabled",
		},
		{
			name: "ssl.enabled false",
			root: map[string]any{
				"ssl": map[string]any{
					"enabled": false,
				},
			},
			wantPath: "ssl.enabled",
		},
		{
			name: "insecure skip verify",
			root: map[string]any{
				"client": map[string]any{
					"insecure_skip_verify": true,
				},
			},
			wantPath: "client.insecure_skip_verify",
		},
		{
			name: "verify_tls false",
			root: map[string]any{
				"client": map[string]any{
					"verify_tls": false,
				},
			},
			wantPath: "client.verify_tls",
		},
		{
			name: "tls_verify false",
			root: map[string]any{
				"client": map[string]any{
					"tls_verify": false,
				},
			},
			wantPath: "client.tls_verify",
		},
		{
			name: "ssl_verify false",
			root: map[string]any{
				"client": map[string]any{
					"ssl_verify": false,
				},
			},
			wantPath: "client.ssl_verify",
		},
		{
			name: "verify_certificate false",
			root: map[string]any{
				"client": map[string]any{
					"verify_certificate": false,
				},
			},
			wantPath: "client.verify_certificate",
		},
		{
			name: "skip_tls_verify true",
			root: map[string]any{
				"client": map[string]any{
					"skip_tls_verify": true,
				},
			},
			wantPath: "client.skip_tls_verify",
		},
		{
			name: "disable_tls true",
			root: map[string]any{
				"client": map[string]any{
					"disable_tls": true,
				},
			},
			wantPath: "client.disable_tls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(TLSDisabledRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Path != tt.wantPath {
				t.Fatalf("expected %q, got %q", tt.wantPath, problems[0].Path)
			}
		})
	}
}

func TestTLSDisabledRuleIgnoresSafeTLSSettings(t *testing.T) {
	tests := []struct {
		name string
		root map[string]any
	}{
		{
			name: "tls.enabled true",
			root: map[string]any{
				"tls": map[string]any{
					"enabled": true,
				},
			},
		},
		{
			name: "disable_tls false",
			root: map[string]any{
				"client": map[string]any{
					"disable_tls": false,
				},
			},
		},
		{
			name: "verify_tls true",
			root: map[string]any{
				"client": map[string]any{
					"verify_tls": true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(TLSDisabledRule{}, tt.root)
			if len(problems) != 0 {
				t.Fatalf("expected 0 problems, got %d", len(problems))
			}
		})
	}
}

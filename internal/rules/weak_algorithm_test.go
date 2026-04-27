package rules

import "testing"

func TestWeakAlgorithmRuleFlagsWeakAlgorithmsInAlgorithmFields(t *testing.T) {
	tests := []struct {
		name     string
		root     map[string]any
		wantPath string
	}{
		{
			name:     "digest-algorithm",
			root:     map[string]any{"storage": map[string]any{"digest-algorithm": "MD5"}},
			wantPath: "storage.digest-algorithm",
		},
		{
			name:     "algorithm",
			root:     map[string]any{"algorithm": "sha1"},
			wantPath: "algorithm",
		},
		{
			name:     "digest",
			root:     map[string]any{"digest": "md5"},
			wantPath: "digest",
		},
		{
			name:     "digest_algorithm",
			root:     map[string]any{"digest_algorithm": "3DES"},
			wantPath: "digest_algorithm",
		},
		{
			name:     "hash",
			root:     map[string]any{"hash": "rc4"},
			wantPath: "hash",
		},
		{
			name:     "cipher",
			root:     map[string]any{"cipher": "blowfish"},
			wantPath: "cipher",
		},
		{
			name:     "encryption",
			root:     map[string]any{"encryption": "none"},
			wantPath: "encryption",
		},
		{
			name:     "password_hash",
			root:     map[string]any{"password_hash": "MD5"},
			wantPath: "password_hash",
		},
		{
			name:     "hash_algorithm",
			root:     map[string]any{"hash_algorithm": "SHA1"},
			wantPath: "hash_algorithm",
		},
		{
			name:     "crypto_method",
			root:     map[string]any{"crypto_method": "RC4"},
			wantPath: "crypto_method",
		},
		{
			name:     "signature_hash_algorithm",
			root:     map[string]any{"signature_hash_algorithm": "SHA1"},
			wantPath: "signature_hash_algorithm",
		},
		{
			name:     "db_password_hash",
			root:     map[string]any{"db_password_hash": "MD5"},
			wantPath: "db_password_hash",
		},
		{
			name:     "cipher_suite_name",
			root:     map[string]any{"cipher_suite_name": "RC4"},
			wantPath: "cipher_suite_name",
		},
		{
			name:     "storage_crypto_profile",
			root:     map[string]any{"storage_crypto_profile": "blowfish"},
			wantPath: "storage_crypto_profile",
		},
		{
			name:     "plaintext encryption",
			root:     map[string]any{"encryption": "plaintext"},
			wantPath: "encryption",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(WeakAlgorithmRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Path != tt.wantPath {
				t.Fatalf("expected %q, got %q", tt.wantPath, problems[0].Path)
			}
		})
	}
}

func TestWeakAlgorithmRuleUsesPathContextForNestedSettings(t *testing.T) {
	tests := []struct {
		name     string
		root     map[string]any
		wantPath string
	}{
		{
			name: "crypto method",
			root: map[string]any{
				"crypto": map[string]any{
					"method": "rc4",
				},
			},
			wantPath: "crypto.method",
		},
		{
			name: "digest type",
			root: map[string]any{
				"storage": map[string]any{
					"digest": map[string]any{
						"type": "md5",
					},
				},
			},
			wantPath: "storage.digest.type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(WeakAlgorithmRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Path != tt.wantPath {
				t.Fatalf("expected %q, got %q", tt.wantPath, problems[0].Path)
			}
		})
	}
}

func TestWeakAlgorithmRuleIgnoresUnrelatedOrStrongValues(t *testing.T) {
	tests := []struct {
		name string
		root map[string]any
	}{
		{
			name: "unrelated provider field",
			root: map[string]any{"provider": "md5"},
		},
		{
			name: "generic method field",
			root: map[string]any{"method": "md5"},
		},
		{
			name: "strong hash algorithm",
			root: map[string]any{"hash": "SHA-256"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(WeakAlgorithmRule{}, tt.root)
			if len(problems) != 0 {
				t.Fatalf("expected 0 problems, got %d", len(problems))
			}
		})
	}
}

package rules

import (
	"strings"
	"testing"

	"configaudit/internal/audit"
)

func TestWildcardBindRuleFlagsWildcardBindVariants(t *testing.T) {
	tests := []struct {
		name     string
		root     map[string]any
		wantPath string
	}{
		{
			name: "host",
			root: map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
				},
			},
			wantPath: "server.host",
		},
		{
			name: "listen_address",
			root: map[string]any{
				"server": map[string]any{
					"listen_address": "0.0.0.0",
				},
			},
			wantPath: "server.listen_address",
		},
		{
			name: "bind_address",
			root: map[string]any{
				"server": map[string]any{
					"bind_address": "0.0.0.0",
				},
			},
			wantPath: "server.bind_address",
		},
		{
			name: "bind_addr",
			root: map[string]any{
				"server": map[string]any{
					"bind_addr": "0.0.0.0",
				},
			},
			wantPath: "server.bind_addr",
		},
		{
			name: "listen_addr",
			root: map[string]any{
				"server": map[string]any{
					"listen_addr": "0.0.0.0",
				},
			},
			wantPath: "server.listen_addr",
		},
		{
			name: "listen with port",
			root: map[string]any{
				"server": map[string]any{
					"listen": "0.0.0.0:8080",
				},
			},
			wantPath: "server.listen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(WildcardBindRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Severity != audit.SeverityMedium {
				t.Fatalf("expected MEDIUM severity, got %s", problems[0].Severity)
			}
			if problems[0].Path != tt.wantPath {
				t.Fatalf("expected %q, got %q", tt.wantPath, problems[0].Path)
			}
		})
	}
}

func TestWildcardBindRuleReportsRestrictionContext(t *testing.T) {
	tests := []struct {
		name        string
		root        map[string]any
		wantMessage string
		wantAbsent  string
	}{
		{
			name: "mentions missing restrictions",
			root: map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
				},
			},
			wantMessage: "no obvious access restriction",
		},
		{
			name: "suppresses hint when allowlist exists",
			root: map[string]any{
				"server": map[string]any{
					"host":      "0.0.0.0",
					"allowlist": []any{"10.0.0.0/8"},
				},
			},
			wantAbsent: "no obvious access restriction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(WildcardBindRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if tt.wantMessage != "" && !strings.Contains(problems[0].Message, tt.wantMessage) {
				t.Fatalf("expected message to contain %q, got %q", tt.wantMessage, problems[0].Message)
			}
			if tt.wantAbsent != "" && strings.Contains(problems[0].Message, tt.wantAbsent) {
				t.Fatalf("did not expect message to contain %q, got %q", tt.wantAbsent, problems[0].Message)
			}
		})
	}
}

func TestWildcardBindRuleIgnoresPrivateBinding(t *testing.T) {
	problems := runRule(WildcardBindRule{}, map[string]any{
		"server": map[string]any{
			"host": "127.0.0.1",
		},
	})

	if len(problems) != 0 {
		t.Fatalf("expected 0 problems, got %d", len(problems))
	}
}

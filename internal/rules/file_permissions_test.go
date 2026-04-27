package rules

import (
	"io/fs"
	"testing"

	"configaudit/internal/audit"
)

func TestFilePermissionsRuleFlagsWritableFiles(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		root         map[string]any
		mode         fs.FileMode
		wantSeverity audit.Severity
	}{
		{
			name:         "world-writable config",
			file:         "config.yaml",
			root:         map[string]any{"service": "api"},
			mode:         0o666,
			wantSeverity: audit.SeverityHigh,
		},
		{
			name:         "group-writable config",
			file:         "group.yaml",
			root:         map[string]any{"service": "api"},
			mode:         0o620,
			wantSeverity: audit.SeverityMedium,
		},
		{
			name:         "world-writable empty config",
			file:         "empty.json",
			root:         map[string]any{},
			mode:         0o666,
			wantSeverity: audit.SeverityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := scanPermissionProblems(tt.file, tt.root, tt.mode)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Severity != tt.wantSeverity {
				t.Fatalf("expected %s severity, got %s", tt.wantSeverity, problems[0].Severity)
			}
			if problems[0].File != tt.file {
				t.Fatalf("expected file %q, got %q", tt.file, problems[0].File)
			}
		})
	}
}

func TestFilePermissionsRuleIgnoresRestrictedFile(t *testing.T) {
	problems := scanPermissionProblems("config.yaml", map[string]any{"service": "api"}, 0o600)
	if len(problems) != 0 {
		t.Fatalf("expected 0 problems, got %d", len(problems))
	}
}

func scanPermissionProblems(file string, root map[string]any, mode fs.FileMode) []audit.Problem {
	engine := audit.NewEngine(FilePermissionsRule{})
	return engine.Scan(audit.Context{
		File:        file,
		Root:        root,
		FileMode:    mode,
		HasFileMode: true,
	})
}

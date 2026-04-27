package rules

import (
	"testing"

	"configaudit/internal/audit"
)

func TestDebugLoggingRuleFlagsSupportedLoggingLevels(t *testing.T) {
	tests := []struct {
		name     string
		root     map[string]any
		wantPath string
	}{
		{
			name: "log.level debug",
			root: map[string]any{
				"log": map[string]any{
					"level": "debug",
				},
			},
			wantPath: "log.level",
		},
		{
			name: "logging.level trace",
			root: map[string]any{
				"logging": map[string]any{
					"level": "TRACE",
				},
			},
			wantPath: "logging.level",
		},
		{
			name: "root level debug",
			root: map[string]any{
				"level": "debug",
			},
			wantPath: "level",
		},
		{
			name: "logger.mode debug",
			root: map[string]any{
				"logger": map[string]any{
					"mode": "debug",
				},
			},
			wantPath: "logger.mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(DebugLoggingRule{}, tt.root)
			if len(problems) != 1 {
				t.Fatalf("expected 1 problem, got %d", len(problems))
			}
			if problems[0].Severity != audit.SeverityLow {
				t.Fatalf("expected LOW severity, got %s", problems[0].Severity)
			}
			if problems[0].Path != tt.wantPath {
				t.Fatalf("expected %q, got %q", tt.wantPath, problems[0].Path)
			}
		})
	}
}

func TestDebugLoggingRuleIgnoresNonLoggingLevels(t *testing.T) {
	tests := []struct {
		name string
		root map[string]any
	}{
		{
			name: "info level under log",
			root: map[string]any{
				"log": map[string]any{
					"level": "info",
				},
			},
		},
		{
			name: "debug level under cache",
			root: map[string]any{
				"cache": map[string]any{
					"level": "debug",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := runRule(DebugLoggingRule{}, tt.root)
			if len(problems) != 0 {
				t.Fatalf("expected 0 problems, got %d", len(problems))
			}
		})
	}
}

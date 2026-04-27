package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRejectsInvalidInvocation(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		stdin      string
		wantExit   int
		wantStderr string
	}{
		{
			name:       "requires file path or stdin",
			wantExit:   2,
			wantStderr: "configuration file path is required",
		},
		{
			name:       "rejects unsupported format",
			args:       []string{"--format", "toml", "--stdin"},
			stdin:      "{}",
			wantExit:   2,
			wantStderr: "invalid format",
		},
		{
			name:       "rejects recursive scan from stdin",
			args:       []string{"--stdin", "--recursive"},
			stdin:      "{}",
			wantExit:   2,
			wantStderr: "--recursive cannot be used together with --stdin",
		},
		{
			name:       "server mode requires transport address",
			args:       []string{"server"},
			wantExit:   2,
			wantStderr: "at least one of --http or --grpc must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := Run(tt.args, strings.NewReader(tt.stdin), &stdout, &stderr)
			if code != tt.wantExit {
				t.Fatalf("expected exit code %d, got %d", tt.wantExit, code)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("expected stderr to contain %q, got:\n%s", tt.wantStderr, stderr.String())
			}
		})
	}
}

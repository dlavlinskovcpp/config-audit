package parser

import "testing"

func TestParseAcceptsStructuredJSONAndYAML(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		filename string
		format   Format
		want     string
	}{
		{
			name:     "auto-detects JSON",
			data:     []byte(`{"log":{"level":"debug"}}`),
			filename: "config.json",
			format:   FormatAuto,
			want:     "debug",
		},
		{
			name:     "auto-detects YAML",
			data:     []byte("log:\n  level: info\n"),
			filename: "config.yaml",
			format:   FormatAuto,
			want:     "info",
		},
		{
			name:     "honors explicit format over file extension",
			data:     []byte(`{"log":{"level":"debug"}}`),
			filename: "config.yaml",
			format:   FormatJSON,
			want:     "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Parse(tt.data, tt.filename, tt.format)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			logConfig := root["log"].(map[string]any)
			if got := logConfig["level"]; got != tt.want {
				t.Fatalf("expected %q, got %v", tt.want, got)
			}
		})
	}
}

func TestParseRejectsInvalidConfigurations(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		filename string
		format   Format
	}{
		{
			name:     "truncated JSON",
			data:     []byte(`{"log":`),
			filename: "config.json",
			format:   FormatAuto,
		},
		{
			name:     "malformed YAML",
			data:     []byte("storage:\n  digest-algorithm: [\n"),
			filename: "config.yaml",
			format:   FormatAuto,
		},
		{
			name:     "top-level array",
			data:     []byte(`[{"log":{"level":"debug"}}]`),
			filename: "config.json",
			format:   FormatAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.data, tt.filename, tt.format); err == nil {
				t.Fatal("expected parse error")
			}
		})
	}
}

func TestParseFormatRecognizesSupportedValues(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
	}{
		{name: "auto", input: "auto", want: FormatAuto},
		{name: "empty means auto", input: "", want: FormatAuto},
		{name: "json", input: "json", want: FormatJSON},
		{name: "yaml", input: "yaml", want: FormatYAML},
		{name: "rejects unknown value", input: "toml", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFormat returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

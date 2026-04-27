package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatAuto Format = "auto"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

type candidate struct {
	name  string
	parse func([]byte) (map[string]any, error)
}

func ParseFormat(value string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return FormatAuto, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("invalid format %q, expected auto, json, or yaml", value)
	}
}

func Parse(data []byte, filename string, format Format) (map[string]any, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("configuration is empty")
	}

	parsers := parsersFor(filename, format)
	var parseErrors []string

	for _, parser := range parsers {
		root, err := parser.parse(data)
		if err == nil {
			return root, nil
		}
		parseErrors = append(parseErrors, fmt.Sprintf("%s: %v", parser.name, err))
	}

	return nil, fmt.Errorf("parse configuration: %s", strings.Join(parseErrors, "; "))
}

func parsersFor(filename string, format Format) []candidate {
	jsonParser := candidate{name: "json", parse: parseJSON}
	yamlParser := candidate{name: "yaml", parse: parseYAML}

	switch format {
	case FormatJSON:
		return []candidate{jsonParser}
	case FormatYAML:
		return []candidate{yamlParser}
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		return []candidate{jsonParser, yamlParser}
	case ".yaml", ".yml":
		return []candidate{yamlParser, jsonParser}
	default:
		return []candidate{jsonParser, yamlParser}
	}
}

func parseJSON(data []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var root any
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values are not allowed")
		}
		return nil, err
	}

	return normalizeRoot(root)
}

func parseYAML(data []byte) (map[string]any, error) {
	var root any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	return normalizeRoot(root)
}

func normalizeRoot(value any) (map[string]any, error) {
	normalized, err := normalize(value)
	if err != nil {
		return nil, err
	}

	root, ok := normalized.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("top-level value must be an object")
	}

	return root, nil
}

func normalize(value any) (any, error) {
	switch current := value.(type) {
	case map[string]any:
		normalized := make(map[string]any, len(current))
		for key, child := range current {
			converted, err := normalize(child)
			if err != nil {
				return nil, err
			}
			normalized[key] = converted
		}
		return normalized, nil

	case map[any]any:
		normalized := make(map[string]any, len(current))
		for key, child := range current {
			stringKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("map key %v is not a string", key)
			}
			converted, err := normalize(child)
			if err != nil {
				return nil, err
			}
			normalized[stringKey] = converted
		}
		return normalized, nil

	case []any:
		normalized := make([]any, len(current))
		for i, child := range current {
			converted, err := normalize(child)
			if err != nil {
				return nil, err
			}
			normalized[i] = converted
		}
		return normalized, nil

	default:
		return current, nil
	}
}

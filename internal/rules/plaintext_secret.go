package rules

import (
	"regexp"
	"strings"

	"configaudit/internal/audit"
)

type PlaintextSecretRule struct{}

var envVarNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (PlaintextSecretRule) ID() string {
	return "plaintext-secret"
}

func (PlaintextSecretRule) Check(_ audit.Context, node audit.Node) []audit.Problem {
	if !looksSensitiveKey(node.Key) {
		return nil
	}

	value, ok := node.Value.(string)
	if !ok {
		return nil
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" || isSecretReference(trimmed) {
		return nil
	}

	return []audit.Problem{{
		Severity:       audit.SeverityHigh,
		Path:           node.Path,
		Message:        "Plaintext secret detected in configuration.",
		Recommendation: "Do not store secrets in plaintext configuration files. Use environment variables, a secret manager, or encrypted storage.",
	}}
}

func looksSensitiveKey(key string) bool {
	normalized := normalizeIdentifier(key)
	if normalized == "" {
		return false
	}

	for _, candidate := range []string{
		"password",
		"passwd",
		"pwd",
		"secret",
		"token",
		"apikey",
		"accesskey",
		"privatekey",
	} {
		if strings.Contains(normalized, candidate) {
			return true
		}
	}

	return false
}

func isSecretReference(value string) bool {
	switch {
	case strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}"):
		return true
	case strings.HasPrefix(value, "$") && envVarNamePattern.MatchString(strings.TrimPrefix(value, "$")):
		return true
	}

	lower := strings.ToLower(value)
	for _, prefix := range []string{"env:", "vault:", "secret:", "file:"} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	return false
}

package rules

import (
	"fmt"
	"strings"

	"configaudit/internal/audit"
)

type WeakAlgorithmRule struct{}

func (WeakAlgorithmRule) ID() string {
	return "weak-algorithm"
}

func (WeakAlgorithmRule) Check(_ audit.Context, node audit.Node) []audit.Problem {
	if !looksLikeAlgorithmField(node) {
		return nil
	}

	value, ok := node.Value.(string)
	if !ok {
		return nil
	}

	trimmed := strings.TrimSpace(value)
	if !isWeakAlgorithm(trimmed) {
		return nil
	}

	return []audit.Problem{{
		Severity:       audit.SeverityHigh,
		Path:           node.Path,
		Message:        fmt.Sprintf("Weak algorithm %s detected.", strings.ToUpper(trimmed)),
		Recommendation: "Replace it with a modern secure alternative such as SHA-256, SHA-512, bcrypt, scrypt, Argon2, AES-GCM, depending on the use case.",
	}}
}

func looksLikeAlgorithmField(node audit.Node) bool {
	return containsAlgorithmMarker(node.Key) || containsAlgorithmMarker(node.Path)
}

func containsAlgorithmMarker(value string) bool {
	normalized := normalizeIdentifier(value)
	if normalized == "" {
		return false
	}

	for _, marker := range []string{
		"algo",
		"algorithm",
		"digest",
		"hash",
		"cipher",
		"encrypt",
		"crypto",
		"crypt",
		"checksum",
		"signature",
		"signing",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	return false
}

func isWeakAlgorithm(value string) bool {
	switch normalizeIdentifier(value) {
	case "md5", "sha1", "des", "3des", "rc4", "blowfish", "none", "plaintext":
		return true
	default:
		return false
	}
}

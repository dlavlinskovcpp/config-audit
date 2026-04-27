package rules

import (
	"strings"

	"configaudit/internal/audit"
)

type WildcardBindRule struct{}

func (WildcardBindRule) ID() string {
	return "wildcard-bind"
}

func (WildcardBindRule) Check(_ audit.Context, node audit.Node) []audit.Problem {
	if !looksLikeBindKey(node.Key) {
		return nil
	}

	value, ok := node.Value.(string)
	if !ok || !isWildcardBindValue(value) {
		return nil
	}

	message := "Service is bound to 0.0.0.0."
	if !containsRestrictionHint(node.Parent) {
		message = "Service is bound to 0.0.0.0 and no obvious access restriction was found in the same configuration block."
	}

	return []audit.Problem{{
		Severity:       audit.SeverityMedium,
		Path:           node.Path,
		Message:        message,
		Recommendation: "Do not expose services on all interfaces unless it is required. Restrict access using firewall rules, bind to a private interface, or configure an allowlist.",
	}}
}

func looksLikeBindKey(key string) bool {
	switch normalizeIdentifier(key) {
	case "host", "bind", "listen", "address", "addr", "listenaddress", "bindaddress", "bindaddr", "listenaddr":
		return true
	default:
		return false
	}
}

func isWildcardBindValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "0.0.0.0" || strings.HasPrefix(trimmed, "0.0.0.0:")
}

func containsRestrictionHint(value any) bool {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			switch normalizeIdentifier(key) {
			case "allowedips", "allowlist", "whitelist", "trustedproxies", "firewall", "auth":
				return true
			}
			if containsRestrictionHint(child) {
				return true
			}
		}
	case []any:
		for _, child := range current {
			if containsRestrictionHint(child) {
				return true
			}
		}
	}

	return false
}

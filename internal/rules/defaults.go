package rules

import "configaudit/internal/audit"

func Default(includePermissions bool) []audit.Rule {
	defaultRules := []audit.Rule{
		DebugLoggingRule{},
		PlaintextSecretRule{},
		WildcardBindRule{},
		TLSDisabledRule{},
		WeakAlgorithmRule{},
	}

	if includePermissions {
		defaultRules = append(defaultRules, FilePermissionsRule{})
	}

	return defaultRules
}

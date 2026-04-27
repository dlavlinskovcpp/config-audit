package rules

import "configaudit/internal/audit"

type FilePermissionsRule struct{}

func (FilePermissionsRule) ID() string {
	return "file-permissions"
}

func (FilePermissionsRule) Check(ctx audit.Context, node audit.Node) []audit.Problem {
	if node.Path != "" || !ctx.HasFileMode {
		return nil
	}

	mode := ctx.FileMode.Perm()
	switch {
	case mode&0o002 != 0:
		return []audit.Problem{{
			Severity:       audit.SeverityHigh,
			Message:        "Configuration file is world-writable.",
			Recommendation: "Restrict configuration file permissions. Sensitive configuration files should not be writable by group or others.",
		}}
	case mode&0o020 != 0:
		return []audit.Problem{{
			Severity:       audit.SeverityMedium,
			Message:        "Configuration file is group-writable.",
			Recommendation: "Restrict configuration file permissions. Sensitive configuration files should not be writable by group or others.",
		}}
	default:
		return nil
	}
}

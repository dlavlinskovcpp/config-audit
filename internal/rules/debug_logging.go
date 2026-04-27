package rules

import (
	"fmt"
	"strings"

	"configaudit/internal/audit"
)

type DebugLoggingRule struct{}

func (DebugLoggingRule) ID() string {
	return "debug-logging"
}

func (DebugLoggingRule) Check(_ audit.Context, node audit.Node) []audit.Problem {
	value, ok := node.Value.(string)
	if !ok || !isLoggingFieldPath(node.Path) {
		return nil
	}

	level := strings.TrimSpace(value)
	switch {
	case strings.EqualFold(level, "debug"), strings.EqualFold(level, "trace"):
		return []audit.Problem{{
			Severity:       audit.SeverityLow,
			Path:           node.Path,
			Message:        fmt.Sprintf("Debug or trace logging is enabled (%s).", strings.ToUpper(level)),
			Recommendation: "Do not use debug/trace logging in production. Use info or a more restrictive level.",
		}}
	default:
		return nil
	}
}

func isLoggingFieldPath(path string) bool {
	return path == "level" ||
		pathHasSuffix(path, "log.level") ||
		pathHasSuffix(path, "logging.level") ||
		pathHasSuffix(path, "logger.level") ||
		pathHasSuffix(path, "log.mode") ||
		pathHasSuffix(path, "logging.mode") ||
		pathHasSuffix(path, "logger.mode")
}

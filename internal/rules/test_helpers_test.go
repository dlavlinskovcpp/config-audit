package rules

import (
	"configaudit/internal/audit"
)

func runRule(rule audit.Rule, root map[string]any) []audit.Problem {
	engine := audit.NewEngine(rule)
	return engine.Scan(audit.Context{Root: root})
}

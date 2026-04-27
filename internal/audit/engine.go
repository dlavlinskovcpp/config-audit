package audit

import "sort"

type Engine struct {
	rules []Rule
}

func NewEngine(rules ...Rule) *Engine {
	copied := make([]Rule, len(rules))
	copy(copied, rules)

	return &Engine{rules: copied}
}

func (e *Engine) Scan(ctx Context) []Problem {
	if ctx.Root == nil {
		return nil
	}

	var problems []Problem

	Walk(ctx.Root, func(node Node) {
		for _, rule := range e.rules {
			ruleProblems := rule.Check(ctx, node)
			for _, problem := range ruleProblems {
				if problem.RuleID == "" {
					problem.RuleID = rule.ID()
				}
				if problem.File == "" {
					problem.File = ctx.File
				}
				if problem.Path == "" {
					problem.Path = node.Path
				}
				problems = append(problems, problem)
			}
		}
	})

	SortProblems(problems)

	return problems
}

func SortProblems(problems []Problem) {
	sort.SliceStable(problems, func(i, j int) bool {
		left := problems[i]
		right := problems[j]

		if severityRank(left.Severity) != severityRank(right.Severity) {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Message != right.Message {
			return left.Message < right.Message
		}

		return left.Recommendation < right.Recommendation
	})
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityHigh:
		return 0
	case SeverityMedium:
		return 1
	case SeverityLow:
		return 2
	default:
		return 3
	}
}

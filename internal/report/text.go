package report

import (
	"fmt"
	"io"

	"configaudit/internal/audit"
)

type TextOptions struct {
	AlwaysIncludeFile bool
}

func WriteText(w io.Writer, problems []audit.Problem, options ...TextOptions) {
	if len(problems) == 0 {
		fmt.Fprintln(w, "No problems found.")
		return
	}

	fmt.Fprintf(w, "Found %d problem(s)\n\n", len(problems))

	opts := TextOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	includeFile := opts.AlwaysIncludeFile || multipleFiles(problems)
	for i, problem := range problems {
		fmt.Fprintf(w, "[%s] %s\n", problem.Severity, problem.Location(includeFile))
		fmt.Fprintln(w, problem.Message)
		fmt.Fprintf(w, "Recommendation: %s\n", problem.Recommendation)

		if i < len(problems)-1 {
			fmt.Fprintln(w)
		}
	}
}

func multipleFiles(problems []audit.Problem) bool {
	seen := make(map[string]struct{})
	for _, problem := range problems {
		if problem.File == "" {
			continue
		}
		seen[problem.File] = struct{}{}
		if len(seen) > 1 {
			return true
		}
	}

	return false
}

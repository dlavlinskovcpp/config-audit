package app

import (
	"fmt"
	"io"

	"configaudit/internal/audit"
	"configaudit/internal/parser"
	"configaudit/internal/report"
)

type RunOptions struct {
	Path             string
	UseStdin         bool
	Silent           bool
	Recursive        bool
	CheckPermissions bool
	Format           parser.Format
	Stdin            io.Reader
	Stdout           io.Writer
	Stderr           io.Writer
}

func Run(opts RunOptions) int {
	stdout := writerOrDiscard(opts.Stdout)
	stderr := writerOrDiscard(opts.Stderr)

	service := NewService()

	var (
		problems []audit.Problem
		err      error
	)

	switch {
	case opts.UseStdin:
		if opts.Stdin == nil {
			fmt.Fprintln(stderr, "error: stdin is not available")
			return 2
		}
		var scanned []audit.Problem
		scanned, err = service.ScanReader(opts.Stdin, opts.Format)
		problems = scanned
	default:
		var scanned []audit.Problem
		scanned, err = service.ScanPath(opts.Path, ScanOptions{
			Format:           opts.Format,
			Recursive:        opts.Recursive,
			CheckPermissions: opts.CheckPermissions,
		})
		problems = scanned
	}

	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	report.WriteText(stdout, problems, report.TextOptions{
		AlwaysIncludeFile: opts.Recursive,
	})

	if len(problems) > 0 && !opts.Silent {
		return 1
	}

	return 0
}

func writerOrDiscard(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}

	return w
}

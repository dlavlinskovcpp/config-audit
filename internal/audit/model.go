package audit

import "io/fs"

type Severity string

const (
	SeverityLow    Severity = "LOW"
	SeverityMedium Severity = "MEDIUM"
	SeverityHigh   Severity = "HIGH"
)

type Problem struct {
	RuleID         string   `json:"rule_id,omitempty"`
	Severity       Severity `json:"severity"`
	File           string   `json:"file,omitempty"`
	Path           string   `json:"path,omitempty"`
	Message        string   `json:"message"`
	Recommendation string   `json:"recommendation"`
}

func (p Problem) Location(includeFile bool) string {
	switch {
	case includeFile && p.File != "" && p.Path != "":
		return p.File + ":" + p.Path
	case includeFile && p.File != "":
		return p.File
	case p.Path != "":
		return p.Path
	case p.File != "":
		return p.File
	default:
		return "<root>"
	}
}

type Context struct {
	File        string
	Root        map[string]any
	FileMode    fs.FileMode
	HasFileMode bool
}

type Node struct {
	Path       string
	Key        string
	Value      any
	Parent     any
	ParentPath string
}

type Rule interface {
	ID() string
	Check(ctx Context, node Node) []Problem
}

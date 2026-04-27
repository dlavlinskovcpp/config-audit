package app

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"configaudit/internal/audit"
	"configaudit/internal/parser"
	"configaudit/internal/rules"
)

type ScanOptions struct {
	Format           parser.Format
	Recursive        bool
	CheckPermissions bool
}

type Scanner interface {
	ScanContent(name string, data []byte, format parser.Format) ([]audit.Problem, error)
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ScanReader(reader io.Reader, format parser.Format) ([]audit.Problem, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}

	return s.ScanContent("", data, format)
}

func (s *Service) ScanContent(name string, data []byte, format parser.Format) ([]audit.Problem, error) {
	root, err := parser.Parse(data, name, format)
	if err != nil {
		return nil, err
	}

	engine := audit.NewEngine(rules.Default(false)...)

	return engine.Scan(audit.Context{
		File: name,
		Root: root,
	}), nil
}

func (s *Service) ScanPath(path string, opts ScanOptions) ([]audit.Problem, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}

	if info.IsDir() {
		if !opts.Recursive {
			return nil, fmt.Errorf("%q is a directory; use --recursive to scan directories", path)
		}

		return s.scanDirectory(path, opts)
	}

	return s.scanFile(path, opts)
}

func (s *Service) scanDirectory(root string, opts ScanOptions) ([]audit.Problem, error) {
	files, err := discoverConfigFiles(root)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .json, .yaml, or .yml files found under %q", root)
	}

	var allProblems []audit.Problem
	var scanErrors []error

	for _, path := range files {
		problems, err := s.scanFile(path, opts)
		if err != nil {
			scanErrors = append(scanErrors, err)
			continue
		}
		allProblems = append(allProblems, problems...)
	}

	audit.SortProblems(allProblems)

	if len(scanErrors) > 0 {
		return allProblems, errors.Join(scanErrors...)
	}

	return allProblems, nil
}

func discoverConfigFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !isConfigFile(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %q: %w", root, err)
	}

	sort.Strings(files)

	return files, nil
}

func isConfigFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func (s *Service) scanFile(path string, opts ScanOptions) ([]audit.Problem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	root, err := parser.Parse(data, path, opts.Format)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	ctx := audit.Context{
		File: path,
		Root: root,
	}

	if opts.CheckPermissions {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", path, err)
		}
		ctx.FileMode = info.Mode()
		ctx.HasFileMode = true
	}

	engine := audit.NewEngine(rules.Default(opts.CheckPermissions)...)

	return engine.Scan(ctx), nil
}

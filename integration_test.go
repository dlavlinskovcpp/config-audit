package configaudit_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce  sync.Once
	binaryPath string
	buildErr   error
)

func TestCLIReportsFindingsWithExitCodeOne(t *testing.T) {
	stdout, _, code := runConfigAudit(t, nil, "testdata/debug-log.json")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	requireContains(t, stdout, "[LOW] log.level")
}

func TestCLISilentKeepsSuccessfulExitCode(t *testing.T) {
	stdout, _, code := runConfigAudit(t, nil, "--silent", "testdata/debug-log.json")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	requireContains(t, stdout, "[LOW] log.level")
}

func TestCLIReadsFromStdin(t *testing.T) {
	stdout, _, code := runConfigAudit(t, strings.NewReader("storage:\n  digest-algorithm: MD5\n"), "--stdin")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	requireContains(t, stdout, "storage.digest-algorithm")
}

func TestCLIRejectsInvalidConfig(t *testing.T) {
	file := writeTempFile(t, "invalid.yaml", "storage:\n  digest-algorithm: [\n", 0o600)

	_, stderr, code := runConfigAudit(t, nil, file)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("expected stderr to contain parse error")
	}
}

func TestCLIReturnsZeroForCleanConfig(t *testing.T) {
	stdout, _, code := runConfigAudit(t, nil, "testdata/clean.yaml")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	requireContains(t, stdout, "No problems found.")
}

func TestCLIRecursivelyScansDirectories(t *testing.T) {
	root := t.TempDir()
	nestedDir := filepath.Join(root, "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	writeFile(t, filepath.Join(root, "clean.yaml"), "log:\n  level: info\n", 0o600)
	vulnerableFile := filepath.Join(nestedDir, "prod.yaml")
	writeFile(t, vulnerableFile, "storage:\n  digest-algorithm: MD5\n", 0o600)

	stdout, _, code := runConfigAudit(t, nil, "--recursive", root)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	requireContains(t, stdout, vulnerableFile+":storage.digest-algorithm")
}

func TestCLIChecksFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not reliable on Windows")
	}

	file := writeTempFile(t, "config.yaml", "{}\n", 0o666)
	if err := os.Chmod(file, 0o666); err != nil {
		t.Fatalf("Chmod returned error: %v", err)
	}

	stdout, _, code := runConfigAudit(t, nil, "--check-permissions", file)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	requireContains(t, stdout, "Configuration file is world-writable.")
}

func buildBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		name := "configaudit"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}

		dir, err := os.MkdirTemp("", "configaudit-bin-*")
		if err != nil {
			buildErr = err
			return
		}
		binaryPath = filepath.Join(dir, name)

		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/configaudit")

		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			buildErr = &commandError{err: err, stderr: stderr.String()}
		}
	})

	if buildErr != nil {
		t.Fatalf("failed to build binary: %v", buildErr)
	}

	return binaryPath
}

func runConfigAudit(t *testing.T, stdin io.Reader, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(buildBinary(t), args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != nil {
		cmd.Stdin = stdin
	}

	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("command failed unexpectedly: %v", err)
	}

	return stdout.String(), stderr.String(), exitErr.ExitCode()
}

func writeTempFile(t *testing.T, name, content string, mode os.FileMode) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	writeFile(t, path, content, mode)
	return path
}

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

func requireContains(t *testing.T, actual string, expected string) {
	t.Helper()

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected output to contain %q, got:\n%s", expected, actual)
	}
}

type commandError struct {
	err    error
	stderr string
}

func (e *commandError) Error() string {
	return e.err.Error() + "\n" + e.stderr
}

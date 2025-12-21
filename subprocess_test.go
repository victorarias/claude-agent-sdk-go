package sdk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindCLI_WithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "my-claude")
	if err := os.WriteFile(mockCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findCLI(mockCLI, "")
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != mockCLI {
		t.Errorf("got %q, want %q", path, mockCLI)
	}
}

func TestFindCLI_WithBundledPath(t *testing.T) {
	tmpDir := t.TempDir()
	bundledCLI := filepath.Join(tmpDir, "bundled-claude")
	if err := os.WriteFile(bundledCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findCLI("", bundledCLI)
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != bundledCLI {
		t.Errorf("got %q, want %q", path, bundledCLI)
	}
}

func TestFindCLI_WithEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(mockCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	path, err := findCLI("", "")
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != mockCLI {
		t.Errorf("got %q, want %q", path, mockCLI)
	}
}

func TestFindCLI_NotFound(t *testing.T) {
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	_, err := findCLI("", "")
	if err == nil {
		t.Error("expected error when CLI not found")
	}

	var notFoundErr *CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}

func TestFindCLI_ExplicitPathNotExists(t *testing.T) {
	_, err := findCLI("/nonexistent/path/to/claude", "")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}

	var notFoundErr *CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}

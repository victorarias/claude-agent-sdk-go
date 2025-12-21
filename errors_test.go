package sdk

import (
	"errors"
	"testing"
)

func TestSDKError(t *testing.T) {
	err := &SDKError{Message: "test error"}
	if err.Error() != "sdk: test error" {
		t.Errorf("got %q, want %q", err.Error(), "sdk: test error")
	}
}

func TestCLINotFoundError(t *testing.T) {
	err := &CLINotFoundError{SearchedPaths: []string{"/usr/bin/claude", "/usr/local/bin/claude"}}
	if !errors.Is(err, ErrCLINotFound) {
		t.Error("CLINotFoundError should match ErrCLINotFound")
	}
}

func TestConnectionError(t *testing.T) {
	err := &ConnectionError{Message: "pipe closed"}
	if !errors.Is(err, ErrConnection) {
		t.Error("ConnectionError should match ErrConnection")
	}
}

func TestProcessError(t *testing.T) {
	err := &ProcessError{ExitCode: 1, Stderr: "error output"}
	if err.ExitCode != 1 {
		t.Errorf("got exit code %d, want 1", err.ExitCode)
	}
}

func TestJSONDecodeError(t *testing.T) {
	origErr := errors.New("unexpected token")
	err := &JSONDecodeError{Line: `{"invalid`, OriginalError: origErr}
	if !errors.Is(err, ErrParse) {
		t.Error("JSONDecodeError should match ErrParse")
	}
}

func TestMessageParseError(t *testing.T) {
	err := &MessageParseError{Message: "unknown type", Data: map[string]any{"type": "unknown"}}
	if !errors.Is(err, ErrParse) {
		t.Error("MessageParseError should match ErrParse")
	}
}

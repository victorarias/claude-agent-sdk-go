package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Sentinel errors for error checking with errors.Is
var (
	ErrCLINotFound = errors.New("claude CLI not found")
	ErrCLIVersion  = errors.New("CLI version too old")
	ErrConnection  = errors.New("connection error")
	ErrProcess     = errors.New("process error")
	ErrParse       = errors.New("parse error")
	ErrTimeout     = errors.New("timeout error")
	ErrClosed      = errors.New("transport closed")
)

// SDKError is the base error type for all SDK errors.
type SDKError struct {
	Message string
	Cause   error
}

func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("sdk: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("sdk: %s", e.Message)
}

func (e *SDKError) Unwrap() error {
	return e.Cause
}

// CLINotFoundError is returned when the Claude CLI cannot be found.
type CLINotFoundError struct {
	SearchedPaths []string
	CLIPath       string // The explicit path that was tried (if any)
}

func (e *CLINotFoundError) Error() string {
	if e.CLIPath != "" {
		return fmt.Sprintf("claude CLI not found at: %s", e.CLIPath)
	}
	return fmt.Sprintf("claude CLI not found, searched: %s", strings.Join(e.SearchedPaths, ", "))
}

func (e *CLINotFoundError) Is(target error) bool {
	return target == ErrCLINotFound
}

// ConnectionError is returned when the connection to Claude fails.
type ConnectionError struct {
	Message string
	Cause   error
}

func (e *ConnectionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("connection error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("connection error: %s", e.Message)
}

func (e *ConnectionError) Is(target error) bool {
	return target == ErrConnection
}

func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// ProcessError is returned when the CLI process fails.
type ProcessError struct {
	ExitCode int
	Stderr   string
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("process exited with code %d: %s", e.ExitCode, e.Stderr)
}

func (e *ProcessError) Is(target error) bool {
	return target == ErrProcess
}

// JSONDecodeError is returned when JSON from CLI cannot be decoded.
type JSONDecodeError struct {
	Line          string
	OriginalError error
}

func (e *JSONDecodeError) Error() string {
	return fmt.Sprintf("JSON decode error on line %q: %v", e.Line, e.OriginalError)
}

func (e *JSONDecodeError) Is(target error) bool {
	return target == ErrParse
}

func (e *JSONDecodeError) Unwrap() error {
	return e.OriginalError
}

// MessageParseError is returned when a message cannot be parsed.
type MessageParseError struct {
	Message string
	Data    map[string]any
}

func (e *MessageParseError) Error() string {
	return fmt.Sprintf("message parse error: %s", e.Message)
}

func (e *MessageParseError) Is(target error) bool {
	return target == ErrParse
}

// CLIVersionError is returned when the CLI version is too old.
type CLIVersionError struct {
	InstalledVersion string
	MinimumVersion   string
}

func (e *CLIVersionError) Error() string {
	return fmt.Sprintf("CLI version %s is below minimum required version %s", e.InstalledVersion, e.MinimumVersion)
}

func (e *CLIVersionError) Is(target error) bool {
	return target == ErrCLIVersion
}

// TimeoutError is returned when an operation times out.
type TimeoutError struct {
	Operation string        // What operation timed out
	Duration  time.Duration // How long we waited
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout after %v: %s", e.Duration, e.Operation)
}

func (e *TimeoutError) Is(target error) bool {
	return target == ErrTimeout
}

// ClosedError is returned when an operation is attempted on a closed resource.
type ClosedError struct {
	Resource string // What resource was closed
}

func (e *ClosedError) Error() string {
	return fmt.Sprintf("resource closed: %s", e.Resource)
}

func (e *ClosedError) Is(target error) bool {
	return target == ErrClosed
}

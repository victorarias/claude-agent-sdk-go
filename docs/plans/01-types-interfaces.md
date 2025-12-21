# Plan 01: Types & Interfaces

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Define all Go types that mirror the Python SDK's type system.

**Architecture:** Use Go structs with JSON tags for serialization. Use interfaces for polymorphic types (Message, ContentBlock). Use functional options pattern for ClaudeAgentOptions.

**Tech Stack:** Go 1.21+, encoding/json

---

## Task 1: Initialize Go Module

**Files:**
- Create: `go.mod`
- Create: `sdk.go`

**Step 1: Create go.mod**

```bash
mkdir -p ~/projects/claude-agent-sdk-go
cd ~/projects/claude-agent-sdk-go
go mod init github.com/victorarias/claude-agent-sdk-go
```

**Step 2: Create package file**

Create `sdk.go`:

```go
// Package sdk provides a Go client for the Claude Agent SDK.
//
// The SDK spawns the Claude CLI as a subprocess and communicates
// via JSON streaming for bidirectional control protocol.
package sdk

// Version is the SDK version.
const Version = "0.1.0"
```

**Step 3: Verify module**

```bash
go build ./...
```

Expected: No errors

**Step 4: Commit**

```bash
git init
git add go.mod sdk.go
git commit -m "feat: initialize claude-agent-sdk-go module"
```

---

## Task 2: Define Error Types

**Files:**
- Create: `errors.go`
- Create: `errors_test.go`

**Step 1: Write failing test**

Create `errors_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestSDKError -v
```

Expected: FAIL - types not defined

**Step 3: Write implementation**

Create `errors.go`:

```go
package sdk

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for error checking with errors.Is
var (
	ErrCLINotFound = errors.New("claude CLI not found")
	ErrConnection  = errors.New("connection error")
	ErrProcess     = errors.New("process error")
	ErrParse       = errors.New("parse error")
	ErrTimeout     = errors.New("timeout error")
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
}

func (e *CLINotFoundError) Error() string {
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

// ParseError is returned when message parsing fails.
type ParseError struct {
	Message string
	RawData string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %s", e.Message)
}

func (e *ParseError) Is(target error) bool {
	return target == ErrParse
}
```

**Step 4: Run tests**

```bash
go test -run "TestSDKError|TestCLINotFoundError|TestConnectionError|TestProcessError" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add errors.go errors_test.go
git commit -m "feat: add error types"
```

---

## Task 3: Define Content Block Types

**Files:**
- Create: `types.go`
- Create: `types_test.go`

**Step 1: Write failing test**

Create `types_test.go`:

```go
package sdk

import (
	"encoding/json"
	"testing"
)

func TestTextBlock(t *testing.T) {
	block := TextBlock{Text: "hello"}
	if block.BlockType() != "text" {
		t.Errorf("got %q, want %q", block.BlockType(), "text")
	}
}

func TestToolUseBlock(t *testing.T) {
	block := ToolUseBlock{
		ID:    "tool_123",
		Name:  "Bash",
		Input: map[string]any{"command": "ls"},
	}
	if block.BlockType() != "tool_use" {
		t.Errorf("got %q, want %q", block.BlockType(), "tool_use")
	}
}

func TestThinkingBlock(t *testing.T) {
	block := ThinkingBlock{
		Thinking:  "Let me think...",
		Signature: "sig123",
	}
	if block.BlockType() != "thinking" {
		t.Errorf("got %q, want %q", block.BlockType(), "thinking")
	}
}

func TestToolResultBlock(t *testing.T) {
	block := ToolResultBlock{
		ToolUseID: "tool_123",
		Content:   "output",
		IsError:   false,
	}
	if block.BlockType() != "tool_result" {
		t.Errorf("got %q, want %q", block.BlockType(), "tool_result")
	}
}

func TestContentBlockJSON(t *testing.T) {
	input := `{"type":"text","text":"hello"}`
	var raw map[string]any
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatal(err)
	}

	block, err := ParseContentBlock(raw)
	if err != nil {
		t.Fatal(err)
	}

	textBlock, ok := block.(*TextBlock)
	if !ok {
		t.Fatalf("expected *TextBlock, got %T", block)
	}
	if textBlock.Text != "hello" {
		t.Errorf("got %q, want %q", textBlock.Text, "hello")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestTextBlock -v
```

Expected: FAIL - types not defined

**Step 3: Write implementation**

Create `types.go`:

```go
package sdk

import (
	"encoding/json"
	"fmt"
)

// ContentBlock represents a block of content in a message.
type ContentBlock interface {
	BlockType() string
}

// TextBlock contains text content.
type TextBlock struct {
	Text string `json:"text"`
}

func (b *TextBlock) BlockType() string { return "text" }

// ThinkingBlock contains Claude's thinking content.
type ThinkingBlock struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

func (b *ThinkingBlock) BlockType() string { return "thinking" }

// ToolUseBlock represents a tool invocation.
type ToolUseBlock struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

func (b *ToolUseBlock) BlockType() string { return "tool_use" }

// ToolResultBlock contains the result of a tool execution.
type ToolResultBlock struct {
	ToolUseID string `json:"tool_use_id"`
	Content   any    `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

func (b *ToolResultBlock) BlockType() string { return "tool_result" }

// ParseContentBlock parses a raw JSON map into a ContentBlock.
func ParseContentBlock(raw map[string]any) (ContentBlock, error) {
	blockType, ok := raw["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'type' field")
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal block: %w", err)
	}

	switch blockType {
	case "text":
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse text block: %w", err)
		}
		return &block, nil

	case "thinking":
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse thinking block: %w", err)
		}
		return &block, nil

	case "tool_use":
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_use block: %w", err)
		}
		return &block, nil

	case "tool_result":
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_result block: %w", err)
		}
		return &block, nil

	default:
		return nil, fmt.Errorf("unknown block type: %s", blockType)
	}
}
```

**Step 4: Run tests**

```bash
go test -run "TestTextBlock|TestToolUseBlock|TestThinkingBlock|TestToolResultBlock|TestContentBlockJSON" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add content block types"
```

---

## Task 4: Define Message Types

**Files:**
- Modify: `types.go`
- Modify: `types_test.go`

**Step 1: Write failing test**

Add to `types_test.go`:

```go
func TestUserMessage(t *testing.T) {
	msg := &UserMessage{Content: "hello"}
	if msg.MessageType() != "user" {
		t.Errorf("got %q, want %q", msg.MessageType(), "user")
	}
}

func TestAssistantMessage(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{&TextBlock{Text: "hello"}},
		Model:   "claude-sonnet-4-5",
	}
	if msg.MessageType() != "assistant" {
		t.Errorf("got %q, want %q", msg.MessageType(), "assistant")
	}
}

func TestSystemMessage(t *testing.T) {
	msg := &SystemMessage{
		Subtype: "init",
		Data:    map[string]any{"version": "1.0"},
	}
	if msg.MessageType() != "system" {
		t.Errorf("got %q, want %q", msg.MessageType(), "system")
	}
}

func TestResultMessage(t *testing.T) {
	msg := &ResultMessage{
		Subtype:      "success",
		DurationMS:   1000,
		DurationAPI:  800,
		IsError:      false,
		NumTurns:     3,
		SessionID:    "sess_123",
		TotalCostUSD: floatPtr(0.05),
	}
	if msg.MessageType() != "result" {
		t.Errorf("got %q, want %q", msg.MessageType(), "result")
	}
}

func floatPtr(f float64) *float64 { return &f }
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestUserMessage -v
```

Expected: FAIL - Message type not defined

**Step 3: Write implementation**

Add to `types.go`:

```go
// Message represents a message in the conversation.
type Message interface {
	MessageType() string
}

// UserMessage represents a user's message.
type UserMessage struct {
	Content         any     `json:"content"` // string or []ContentBlock
	UUID            *string `json:"uuid,omitempty"`
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`
}

func (m *UserMessage) MessageType() string { return "user" }

// AssistantMessage represents Claude's response.
type AssistantMessage struct {
	Content         []ContentBlock `json:"content"`
	Model           string         `json:"model"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
	Error           *string        `json:"error,omitempty"`
}

func (m *AssistantMessage) MessageType() string { return "assistant" }

// SystemMessage represents a system message.
type SystemMessage struct {
	Subtype string         `json:"subtype"`
	Data    map[string]any `json:"data,omitempty"`
}

func (m *SystemMessage) MessageType() string { return "system" }

// ResultMessage represents the final result of a query.
type ResultMessage struct {
	Subtype          string         `json:"subtype"`
	DurationMS       int            `json:"duration_ms"`
	DurationAPI      int            `json:"duration_api_ms"`
	IsError          bool           `json:"is_error"`
	NumTurns         int            `json:"num_turns"`
	SessionID        string         `json:"session_id"`
	TotalCostUSD     *float64       `json:"total_cost_usd,omitempty"`
	Usage            map[string]any `json:"usage,omitempty"`
	Result           *string        `json:"result,omitempty"`
	StructuredOutput any            `json:"structured_output,omitempty"`
}

func (m *ResultMessage) MessageType() string { return "result" }

// StreamEvent represents a streaming event.
type StreamEvent struct {
	UUID            string         `json:"uuid"`
	SessionID       string         `json:"session_id"`
	Event           map[string]any `json:"event"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
}

func (m *StreamEvent) MessageType() string { return "stream_event" }
```

**Step 4: Run tests**

```bash
go test -run "TestUserMessage|TestAssistantMessage|TestSystemMessage|TestResultMessage" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add message types"
```

---

## Task 5: Define Options Types

**Files:**
- Create: `options.go`
- Create: `options_test.go`

**Step 1: Write failing test**

Create `options_test.go`:

```go
package sdk

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.PermissionMode != "" {
		t.Errorf("expected empty permission mode, got %q", opts.PermissionMode)
	}
}

func TestOptionsWithModel(t *testing.T) {
	opts := DefaultOptions()
	WithModel("claude-opus-4")(opts)
	if opts.Model != "claude-opus-4" {
		t.Errorf("got %q, want %q", opts.Model, "claude-opus-4")
	}
}

func TestOptionsWithCwd(t *testing.T) {
	opts := DefaultOptions()
	WithCwd("/tmp/test")(opts)
	if opts.Cwd != "/tmp/test" {
		t.Errorf("got %q, want %q", opts.Cwd, "/tmp/test")
	}
}

func TestOptionsWithPermissionMode(t *testing.T) {
	opts := DefaultOptions()
	WithPermissionMode(PermissionBypass)(opts)
	if opts.PermissionMode != PermissionBypass {
		t.Errorf("got %q, want %q", opts.PermissionMode, PermissionBypass)
	}
}

func TestOptionsWithEnv(t *testing.T) {
	opts := DefaultOptions()
	WithEnv(map[string]string{"FOO": "bar"})(opts)
	if opts.Env["FOO"] != "bar" {
		t.Errorf("got %q, want %q", opts.Env["FOO"], "bar")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestDefaultOptions -v
```

Expected: FAIL - Options not defined

**Step 3: Write implementation**

Create `options.go`:

```go
package sdk

// PermissionMode controls how tool permissions are handled.
type PermissionMode string

const (
	PermissionDefault  PermissionMode = "default"
	PermissionAccept   PermissionMode = "acceptEdits"
	PermissionPlan     PermissionMode = "plan"
	PermissionBypass   PermissionMode = "bypassPermissions"
)

// Options configures the Claude SDK client.
type Options struct {
	// Tools configuration
	Tools           []string `json:"tools,omitempty"`
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`

	// System prompt
	SystemPrompt       string `json:"system_prompt,omitempty"`
	AppendSystemPrompt string `json:"append_system_prompt,omitempty"`

	// Model configuration
	Model         string `json:"model,omitempty"`
	FallbackModel string `json:"fallback_model,omitempty"`

	// Permission settings
	PermissionMode           PermissionMode `json:"permission_mode,omitempty"`
	PermissionPromptToolName string         `json:"permission_prompt_tool_name,omitempty"`

	// Session settings
	ContinueConversation bool   `json:"continue_conversation,omitempty"`
	Resume               string `json:"resume,omitempty"`
	ForkSession          bool   `json:"fork_session,omitempty"`

	// Limits
	MaxTurns          int     `json:"max_turns,omitempty"`
	MaxBudgetUSD      float64 `json:"max_budget_usd,omitempty"`
	MaxThinkingTokens int     `json:"max_thinking_tokens,omitempty"`

	// Paths
	Cwd     string `json:"cwd,omitempty"`
	CLIPath string `json:"cli_path,omitempty"`
	AddDirs []string `json:"add_dirs,omitempty"`

	// Environment
	Env map[string]string `json:"env,omitempty"`

	// MCP Servers
	MCPServers map[string]MCPServerConfig `json:"mcp_servers,omitempty"`

	// Streaming
	IncludePartialMessages bool `json:"include_partial_messages,omitempty"`

	// File checkpointing
	EnableFileCheckpointing bool `json:"enable_file_checkpointing,omitempty"`

	// Output format
	OutputFormat map[string]any `json:"output_format,omitempty"`

	// Extra CLI arguments
	ExtraArgs map[string]string `json:"extra_args,omitempty"`

	// Betas
	Betas []string `json:"betas,omitempty"`

	// Settings sources
	SettingSources []string `json:"setting_sources,omitempty"`
	Settings       string   `json:"settings,omitempty"`
}

// MCPServerConfig defines an MCP server.
type MCPServerConfig struct {
	Type    string            `json:"type"` // "stdio", "sse", "http"
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Env: make(map[string]string),
	}
}

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithCwd sets the working directory.
func WithCwd(cwd string) Option {
	return func(o *Options) {
		o.Cwd = cwd
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = mode
	}
}

// WithEnv sets environment variables.
func WithEnv(env map[string]string) Option {
	return func(o *Options) {
		for k, v := range env {
			o.Env[k] = v
		}
	}
}

// WithSystemPrompt sets a custom system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = prompt
	}
}

// WithMaxTurns limits the number of conversation turns.
func WithMaxTurns(n int) Option {
	return func(o *Options) {
		o.MaxTurns = n
	}
}

// WithMaxBudget limits spending in USD.
func WithMaxBudget(usd float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = usd
	}
}

// WithTools specifies which tools to enable.
func WithTools(tools ...string) Option {
	return func(o *Options) {
		o.Tools = tools
	}
}

// WithCLIPath sets a custom path to the Claude CLI.
func WithCLIPath(path string) Option {
	return func(o *Options) {
		o.CLIPath = path
	}
}

// WithResume resumes a previous session.
func WithResume(sessionID string) Option {
	return func(o *Options) {
		o.Resume = sessionID
	}
}

// WithContinue continues the last conversation.
func WithContinue() Option {
	return func(o *Options) {
		o.ContinueConversation = true
	}
}

// ApplyOptions applies functional options to an Options struct.
func ApplyOptions(opts *Options, options ...Option) {
	for _, opt := range options {
		opt(opts)
	}
}
```

**Step 4: Run tests**

```bash
go test -run "TestDefaultOptions|TestOptionsWith" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add options.go options_test.go
git commit -m "feat: add options types with functional options pattern"
```

---

## Task 6: Define Hook Types

**Files:**
- Modify: `types.go`
- Modify: `types_test.go`

**Step 1: Write failing test**

Add to `types_test.go`:

```go
func TestHookEvent(t *testing.T) {
	events := []HookEvent{
		HookPreToolUse,
		HookPostToolUse,
		HookUserPromptSubmit,
		HookStop,
		HookSubagentStop,
		HookPreCompact,
	}

	if len(events) != 6 {
		t.Errorf("expected 6 hook events, got %d", len(events))
	}
}

func TestHookInput(t *testing.T) {
	input := &PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "sess_123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls"},
	}

	if input.SessionID != "sess_123" {
		t.Errorf("got %q, want %q", input.SessionID, "sess_123")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestHookEvent -v
```

Expected: FAIL - HookEvent not defined

**Step 3: Write implementation**

Add to `types.go`:

```go
// HookEvent represents the type of hook event.
type HookEvent string

const (
	HookPreToolUse        HookEvent = "PreToolUse"
	HookPostToolUse       HookEvent = "PostToolUse"
	HookUserPromptSubmit  HookEvent = "UserPromptSubmit"
	HookStop              HookEvent = "Stop"
	HookSubagentStop      HookEvent = "SubagentStop"
	HookPreCompact        HookEvent = "PreCompact"
)

// BaseHookInput contains fields common to all hook inputs.
type BaseHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode,omitempty"`
}

// PreToolUseHookInput is the input for PreToolUse hooks.
type PreToolUseHookInput struct {
	BaseHookInput
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
}

// PostToolUseHookInput is the input for PostToolUse hooks.
type PostToolUseHookInput struct {
	BaseHookInput
	ToolName     string         `json:"tool_name"`
	ToolInput    map[string]any `json:"tool_input"`
	ToolResponse any            `json:"tool_response"`
}

// UserPromptSubmitHookInput is the input for UserPromptSubmit hooks.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	Prompt string `json:"prompt"`
}

// StopHookInput is the input for Stop hooks.
type StopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

// SubagentStopHookInput is the input for SubagentStop hooks.
type SubagentStopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

// PreCompactHookInput is the input for PreCompact hooks.
type PreCompactHookInput struct {
	BaseHookInput
	Trigger            string  `json:"trigger"` // "manual" or "auto"
	CustomInstructions *string `json:"custom_instructions,omitempty"`
}

// HookOutput is the output from a hook callback.
type HookOutput struct {
	Continue       bool                   `json:"continue,omitempty"`
	SuppressOutput bool                   `json:"suppressOutput,omitempty"`
	StopReason     string                 `json:"stopReason,omitempty"`
	Decision       string                 `json:"decision,omitempty"` // "block"
	SystemMessage  string                 `json:"systemMessage,omitempty"`
	Reason         string                 `json:"reason,omitempty"`
	HookSpecific   map[string]any         `json:"hookSpecificOutput,omitempty"`
}

// HookCallback is the signature for hook callback functions.
type HookCallback func(input any, toolUseID *string) (*HookOutput, error)

// HookMatcher configures which hooks to run for an event.
type HookMatcher struct {
	Matcher  string         `json:"matcher,omitempty"` // e.g., "Bash" or "Write|Edit"
	Hooks    []HookCallback `json:"-"`                 // Callbacks (not serialized)
	Timeout  float64        `json:"timeout,omitempty"` // Timeout in seconds
}
```

**Step 4: Run tests**

```bash
go test -run "TestHookEvent|TestHookInput" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add hook types"
```

---

## Task 7: Define Control Protocol Types

**Files:**
- Modify: `types.go`
- Modify: `types_test.go`

**Step 1: Write failing test**

Add to `types_test.go`:

```go
func TestControlRequest(t *testing.T) {
	req := &ControlRequest{
		Type:      "control_request",
		RequestID: "req_123",
		Request: map[string]any{
			"subtype": "interrupt",
		},
	}

	if req.RequestID != "req_123" {
		t.Errorf("got %q, want %q", req.RequestID, "req_123")
	}
}

func TestControlResponse(t *testing.T) {
	resp := &ControlResponse{
		Type: "control_response",
		Response: ControlResponseData{
			Subtype:   "success",
			RequestID: "req_123",
			Response:  map[string]any{"status": "ok"},
		},
	}

	if resp.Response.Subtype != "success" {
		t.Errorf("got %q, want %q", resp.Response.Subtype, "success")
	}
}

func TestPermissionResult(t *testing.T) {
	allow := &PermissionResultAllow{
		Behavior:     "allow",
		UpdatedInput: map[string]any{"command": "ls -la"},
	}

	if allow.Behavior != "allow" {
		t.Errorf("got %q, want %q", allow.Behavior, "allow")
	}

	deny := &PermissionResultDeny{
		Behavior:  "deny",
		Message:   "not allowed",
		Interrupt: true,
	}

	if deny.Behavior != "deny" {
		t.Errorf("got %q, want %q", deny.Behavior, "deny")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestControlRequest -v
```

Expected: FAIL - ControlRequest not defined

**Step 3: Write implementation**

Add to `types.go`:

```go
// ControlRequest is sent to the CLI for control operations.
type ControlRequest struct {
	Type      string         `json:"type"` // "control_request"
	RequestID string         `json:"request_id"`
	Request   map[string]any `json:"request"`
}

// ControlResponseData contains the response data.
type ControlResponseData struct {
	Subtype   string         `json:"subtype"` // "success" or "error"
	RequestID string         `json:"request_id"`
	Response  map[string]any `json:"response,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// ControlResponse is received from the CLI for control operations.
type ControlResponse struct {
	Type     string              `json:"type"` // "control_response"
	Response ControlResponseData `json:"response"`
}

// PermissionResultAllow allows a tool to run.
type PermissionResultAllow struct {
	Behavior           string           `json:"behavior"` // "allow"
	UpdatedInput       map[string]any   `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate `json:"updatedPermissions,omitempty"`
}

// PermissionResultDeny denies a tool from running.
type PermissionResultDeny struct {
	Behavior  string `json:"behavior"` // "deny"
	Message   string `json:"message,omitempty"`
	Interrupt bool   `json:"interrupt,omitempty"`
}

// PermissionUpdate describes a permission change.
type PermissionUpdate struct {
	Type        string   `json:"type"` // "addRules", "replaceRules", etc.
	Rules       []PermissionRule `json:"rules,omitempty"`
	Behavior    string   `json:"behavior,omitempty"` // "allow", "deny", "ask"
	Mode        string   `json:"mode,omitempty"`
	Directories []string `json:"directories,omitempty"`
	Destination string   `json:"destination,omitempty"` // "userSettings", "projectSettings", etc.
}

// PermissionRule defines a permission rule.
type PermissionRule struct {
	ToolName    string  `json:"toolName"`
	RuleContent *string `json:"ruleContent,omitempty"`
}

// ToolPermissionContext provides context for permission callbacks.
type ToolPermissionContext struct {
	Signal      any                `json:"-"` // Future: abort signal
	Suggestions []PermissionUpdate `json:"suggestions,omitempty"`
}

// CanUseToolCallback is called when a tool needs permission.
type CanUseToolCallback func(
	toolName string,
	input map[string]any,
	ctx *ToolPermissionContext,
) (any, error) // Returns PermissionResultAllow or PermissionResultDeny
```

**Step 4: Run tests**

```bash
go test -run "TestControlRequest|TestControlResponse|TestPermissionResult" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add control protocol types"
```

---

## Task 8: Define Transport Interface

**Files:**
- Create: `transport.go`
- Create: `transport_test.go`

**Step 1: Write failing test**

Create `transport_test.go`:

```go
package sdk

import (
	"context"
	"testing"
)

// MockTransport implements Transport for testing.
type MockTransport struct {
	connected  bool
	messages   chan map[string]any
	written    []string
	closeError error
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		messages: make(chan map[string]any, 10),
	}
}

func (t *MockTransport) Connect(ctx context.Context) error {
	t.connected = true
	return nil
}

func (t *MockTransport) Close() error {
	t.connected = false
	close(t.messages)
	return t.closeError
}

func (t *MockTransport) Write(data string) error {
	t.written = append(t.written, data)
	return nil
}

func (t *MockTransport) EndInput() error {
	return nil
}

func (t *MockTransport) Messages() <-chan map[string]any {
	return t.messages
}

func (t *MockTransport) IsReady() bool {
	return t.connected
}

func TestMockTransportImplementsInterface(t *testing.T) {
	var _ Transport = (*MockTransport)(nil)
}

func TestMockTransportConnect(t *testing.T) {
	transport := NewMockTransport()

	if transport.IsReady() {
		t.Error("should not be ready before connect")
	}

	err := transport.Connect(context.Background())
	if err != nil {
		t.Errorf("connect failed: %v", err)
	}

	if !transport.IsReady() {
		t.Error("should be ready after connect")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestMockTransportImplementsInterface -v
```

Expected: FAIL - Transport not defined

**Step 3: Write implementation**

Create `transport.go`:

```go
package sdk

import (
	"context"
)

// Transport defines the interface for communicating with Claude.
type Transport interface {
	// Connect establishes the connection.
	Connect(ctx context.Context) error

	// Close terminates the connection and cleans up resources.
	Close() error

	// Write sends data to the CLI.
	Write(data string) error

	// EndInput signals that no more input will be sent.
	EndInput() error

	// Messages returns a channel of parsed JSON messages from the CLI.
	Messages() <-chan map[string]any

	// IsReady returns true if the transport is connected and ready.
	IsReady() bool
}
```

**Step 4: Run tests**

```bash
go test -run "TestMockTransportImplementsInterface|TestMockTransportConnect" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add transport.go transport_test.go
git commit -m "feat: add Transport interface"
```

---

## Summary

After completing Plan 01, you have:

- [x] Go module initialized
- [x] Error types with sentinel errors
- [x] Content block types (Text, Thinking, ToolUse, ToolResult)
- [x] Message types (User, Assistant, System, Result, StreamEvent)
- [x] Options with functional options pattern
- [x] Hook types
- [x] Control protocol types
- [x] Transport interface

**Next:** Plan 02 - Transport Layer (subprocess implementation)

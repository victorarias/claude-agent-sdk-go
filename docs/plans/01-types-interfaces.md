# Plan 01: Types & Interfaces

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Define all Go types that mirror the Python SDK's type system with complete feature parity.

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

// MinimumCLIVersion is the minimum supported CLI version.
const MinimumCLIVersion = "2.0.0"
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
```

**Step 4: Run tests**

```bash
go test -run "TestSDKError|TestCLINotFoundError|TestConnectionError|TestProcessError|TestJSONDecodeError|TestMessageParseError" -v
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
	"strings"
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

// Text returns all text content concatenated from an AssistantMessage.
func (m *AssistantMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			parts = append(parts, textBlock.Text)
		}
	}
	return strings.Join(parts, "")
}

// ToolCalls returns all tool use blocks from an AssistantMessage.
func (m *AssistantMessage) ToolCalls() []*ToolUseBlock {
	var tools []*ToolUseBlock
	for _, block := range m.Content {
		if toolBlock, ok := block.(*ToolUseBlock); ok {
			tools = append(tools, toolBlock)
		}
	}
	return tools
}

// Thinking returns the thinking content if present.
func (m *AssistantMessage) Thinking() string {
	for _, block := range m.Content {
		if thinkingBlock, ok := block.(*ThinkingBlock); ok {
			return thinkingBlock.Thinking
		}
	}
	return ""
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
	if !msg.IsSuccess() {
		t.Error("expected IsSuccess() to return true")
	}
	if msg.Cost() != 0.05 {
		t.Errorf("got cost %f, want 0.05", msg.Cost())
	}
}

func TestStreamEvent(t *testing.T) {
	msg := &StreamEvent{
		UUID:      "uuid_123",
		SessionID: "sess_123",
		Event:     map[string]any{"type": "content_block_delta"},
	}
	if msg.MessageType() != "stream_event" {
		t.Errorf("got %q, want %q", msg.MessageType(), "stream_event")
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

// AssistantMessageError represents error types for assistant messages.
type AssistantMessageError string

const (
	ErrorAuthenticationFailed AssistantMessageError = "authentication_failed"
	ErrorBillingError         AssistantMessageError = "billing_error"
	ErrorRateLimit            AssistantMessageError = "rate_limit"
	ErrorInvalidRequest       AssistantMessageError = "invalid_request"
	ErrorServerError          AssistantMessageError = "server_error"
	ErrorUnknown              AssistantMessageError = "unknown"
)

// AssistantMessage represents Claude's response.
type AssistantMessage struct {
	Content         []ContentBlock         `json:"content"`
	Model           string                 `json:"model"`
	ParentToolUseID *string                `json:"parent_tool_use_id,omitempty"`
	Error           *AssistantMessageError `json:"error,omitempty"`
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

// IsSuccess returns true if the result is successful.
func (m *ResultMessage) IsSuccess() bool {
	return !m.IsError && m.Subtype == "success"
}

// Cost returns the cost in USD.
func (m *ResultMessage) Cost() float64 {
	if m.TotalCostUSD != nil {
		return *m.TotalCostUSD
	}
	return 0
}

// StreamEvent represents a streaming event with partial updates.
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
go test -run "TestUserMessage|TestAssistantMessage|TestSystemMessage|TestResultMessage|TestStreamEvent" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add message types"
```

---

## Task 5: Define Options Types (Complete)

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

func TestOptionsWithHooks(t *testing.T) {
	opts := DefaultOptions()
	callback := func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: true}, nil
	}
	WithHook(HookPreToolUse, HookMatcher{
		Matcher: strPtr("Bash"),
		Hooks:   []HookCallback{callback},
	})(opts)
	if len(opts.Hooks[HookPreToolUse]) != 1 {
		t.Error("expected 1 hook matcher")
	}
}

func TestOptionsWithCanUseTool(t *testing.T) {
	opts := DefaultOptions()
	called := false
	WithCanUseTool(func(toolName string, input map[string]any, ctx *ToolPermissionContext) (PermissionResult, error) {
		called = true
		return &PermissionResultAllow{}, nil
	})(opts)
	if opts.CanUseTool == nil {
		t.Error("expected CanUseTool to be set")
	}
}

func TestOptionsWithSandbox(t *testing.T) {
	opts := DefaultOptions()
	WithSandbox(SandboxSettings{Enabled: true})(opts)
	if opts.Sandbox == nil || !opts.Sandbox.Enabled {
		t.Error("expected sandbox to be enabled")
	}
}

func TestOptionsWithAgents(t *testing.T) {
	opts := DefaultOptions()
	WithAgents(map[string]AgentDefinition{
		"test": {Description: "Test agent", Prompt: "You are a test"},
	})(opts)
	if opts.Agents["test"].Description != "Test agent" {
		t.Error("expected agent to be set")
	}
}

func strPtr(s string) *string { return &s }
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
	PermissionDefault PermissionMode = "default"
	PermissionAccept  PermissionMode = "acceptEdits"
	PermissionPlan    PermissionMode = "plan"
	PermissionBypass  PermissionMode = "bypassPermissions"
)

// SettingSource specifies where settings come from.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// SdkBeta represents beta feature flags.
type SdkBeta string

const (
	BetaContext1M SdkBeta = "context-1m-2025-08-07"
)

// AgentModel specifies the model for custom agents.
type AgentModel string

const (
	AgentModelSonnet  AgentModel = "sonnet"
	AgentModelOpus    AgentModel = "opus"
	AgentModelHaiku   AgentModel = "haiku"
	AgentModelInherit AgentModel = "inherit"
)

// AgentDefinition defines a custom agent.
type AgentDefinition struct {
	Description string     `json:"description"`
	Prompt      string     `json:"prompt"`
	Tools       []string   `json:"tools,omitempty"`
	Model       AgentModel `json:"model,omitempty"`
}

// PluginConfig defines a plugin configuration.
type PluginConfig struct {
	Type string `json:"type"` // "local"
	Path string `json:"path"`
}

// SandboxNetworkConfig defines network isolation settings.
type SandboxNetworkConfig struct {
	AllowUnixSockets    []string `json:"allowUnixSockets,omitempty"`
	AllowAllUnixSockets bool     `json:"allowAllUnixSockets,omitempty"`
	AllowLocalBinding   bool     `json:"allowLocalBinding,omitempty"`
	HTTPProxyPort       *int     `json:"httpProxyPort,omitempty"`
	SocksProxyPort      *int     `json:"socksProxyPort,omitempty"`
}

// SandboxIgnoreViolations defines violation ignore rules.
type SandboxIgnoreViolations struct {
	File    []string `json:"file,omitempty"`
	Network []string `json:"network,omitempty"`
}

// SandboxSettings defines sandbox configuration for isolation.
type SandboxSettings struct {
	Enabled                   bool                     `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed  bool                     `json:"autoAllowBashIfSandboxed,omitempty"`
	ExcludedCommands          []string                 `json:"excludedCommands,omitempty"`
	AllowUnsandboxedCommands  bool                     `json:"allowUnsandboxedCommands,omitempty"`
	Network                   *SandboxNetworkConfig    `json:"network,omitempty"`
	IgnoreViolations          *SandboxIgnoreViolations `json:"ignoreViolations,omitempty"`
	EnableWeakerNestedSandbox bool                     `json:"enableWeakerNestedSandbox,omitempty"`
}

// SystemPromptPreset allows using preset system prompts.
type SystemPromptPreset struct {
	Type   string  `json:"type"`   // "preset"
	Preset string  `json:"preset"` // "claude_code"
	Append *string `json:"append,omitempty"`
}

// ToolsPreset allows using preset tool configurations.
type ToolsPreset struct {
	Type   string `json:"type"`   // "preset"
	Preset string `json:"preset"` // "claude_code"
}

// Options configures the Claude SDK client.
type Options struct {
	// Tools configuration
	Tools           []string `json:"tools,omitempty"`
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`

	// System prompt - can be string or SystemPromptPreset
	SystemPrompt       any    `json:"system_prompt,omitempty"`
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
	MaxBufferSize     int     `json:"max_buffer_size,omitempty"` // Max bytes for stdout buffering (default 1MB)

	// Paths
	Cwd     string   `json:"cwd,omitempty"`
	CLIPath string   `json:"cli_path,omitempty"`
	AddDirs []string `json:"add_dirs,omitempty"`

	// Environment
	Env map[string]string `json:"env,omitempty"`

	// MCP Servers - can be dict or path to config file
	MCPServers any `json:"mcp_servers,omitempty"`

	// SDK MCP Servers (in-process)
	SDKMCPServers map[string]*MCPServer `json:"-"`

	// Streaming
	IncludePartialMessages bool `json:"include_partial_messages,omitempty"`

	// File checkpointing
	EnableFileCheckpointing bool `json:"enable_file_checkpointing,omitempty"`

	// Output format for structured outputs
	OutputFormat map[string]any `json:"output_format,omitempty"`

	// Extra CLI arguments
	ExtraArgs map[string]string `json:"extra_args,omitempty"`

	// Beta features
	Betas []SdkBeta `json:"betas,omitempty"`

	// Settings sources
	SettingSources []SettingSource `json:"setting_sources,omitempty"`
	Settings       string          `json:"settings,omitempty"`

	// User identifier
	User string `json:"user,omitempty"`

	// Custom agents
	Agents map[string]AgentDefinition `json:"agents,omitempty"`

	// Sandbox configuration
	Sandbox *SandboxSettings `json:"sandbox,omitempty"`

	// Plugins
	Plugins []PluginConfig `json:"plugins,omitempty"`

	// Hooks
	Hooks map[HookEvent][]HookMatcher `json:"-"`

	// Permission callback
	CanUseTool CanUseToolCallback `json:"-"`

	// Stderr callback for debugging
	StderrCallback func(string) `json:"-"`

	// Internal: custom transport for testing
	customTransport Transport `json:"-"`
}

// MCPServerConfig defines an external MCP server.
type MCPServerConfig struct {
	Type    string            `json:"type"` // "stdio", "sse", "http", "sdk"
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// MCPServer represents an in-process SDK MCP server.
type MCPServer struct {
	Name    string
	Version string
	Tools   []*MCPTool
}

// MCPTool represents a tool in an MCP server.
type MCPTool struct {
	Name        string
	Description string
	Schema      map[string]any
	Handler     MCPToolHandler
}

// MCPToolHandler is the function signature for MCP tool handlers.
type MCPToolHandler func(args map[string]any) (*MCPToolResult, error)

// MCPToolResult is the result of an MCP tool invocation.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
}

// MCPContent represents content in an MCP result.
type MCPContent struct {
	Type string `json:"type"` // "text", "image", etc.
	Text string `json:"text,omitempty"`
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Env:           make(map[string]string),
		Hooks:         make(map[HookEvent][]HookMatcher),
		MaxBufferSize: 1024 * 1024, // 1MB default
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

// WithSystemPromptPreset sets a system prompt preset.
func WithSystemPromptPreset(preset SystemPromptPreset) Option {
	return func(o *Options) {
		o.SystemPrompt = preset
	}
}

// WithAppendSystemPrompt appends to the system prompt.
func WithAppendSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.AppendSystemPrompt = prompt
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

// WithMaxThinkingTokens sets max thinking tokens.
func WithMaxThinkingTokens(tokens int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = tokens
	}
}

// WithTools specifies which tools to enable.
func WithTools(tools ...string) Option {
	return func(o *Options) {
		o.Tools = tools
	}
}

// WithAllowedTools specifies allowed tools.
func WithAllowedTools(tools ...string) Option {
	return func(o *Options) {
		o.AllowedTools = tools
	}
}

// WithDisallowedTools specifies disallowed tools.
func WithDisallowedTools(tools ...string) Option {
	return func(o *Options) {
		o.DisallowedTools = tools
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

// WithForkSession forks the session.
func WithForkSession() Option {
	return func(o *Options) {
		o.ForkSession = true
	}
}

// WithFileCheckpointing enables file checkpointing.
func WithFileCheckpointing() Option {
	return func(o *Options) {
		o.EnableFileCheckpointing = true
	}
}

// WithPartialMessages enables partial message streaming.
func WithPartialMessages() Option {
	return func(o *Options) {
		o.IncludePartialMessages = true
	}
}

// WithOutputFormat sets structured output format.
func WithOutputFormat(format map[string]any) Option {
	return func(o *Options) {
		o.OutputFormat = format
	}
}

// WithBetas enables beta features.
func WithBetas(betas ...SdkBeta) Option {
	return func(o *Options) {
		o.Betas = betas
	}
}

// WithSettingSources sets the setting sources.
func WithSettingSources(sources ...SettingSource) Option {
	return func(o *Options) {
		o.SettingSources = sources
	}
}

// WithSettings sets the settings path or JSON.
func WithSettings(settings string) Option {
	return func(o *Options) {
		o.Settings = settings
	}
}

// WithSandbox enables sandbox mode.
func WithSandbox(sandbox SandboxSettings) Option {
	return func(o *Options) {
		o.Sandbox = &sandbox
	}
}

// WithAgents sets custom agent definitions.
func WithAgents(agents map[string]AgentDefinition) Option {
	return func(o *Options) {
		o.Agents = agents
	}
}

// WithPlugins sets plugin configurations.
func WithPlugins(plugins ...PluginConfig) Option {
	return func(o *Options) {
		o.Plugins = plugins
	}
}

// WithHook adds a hook for an event.
func WithHook(event HookEvent, matcher HookMatcher) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[event] = append(o.Hooks[event], matcher)
	}
}

// WithCanUseTool sets the tool permission callback.
func WithCanUseTool(callback CanUseToolCallback) Option {
	return func(o *Options) {
		o.CanUseTool = callback
	}
}

// WithStderrCallback sets the stderr callback.
func WithStderrCallback(callback func(string)) Option {
	return func(o *Options) {
		o.StderrCallback = callback
	}
}

// WithMaxBufferSize sets the max buffer size for stdout.
func WithMaxBufferSize(size int) Option {
	return func(o *Options) {
		o.MaxBufferSize = size
	}
}

// WithMCPServers sets external MCP server configurations.
func WithMCPServers(servers map[string]MCPServerConfig) Option {
	return func(o *Options) {
		o.MCPServers = servers
	}
}

// WithSDKMCPServer adds an in-process MCP server.
func WithSDKMCPServer(name string, server *MCPServer) Option {
	return func(o *Options) {
		if o.SDKMCPServers == nil {
			o.SDKMCPServers = make(map[string]*MCPServer)
		}
		o.SDKMCPServers[name] = server
	}
}

// WithTransport sets a custom transport (for testing).
func WithTransport(t Transport) Option {
	return func(o *Options) {
		o.customTransport = t
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
git commit -m "feat: add complete options types with functional options pattern"
```

---

## Task 6: Define Hook Types (Complete)

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
			HookEventName:  "PreToolUse",
		},
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls"},
	}

	if input.SessionID != "sess_123" {
		t.Errorf("got %q, want %q", input.SessionID, "sess_123")
	}
	if input.HookEventName != "PreToolUse" {
		t.Errorf("got %q, want %q", input.HookEventName, "PreToolUse")
	}
}

func TestHookOutput(t *testing.T) {
	output := &SyncHookOutput{
		Continue:       boolPtr(true),
		SuppressOutput: false,
		Decision:       "allow",
		HookSpecificOutput: map[string]any{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
		},
	}
	if output.Continue == nil || !*output.Continue {
		t.Error("expected Continue to be true")
	}
}

func TestAsyncHookOutput(t *testing.T) {
	output := &AsyncHookOutput{
		Async:        true,
		AsyncTimeout: intPtr(30),
	}
	if !output.Async {
		t.Error("expected Async to be true")
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
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
	HookPreToolUse       HookEvent = "PreToolUse"
	HookPostToolUse      HookEvent = "PostToolUse"
	HookUserPromptSubmit HookEvent = "UserPromptSubmit"
	HookStop             HookEvent = "Stop"
	HookSubagentStop     HookEvent = "SubagentStop"
	HookPreCompact       HookEvent = "PreCompact"
)

// HookContext provides context for hook callbacks.
type HookContext struct {
	Signal any // Future: abort signal support
}

// BaseHookInput contains fields common to all hook inputs.
type BaseHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode,omitempty"`
	HookEventName  string `json:"hook_event_name"` // Discriminator field for type safety
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

// AsyncHookOutput indicates the hook runs asynchronously.
type AsyncHookOutput struct {
	Async        bool `json:"async"`
	AsyncTimeout *int `json:"asyncTimeout,omitempty"`
}

// SyncHookOutput is the output from a synchronous hook.
type SyncHookOutput struct {
	Continue           *bool          `json:"continue,omitempty"`
	SuppressOutput     bool           `json:"suppressOutput,omitempty"`
	StopReason         string         `json:"stopReason,omitempty"`
	Decision           string         `json:"decision,omitempty"` // "block", "allow"
	SystemMessage      string         `json:"systemMessage,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	HookSpecificOutput map[string]any `json:"hookSpecificOutput,omitempty"`
}

// HookOutput is the combined output from a hook callback (sync or async).
type HookOutput struct {
	// For sync hooks
	Continue           bool           `json:"continue,omitempty"`
	SuppressOutput     bool           `json:"suppressOutput,omitempty"`
	StopReason         string         `json:"stopReason,omitempty"`
	Decision           string         `json:"decision,omitempty"`
	SystemMessage      string         `json:"systemMessage,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	HookSpecific       map[string]any `json:"hookSpecificOutput,omitempty"`

	// For async hooks
	Async        bool `json:"async,omitempty"`
	AsyncTimeout *int `json:"asyncTimeout,omitempty"`
}

// PreToolUseHookSpecificOutput is the hook-specific output for PreToolUse.
type PreToolUseHookSpecificOutput struct {
	HookEventName            string         `json:"hookEventName"` // "PreToolUse"
	PermissionDecision       string         `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             map[string]any `json:"updatedInput,omitempty"`
}

// HookCallback is the signature for hook callback functions.
type HookCallback func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// HookMatcher configures which hooks to run for an event.
type HookMatcher struct {
	Matcher  *string        `json:"matcher,omitempty"` // e.g., "Bash" or "Write|Edit"
	Hooks    []HookCallback `json:"-"`                 // Callbacks (not serialized)
	Timeout  *float64       `json:"timeout,omitempty"` // Timeout in seconds
}
```

**Step 4: Run tests**

```bash
go test -run "TestHookEvent|TestHookInput|TestHookOutput|TestAsyncHookOutput" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add complete hook types with async support"
```

---

## Task 7: Define Control Protocol Types (Complete)

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

func TestControlRequestSubtypes(t *testing.T) {
	subtypes := []string{
		"initialize",
		"interrupt",
		"set_permission_mode",
		"set_model",
		"rewind_files",
		"can_use_tool",
		"hook_callback",
		"mcp_message",
	}
	if len(subtypes) != 8 {
		t.Errorf("expected 8 control request subtypes")
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

// ControlErrorResponse is an error response from control operations.
type ControlErrorResponse struct {
	Subtype   string `json:"subtype"` // "error"
	RequestID string `json:"request_id"`
	Error     string `json:"error"`
}

// Specific control request types

// ControlInitializeRequest initializes the session.
type ControlInitializeRequest struct {
	Subtype string         `json:"subtype"` // "initialize"
	Hooks   map[string]any `json:"hooks,omitempty"`
}

// ControlInterruptRequest interrupts the current operation.
type ControlInterruptRequest struct {
	Subtype string `json:"subtype"` // "interrupt"
}

// ControlSetPermissionModeRequest changes the permission mode.
type ControlSetPermissionModeRequest struct {
	Subtype string `json:"subtype"` // "set_permission_mode"
	Mode    string `json:"mode"`
}

// ControlSetModelRequest changes the AI model.
type ControlSetModelRequest struct {
	Subtype string  `json:"subtype"` // "set_model"
	Model   *string `json:"model"`   // nil to reset
}

// ControlRewindFilesRequest rewinds files to a checkpoint.
type ControlRewindFilesRequest struct {
	Subtype       string `json:"subtype"` // "rewind_files"
	UserMessageID string `json:"user_message_id"`
}

// ControlCanUseToolRequest asks for tool permission.
type ControlCanUseToolRequest struct {
	Subtype               string             `json:"subtype"` // "can_use_tool"
	ToolName              string             `json:"tool_name"`
	Input                 map[string]any     `json:"input"`
	PermissionSuggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`
	BlockedPath           *string            `json:"blocked_path,omitempty"`
}

// ControlHookCallbackRequest invokes a hook callback.
type ControlHookCallbackRequest struct {
	Subtype    string `json:"subtype"` // "hook_callback"
	CallbackID string `json:"callback_id"`
	Input      any    `json:"input"`
	ToolUseID  string `json:"tool_use_id,omitempty"`
}

// ControlMCPMessageRequest routes an MCP message.
type ControlMCPMessageRequest struct {
	Subtype    string         `json:"subtype"` // "mcp_message"
	ServerName string         `json:"server_name"`
	Message    map[string]any `json:"message"` // JSONRPC message
}

// PermissionResult is the interface for permission results.
type PermissionResult interface {
	isPermissionResult()
}

// PermissionResultAllow allows a tool to run.
type PermissionResultAllow struct {
	Behavior           string             `json:"behavior"` // "allow"
	UpdatedInput       map[string]any     `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate `json:"updatedPermissions,omitempty"`
}

func (r *PermissionResultAllow) isPermissionResult() {}

// PermissionResultDeny denies a tool from running.
type PermissionResultDeny struct {
	Behavior  string `json:"behavior"` // "deny"
	Message   string `json:"message,omitempty"`
	Interrupt bool   `json:"interrupt,omitempty"`
}

func (r *PermissionResultDeny) isPermissionResult() {}

// PermissionUpdateType represents the type of permission update.
type PermissionUpdateType string

const (
	PermissionAddRules         PermissionUpdateType = "addRules"
	PermissionReplaceRules     PermissionUpdateType = "replaceRules"
	PermissionRemoveRules      PermissionUpdateType = "removeRules"
	PermissionSetMode          PermissionUpdateType = "setMode"
	PermissionAddDirectories   PermissionUpdateType = "addDirectories"
	PermissionRemoveDirectories PermissionUpdateType = "removeDirectories"
)

// PermissionUpdateDestination represents where to apply the update.
type PermissionUpdateDestination string

const (
	DestinationUserSettings    PermissionUpdateDestination = "userSettings"
	DestinationProjectSettings PermissionUpdateDestination = "projectSettings"
	DestinationLocalSettings   PermissionUpdateDestination = "localSettings"
	DestinationSession         PermissionUpdateDestination = "session"
)

// PermissionUpdate describes a permission change.
type PermissionUpdate struct {
	Type        PermissionUpdateType        `json:"type"`
	Rules       []PermissionRule            `json:"rules,omitempty"`
	Behavior    string                      `json:"behavior,omitempty"` // "allow", "deny", "ask"
	Mode        string                      `json:"mode,omitempty"`
	Directories []string                    `json:"directories,omitempty"`
	Destination PermissionUpdateDestination `json:"destination,omitempty"`
}

// ToDict converts PermissionUpdate to a map for control protocol.
func (p *PermissionUpdate) ToDict() map[string]any {
	result := map[string]any{
		"type": p.Type,
	}
	if len(p.Rules) > 0 {
		rules := make([]map[string]any, len(p.Rules))
		for i, r := range p.Rules {
			rules[i] = map[string]any{"toolName": r.ToolName}
			if r.RuleContent != nil {
				rules[i]["ruleContent"] = *r.RuleContent
			}
		}
		result["rules"] = rules
	}
	if p.Behavior != "" {
		result["behavior"] = p.Behavior
	}
	if p.Mode != "" {
		result["mode"] = p.Mode
	}
	if len(p.Directories) > 0 {
		result["directories"] = p.Directories
	}
	if p.Destination != "" {
		result["destination"] = p.Destination
	}
	return result
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
	BlockedPath *string            `json:"blocked_path,omitempty"`
}

// CanUseToolCallback is called when a tool needs permission.
type CanUseToolCallback func(
	toolName string,
	input map[string]any,
	ctx *ToolPermissionContext,
) (PermissionResult, error)
```

**Step 4: Run tests**

```bash
go test -run "TestControlRequest|TestControlResponse|TestPermissionResult|TestControlRequestSubtypes" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add complete control protocol types"
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
- [x] Error types with sentinel errors (SDKError, CLINotFoundError, ConnectionError, ProcessError, JSONDecodeError, MessageParseError)
- [x] Content block types (Text, Thinking, ToolUse, ToolResult)
- [x] Message types (User, Assistant, System, Result, StreamEvent) with AssistantMessageError enum
- [x] Complete Options with:
  - All 50+ configuration fields matching Python
  - Functional options pattern
  - Sandbox configuration
  - Agent definitions
  - Plugin configuration
  - MCP server configuration
- [x] Complete Hook types:
  - All 6 hook events
  - Async and sync hook outputs
  - Hook-specific outputs
  - HookContext with abort signal support
- [x] Complete Control protocol types:
  - All 8 control request subtypes
  - Permission results (Allow/Deny)
  - Permission updates with ToDict()
  - Tool permission context
- [x] Transport interface

**Next:** Plan 02 - Transport Layer (subprocess implementation)

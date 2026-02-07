// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"fmt"
)

// ContentBlock represents a block of content in a message.
type ContentBlock interface {
	BlockType() string
	Type() string // Alias for BlockType for compatibility
}

// TextBlock contains text content.
type TextBlock struct {
	TextContent string `json:"text"`
}

func (b *TextBlock) BlockType() string { return "text" }
func (b *TextBlock) Type() string      { return "text" }

// Text returns the text content.
func (b *TextBlock) Text() string { return b.TextContent }

// ThinkingBlock contains Claude's thinking content.
type ThinkingBlock struct {
	ThinkingContent string `json:"thinking"`
	Signature       string `json:"signature"`
}

func (b *ThinkingBlock) BlockType() string { return "thinking" }
func (b *ThinkingBlock) Type() string      { return "thinking" }

// Thinking returns the thinking content.
func (b *ThinkingBlock) Thinking() string { return b.ThinkingContent }

// ToolUseBlock represents a tool invocation.
type ToolUseBlock struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	ToolInput map[string]any `json:"input"`
}

func (b *ToolUseBlock) BlockType() string { return "tool_use" }
func (b *ToolUseBlock) Type() string      { return "tool_use" }

// Input returns the tool input.
func (b *ToolUseBlock) Input() map[string]any { return b.ToolInput }

// ToolResultBlock contains the result of a tool execution.
type ToolResultBlock struct {
	ToolUseID     string `json:"tool_use_id"`
	ResultContent string `json:"content,omitempty"`
	IsError       bool   `json:"is_error,omitempty"`
}

func (b *ToolResultBlock) BlockType() string { return "tool_result" }
func (b *ToolResultBlock) Type() string      { return "tool_result" }

// Content returns the result content as string.
func (b *ToolResultBlock) Content() string { return b.ResultContent }

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
		var block textBlockJSON
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse text block: %w", err)
		}
		return &TextBlock{TextContent: block.Text}, nil

	case "thinking":
		var block thinkingBlockJSON
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse thinking block: %w", err)
		}
		return &ThinkingBlock{ThinkingContent: block.Thinking, Signature: block.Signature}, nil

	case "tool_use":
		var block toolUseBlockJSON
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_use block: %w", err)
		}
		return &ToolUseBlock{ID: block.ID, Name: block.Name, ToolInput: block.Input}, nil

	case "tool_result":
		var block toolResultBlockJSON
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_result block: %w", err)
		}
		return &ToolResultBlock{ToolUseID: block.ToolUseID, ResultContent: block.Content, IsError: block.IsError}, nil

	default:
		return nil, fmt.Errorf("unknown block type: %s", blockType)
	}
}

// JSON unmarshaling helper types
type textBlockJSON struct {
	Text string `json:"text"`
}

type thinkingBlockJSON struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

type toolUseBlockJSON struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

type toolResultBlockJSON struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// Message represents a message in the conversation.
type Message interface {
	MessageType() string
}

// UserMessage represents a user's message.
type UserMessage struct {
	Content         []ContentBlock `json:"content"`
	Role            string         `json:"role,omitempty"`
	UUID            string         `json:"uuid,omitempty"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
	ToolUseResult   map[string]any `json:"tool_use_result,omitempty"`
}

func (m *UserMessage) MessageType() string { return "user" }

// Text returns all text content concatenated from a UserMessage.
func (m *UserMessage) Text() string {
	var result string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			result += textBlock.TextContent
		}
	}
	return result
}

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
	StopReason      string                 `json:"stop_reason,omitempty"`
	ParentToolUseID *string                `json:"parent_tool_use_id,omitempty"`
	Error           *AssistantMessageError `json:"error,omitempty"`
}

func (m *AssistantMessage) MessageType() string { return "assistant" }

// Text returns all text content concatenated from an AssistantMessage.
func (m *AssistantMessage) Text() string {
	var result string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			result += textBlock.TextContent
		}
	}
	return result
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

// GetThinking returns the thinking content if present.
func (m *AssistantMessage) GetThinking() string {
	for _, block := range m.Content {
		if thinkingBlock, ok := block.(*ThinkingBlock); ok {
			return thinkingBlock.ThinkingContent
		}
	}
	return ""
}

// Thinking returns the thinking content if present (alias for GetThinking).
func (m *AssistantMessage) Thinking() string {
	return m.GetThinking()
}

// HasToolCalls returns true if the message contains tool calls.
func (m *AssistantMessage) HasToolCalls() bool {
	for _, block := range m.Content {
		if _, ok := block.(*ToolUseBlock); ok {
			return true
		}
	}
	return false
}

// SystemMessage represents a system message.
type SystemMessage struct {
	Subtype   string         `json:"subtype"`
	SessionID string         `json:"session_id,omitempty"`
	Version   string         `json:"version,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
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
	EventType       string         `json:"event_type,omitempty"`
	Index           *int           `json:"index,omitempty"`
	Delta           map[string]any `json:"delta,omitempty"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
}

func (m *StreamEvent) MessageType() string { return "stream_event" }

// AuthStatusMessage reports authentication progress from the CLI.
type AuthStatusMessage struct {
	IsAuthenticating bool     `json:"isAuthenticating"`
	Output           []string `json:"output,omitempty"`
	Error            string   `json:"error,omitempty"`
	UUID             string   `json:"uuid,omitempty"`
	SessionID        string   `json:"session_id,omitempty"`
}

func (m *AuthStatusMessage) MessageType() string { return "auth_status" }

// ToolProgressMessage reports progress for a long-running tool call.
type ToolProgressMessage struct {
	ToolUseID          string  `json:"tool_use_id"`
	ToolName           string  `json:"tool_name"`
	ParentToolUseID    *string `json:"parent_tool_use_id,omitempty"`
	ElapsedTimeSeconds float64 `json:"elapsed_time_seconds,omitempty"`
	UUID               string  `json:"uuid,omitempty"`
	SessionID          string  `json:"session_id,omitempty"`
}

func (m *ToolProgressMessage) MessageType() string { return "tool_progress" }

// ToolUseSummaryMessage reports condensed tool usage summaries.
type ToolUseSummaryMessage struct {
	Summary             string   `json:"summary"`
	PrecedingToolUseIDs []string `json:"preceding_tool_use_ids,omitempty"`
	UUID                string   `json:"uuid,omitempty"`
	SessionID           string   `json:"session_id,omitempty"`
}

func (m *ToolUseSummaryMessage) MessageType() string { return "tool_use_summary" }

// TaskNotificationMessage reports background task completion/failure.
type TaskNotificationMessage struct {
	Subtype    string `json:"subtype"`
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	OutputFile string `json:"output_file,omitempty"`
	Summary    string `json:"summary,omitempty"`
	UUID       string `json:"uuid,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
}

func (m *TaskNotificationMessage) MessageType() string { return "system" }

// FilesPersistedFile describes a successfully persisted file.
type FilesPersistedFile struct {
	Filename string `json:"filename"`
	FileID   string `json:"file_id"`
}

// FilesPersistedFailure describes a file that failed to persist.
type FilesPersistedFailure struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// FilesPersistedMessage reports persisted-file snapshots.
type FilesPersistedMessage struct {
	Subtype     string                  `json:"subtype"`
	Files       []FilesPersistedFile    `json:"files,omitempty"`
	Failed      []FilesPersistedFailure `json:"failed,omitempty"`
	ProcessedAt string                  `json:"processed_at,omitempty"`
	UUID        string                  `json:"uuid,omitempty"`
	SessionID   string                  `json:"session_id,omitempty"`
}

func (m *FilesPersistedMessage) MessageType() string { return "system" }

// HookStartedMessage reports when a hook invocation starts.
type HookStartedMessage struct {
	Subtype   string `json:"subtype"`
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	UUID      string `json:"uuid,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (m *HookStartedMessage) MessageType() string { return "system" }

// HookProgressMessage reports streaming output from a running hook.
type HookProgressMessage struct {
	Subtype   string `json:"subtype"`
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	Output    string `json:"output,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (m *HookProgressMessage) MessageType() string { return "system" }

// HookResponseMessage reports completion status for a hook invocation.
type HookResponseMessage struct {
	Subtype   string `json:"subtype"`
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	Output    string `json:"output,omitempty"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	ExitCode  *int   `json:"exit_code,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (m *HookResponseMessage) MessageType() string { return "system" }

// HookEvent represents the type of hook event.
type HookEvent string

const (
	// HookPreToolUse is triggered before a tool is invoked. This allows inspection,
	// modification, or blocking of tool calls before they execute.
	HookPreToolUse HookEvent = "PreToolUse"

	// HookPostToolUse is triggered after a tool has been invoked. This allows
	// inspection or modification of tool results before they are returned to Claude.
	HookPostToolUse HookEvent = "PostToolUse"

	// HookPostToolUseFailure is triggered when a tool invocation fails.
	HookPostToolUseFailure HookEvent = "PostToolUseFailure"

	// HookUserPromptSubmit is triggered when a user submits a prompt. This allows
	// inspection or modification of user input before it is sent to Claude.
	HookUserPromptSubmit HookEvent = "UserPromptSubmit"

	// HookStop is triggered when a session is about to stop. This allows cleanup
	// or logging before the session ends.
	HookStop HookEvent = "Stop"

	// HookSubagentStop is triggered when a subagent session is about to stop.
	// This allows cleanup specific to subagent termination.
	HookSubagentStop HookEvent = "SubagentStop"

	// HookPreCompact is triggered before the conversation history is compacted.
	// This allows inspection or archival of messages before they are removed.
	HookPreCompact HookEvent = "PreCompact"

	// HookNotification is triggered for CLI/user-facing notifications.
	HookNotification HookEvent = "Notification"

	// HookSubagentStart is triggered when a subagent starts.
	HookSubagentStart HookEvent = "SubagentStart"

	// HookPermissionRequest is triggered for permission request hook events.
	HookPermissionRequest HookEvent = "PermissionRequest"

	// HookSessionStart is triggered when a session starts.
	HookSessionStart HookEvent = "SessionStart"

	// HookSessionEnd is triggered when a session ends.
	HookSessionEnd HookEvent = "SessionEnd"

	// HookSetup is triggered during setup operations.
	HookSetup HookEvent = "Setup"

	// HookTeammateIdle is triggered when a teammate agent becomes idle.
	HookTeammateIdle HookEvent = "TeammateIdle"

	// HookTaskCompleted is triggered when a task completes.
	HookTaskCompleted HookEvent = "TaskCompleted"
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
	HookEventName  string `json:"hook_event_name"`
}

// PreToolUseHookInput is the input for PreToolUse hooks.
type PreToolUseHookInput struct {
	BaseHookInput
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
	ToolUseID string         `json:"tool_use_id"`
}

// PostToolUseHookInput is the input for PostToolUse hooks.
type PostToolUseHookInput struct {
	BaseHookInput
	ToolName     string         `json:"tool_name"`
	ToolInput    map[string]any `json:"tool_input"`
	ToolResponse any            `json:"tool_response"`
	ToolUseID    string         `json:"tool_use_id"`
}

// PostToolUseFailureHookInput is the input for PostToolUseFailure hooks.
type PostToolUseFailureHookInput struct {
	BaseHookInput
	ToolName    string         `json:"tool_name"`
	ToolInput   map[string]any `json:"tool_input"`
	ToolUseID   string         `json:"tool_use_id"`
	Error       string         `json:"error"`
	IsInterrupt *bool          `json:"is_interrupt,omitempty"`
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
	StopHookActive      bool   `json:"stop_hook_active"`
	AgentID             string `json:"agent_id"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
	AgentType           string `json:"agent_type"`
}

// PreCompactHookInput is the input for PreCompact hooks.
type PreCompactHookInput struct {
	BaseHookInput
	Trigger            string  `json:"trigger"`
	CustomInstructions *string `json:"custom_instructions,omitempty"`
}

// NotificationHookInput is the input for Notification hooks.
type NotificationHookInput struct {
	BaseHookInput
	Message          string  `json:"message"`
	Title            *string `json:"title,omitempty"`
	NotificationType string  `json:"notification_type"`
}

// SubagentStartHookInput is the input for SubagentStart hooks.
type SubagentStartHookInput struct {
	BaseHookInput
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
}

// PermissionRequestHookInput is the input for PermissionRequest hooks.
type PermissionRequestHookInput struct {
	BaseHookInput
	ToolName              string         `json:"tool_name"`
	ToolInput             map[string]any `json:"tool_input"`
	PermissionSuggestions []any          `json:"permission_suggestions,omitempty"`
}

// SessionStartHookInput is the input for SessionStart hooks.
type SessionStartHookInput struct {
	BaseHookInput
	Source    string  `json:"source"`
	AgentType *string `json:"agent_type,omitempty"`
	Model     *string `json:"model,omitempty"`
}

// SessionEndHookInput is the input for SessionEnd hooks.
type SessionEndHookInput struct {
	BaseHookInput
	Reason string `json:"reason"`
}

// SetupHookInput is the input for Setup hooks.
type SetupHookInput struct {
	BaseHookInput
	Trigger string `json:"trigger"`
}

// TeammateIdleHookInput is the input for TeammateIdle hooks.
type TeammateIdleHookInput struct {
	BaseHookInput
	TeammateName string `json:"teammate_name"`
	TeamName     string `json:"team_name"`
}

// TaskCompletedHookInput is the input for TaskCompleted hooks.
type TaskCompletedHookInput struct {
	BaseHookInput
	TaskID          string  `json:"task_id"`
	TaskSubject     string  `json:"task_subject"`
	TaskDescription *string `json:"task_description,omitempty"`
	TeammateName    *string `json:"teammate_name,omitempty"`
	TeamName        *string `json:"team_name,omitempty"`
}

// HookOutput is the output from a hook callback.
type HookOutput struct {
	Continue       *bool          `json:"continue,omitempty"`
	SuppressOutput bool           `json:"suppressOutput,omitempty"`
	StopReason     string         `json:"stopReason,omitempty"`
	Decision       string         `json:"decision,omitempty"`
	SystemMessage  string         `json:"systemMessage,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	HookSpecific   map[string]any `json:"hookSpecificOutput,omitempty"`
	Async          bool           `json:"async,omitempty"`
	AsyncTimeout   *int           `json:"asyncTimeout,omitempty"`
}

// HookCallback is the signature for hook callback functions.
type HookCallback func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// Type-safe hook callback signatures for each hook event type.
// These eliminate the need for type assertions in hook implementations.

// PreToolUseCallback is a type-safe callback for PreToolUse hooks.
type PreToolUseCallback func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// PostToolUseCallback is a type-safe callback for PostToolUse hooks.
type PostToolUseCallback func(input *PostToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// PostToolUseFailureCallback is a type-safe callback for PostToolUseFailure hooks.
type PostToolUseFailureCallback func(input *PostToolUseFailureHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// UserPromptSubmitCallback is a type-safe callback for UserPromptSubmit hooks.
type UserPromptSubmitCallback func(input *UserPromptSubmitHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// StopCallback is a type-safe callback for Stop hooks.
type StopCallback func(input *StopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// SubagentStopCallback is a type-safe callback for SubagentStop hooks.
type SubagentStopCallback func(input *SubagentStopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// PreCompactCallback is a type-safe callback for PreCompact hooks.
type PreCompactCallback func(input *PreCompactHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// NotificationCallback is a type-safe callback for Notification hooks.
type NotificationCallback func(input *NotificationHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// SubagentStartCallback is a type-safe callback for SubagentStart hooks.
type SubagentStartCallback func(input *SubagentStartHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// PermissionRequestCallback is a type-safe callback for PermissionRequest hooks.
type PermissionRequestCallback func(input *PermissionRequestHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// SessionStartCallback is a type-safe callback for SessionStart hooks.
type SessionStartCallback func(input *SessionStartHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// SessionEndCallback is a type-safe callback for SessionEnd hooks.
type SessionEndCallback func(input *SessionEndHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// SetupCallback is a type-safe callback for Setup hooks.
type SetupCallback func(input *SetupHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// TeammateIdleCallback is a type-safe callback for TeammateIdle hooks.
type TeammateIdleCallback func(input *TeammateIdleHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// TaskCompletedCallback is a type-safe callback for TaskCompleted hooks.
type TaskCompletedCallback func(input *TaskCompletedHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// ToGenericCallback converts a type-safe callback to a generic HookCallback.
// This allows using type-safe callbacks with the existing hook infrastructure.
func ToGenericCallback[T any](callback func(*T, *string, *HookContext) (*HookOutput, error)) HookCallback {
	return func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		if typedInput, ok := input.(*T); ok {
			return callback(typedInput, toolUseID, ctx)
		}
		if typedValue, ok := input.(T); ok {
			v := typedValue
			return callback(&v, toolUseID, ctx)
		}
		if rawMap, ok := input.(map[string]any); ok {
			data, err := json.Marshal(rawMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal hook input map: %w", err)
			}
			var typed T
			if err := json.Unmarshal(data, &typed); err != nil {
				return nil, fmt.Errorf("failed to parse hook input as typed payload: %w", err)
			}
			return callback(&typed, toolUseID, ctx)
		}
		return nil, fmt.Errorf("invalid input type: expected *%T/%T/map[string]any, got %T", new(T), *new(T), input)
	}
}

// HookMatcher configures which hooks to run for an event.
type HookMatcher struct {
	Matcher map[string]any `json:"matcher,omitempty"` // e.g., {"tool_name": "Bash"} or nil for all
	Hooks   []HookCallback `json:"-"`
	Timeout *float64       `json:"timeout,omitempty"`
}

// HookBuilder provides a fluent API for building hooks.
// This unifies the three different ways to register hooks into a single, discoverable pattern.
//
// Example usage:
//
//	hook := NewHookBuilder().
//	    ForEvent(types.HookPreToolUse).
//	    WithMatcher(map[string]any{"tool_name": "Bash"}).
//	    WithCallback(func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
//	        // Your hook logic here
//	        return &HookOutput{Continue: boolPtr(true)}, nil
//	    }).
//	    WithTimeout(5.0).
//	    Build()
type HookBuilder struct {
	event     HookEvent
	matcher   map[string]any
	callbacks []HookCallback
	timeout   *float64
}

// NewHookBuilder creates a new HookBuilder.
func NewHookBuilder() *HookBuilder {
	return &HookBuilder{
		callbacks: make([]HookCallback, 0),
	}
}

// ForEvent sets the hook event type.
func (b *HookBuilder) ForEvent(event HookEvent) *HookBuilder {
	b.event = event
	return b
}

// WithMatcher sets the matcher for the hook.
// The matcher filters which tool calls this hook applies to.
// For example: map[string]any{"tool_name": "Bash"} only matches Bash tool calls.
func (b *HookBuilder) WithMatcher(matcher map[string]any) *HookBuilder {
	b.matcher = matcher
	return b
}

// MatchAll configures the hook to match all tool calls (no filtering).
func (b *HookBuilder) MatchAll() *HookBuilder {
	b.matcher = nil
	return b
}

// WithCallback adds a type-safe callback to the hook.
// The callback type should match the event type (e.g., PreToolUseCallback for HookPreToolUse).
// The callback is automatically converted to a generic HookCallback.
func (b *HookBuilder) WithCallback(callback any) *HookBuilder {
	// Convert type-safe callback to generic HookCallback
	var genericCallback HookCallback

	switch cb := callback.(type) {
	case PreToolUseCallback:
		genericCallback = ToGenericCallback[PreToolUseHookInput](cb)
	case PostToolUseCallback:
		genericCallback = ToGenericCallback[PostToolUseHookInput](cb)
	case PostToolUseFailureCallback:
		genericCallback = ToGenericCallback[PostToolUseFailureHookInput](cb)
	case UserPromptSubmitCallback:
		genericCallback = ToGenericCallback[UserPromptSubmitHookInput](cb)
	case StopCallback:
		genericCallback = ToGenericCallback[StopHookInput](cb)
	case SubagentStopCallback:
		genericCallback = ToGenericCallback[SubagentStopHookInput](cb)
	case PreCompactCallback:
		genericCallback = ToGenericCallback[PreCompactHookInput](cb)
	case NotificationCallback:
		genericCallback = ToGenericCallback[NotificationHookInput](cb)
	case SubagentStartCallback:
		genericCallback = ToGenericCallback[SubagentStartHookInput](cb)
	case PermissionRequestCallback:
		genericCallback = ToGenericCallback[PermissionRequestHookInput](cb)
	case SessionStartCallback:
		genericCallback = ToGenericCallback[SessionStartHookInput](cb)
	case SessionEndCallback:
		genericCallback = ToGenericCallback[SessionEndHookInput](cb)
	case SetupCallback:
		genericCallback = ToGenericCallback[SetupHookInput](cb)
	case TeammateIdleCallback:
		genericCallback = ToGenericCallback[TeammateIdleHookInput](cb)
	case TaskCompletedCallback:
		genericCallback = ToGenericCallback[TaskCompletedHookInput](cb)
	case HookCallback:
		// Already a generic callback
		genericCallback = cb
	default:
		// Try to convert using reflection-based approach for function types
		// This handles inline function definitions
		if fn, ok := callback.(func(*PreToolUseHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[PreToolUseHookInput](fn)
		} else if fn, ok := callback.(func(*PostToolUseHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[PostToolUseHookInput](fn)
		} else if fn, ok := callback.(func(*PostToolUseFailureHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[PostToolUseFailureHookInput](fn)
		} else if fn, ok := callback.(func(*UserPromptSubmitHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[UserPromptSubmitHookInput](fn)
		} else if fn, ok := callback.(func(*StopHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[StopHookInput](fn)
		} else if fn, ok := callback.(func(*SubagentStopHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[SubagentStopHookInput](fn)
		} else if fn, ok := callback.(func(*PreCompactHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[PreCompactHookInput](fn)
		} else if fn, ok := callback.(func(*NotificationHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[NotificationHookInput](fn)
		} else if fn, ok := callback.(func(*SubagentStartHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[SubagentStartHookInput](fn)
		} else if fn, ok := callback.(func(*PermissionRequestHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[PermissionRequestHookInput](fn)
		} else if fn, ok := callback.(func(*SessionStartHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[SessionStartHookInput](fn)
		} else if fn, ok := callback.(func(*SessionEndHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[SessionEndHookInput](fn)
		} else if fn, ok := callback.(func(*SetupHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[SetupHookInput](fn)
		} else if fn, ok := callback.(func(*TeammateIdleHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[TeammateIdleHookInput](fn)
		} else if fn, ok := callback.(func(*TaskCompletedHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[TaskCompletedHookInput](fn)
		} else {
			panic(fmt.Sprintf("unsupported callback type: %T", callback))
		}
	}

	b.callbacks = append(b.callbacks, genericCallback)
	return b
}

// WithGenericCallback adds a generic HookCallback directly.
// This is useful for backward compatibility with existing code.
func (b *HookBuilder) WithGenericCallback(callback HookCallback) *HookBuilder {
	b.callbacks = append(b.callbacks, callback)
	return b
}

// WithTimeout sets the timeout for the hook in seconds.
func (b *HookBuilder) WithTimeout(timeout float64) *HookBuilder {
	b.timeout = &timeout
	return b
}

// Build creates the HookMatcher from the builder configuration.
func (b *HookBuilder) Build() HookMatcher {
	return HookMatcher{
		Matcher: b.matcher,
		Hooks:   b.callbacks,
		Timeout: b.timeout,
	}
}

// ToOption converts the builder to an Option that can be passed to NewClient.
// This allows using HookBuilder with the existing options pattern.
func (b *HookBuilder) ToOption() Option {
	matcher := b.Build()
	event := b.event

	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[event] = append(o.Hooks[event], matcher)
	}
}

// BuildForOptions builds the hook and directly adds it to the given Options.
// This is a convenience method that combines Build() and adding to Options.
func (b *HookBuilder) BuildForOptions(opts *Options) {
	if opts.Hooks == nil {
		opts.Hooks = make(map[HookEvent][]HookMatcher)
	}
	opts.Hooks[b.event] = append(opts.Hooks[b.event], b.Build())
}

// ControlRequest is sent to the CLI for control operations.
type ControlRequest struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Request   map[string]any `json:"request"`
}

// ControlResponseData contains the response data.
type ControlResponseData struct {
	Subtype   string         `json:"subtype"`
	RequestID string         `json:"request_id"`
	Response  map[string]any `json:"response,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// ControlResponse is received from the CLI for control operations.
type ControlResponse struct {
	Type     string              `json:"type"`
	Response ControlResponseData `json:"response"`
}

// SDKControlRequest is the interface for control request subtypes.
// Each control request type implements this interface via ControlRequestType().
type SDKControlRequest interface {
	ControlRequestType() string
}

// SDKControlInterruptRequest requests an interrupt.
type SDKControlInterruptRequest struct {
	Subtype string `json:"subtype"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlInterruptRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlPermissionRequest requests permission for a tool.
type SDKControlPermissionRequest struct {
	Subtype               string             `json:"subtype"`
	ToolName              string             `json:"tool_name"`
	Input                 map[string]any     `json:"input"`
	PermissionSuggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`
	BlockedPath           *string            `json:"blocked_path,omitempty"`
	DecisionReason        *string            `json:"decision_reason,omitempty"`
	ToolUseID             string             `json:"tool_use_id,omitempty"`
	AgentID               *string            `json:"agent_id,omitempty"`
	Description           *string            `json:"description,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlPermissionRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlInitializeRequest initializes the SDK session.
type SDKControlInitializeRequest struct {
	Subtype string                    `json:"subtype"`
	Hooks   map[HookEvent]any         `json:"hooks,omitempty"`
	Agents  map[string]map[string]any `json:"agents,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlInitializeRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlSetPermissionModeRequest sets the permission mode.
type SDKControlSetPermissionModeRequest struct {
	Subtype string `json:"subtype"`
	Mode    string `json:"mode"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlSetPermissionModeRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlSetModelRequest sets the active model.
type SDKControlSetModelRequest struct {
	Subtype string `json:"subtype"`
	Model   string `json:"model,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlSetModelRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlSetMaxThinkingTokensRequest sets or clears max thinking tokens.
type SDKControlSetMaxThinkingTokensRequest struct {
	Subtype           string `json:"subtype"`
	MaxThinkingTokens *int   `json:"max_thinking_tokens"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlSetMaxThinkingTokensRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKHookCallbackRequest invokes a hook callback.
type SDKHookCallbackRequest struct {
	Subtype    string  `json:"subtype"`
	CallbackID string  `json:"callback_id"`
	Input      any     `json:"input,omitempty"`
	ToolUseID  *string `json:"tool_use_id,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKHookCallbackRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlMcpMessageRequest sends a message to an MCP server.
type SDKControlMcpMessageRequest struct {
	Subtype    string `json:"subtype"`
	ServerName string `json:"server_name"`
	Message    any    `json:"message,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlMcpMessageRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlMcpToolCallRequest calls a tool on an MCP server.
type SDKControlMcpToolCallRequest struct {
	Subtype    string         `json:"subtype"`
	ServerName string         `json:"server_name"`
	ToolName   string         `json:"tool_name"`
	Input      map[string]any `json:"input,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlMcpToolCallRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlRewindFilesRequest rewinds files to a previous message.
type SDKControlRewindFilesRequest struct {
	Subtype       string `json:"subtype"`
	UserMessageID string `json:"user_message_id"`
	DryRun        *bool  `json:"dry_run,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlRewindFilesRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlMcpStatusRequest requests MCP server status.
type SDKControlMcpStatusRequest struct {
	Subtype string `json:"subtype"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlMcpStatusRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlMcpSetServersRequest sets dynamic MCP servers.
type SDKControlMcpSetServersRequest struct {
	Subtype string         `json:"subtype"`
	Servers map[string]any `json:"servers"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlMcpSetServersRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlMcpReconnectRequest reconnects an MCP server.
type SDKControlMcpReconnectRequest struct {
	Subtype    string `json:"subtype"`
	ServerName string `json:"serverName"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlMcpReconnectRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlMcpToggleRequest toggles an MCP server enabled state.
type SDKControlMcpToggleRequest struct {
	Subtype    string `json:"subtype"`
	ServerName string `json:"serverName"`
	Enabled    bool   `json:"enabled"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlMcpToggleRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlInitializeResponse contains initialization metadata returned by the CLI.
type SDKControlInitializeResponse struct {
	Commands              []SlashCommand `json:"commands,omitempty"`
	OutputStyle           string         `json:"output_style,omitempty"`
	AvailableOutputStyles []string       `json:"available_output_styles,omitempty"`
	Models                []ModelInfo    `json:"models,omitempty"`
	Account               AccountInfo    `json:"account,omitempty"`
}

// SlashCommand describes an available slash command.
type SlashCommand struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ArgumentHint string `json:"argumentHint"`
}

// ModelInfo describes an available model.
type ModelInfo struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// AccountInfo describes account metadata for the authenticated user.
type AccountInfo struct {
	Email            string `json:"email,omitempty"`
	Organization     string `json:"organization,omitempty"`
	SubscriptionType string `json:"subscriptionType,omitempty"`
	TokenSource      string `json:"tokenSource,omitempty"`
	APIKeySource     string `json:"apiKeySource,omitempty"`
}

// RewindFilesResult contains file rewind operation details.
type RewindFilesResult struct {
	CanRewind    bool     `json:"canRewind"`
	Error        string   `json:"error,omitempty"`
	FilesChanged []string `json:"filesChanged,omitempty"`
	Insertions   int      `json:"insertions,omitempty"`
	Deletions    int      `json:"deletions,omitempty"`
}

// MCPServerStatus describes a server status returned by mcp_status.
type MCPServerStatus struct {
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Scope  string         `json:"scope,omitempty"`
	Error  string         `json:"error,omitempty"`
	Config map[string]any `json:"config,omitempty"`
}

// MCPSetServersResult describes dynamic MCP update results.
type MCPSetServersResult struct {
	Added   []string          `json:"added"`
	Removed []string          `json:"removed"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// ParseSDKControlRequest parses a raw JSON map into a typed SDKControlRequest.
func ParseSDKControlRequest(raw map[string]any) (SDKControlRequest, error) {
	subtype, ok := raw["subtype"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'subtype' field")
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	switch subtype {
	case "interrupt":
		var req SDKControlInterruptRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse interrupt request: %w", err)
		}
		return &req, nil

	case "can_use_tool":
		var req SDKControlPermissionRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse permission request: %w", err)
		}
		return &req, nil

	case "initialize":
		var req SDKControlInitializeRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse initialize request: %w", err)
		}
		return &req, nil

	case "set_permission_mode":
		var req SDKControlSetPermissionModeRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse set permission mode request: %w", err)
		}
		return &req, nil

	case "set_model":
		var req SDKControlSetModelRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse set model request: %w", err)
		}
		return &req, nil

	case "set_max_thinking_tokens":
		var req SDKControlSetMaxThinkingTokensRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse set max thinking tokens request: %w", err)
		}
		return &req, nil

	case "hook_callback":
		var req SDKHookCallbackRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse hook callback request: %w", err)
		}
		return &req, nil

	case "mcp_message":
		var req SDKControlMcpMessageRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse mcp message request: %w", err)
		}
		return &req, nil

	case "mcp_tool_call":
		var req SDKControlMcpToolCallRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse mcp tool call request: %w", err)
		}
		return &req, nil

	case "rewind_files":
		var req SDKControlRewindFilesRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse rewind files request: %w", err)
		}
		return &req, nil

	case "mcp_status":
		var req SDKControlMcpStatusRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse mcp status request: %w", err)
		}
		return &req, nil

	case "mcp_set_servers":
		var req SDKControlMcpSetServersRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse mcp set servers request: %w", err)
		}
		return &req, nil

	case "mcp_reconnect":
		var req SDKControlMcpReconnectRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse mcp reconnect request: %w", err)
		}
		return &req, nil

	case "mcp_toggle":
		var req SDKControlMcpToggleRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to parse mcp toggle request: %w", err)
		}
		return &req, nil

	default:
		return nil, fmt.Errorf("unknown control request subtype: %s", subtype)
	}
}

// PermissionResult is the interface for permission results.
type PermissionResult interface {
	isPermissionResult()
}

// PermissionResultAllow allows a tool to run.
type PermissionResultAllow struct {
	Behavior           string             `json:"behavior"`
	UpdatedInput       map[string]any     `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate `json:"updatedPermissions,omitempty"`
	ToolUseID          *string            `json:"toolUseID,omitempty"`
}

func (r *PermissionResultAllow) isPermissionResult() {}

// PermissionResultDeny denies a tool from running.
type PermissionResultDeny struct {
	Behavior  string  `json:"behavior"`
	Message   string  `json:"message,omitempty"`
	Interrupt bool    `json:"interrupt,omitempty"`
	ToolUseID *string `json:"toolUseID,omitempty"`
}

func (r *PermissionResultDeny) isPermissionResult() {}

// PermissionUpdateType represents the type of permission update.
type PermissionUpdateType string

const (
	PermissionAddRules          PermissionUpdateType = "addRules"
	PermissionReplaceRules      PermissionUpdateType = "replaceRules"
	PermissionRemoveRules       PermissionUpdateType = "removeRules"
	PermissionSetMode           PermissionUpdateType = "setMode"
	PermissionAddDirectories    PermissionUpdateType = "addDirectories"
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

// PermissionRule defines a permission rule.
type PermissionRule struct {
	ToolName    string  `json:"toolName"`
	RuleContent *string `json:"ruleContent,omitempty"`
}

// PermissionUpdate describes a permission change.
type PermissionUpdate struct {
	Type        PermissionUpdateType        `json:"type"`
	Rules       []PermissionRule            `json:"rules,omitempty"`
	Behavior    string                      `json:"behavior,omitempty"`
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

// ToolPermissionContext provides context for permission callbacks.
type ToolPermissionContext struct {
	Signal         any                `json:"-"`
	Suggestions    []PermissionUpdate `json:"suggestions,omitempty"`
	BlockedPath    *string            `json:"blocked_path,omitempty"`
	DecisionReason *string            `json:"decision_reason,omitempty"`
	ToolUseID      string             `json:"tool_use_id,omitempty"`
	AgentID        *string            `json:"agent_id,omitempty"`
	Description    *string            `json:"description,omitempty"`
}

// CanUseToolCallback is called when a tool needs permission.
type CanUseToolCallback func(
	toolName string,
	input map[string]any,
	ctx *ToolPermissionContext,
) (PermissionResult, error)

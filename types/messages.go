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

// HookEvent represents the type of hook event.
type HookEvent string

const (
	// HookPreToolUse is triggered before a tool is invoked. This allows inspection,
	// modification, or blocking of tool calls before they execute.
	HookPreToolUse HookEvent = "PreToolUse"

	// HookPostToolUse is triggered after a tool has been invoked. This allows
	// inspection or modification of tool results before they are returned to Claude.
	HookPostToolUse HookEvent = "PostToolUse"

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
	Trigger            string  `json:"trigger"`
	CustomInstructions *string `json:"custom_instructions,omitempty"`
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

// UserPromptSubmitCallback is a type-safe callback for UserPromptSubmit hooks.
type UserPromptSubmitCallback func(input *UserPromptSubmitHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// StopCallback is a type-safe callback for Stop hooks.
type StopCallback func(input *StopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// SubagentStopCallback is a type-safe callback for SubagentStop hooks.
type SubagentStopCallback func(input *SubagentStopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// PreCompactCallback is a type-safe callback for PreCompact hooks.
type PreCompactCallback func(input *PreCompactHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error)

// ToGenericCallback converts a type-safe callback to a generic HookCallback.
// This allows using type-safe callbacks with the existing hook infrastructure.
func ToGenericCallback[T any](callback func(*T, *string, *HookContext) (*HookOutput, error)) HookCallback {
	return func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		typedInput, ok := input.(*T)
		if !ok {
			return nil, fmt.Errorf("invalid input type: expected *%T, got %T", new(T), input)
		}
		return callback(typedInput, toolUseID, ctx)
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
	case UserPromptSubmitCallback:
		genericCallback = ToGenericCallback[UserPromptSubmitHookInput](cb)
	case StopCallback:
		genericCallback = ToGenericCallback[StopHookInput](cb)
	case SubagentStopCallback:
		genericCallback = ToGenericCallback[SubagentStopHookInput](cb)
	case PreCompactCallback:
		genericCallback = ToGenericCallback[PreCompactHookInput](cb)
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
		} else if fn, ok := callback.(func(*UserPromptSubmitHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[UserPromptSubmitHookInput](fn)
		} else if fn, ok := callback.(func(*StopHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[StopHookInput](fn)
		} else if fn, ok := callback.(func(*SubagentStopHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[SubagentStopHookInput](fn)
		} else if fn, ok := callback.(func(*PreCompactHookInput, *string, *HookContext) (*HookOutput, error)); ok {
			genericCallback = ToGenericCallback[PreCompactHookInput](fn)
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
	Subtype               string              `json:"subtype"`
	ToolName              string              `json:"tool_name"`
	Input                 map[string]any      `json:"input"`
	PermissionSuggestions []PermissionUpdate  `json:"permission_suggestions,omitempty"`
	BlockedPath           *string             `json:"blocked_path,omitempty"`
}

// ControlRequestType returns the request subtype.
func (r *SDKControlPermissionRequest) ControlRequestType() string {
	return r.Subtype
}

// SDKControlInitializeRequest initializes the SDK session.
type SDKControlInitializeRequest struct {
	Subtype string              `json:"subtype"`
	Hooks   map[HookEvent]any   `json:"hooks,omitempty"`
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
}

// ControlRequestType returns the request subtype.
func (r *SDKControlRewindFilesRequest) ControlRequestType() string {
	return r.Subtype
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
}

func (r *PermissionResultAllow) isPermissionResult() {}

// PermissionResultDeny denies a tool from running.
type PermissionResultDeny struct {
	Behavior  string `json:"behavior"`
	Message   string `json:"message,omitempty"`
	Interrupt bool   `json:"interrupt,omitempty"`
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
	Signal      any                `json:"-"`
	Suggestions []PermissionUpdate `json:"suggestions,omitempty"`
	BlockedPath *string            `json:"blocked_path,omitempty"`
}

// CanUseToolCallback is called when a tool needs permission.
type CanUseToolCallback func(
	toolName string,
	input map[string]any,
	ctx *ToolPermissionContext,
) (PermissionResult, error)

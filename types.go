package sdk

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

// HookMatcher configures which hooks to run for an event.
type HookMatcher struct {
	Matcher *string        `json:"matcher,omitempty"`
	Hooks   []HookCallback `json:"-"`
	Timeout *float64       `json:"timeout,omitempty"`
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

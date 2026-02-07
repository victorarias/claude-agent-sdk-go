// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

// NewPreToolUseOutput creates a HookOutput for PreToolUse events with hook-specific fields.
// The decision parameter should be "allow", "deny", or "ask".
// The reason parameter is optional and provides context for the decision.
// The updatedInput parameter is optional and allows modifying tool input parameters.
func NewPreToolUseOutput(decision, reason string, updatedInput map[string]any) *HookOutput {
	return NewPreToolUseOutputWithContext(decision, reason, updatedInput, "")
}

// NewPreToolUseOutputWithContext creates a HookOutput for PreToolUse events with optional additional context.
func NewPreToolUseOutputWithContext(decision, reason string, updatedInput map[string]any, additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName":      "PreToolUse",
		"permissionDecision": decision,
	}

	if reason != "" {
		hookSpecific["permissionDecisionReason"] = reason
	}

	if updatedInput != nil {
		hookSpecific["updatedInput"] = updatedInput
	}
	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewPostToolUseOutput creates a HookOutput for PostToolUse events with hook-specific fields.
// The additionalContext parameter is optional and provides context about the tool execution.
func NewPostToolUseOutput(additionalContext string) *HookOutput {
	return NewPostToolUseOutputWithUpdate(additionalContext, nil)
}

// NewPostToolUseOutputWithUpdate creates a HookOutput for PostToolUse with optional updated MCP tool output.
func NewPostToolUseOutputWithUpdate(additionalContext string, updatedMCPToolOutput any) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "PostToolUse",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}
	if updatedMCPToolOutput != nil {
		hookSpecific["updatedMCPToolOutput"] = updatedMCPToolOutput
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewPostToolUseFailureOutput creates a HookOutput for PostToolUseFailure events.
func NewPostToolUseFailureOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "PostToolUseFailure",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewUserPromptSubmitOutput creates a HookOutput for UserPromptSubmit events with hook-specific fields.
// The additionalContext parameter is optional and provides context about the prompt submission.
func NewUserPromptSubmitOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "UserPromptSubmit",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewSessionStartOutput creates a HookOutput for SessionStart events with hook-specific fields.
// The additionalContext parameter is optional and provides context about the session start.
func NewSessionStartOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "SessionStart",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewSetupOutput creates a HookOutput for Setup events.
func NewSetupOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "Setup",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewStopOutput creates a HookOutput for Stop events with hook-specific fields.
func NewStopOutput() *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "Stop",
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewSubagentStopOutput creates a HookOutput for SubagentStop events with hook-specific fields.
func NewSubagentStopOutput() *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "SubagentStop",
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewNotificationOutput creates a HookOutput for Notification events.
func NewNotificationOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "Notification",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewSubagentStartOutput creates a HookOutput for SubagentStart events.
func NewSubagentStartOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "SubagentStart",
	}

	if additionalContext != "" {
		hookSpecific["additionalContext"] = additionalContext
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewPermissionRequestOutput creates a HookOutput for PermissionRequest events.
func NewPermissionRequestOutput(decision map[string]any) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "PermissionRequest",
	}

	if decision != nil {
		hookSpecific["decision"] = decision
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewPreCompactOutput creates a HookOutput for PreCompact events with hook-specific fields.
// The customInstructions parameter is optional and provides additional instructions for compacting.
func NewPreCompactOutput(customInstructions string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "PreCompact",
	}

	if customInstructions != "" {
		hookSpecific["customInstructions"] = customInstructions
	}

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

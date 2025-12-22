// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

// NewPreToolUseOutput creates a HookOutput for PreToolUse events with hook-specific fields.
// The decision parameter should be "allow", "deny", or "ask".
// The reason parameter is optional and provides context for the decision.
// The updatedInput parameter is optional and allows modifying tool input parameters.
func NewPreToolUseOutput(decision, reason string, updatedInput map[string]any) *HookOutput {
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

	return &HookOutput{
		HookSpecific: hookSpecific,
	}
}

// NewPostToolUseOutput creates a HookOutput for PostToolUse events with hook-specific fields.
// The additionalContext parameter is optional and provides context about the tool execution.
func NewPostToolUseOutput(additionalContext string) *HookOutput {
	hookSpecific := map[string]any{
		"hookEventName": "PostToolUse",
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

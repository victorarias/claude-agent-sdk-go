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

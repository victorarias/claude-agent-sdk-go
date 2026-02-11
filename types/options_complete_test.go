// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestOptionsCompleteness verifies that all fields from Python's ClaudeAgentOptions
// are present in Go's Options struct.
func TestOptionsCompleteness(t *testing.T) {
	// Create an Options struct and verify all required fields exist
	opts := &Options{}

	// Get the reflect.Type of Options
	optType := reflect.TypeOf(opts).Elem()

	// Python ClaudeAgentOptions fields that should exist in Go Options.
	// Based on reference/src/claude_agent_sdk/types.py lines 617-681
	// Format: GoFieldName -> PythonFieldName
	requiredFields := map[string]string{
		// Line 620: tools: list[str] | ToolsPreset | None = None
		"Tools": "tools",
		// Line 621: allowed_tools: list[str] = field(default_factory=list)
		"AllowedTools": "allowed_tools",
		// Line 622: system_prompt: str | SystemPromptPreset | None = None
		"SystemPrompt": "system_prompt",
		// Line 623: mcp_servers: dict[str, McpServerConfig] | str | Path = field(default_factory=dict)
		"MCPServers": "mcp_servers",
		// Line 624: permission_mode: PermissionMode | None = None
		"PermissionMode": "permission_mode",
		// Line 625: continue_conversation: bool = False
		"ContinueConversation": "continue_conversation",
		// Line 626: resume: str | None = None
		"Resume": "resume",
		// Line 627: max_turns: int | None = None
		"MaxTurns": "max_turns",
		// Line 628: max_budget_usd: float | None = None
		"MaxBudgetUSD": "max_budget_usd",
		// Line 629: disallowed_tools: list[str] = field(default_factory=list)
		"DisallowedTools": "disallowed_tools",
		// Line 630: model: str | None = None
		"Model": "model",
		// Line 631: fallback_model: str | None = None
		"FallbackModel": "fallback_model",
		// Line 633: betas: list[SdkBeta] = field(default_factory=list)
		"Betas": "betas",
		// Line 634: permission_prompt_tool_name: str | None = None
		"PermissionPromptToolName": "permission_prompt_tool_name",
		// Line 635: cwd: str | Path | None = None
		"Cwd": "cwd",
		// Line 636: cli_path: str | Path | None = None
		"CLIPath": "cli_path",
		// Line 637: settings: str | None = None
		"Settings": "settings",
		// Line 638: add_dirs: list[str | Path] = field(default_factory=list)
		"AddDirs": "add_dirs",
		// Line 639: env: dict[str, str] = field(default_factory=dict)
		"Env": "env",
		// Line 640-642: extra_args: dict[str, str | None] = field(default_factory=dict)
		"ExtraArgs": "extra_args",
		// Line 643: max_buffer_size: int | None = None
		"MaxBufferSize": "max_buffer_size",
		// Line 647: stderr: Callable[[str], None] | None = None
		"StderrCallback": "stderr",
		// Line 650: can_use_tool: CanUseTool | None = None
		"CanUseTool": "can_use_tool",
		// Line 653: hooks: dict[HookEvent, list[HookMatcher]] | None = None
		"Hooks": "hooks",
		// Line 655: user: str | None = None
		"User": "user",
		// Line 658: include_partial_messages: bool = False
		"IncludePartialMessages": "include_partial_messages",
		// Line 661: fork_session: bool = False
		"ForkSession": "fork_session",
		// Line 663: agents: dict[str, AgentDefinition] | None = None
		"Agents": "agents",
		// Line 665: setting_sources: list[SettingSource] | None = None
		"SettingSources": "setting_sources",
		// Line 669: sandbox: SandboxSettings | None = None
		"Sandbox": "sandbox",
		// Line 671: plugins: list[SdkPluginConfig] = field(default_factory=list)
		"Plugins": "plugins",
		// Line 673: max_thinking_tokens: int | None = None
		"MaxThinkingTokens": "max_thinking_tokens",
		// Line 676: output_format: dict[str, Any] | None = None
		"OutputFormat": "output_format",
		// Line 680: enable_file_checkpointing: bool = False
		"EnableFileCheckpointing": "enable_file_checkpointing",
	}

	// Track missing fields
	missingFields := []string{}

	// Check that each required field exists
	for goFieldName, pythonFieldName := range requiredFields {
		field, found := optType.FieldByName(goFieldName)
		if !found {
			missingFields = append(missingFields, goFieldName+" (Python: "+pythonFieldName+")")
			continue
		}

		// Verify the field has a json tag if it's not a non-serializable type
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			// Field is intentionally not serialized, that's OK for callbacks/hooks
			continue
		}

		// For serializable fields, verify the JSON tag matches Python field name
		if jsonTag == "" && field.Type.Kind() != reflect.Func {
			t.Errorf("Field %s should have a json tag", goFieldName)
		}
	}

	// Report all missing fields at once
	if len(missingFields) > 0 {
		t.Errorf("Missing %d fields in Options struct:\n  - %s",
			len(missingFields),
			missingFields[0])
		for i := 1; i < len(missingFields); i++ {
			t.Errorf("  - %s", missingFields[i])
		}
	}
}

// TestOptionsJSONTags verifies that JSON tags match expected Python field names.
func TestOptionsJSONTags(t *testing.T) {
	// Map of Go field names to expected JSON tag values (without omitempty)
	expectedTags := map[string]string{
		"Model":                      "model",
		"MaxTurns":                   "max_turns",
		"MaxThinkingTokens":          "max_thinking_tokens",
		"SystemPrompt":               "system_prompt",
		"AppendSystemPrompt":         "append_system_prompt",
		"Tools":                      "tools",
		"AllowedTools":               "allowed_tools",
		"DisallowedTools":            "disallowed_tools",
		"MCPServers":                 "mcp_servers",
		"PermissionMode":             "permission_mode",
		"PermissionPromptToolName":   "permission_prompt_tool_name",
		"Env":                        "env",
		"Cwd":                        "cwd",
		"CLIPath":                    "cli_path",
		"PathToClaudeCodeExecutable": "path_to_claude_code_executable",
		"Executable":                 "executable",
		"ExecutableArgs":             "executable_args",
		"AddDirs":                    "add_dirs",
		"Resume":                     "resume",
		"ForkSession":                "fork_session",
		"ContinueConversation":       "continue_conversation",
		"EnableFileCheckpointing":    "enable_file_checkpointing",
		"Sandbox":                    "sandbox",
		"Agents":                     "agents",
		"Plugins":                    "plugins",
		"Betas":                      "betas",
		"SettingSources":             "setting_sources",
		"Settings":                   "settings",
		"User":                       "user",
		"MaxBudgetUSD":               "max_budget_usd",
		"OutputFormat":               "output_format",
		"ExtraArgs":                  "extra_args",
		"MaxBufferSize":              "max_buffer_size",
		"IncludePartialMessages":     "include_partial_messages",
		"FallbackModel":              "fallback_model",
		"BundledCLIPath":             "bundled_cli_path",
	}

	optType := reflect.TypeOf(Options{})

	for fieldName, expectedTag := range expectedTags {
		field, found := optType.FieldByName(fieldName)
		if !found {
			t.Errorf("Field %s not found in Options struct", fieldName)
			continue
		}

		jsonTag := field.Tag.Get("json")
		// Remove ,omitempty or other options from tag
		if jsonTag != "" {
			// Parse the tag - first part is the name
			var tagName string
			for i, c := range jsonTag {
				if c == ',' {
					tagName = jsonTag[:i]
					break
				}
			}
			if tagName == "" {
				tagName = jsonTag
			}

			if tagName != expectedTag && tagName != "-" {
				t.Errorf("Field %s has JSON tag %q, expected %q", fieldName, tagName, expectedTag)
			}
		}
	}
}

// TestOptionsFieldTypes verifies that field types are appropriate for their use.
func TestOptionsFieldTypes(t *testing.T) {
	tests := []struct {
		fieldName    string
		expectedKind reflect.Kind
		description  string
		allowAny     bool // Allow 'any' type (interface{})
	}{
		{"Model", reflect.String, "model should be a string", false},
		{"MaxTurns", reflect.Int, "max_turns should be an int", false},
		{"MaxThinkingTokens", reflect.Int, "max_thinking_tokens should be an int", false},
		{"MaxBudgetUSD", reflect.Float64, "max_budget_usd should be a float64", false},
		{"Tools", reflect.Slice, "tools should be a slice", true}, // Can also be ToolsPreset in Python
		{"AllowedTools", reflect.Slice, "allowed_tools should be a slice", false},
		{"DisallowedTools", reflect.Slice, "disallowed_tools should be a slice", false},
		{"Env", reflect.Map, "env should be a map", false},
		{"Cwd", reflect.String, "cwd should be a string", false},
		{"PathToClaudeCodeExecutable", reflect.String, "path_to_claude_code_executable should be a string", false},
		{"Executable", reflect.String, "executable should be a string", false},
		{"ExecutableArgs", reflect.Slice, "executable_args should be a slice", false},
		{"Resume", reflect.String, "resume should be a string", false},
		{"ForkSession", reflect.Bool, "fork_session should be a bool", false},
		{"ContinueConversation", reflect.Bool, "continue_conversation should be a bool", false},
		{"EnableFileCheckpointing", reflect.Bool, "enable_file_checkpointing should be a bool", false},
		{"IncludePartialMessages", reflect.Bool, "include_partial_messages should be a bool", false},
		{"Agents", reflect.Map, "agents should be a map", false},
		{"Plugins", reflect.Slice, "plugins should be a slice", false},
		{"Betas", reflect.Slice, "betas should be a slice", false},
		{"SettingSources", reflect.Slice, "setting_sources should be a slice", false},
		{"User", reflect.String, "user should be a string", false},
		{"OutputFormat", reflect.Map, "output_format should be a map", false},
		{"ExtraArgs", reflect.Map, "extra_args should be a map", false},
		{"MaxBufferSize", reflect.Int, "max_buffer_size should be an int", false},
	}

	optType := reflect.TypeOf(Options{})

	for _, tt := range tests {
		field, found := optType.FieldByName(tt.fieldName)
		if !found {
			t.Errorf("Field %s not found: %s", tt.fieldName, tt.description)
			continue
		}

		// For pointer types, get the underlying type
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Allow Interface (any) type if specified
		if tt.allowAny && fieldType.Kind() == reflect.Interface {
			continue
		}

		if fieldType.Kind() != tt.expectedKind {
			t.Errorf("Field %s: expected kind %v, got %v (%s)",
				tt.fieldName, tt.expectedKind, fieldType.Kind(), tt.description)
		}
	}
}

// TestOptionsSerializationRoundTrip verifies that Options can be marshaled
// and unmarshaled without losing data for basic fields.
func TestOptionsSerializationRoundTrip(t *testing.T) {
	original := &Options{
		Model:                      "claude-opus-4",
		MaxTurns:                   10,
		MaxThinkingTokens:          5000,
		MaxBudgetUSD:               1.5,
		Tools:                      []string{"bash", "edit"},
		AllowedTools:               []string{"bash"},
		DisallowedTools:            []string{"danger"},
		Env:                        map[string]string{"FOO": "bar"},
		Cwd:                        "/tmp",
		Resume:                     "session-123",
		ForkSession:                true,
		ContinueConversation:       false,
		EnableFileCheckpointing:    true,
		IncludePartialMessages:     true,
		User:                       "test-user",
		MaxBufferSize:              2048,
		FallbackModel:              "claude-sonnet-3.5",
		AppendSystemPrompt:         "Be helpful",
		PermissionMode:             PermissionDefault,
		CLIPath:                    "/usr/bin/claude",
		PathToClaudeCodeExecutable: "/tmp/cli.js",
		Executable:                 "node",
		ExecutableArgs:             []string{"--no-warnings"},
		AddDirs:                    []string{"/tmp/dir1", "/tmp/dir2"},
		Settings:                   "/path/to/settings.json",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Options: %v", err)
	}

	// Unmarshal back
	var decoded Options
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Options: %v", err)
	}

	// Compare key fields
	if decoded.Model != original.Model {
		t.Errorf("Model: got %q, want %q", decoded.Model, original.Model)
	}
	if decoded.MaxTurns != original.MaxTurns {
		t.Errorf("MaxTurns: got %d, want %d", decoded.MaxTurns, original.MaxTurns)
	}
	if decoded.MaxThinkingTokens != original.MaxThinkingTokens {
		t.Errorf("MaxThinkingTokens: got %d, want %d", decoded.MaxThinkingTokens, original.MaxThinkingTokens)
	}
	if decoded.MaxBudgetUSD != original.MaxBudgetUSD {
		t.Errorf("MaxBudgetUSD: got %f, want %f", decoded.MaxBudgetUSD, original.MaxBudgetUSD)
	}
	if decoded.ForkSession != original.ForkSession {
		t.Errorf("ForkSession: got %v, want %v", decoded.ForkSession, original.ForkSession)
	}
	if decoded.EnableFileCheckpointing != original.EnableFileCheckpointing {
		t.Errorf("EnableFileCheckpointing: got %v, want %v", decoded.EnableFileCheckpointing, original.EnableFileCheckpointing)
	}
	if decoded.User != original.User {
		t.Errorf("User: got %q, want %q", decoded.User, original.User)
	}
}

// TestSystemPromptTypes verifies that SystemPrompt can accept both string and preset types.
func TestSystemPromptTypes(t *testing.T) {
	// Test with string
	opts1 := &Options{
		SystemPrompt: "custom prompt",
	}
	if opts1.SystemPrompt == nil {
		t.Error("SystemPrompt should not be nil when set to string")
	}

	// Test with preset
	preset := SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
	}
	opts2 := &Options{
		SystemPrompt: preset,
	}
	if opts2.SystemPrompt == nil {
		t.Error("SystemPrompt should not be nil when set to preset")
	}
}

// TestToolsTypes verifies that Tools field can accept string slice.
// TODO: In Python, Tools can also be ToolsPreset. Consider making this field `any` like SystemPrompt.
func TestToolsTypes(t *testing.T) {
	// Test with string slice
	opts1 := &Options{
		Tools: []string{"bash", "edit"},
	}
	if opts1.Tools == nil {
		t.Error("Tools should not be nil when set to string slice")
	}
}

// TestOptionsDefaultValues verifies DefaultOptions returns reasonable defaults.
func TestOptionsDefaultValues(t *testing.T) {
	opts := DefaultOptions()

	if opts == nil {
		t.Fatal("DefaultOptions() returned nil")
	}

	if opts.Env == nil {
		t.Error("Default Env should be initialized to empty map")
	}

	if opts.MaxBufferSize <= 0 {
		t.Error("Default MaxBufferSize should be positive")
	}
}

// TestPythonGoFieldDifferences documents intentional differences between
// Python and Go implementations for maintainability.
func TestPythonGoFieldDifferences(t *testing.T) {
	// This test documents known differences between Python's ClaudeAgentOptions
	// and Go's Options struct. These are intentional design decisions.

	t.Run("AppendSystemPrompt exists in Go but not Python", func(t *testing.T) {
		// In Python, appending to system prompt is done via SystemPromptPreset.append field
		// In Go, we have a separate AppendSystemPrompt field for convenience
		opts := &Options{
			AppendSystemPrompt: "Additional instructions",
		}
		if opts.AppendSystemPrompt == "" {
			t.Error("AppendSystemPrompt should be settable")
		}
	})

	t.Run("Tools field type difference", func(t *testing.T) {
		// Python: tools: list[str] | ToolsPreset | None
		// Go: Tools []string (currently)
		// TODO: Consider making Go's Tools field `any` type like SystemPrompt
		// to support both []string and ToolsPreset
		opts := &Options{
			Tools: []string{"bash", "edit"},
		}
		if opts.Tools == nil {
			t.Error("Tools should support string slice")
		}
	})

	t.Run("SDK-specific fields in Go", func(t *testing.T) {
		// Go has additional fields not in Python ClaudeAgentOptions:
		// - SDKMCPServers (for in-process MCP servers)
		// - BundledCLIPath
		// - customTransport (for testing)
		// These are Go SDK implementation details

		opts := &Options{}

		// Verify these exist
		optType := reflect.TypeOf(opts).Elem()

		if _, found := optType.FieldByName("SDKMCPServers"); !found {
			t.Error("SDKMCPServers should exist for in-process MCP servers")
		}

		if _, found := optType.FieldByName("BundledCLIPath"); !found {
			t.Error("BundledCLIPath should exist")
		}
	})

	t.Run("Python debug_stderr vs Go StderrCallback", func(t *testing.T) {
		// Python has both:
		//   - debug_stderr: Any (file-like object, deprecated)
		//   - stderr: Callable[[str], None] | None
		// Go has:
		//   - StderrCallback: func(string)
		// Go uses only the callback approach for simplicity

		opts := &Options{
			StderrCallback: func(s string) {
				// Callback for stderr
			},
		}
		if opts.StderrCallback == nil {
			t.Error("StderrCallback should be settable")
		}
	})
}

// TestOptionsFieldCount verifies we have roughly the expected number of fields.
func TestOptionsFieldCount(t *testing.T) {
	optType := reflect.TypeOf(Options{})
	fieldCount := optType.NumField()

	// Python ClaudeAgentOptions has ~35 fields
	// Go Options should have at least that many (may have a few extra for SDK-specific features)
	if fieldCount < 35 {
		t.Errorf("Options has %d fields, expected at least 35 to match Python", fieldCount)
	}

	// Sanity check: shouldn't have way more than Python
	if fieldCount > 50 {
		t.Logf("Warning: Options has %d fields, Python has ~35. Consider if all fields are necessary.", fieldCount)
	}
}

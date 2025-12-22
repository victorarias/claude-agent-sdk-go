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
	requiredFields := map[string]string{
		// Core functionality
		"Model":                     "model",
		"MaxTurns":                  "max_turns",
		"MaxThinkingTokens":         "max_thinking_tokens",
		"SystemPrompt":              "system_prompt",
		"AppendSystemPrompt":        "append_system_prompt",

		// Tools configuration
		"Tools":                     "tools",
		"AllowedTools":              "allowed_tools",
		"DisallowedTools":           "disallowed_tools",

		// MCP servers
		"MCPServers":                "mcp_servers",

		// Permission settings
		"PermissionMode":            "permission_mode",
		"PermissionPromptToolName":  "permission_prompt_tool_name",
		"CanUseTool":                "can_use_tool",

		// Hooks
		"Hooks":                     "hooks",

		// Environment and paths
		"Env":                       "env",
		"Cwd":                       "cwd",
		"CLIPath":                   "cli_path",
		"AddDirs":                   "add_dirs",

		// Session management
		"Resume":                    "resume",
		"ForkSession":               "fork_session",
		"ContinueConversation":      "continue_conversation",

		// File checkpointing
		"EnableFileCheckpointing":   "enable_file_checkpointing",

		// Sandbox
		"Sandbox":                   "sandbox",

		// Custom agents
		"Agents":                    "agents",

		// Plugins
		"Plugins":                   "plugins",

		// Beta features
		"Betas":                     "betas",

		// Settings
		"SettingSources":            "setting_sources",
		"Settings":                  "settings",

		// User identifier
		"User":                      "user",

		// Additional limits
		"MaxBudgetUSD":              "max_budget_usd",

		// Output format
		"OutputFormat":              "output_format",

		// Extra args
		"ExtraArgs":                 "extra_args",

		// Buffer size
		"MaxBufferSize":             "max_buffer_size",

		// Streaming
		"IncludePartialMessages":    "include_partial_messages",

		// Fallback model
		"FallbackModel":             "fallback_model",

		// Stderr callback (Python has both debug_stderr and stderr)
		"StderrCallback":            "stderr",
	}

	// Check that each required field exists
	for goFieldName, pythonFieldName := range requiredFields {
		field, found := optType.FieldByName(goFieldName)
		if !found {
			t.Errorf("Field %s (Python: %s) not found in Options struct", goFieldName, pythonFieldName)
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
}

// TestOptionsJSONTags verifies that JSON tags match expected Python field names.
func TestOptionsJSONTags(t *testing.T) {
	// Map of Go field names to expected JSON tag values (without omitempty)
	expectedTags := map[string]string{
		"Model":                    "model",
		"MaxTurns":                 "max_turns",
		"MaxThinkingTokens":        "max_thinking_tokens",
		"SystemPrompt":             "system_prompt",
		"AppendSystemPrompt":       "append_system_prompt",
		"Tools":                    "tools",
		"AllowedTools":             "allowed_tools",
		"DisallowedTools":          "disallowed_tools",
		"MCPServers":               "mcp_servers",
		"PermissionMode":           "permission_mode",
		"PermissionPromptToolName": "permission_prompt_tool_name",
		"Env":                      "env",
		"Cwd":                      "cwd",
		"CLIPath":                  "cli_path",
		"AddDirs":                  "add_dirs",
		"Resume":                   "resume",
		"ForkSession":              "fork_session",
		"ContinueConversation":     "continue_conversation",
		"EnableFileCheckpointing":  "enable_file_checkpointing",
		"Sandbox":                  "sandbox",
		"Agents":                   "agents",
		"Plugins":                  "plugins",
		"Betas":                    "betas",
		"SettingSources":           "setting_sources",
		"Settings":                 "settings",
		"User":                     "user",
		"MaxBudgetUSD":             "max_budget_usd",
		"OutputFormat":             "output_format",
		"ExtraArgs":                "extra_args",
		"MaxBufferSize":            "max_buffer_size",
		"IncludePartialMessages":   "include_partial_messages",
		"FallbackModel":            "fallback_model",
		"BundledCLIPath":           "bundled_cli_path",
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
		Model:                   "claude-opus-4",
		MaxTurns:                10,
		MaxThinkingTokens:       5000,
		MaxBudgetUSD:            1.5,
		Tools:                   []string{"bash", "edit"},
		AllowedTools:            []string{"bash"},
		DisallowedTools:         []string{"danger"},
		Env:                     map[string]string{"FOO": "bar"},
		Cwd:                     "/tmp",
		Resume:                  "session-123",
		ForkSession:             true,
		ContinueConversation:    false,
		EnableFileCheckpointing: true,
		IncludePartialMessages:  true,
		User:                    "test-user",
		MaxBufferSize:           2048,
		FallbackModel:           "claude-sonnet-3.5",
		AppendSystemPrompt:      "Be helpful",
		PermissionMode:          PermissionDefault,
		CLIPath:                 "/usr/bin/claude",
		AddDirs:                 []string{"/tmp/dir1", "/tmp/dir2"},
		Settings:                "/path/to/settings.json",
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

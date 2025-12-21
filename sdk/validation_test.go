package sdk

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestPermissionValidation tests that CanUseTool and PermissionPromptToolName
// are mutually exclusive.
func TestPermissionValidation(t *testing.T) {
	tests := []struct {
		name        string
		options     []types.Option
		expectError bool
		errorMsg    string
	}{
		{
			name: "both CanUseTool and PermissionPromptToolName set - should error",
			options: []types.Option{
				WithCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
					return &types.PermissionResultAllow{Behavior: "allow"}, nil
				}),
				func(o *types.Options) {
					o.PermissionPromptToolName = "stdio"
				},
			},
			expectError: true,
			errorMsg:    "can_use_tool callback cannot be used with permission_prompt_tool_name",
		},
		{
			name: "only CanUseTool set - should succeed",
			options: []types.Option{
				WithCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
					return &types.PermissionResultAllow{Behavior: "allow"}, nil
				}),
			},
			expectError: false,
		},
		{
			name: "only PermissionPromptToolName set - should succeed",
			options: []types.Option{
				func(o *types.Options) {
					o.PermissionPromptToolName = "stdio"
				},
			},
			expectError: false,
		},
		{
			name:        "neither set - should succeed",
			options:     []types.Option{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.options...)
			err := client.validateOptions()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q but got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

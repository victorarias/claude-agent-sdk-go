// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestOptions_CustomTransport tests custom transport methods.
func TestOptions_CustomTransport(t *testing.T) {
	t.Run("CustomTransport returns nil by default", func(t *testing.T) {
		opts := DefaultOptions()
		if opts.CustomTransport() != nil {
			t.Error("Expected CustomTransport to return nil by default")
		}
	})

	t.Run("SetCustomTransport sets transport", func(t *testing.T) {
		opts := DefaultOptions()
		transport := &mockTransport{}
		opts.SetCustomTransport(transport)
		if opts.CustomTransport() != transport {
			t.Error("Expected CustomTransport to return the set transport")
		}
	})
}

// TestOptions_WithSDKMCPServer tests WithSDKMCPServer option.
func TestOptions_WithSDKMCPServer(t *testing.T) {
	t.Run("WithSDKMCPServer adds server to MCPServers map", func(t *testing.T) {
		server := &MCPStdioServerConfig{
			Command: "test-server",
			Args:    []string{"--verbose"},
		}

		opts := DefaultOptions()
		opt := WithSDKMCPServer("test", server)
		opt(opts)

		if opts.MCPServers == nil {
			t.Fatal("Expected MCPServers to be initialized")
		}

		result, ok := opts.MCPServers["test"]
		if !ok {
			t.Fatal("Expected server 'test' to be in MCPServers")
		}

		stdioServer, ok := result.(*MCPStdioServerConfig)
		if !ok {
			t.Fatal("Expected server to be MCPStdioServerConfig")
		}

		if stdioServer.Command != "test-server" {
			t.Errorf("Expected command 'test-server', got %q", stdioServer.Command)
		}
	})
}

// TestToolsMap tests getToolsMap, GetTool, and CallTool methods.
func TestToolsMap(t *testing.T) {
	tool1 := Tool{
		Name:        "calculator",
		Description: "Performs calculations",
	}
	tool2 := Tool{
		Name:        "search",
		Description: "Searches the web",
	}

	t.Run("getToolsMap creates map from tools", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Tools = []Tool{tool1, tool2}

		toolsMap := opts.getToolsMap()
		if len(toolsMap) != 2 {
			t.Errorf("Expected 2 tools in map, got %d", len(toolsMap))
		}

		if toolsMap["calculator"].Name != "calculator" {
			t.Error("Expected calculator tool in map")
		}
		if toolsMap["search"].Name != "search" {
			t.Error("Expected search tool in map")
		}
	})

	t.Run("GetTool retrieves tool by name", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Tools = []Tool{tool1, tool2}

		tool := opts.GetTool("calculator")
		if tool == nil {
			t.Fatal("Expected to find calculator tool")
		}
		if tool.Name != "calculator" {
			t.Errorf("Expected calculator, got %s", tool.Name)
		}
	})

	t.Run("GetTool returns nil for unknown tool", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Tools = []Tool{tool1}

		tool := opts.GetTool("unknown")
		if tool != nil {
			t.Error("Expected nil for unknown tool")
		}
	})

	t.Run("CallTool executes tool callback", func(t *testing.T) {
		called := false
		callbackTool := Tool{
			Name:        "test",
			Description: "Test tool",
			Callback: func(params map[string]any) (any, error) {
				called = true
				return "success", nil
			},
		}

		opts := DefaultOptions()
		opts.Tools = []Tool{callbackTool}

		result, err := opts.CallTool("test", map[string]any{"param": "value"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !called {
			t.Error("Expected tool callback to be called")
		}
		if result != "success" {
			t.Errorf("Expected result 'success', got %v", result)
		}
	})

	t.Run("CallTool returns error for unknown tool", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Tools = []Tool{}

		_, err := opts.CallTool("unknown", nil)
		if err == nil {
			t.Error("Expected error for unknown tool")
		}
	})

	t.Run("CallTool returns error when tool has no callback", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Tools = []Tool{{Name: "nocallback"}}

		_, err := opts.CallTool("nocallback", nil)
		if err == nil {
			t.Error("Expected error when tool has no callback")
		}
	})
}

// TestToConfig tests the ToConfig method.
func TestToConfig(t *testing.T) {
	t.Run("ToConfig creates basic config map", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Model = "claude-3-5-sonnet-20241022"
		opts.MaxTurns = ptr(10)
		opts.PermissionMode = ptr(PermissionModeNone)

		config := opts.ToConfig()

		if config["model"] != "claude-3-5-sonnet-20241022" {
			t.Errorf("Expected model in config, got %v", config["model"])
		}

		if config["max_turns"] != 10 {
			t.Errorf("Expected max_turns=10, got %v", config["max_turns"])
		}

		if config["permission_mode"] != "none" {
			t.Errorf("Expected permission_mode=none, got %v", config["permission_mode"])
		}
	})

	t.Run("ToConfig includes system prompt", func(t *testing.T) {
		opts := DefaultOptions()
		opts.SystemPrompt = &SystemPrompt{
			Text: "You are a helpful assistant",
		}

		config := opts.ToConfig()
		sysprompt, ok := config["system_prompt"].(map[string]any)
		if !ok {
			t.Fatal("Expected system_prompt in config")
		}

		if sysprompt["text"] != "You are a helpful assistant" {
			t.Errorf("Expected system prompt text, got %v", sysprompt["text"])
		}
	})

	t.Run("ToConfig includes tools", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Tools = []Tool{
			{
				Name:        "calculator",
				Description: "Performs calculations",
			},
		}

		config := opts.ToConfig()
		tools, ok := config["tools"].([]map[string]any)
		if !ok {
			t.Fatal("Expected tools in config")
		}

		if len(tools) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(tools))
		}

		if tools[0]["name"] != "calculator" {
			t.Errorf("Expected calculator tool, got %v", tools[0]["name"])
		}
	})

	t.Run("ToConfig omits nil fields", func(t *testing.T) {
		opts := DefaultOptions()
		opts.MaxTurns = nil

		config := opts.ToConfig()
		if _, exists := config["max_turns"]; exists {
			t.Error("Expected max_turns to be omitted when nil")
		}
	})

	t.Run("ToConfig includes MCP servers", func(t *testing.T) {
		opts := DefaultOptions()
		opts.MCPServers = map[string]MCPServerConfig{
			"test": &MCPStdioServerConfig{
				Command: "test-server",
			},
		}

		config := opts.ToConfig()
		servers, ok := config["mcp_servers"].(map[string]map[string]any)
		if !ok {
			t.Fatal("Expected mcp_servers in config")
		}

		testServer, ok := servers["test"]
		if !ok {
			t.Fatal("Expected 'test' server in config")
		}

		if testServer["command"] != "test-server" {
			t.Errorf("Expected command 'test-server', got %v", testServer["command"])
		}
	})

	t.Run("ToConfig includes sandbox settings", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Sandbox = &SandboxSettings{
			Enabled: ptr(true),
		}

		config := opts.ToConfig()
		sandbox, ok := config["sandbox"].(map[string]any)
		if !ok {
			t.Fatal("Expected sandbox in config")
		}

		if sandbox["enabled"] != true {
			t.Errorf("Expected enabled=true, got %v", sandbox["enabled"])
		}
	})
}

// TestMCPServerBuilder_WithTool tests the WithTool method.
func TestMCPServerBuilder_WithTool(t *testing.T) {
	t.Run("WithTool adds tool to builder", func(t *testing.T) {
		builder := NewMCPServerBuilder()
		tool := Tool{
			Name:        "test",
			Description: "Test tool",
		}

		builder.WithTool(tool)
		server := builder.Build()

		if len(server.Tools) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(server.Tools))
		}

		if server.Tools[0].Name != "test" {
			t.Errorf("Expected tool name 'test', got %s", server.Tools[0].Name)
		}
	})

	t.Run("WithTool returns builder for chaining", func(t *testing.T) {
		builder := NewMCPServerBuilder()
		result := builder.WithTool(Tool{Name: "test"})
		if result != builder {
			t.Error("Expected WithTool to return the same builder instance")
		}
	})
}

// mockTransport is a mock transport for testing.
type mockTransport struct{}

func (m *mockTransport) Start() error                     { return nil }
func (m *mockTransport) Stop() error                      { return nil }
func (m *mockTransport) Send(data []byte) error           { return nil }
func (m *mockTransport) Receive() ([]byte, error)         { return nil, nil }
func (m *mockTransport) IsConnected() bool                { return true }
func (m *mockTransport) SetStderrCallback(func([]byte))   {}
func (m *mockTransport) SetMaxBufferSize(size int)        {}
func (m *mockTransport) SetPermissionMode(mode string)    {}
func (m *mockTransport) HandleInterrupt() error           { return nil }
func (m *mockTransport) HandlePermission(result any) error { return nil }

// ptr is a helper function to create pointers.
func ptr[T any](v T) *T {
	return &v
}

// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestMCPTool(t *testing.T) {
	tool := &types.MCPTool{
		Name:        "test_tool",
		Description: "A test tool",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
		},
		Handler: func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "success"}},
			}, nil
		},
	}

	if tool.Name != "test_tool" {
		t.Errorf("expected name test_tool, got %s", tool.Name)
	}

	result, err := tool.Handler(map[string]any{"input": "test"})
	if err != nil {
		t.Errorf("handler failed: %v", err)
	}

	if len(result.Content) != 1 || result.Content[0].Text != "success" {
		t.Errorf("expected success, got %v", result)
	}
}

func TestMCPServer_GetTool(t *testing.T) {
	server := &types.MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
		Tools: []*types.MCPTool{
			{
				Name:        "greet",
				Description: "Greets someone",
				Handler: func(args map[string]any) (*types.MCPToolResult, error) {
					return &types.MCPToolResult{
						Content: []types.MCPContent{{Type: "text", Text: "Hello!"}},
					}, nil
				},
			},
		},
	}

	tool, ok := server.GetTool("greet")
	if !ok {
		t.Fatal("tool 'greet' not found")
	}
	if tool.Name != "greet" {
		t.Errorf("expected tool name greet, got %s", tool.Name)
	}

	_, ok = server.GetTool("nonexistent")
	if ok {
		t.Error("expected nonexistent tool to not be found")
	}
}

func TestMCPServer_CallTool(t *testing.T) {
	server := &types.MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
		Tools: []*types.MCPTool{
			{
				Name:        "echo",
				Description: "Echoes input",
				Handler: func(args map[string]any) (*types.MCPToolResult, error) {
					msg := args["message"].(string)
					return &types.MCPToolResult{
						Content: []types.MCPContent{{Type: "text", Text: msg}},
					}, nil
				},
			},
		},
	}

	result, err := server.CallTool("echo", map[string]any{"message": "hello"})
	if err != nil {
		t.Errorf("CallTool failed: %v", err)
	}

	if len(result.Content) != 1 || result.Content[0].Text != "hello" {
		t.Errorf("expected hello, got %v", result)
	}
}

func TestMCPServer_CallTool_NotFound(t *testing.T) {
	server := &types.MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
	}

	_, err := server.CallTool("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestMCPServer_ToConfig(t *testing.T) {
	server := &types.MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
		Tools: []*types.MCPTool{
			{
				Name:        "greet",
				Description: "Greets someone",
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
				},
			},
		},
	}

	config := server.ToConfig()

	if config["name"] != "test-server" {
		t.Errorf("expected name test-server, got %v", config["name"])
	}
	if config["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", config["version"])
	}

	tools := config["tools"].([]map[string]any)
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}
	if tools[0]["name"] != "greet" {
		t.Errorf("expected tool name greet, got %v", tools[0]["name"])
	}
}

func TestMCPServerBuilder(t *testing.T) {
	server := types.NewMCPServerBuilder("test-server").
		WithVersion("2.0.0").
		WithTool("greet", "Greets a user", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			name := args["name"].(string)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "Hello, " + name + "!"}},
			}, nil
		}).
		Build()

	if server.Name != "test-server" {
		t.Errorf("expected name test-server, got %s", server.Name)
	}

	if server.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", server.Version)
	}

	if len(server.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(server.Tools))
	}

	if server.Tools[0].Name != "greet" {
		t.Errorf("expected tool name greet, got %s", server.Tools[0].Name)
	}

	// Test the handler
	result, err := server.CallTool("greet", map[string]any{"name": "World"})
	if err != nil {
		t.Errorf("CallTool failed: %v", err)
	}

	if len(result.Content) != 1 || result.Content[0].Text != "Hello, World!" {
		t.Errorf("expected Hello, World!, got %v", result)
	}
}

func TestMCPServerBuilder_MultipleTool(t *testing.T) {
	server := types.NewMCPServerBuilder("multi-tool").
		WithTool("add", "Adds numbers", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "result"}},
			}, nil
		}).
		WithTool("multiply", "Multiplies numbers", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "result"}},
			}, nil
		}).
		Build()

	if len(server.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(server.Tools))
	}

	// Test add exists
	_, ok := server.GetTool("add")
	if !ok {
		t.Error("expected add tool to exist")
	}

	// Test multiply exists
	_, ok = server.GetTool("multiply")
	if !ok {
		t.Error("expected multiply tool to exist")
	}
}

func TestMCPServerBuilder_DefaultVersion(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()

	if server.Version != "1.0.0" {
		t.Errorf("expected default version 1.0.0, got %s", server.Version)
	}
}

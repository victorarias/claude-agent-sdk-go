package sdk

import (
	"context"
	"testing"
)

func TestMCPToolHandler(t *testing.T) {
	handler := &MCPToolHandler{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, input map[string]any) (any, error) {
			return map[string]any{"result": "success"}, nil
		},
	}

	if handler.Name != "test_tool" {
		t.Errorf("expected name test_tool, got %s", handler.Name)
	}

	result, err := handler.Handler(context.Background(), map[string]any{"input": "test"})
	if err != nil {
		t.Errorf("handler failed: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["result"] != "success" {
		t.Errorf("expected success, got %v", resultMap["result"])
	}
}

func TestNewMCPServer(t *testing.T) {
	tools := []*MCPToolHandler{
		{
			Name:        "tool1",
			Description: "First tool",
			InputSchema: map[string]any{"type": "object"},
			Handler: func(ctx context.Context, input map[string]any) (any, error) {
				return "tool1 result", nil
			},
		},
		{
			Name:        "tool2",
			Description: "Second tool",
			InputSchema: map[string]any{"type": "object"},
			Handler: func(ctx context.Context, input map[string]any) (any, error) {
				return "tool2 result", nil
			},
		},
	}

	server := NewMCPServer("test-server", tools)

	if server.Name != "test-server" {
		t.Errorf("expected name test-server, got %s", server.Name)
	}

	if len(server.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(server.Tools))
	}
}

func TestMCPServer_GetTool(t *testing.T) {
	server := NewMCPServer("test-server", []*MCPToolHandler{
		{
			Name:        "greet",
			Description: "Greets someone",
			Handler: func(ctx context.Context, input map[string]any) (any, error) {
				return "Hello!", nil
			},
		},
	})

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
	server := NewMCPServer("test-server", []*MCPToolHandler{
		{
			Name:        "echo",
			Description: "Echoes input",
			Handler: func(ctx context.Context, input map[string]any) (any, error) {
				return map[string]any{"echo": input["message"]}, nil
			},
		},
	})

	result, err := server.CallTool(context.Background(), "echo", map[string]any{"message": "hello"})
	if err != nil {
		t.Errorf("CallTool failed: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["echo"] != "hello" {
		t.Errorf("expected echo hello, got %v", resultMap["echo"])
	}
}

func TestMCPServer_CallTool_NotFound(t *testing.T) {
	server := NewMCPServer("test-server", nil)

	_, err := server.CallTool(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestMCPServer_ToConfig(t *testing.T) {
	server := NewMCPServer("test-server", []*MCPToolHandler{
		{
			Name:        "greet",
			Description: "Greets someone",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
	})

	config := server.ToConfig()

	if config["name"] != "test-server" {
		t.Errorf("expected name test-server, got %v", config["name"])
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
	server := NewMCPServerBuilder("test-server").
		WithTool("greet", "Greets a user", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}, func(ctx context.Context, input map[string]any) (any, error) {
			name := input["name"].(string)
			return map[string]any{"greeting": "Hello, " + name + "!"}, nil
		}).
		Build()

	if server.Name != "test-server" {
		t.Errorf("expected name test-server, got %s", server.Name)
	}

	if len(server.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(server.Tools))
	}

	if server.Tools[0].Name != "greet" {
		t.Errorf("expected tool name greet, got %s", server.Tools[0].Name)
	}

	// Test the handler
	result, err := server.CallTool(context.Background(), "greet", map[string]any{"name": "World"})
	if err != nil {
		t.Errorf("CallTool failed: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["greeting"] != "Hello, World!" {
		t.Errorf("expected greeting Hello, World!, got %v", resultMap["greeting"])
	}
}

func TestMCPServerBuilder_MultipleTool(t *testing.T) {
	server := NewMCPServerBuilder("multi-tool").
		WithTool("add", "Adds numbers", map[string]any{
			"type": "object",
		}, func(ctx context.Context, input map[string]any) (any, error) {
			a := input["a"].(float64)
			b := input["b"].(float64)
			return a + b, nil
		}).
		WithTool("multiply", "Multiplies numbers", map[string]any{
			"type": "object",
		}, func(ctx context.Context, input map[string]any) (any, error) {
			a := input["a"].(float64)
			b := input["b"].(float64)
			return a * b, nil
		}).
		Build()

	if len(server.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(server.Tools))
	}

	// Test add
	result, err := server.CallTool(context.Background(), "add", map[string]any{"a": 2.0, "b": 3.0})
	if err != nil {
		t.Errorf("CallTool add failed: %v", err)
	}
	if result.(float64) != 5.0 {
		t.Errorf("expected 5, got %v", result)
	}

	// Test multiply
	result, err = server.CallTool(context.Background(), "multiply", map[string]any{"a": 4.0, "b": 5.0})
	if err != nil {
		t.Errorf("CallTool multiply failed: %v", err)
	}
	if result.(float64) != 20.0 {
		t.Errorf("expected 20, got %v", result)
	}
}

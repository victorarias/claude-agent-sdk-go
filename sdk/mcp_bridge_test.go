package sdk

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestHandleMCPMessage_Initialize tests MCP initialize request
func TestHandleMCPMessage_Initialize(t *testing.T) {
	// Create a test MCP server
	server := types.NewMCPServerBuilder("test-server").
		WithVersion("1.0.0").
		WithTool("test_tool", "A test tool", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "ok"}},
			}, nil
		}).
		Build()

	// Create query and register server
	transport := NewMockTransport()
	query := NewQuery(transport, true)
	query.RegisterMCPServer(server)

	// Create MCP initialize message
	mcpMessage := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	// Call handleMCPMessage
	response, err := query.handleMCPMessage("test-server", mcpMessage)
	if err != nil {
		t.Fatalf("handleMCPMessage failed: %v", err)
	}

	// Verify response structure
	respMap, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("expected map response, got %T", response)
	}

	if respMap["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", respMap["jsonrpc"])
	}

	if respMap["id"] != 1 {
		t.Errorf("expected id 1, got %v", respMap["id"])
	}

	// Check result
	result, ok := respMap["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", respMap["result"])
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion 2024-11-05, got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("expected serverInfo map, got %T", result["serverInfo"])
	}

	if serverInfo["name"] != "test-server" {
		t.Errorf("expected server name test-server, got %v", serverInfo["name"])
	}

	if serverInfo["version"] != "1.0.0" {
		t.Errorf("expected server version 1.0.0, got %v", serverInfo["version"])
	}
}

// TestHandleMCPMessage_NotificationsInitialized tests notifications/initialized
func TestHandleMCPMessage_NotificationsInitialized(t *testing.T) {
	server := types.NewMCPServerBuilder("test-server").Build()

	transport := NewMockTransport()
	query := NewQuery(transport, true)
	query.RegisterMCPServer(server)

	// Notification has no id
	mcpMessage := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}

	// Call handleMCPMessage - notifications return nil
	response, err := query.handleMCPMessage("test-server", mcpMessage)
	if err != nil {
		t.Fatalf("handleMCPMessage failed: %v", err)
	}

	if response != nil {
		t.Errorf("expected nil response for notification, got %v", response)
	}
}

// TestHandleMCPMessage_ToolsList tests tools/list request
func TestHandleMCPMessage_ToolsList(t *testing.T) {
	server := types.NewMCPServerBuilder("test-server").
		WithTool("greet", "Greets someone", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "Hello!"}},
			}, nil
		}).
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "echo"}},
			}, nil
		}).
		Build()

	transport := NewMockTransport()
	query := NewQuery(transport, true)
	query.RegisterMCPServer(server)

	mcpMessage := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}

	response, err := query.handleMCPMessage("test-server", mcpMessage)
	if err != nil {
		t.Fatalf("handleMCPMessage failed: %v", err)
	}

	respMap, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("expected map response, got %T", response)
	}

	if respMap["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", respMap["jsonrpc"])
	}

	if respMap["id"] != 2 {
		t.Errorf("expected id 2, got %v", respMap["id"])
	}

	result, ok := respMap["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", respMap["result"])
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got %T", result["tools"])
	}

	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}

	// Check first tool
	tool0, ok := tools[0].(map[string]any)
	if !ok {
		t.Fatalf("expected tool map, got %T", tools[0])
	}

	if tool0["name"] != "greet" {
		t.Errorf("expected tool name greet, got %v", tool0["name"])
	}

	if tool0["description"] != "Greets someone" {
		t.Errorf("expected description, got %v", tool0["description"])
	}
}

// TestHandleMCPMessage_ToolsCall tests tools/call request
func TestHandleMCPMessage_ToolsCall(t *testing.T) {
	server := types.NewMCPServerBuilder("test-server").
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			message := args["message"].(string)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: message}},
			}, nil
		}).
		Build()

	transport := NewMockTransport()
	query := NewQuery(transport, true)
	query.RegisterMCPServer(server)

	mcpMessage := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "echo",
			"arguments": map[string]any{
				"message": "hello world",
			},
		},
	}

	response, err := query.handleMCPMessage("test-server", mcpMessage)
	if err != nil {
		t.Fatalf("handleMCPMessage failed: %v", err)
	}

	respMap, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("expected map response, got %T", response)
	}

	if respMap["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", respMap["jsonrpc"])
	}

	if respMap["id"] != 3 {
		t.Errorf("expected id 3, got %v", respMap["id"])
	}

	result, ok := respMap["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", respMap["result"])
	}

	content, ok := result["content"].([]any)
	if !ok {
		t.Fatalf("expected content array, got %T", result["content"])
	}

	if len(content) != 1 {
		t.Errorf("expected 1 content item, got %d", len(content))
	}

	contentItem, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("expected content item map, got %T", content[0])
	}

	if contentItem["type"] != "text" {
		t.Errorf("expected type text, got %v", contentItem["type"])
	}

	if contentItem["text"] != "hello world" {
		t.Errorf("expected text 'hello world', got %v", contentItem["text"])
	}
}

// TestHandleMCPMessage_UnknownServer tests error handling for unknown server
func TestHandleMCPMessage_UnknownServer(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	mcpMessage := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	}

	response, err := query.handleMCPMessage("nonexistent-server", mcpMessage)
	if err != nil {
		t.Fatalf("handleMCPMessage failed: %v", err)
	}

	respMap, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("expected map response, got %T", response)
	}

	// Should have error
	mcpError, ok := respMap["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error in response, got %v", respMap)
	}

	code, ok := mcpError["code"].(int)
	if !ok {
		// Try float64 (JSON unmarshaling might give us this)
		codeFloat, okFloat := mcpError["code"].(float64)
		if okFloat {
			code = int(codeFloat)
		} else {
			t.Fatalf("expected int or float64 code, got %T: %v", mcpError["code"], mcpError["code"])
		}
	}
	if code != -32601 {
		t.Errorf("expected error code -32601, got %d", code)
	}

	message, ok := mcpError["message"].(string)
	if !ok || message == "" {
		t.Errorf("expected error message, got %v", mcpError["message"])
	}
}

// TestHandleControlRequest_MCPMessage tests the control protocol integration
func TestHandleControlRequest_MCPMessage(t *testing.T) {
	server := types.NewMCPServerBuilder("test-server").
		WithTool("test", "Test tool", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "success"}},
			}, nil
		}).
		Build()

	transport := NewMockTransport()
	query := NewQuery(transport, true)
	query.RegisterMCPServer(server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := query.Start(ctx); err != nil {
		t.Fatalf("failed to start query: %v", err)
	}
	defer query.Close()

	// Simulate a control request with mcp_message subtype
	controlRequest := map[string]any{
		"type":       "control_request",
		"request_id": "test-req-1",
		"request": map[string]any{
			"subtype":     "mcp_message",
			"server_name": "test-server",
			"message": map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "tools/list",
			},
		},
	}

	// Send the control request
	go query.handleControlRequest(controlRequest)

	// Wait for response to be written
	time.Sleep(100 * time.Millisecond)

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no response written")
	}

	// Parse response
	var response map[string]any
	if err := json.Unmarshal([]byte(written[0]), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["type"] != "control_response" {
		t.Errorf("expected control_response, got %v", response["type"])
	}

	responseData, ok := response["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected response map, got %T", response["response"])
	}

	if responseData["subtype"] != "success" {
		t.Errorf("expected success subtype, got %v", responseData["subtype"])
	}

	// Check the mcp_response
	respPayload, ok := responseData["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected response payload map, got %T", responseData["response"])
	}

	mcpResponse, ok := respPayload["mcp_response"].(map[string]any)
	if !ok {
		t.Fatalf("expected mcp_response map, got %T", respPayload["mcp_response"])
	}

	if mcpResponse["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", mcpResponse["jsonrpc"])
	}

	result, ok := mcpResponse["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", mcpResponse["result"])
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got %T", result["tools"])
	}

	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}
}

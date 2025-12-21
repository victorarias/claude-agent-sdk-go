package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestMCPServerTransport_ProcessOne(t *testing.T) {
	server := types.NewMCPServerBuilder("test").
		WithTool("add", "Add numbers", nil, func(input map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "result"}},
			}, nil
		}).
		Build()

	// Create buffers for testing
	clientToServer := &bytes.Buffer{}
	serverToClient := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, clientToServer, serverToClient)

	// Write initialize request
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
		},
	}
	reqBytes, _ := json.Marshal(initReq)
	clientToServer.Write(reqBytes)
	clientToServer.Write([]byte("\n"))

	// Process one request
	if err := transport.ProcessOne(); err != nil {
		t.Fatalf("ProcessOne failed: %v", err)
	}

	// Check response
	respLine, _ := io.ReadAll(serverToClient)
	var resp MCPResponse
	if err := json.Unmarshal(bytes.TrimSpace(respLine), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCPServerTransport_ToolsCall(t *testing.T) {
	server := types.NewMCPServerBuilder("test").
		WithTool("greet", "Greet someone", nil, func(input map[string]any) (*types.MCPToolResult, error) {
			name := input["name"].(string)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "Hello, " + name + "!"}},
			}, nil
		}).
		Build()

	clientToServer := &bytes.Buffer{}
	serverToClient := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, clientToServer, serverToClient)

	// Write tools/call request
	callReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "greet",
			"arguments": map[string]any{"name": "World"},
		},
	}
	reqBytes, _ := json.Marshal(callReq)
	clientToServer.Write(reqBytes)
	clientToServer.Write([]byte("\n"))

	// Process request
	if err := transport.ProcessOne(); err != nil {
		t.Fatalf("ProcessOne failed: %v", err)
	}

	// Check response contains greeting
	respBytes := serverToClient.Bytes()
	if !bytes.Contains(respBytes, []byte("Hello, World!")) {
		t.Errorf("Expected response to contain greeting, got: %s", string(respBytes))
	}
}

func TestMCPServerTransport_MultipleRequests(t *testing.T) {
	server := types.NewMCPServerBuilder("test").
		WithTool("echo", "Echo input", nil, func(input map[string]any) (*types.MCPToolResult, error) {
			text := input["text"].(string)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: text}},
			}, nil
		}).
		Build()

	clientToServer := &bytes.Buffer{}
	serverToClient := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, clientToServer, serverToClient)

	// Write multiple requests
	for i := 1; i <= 3; i++ {
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      i,
			Method:  "tools/call",
			Params: map[string]any{
				"name":      "echo",
				"arguments": map[string]any{"text": "message"},
			},
		}
		reqBytes, _ := json.Marshal(req)
		clientToServer.Write(reqBytes)
		clientToServer.Write([]byte("\n"))
	}

	// Process all requests
	for i := 0; i < 3; i++ {
		if err := transport.ProcessOne(); err != nil {
			t.Fatalf("ProcessOne %d failed: %v", i, err)
		}
	}

	// Count responses (should be 3 lines)
	responseCount := bytes.Count(serverToClient.Bytes(), []byte("\n"))
	if responseCount != 3 {
		t.Errorf("Expected 3 responses, got %d", responseCount)
	}
}

func TestMCPServerTransport_Run(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()

	// Create a reader that will return EOF
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, input, output)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should exit gracefully when input is exhausted
	err := transport.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestMCPServerTransport_RunContextCancel(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()

	// Create a pipe so we can control the input
	pr, _ := io.Pipe()
	output := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, pr, output)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- transport.Run(ctx)
	}()

	// Cancel immediately
	cancel()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("Unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Transport did not exit after context cancel")
	}
}

func TestMCPServerTransport_EmptyLines(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()

	clientToServer := bytes.NewBufferString("\n\n\n")
	serverToClient := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, clientToServer, serverToClient)

	// Process empty lines - should not produce errors
	for i := 0; i < 3; i++ {
		if err := transport.ProcessOne(); err != nil && err != io.EOF {
			t.Fatalf("ProcessOne failed on empty line: %v", err)
		}
	}

	// No responses should be written for empty lines
	if serverToClient.Len() != 0 {
		t.Errorf("Expected no output for empty lines, got: %s", serverToClient.String())
	}
}

func TestMCPServerTransport_InvalidJSON(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()

	clientToServer := bytes.NewBufferString("not valid json\n")
	serverToClient := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, clientToServer, serverToClient)

	// Process invalid JSON - should return error response
	if err := transport.ProcessOne(); err != nil {
		t.Fatalf("ProcessOne failed: %v", err)
	}

	// Should have an error response
	var resp MCPResponse
	if err := json.Unmarshal(bytes.TrimSpace(serverToClient.Bytes()), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("Expected error response for invalid JSON")
	}

	if resp.Error.Code != MCPErrorParseError {
		t.Errorf("Expected parse error code, got %d", resp.Error.Code)
	}
}

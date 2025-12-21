# Plan 06: MCP Server Support

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement SDK-hosted MCP (Model Context Protocol) server support for custom tools.

**Architecture:** Enable developers to define custom tools that Claude can use, hosted in-process by the SDK. Tools are exposed via MCP protocol over stdio, allowing bidirectional communication for tool invocation and result handling.

**Tech Stack:** Go 1.21+

**Dependencies:** Plans 01-04 must be complete. Plan 03 defines the MCP integration points.

---

## Task 0: MCP Protocol Types

**Files:**
- Create: `mcp_types.go`
- Create: `mcp_types_test.go`

**Step 1: Write test**

Create `mcp_types_test.go`:

```go
package sdk

import (
	"encoding/json"
	"testing"
)

func TestMCPToolDefinition_JSON(t *testing.T) {
	tool := MCPToolDefinition{
		Name:        "calculate",
		Description: "Perform calculations",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "Math expression",
				},
			},
			"required": []string{"expression"},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MCPToolDefinition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != tool.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, tool.Name)
	}
}

func TestMCPRequest_JSON(t *testing.T) {
	tests := []struct {
		name    string
		request MCPRequest
	}{
		{
			name: "initialize",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params: map[string]any{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]any{},
					"clientInfo": map[string]any{
						"name":    "claude-agent-sdk-go",
						"version": Version,
					},
				},
			},
		},
		{
			name: "tools/list",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/list",
			},
		},
		{
			name: "tools/call",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      3,
				Method:  "tools/call",
				Params: map[string]any{
					"name": "calculate",
					"arguments": map[string]any{
						"expression": "2+2",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var decoded MCPRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if decoded.Method != tt.request.Method {
				t.Errorf("Method mismatch: got %s, want %s", decoded.Method, tt.request.Method)
			}
		})
	}
}

func TestMCPResponse_JSON(t *testing.T) {
	tests := []struct {
		name     string
		response MCPResponse
	}{
		{
			name: "success",
			response: MCPResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result: map[string]any{
					"content": []any{
						map[string]any{
							"type": "text",
							"text": "Result: 4",
						},
					},
				},
			},
		},
		{
			name: "error",
			response: MCPResponse{
				JSONRPC: "2.0",
				ID:      2,
				Error: &MCPError{
					Code:    -32600,
					Message: "Invalid Request",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var decoded MCPResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if tt.response.Error != nil && decoded.Error == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestMCP -v
```

Expected: Compilation error (types don't exist)

**Step 3: Implement types**

Create `mcp_types.go`:

```go
package sdk

import "encoding/json"

// MCP Protocol Version
const MCPProtocolVersion = "2024-11-05"

// MCPToolDefinition defines a tool exposed via MCP.
type MCPToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPToolHandler is a function that handles tool invocations.
type MCPToolHandler func(input map[string]any) (any, error)

// MCPTool combines a tool definition with its handler.
type MCPTool struct {
	Definition MCPToolDefinition
	Handler    MCPToolHandler
}

// MCPRequest represents a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard MCP error codes
const (
	MCPErrorParseError     = -32700
	MCPErrorInvalidRequest = -32600
	MCPErrorMethodNotFound = -32601
	MCPErrorInvalidParams  = -32602
	MCPErrorInternal       = -32603
)

// MCPNotification represents a JSON-RPC 2.0 notification (no ID).
type MCPNotification struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

// MCPServerInfo contains server metadata.
type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCPClientInfo contains client metadata.
type MCPClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCPCapabilities describes server capabilities.
type MCPCapabilities struct {
	Tools     *MCPToolsCapability     `json:"tools,omitempty"`
	Resources *MCPResourcesCapability `json:"resources,omitempty"`
	Prompts   *MCPPromptsCapability   `json:"prompts,omitempty"`
}

// MCPToolsCapability describes tool capabilities.
type MCPToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// MCPResourcesCapability describes resource capabilities.
type MCPResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// MCPPromptsCapability describes prompt capabilities.
type MCPPromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// MCPInitializeResult is the result of initialize request.
type MCPInitializeResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    MCPCapabilities `json:"capabilities"`
	ServerInfo      MCPServerInfo   `json:"serverInfo"`
}

// MCPToolCallParams contains parameters for tools/call.
type MCPToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// MCPToolResult contains the result of a tool call.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in tool results.
type MCPContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"` // Base64 for embedded resources
}

// NewMCPTextContent creates a text content item.
func NewMCPTextContent(text string) MCPContent {
	return MCPContent{
		Type: "text",
		Text: text,
	}
}

// NewMCPErrorContent creates an error content item.
func NewMCPErrorContent(err error) MCPContent {
	return MCPContent{
		Type: "text",
		Text: "Error: " + err.Error(),
	}
}

// MCPToolsListResult is the result of tools/list.
type MCPToolsListResult struct {
	Tools []MCPToolDefinition `json:"tools"`
}

// ParseMCPRequest parses a JSON-RPC request from bytes.
func ParseMCPRequest(data []byte) (*MCPRequest, error) {
	var req MCPRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// MarshalMCPResponse serializes a response to JSON.
func MarshalMCPResponse(resp *MCPResponse) ([]byte, error) {
	return json.Marshal(resp)
}
```

**Step 4: Run test**

```bash
go test -run TestMCP -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add mcp_types.go mcp_types_test.go
git commit -m "feat: add MCP protocol types"
```

---

## Task 1: MCP Server Builder

**Files:**
- Create: `mcp_builder.go`
- Create: `mcp_builder_test.go`

**Step 1: Write test**

Create `mcp_builder_test.go`:

```go
package sdk

import (
	"testing"
)

func TestMCPServerBuilder(t *testing.T) {
	server := NewMCPServerBuilder("test-server").
		WithDescription("A test server").
		WithVersion("1.0.0").
		WithTool("greet", "Greet someone", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name to greet",
				},
			},
			"required": []string{"name"},
		}, func(input map[string]any) (any, error) {
			name := input["name"].(string)
			return "Hello, " + name + "!", nil
		}).
		Build()

	if server.Name != "test-server" {
		t.Errorf("Name mismatch: got %s, want test-server", server.Name)
	}

	if server.Description != "A test server" {
		t.Errorf("Description mismatch: got %s", server.Description)
	}

	if len(server.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(server.Tools))
	}

	tool := server.Tools["greet"]
	if tool.Definition.Name != "greet" {
		t.Errorf("Tool name mismatch: got %s", tool.Definition.Name)
	}

	// Test handler
	result, err := tool.Handler(map[string]any{"name": "World"})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Handler result mismatch: got %v", result)
	}
}

func TestMCPServerBuilder_MultipleTools(t *testing.T) {
	server := NewMCPServerBuilder("multi-tool").
		WithTool("add", "Add numbers", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{"type": "number"},
				"b": map[string]any{"type": "number"},
			},
		}, func(input map[string]any) (any, error) {
			a := input["a"].(float64)
			b := input["b"].(float64)
			return a + b, nil
		}).
		WithTool("multiply", "Multiply numbers", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{"type": "number"},
				"b": map[string]any{"type": "number"},
			},
		}, func(input map[string]any) (any, error) {
			a := input["a"].(float64)
			b := input["b"].(float64)
			return a * b, nil
		}).
		Build()

	if len(server.Tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(server.Tools))
	}

	if _, ok := server.Tools["add"]; !ok {
		t.Error("Missing 'add' tool")
	}
	if _, ok := server.Tools["multiply"]; !ok {
		t.Error("Missing 'multiply' tool")
	}
}

func TestMCPServerBuilder_ListTools(t *testing.T) {
	server := NewMCPServerBuilder("list-test").
		WithTool("tool1", "First tool", nil, nil).
		WithTool("tool2", "Second tool", nil, nil).
		Build()

	tools := server.ListTools()
	if len(tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(tools))
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}

	if !names["tool1"] || !names["tool2"] {
		t.Error("Missing expected tools in list")
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestMCPServerBuilder -v
```

Expected: Compilation error

**Step 3: Implement builder**

Create `mcp_builder.go`:

```go
package sdk

// MCPServer represents an SDK-hosted MCP server.
type MCPServer struct {
	Name        string
	Description string
	Version     string
	Tools       map[string]*MCPTool
}

// ListTools returns all tool definitions.
func (s *MCPServer) ListTools() []MCPToolDefinition {
	tools := make([]MCPToolDefinition, 0, len(s.Tools))
	for _, tool := range s.Tools {
		tools = append(tools, tool.Definition)
	}
	return tools
}

// GetTool returns a tool by name.
func (s *MCPServer) GetTool(name string) (*MCPTool, bool) {
	tool, ok := s.Tools[name]
	return tool, ok
}

// CallTool invokes a tool by name with the given input.
func (s *MCPServer) CallTool(name string, input map[string]any) (any, error) {
	tool, ok := s.Tools[name]
	if !ok {
		return nil, &MCPError{
			Code:    MCPErrorMethodNotFound,
			Message: "tool not found: " + name,
		}
	}
	return tool.Handler(input)
}

// MCPServerBuilder provides a fluent API for building MCP servers.
type MCPServerBuilder struct {
	server *MCPServer
}

// NewMCPServerBuilder creates a new builder with the given server name.
func NewMCPServerBuilder(name string) *MCPServerBuilder {
	return &MCPServerBuilder{
		server: &MCPServer{
			Name:    name,
			Version: Version,
			Tools:   make(map[string]*MCPTool),
		},
	}
}

// WithDescription sets the server description.
func (b *MCPServerBuilder) WithDescription(desc string) *MCPServerBuilder {
	b.server.Description = desc
	return b
}

// WithVersion sets the server version.
func (b *MCPServerBuilder) WithVersion(version string) *MCPServerBuilder {
	b.server.Version = version
	return b
}

// WithTool adds a tool to the server.
func (b *MCPServerBuilder) WithTool(name, description string, schema map[string]any, handler MCPToolHandler) *MCPServerBuilder {
	if schema == nil {
		schema = map[string]any{"type": "object"}
	}

	b.server.Tools[name] = &MCPTool{
		Definition: MCPToolDefinition{
			Name:        name,
			Description: description,
			InputSchema: schema,
		},
		Handler: handler,
	}
	return b
}

// Build returns the configured MCP server.
func (b *MCPServerBuilder) Build() *MCPServer {
	return b.server
}
```

**Step 4: Run test**

```bash
go test -run TestMCPServerBuilder -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add mcp_builder.go mcp_builder_test.go
git commit -m "feat: add MCP server builder"
```

---

## Task 2: MCP Message Handler

**Files:**
- Create: `mcp_handler.go`
- Create: `mcp_handler_test.go`

**Step 1: Write test**

Create `mcp_handler_test.go`:

```go
package sdk

import (
	"encoding/json"
	"testing"
)

func TestMCPHandler_Initialize(t *testing.T) {
	server := NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(MCPInitializeResult)
	if !ok {
		t.Fatalf("Expected MCPInitializeResult, got %T", resp.Result)
	}

	if result.ProtocolVersion != MCPProtocolVersion {
		t.Errorf("Protocol version mismatch: got %s", result.ProtocolVersion)
	}
}

func TestMCPHandler_ToolsList(t *testing.T) {
	server := NewMCPServerBuilder("test").
		WithTool("greet", "Greet someone", nil, nil).
		WithTool("calculate", "Do math", nil, nil).
		Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(MCPToolsListResult)
	if !ok {
		t.Fatalf("Expected MCPToolsListResult, got %T", resp.Result)
	}

	if len(result.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(result.Tools))
	}
}

func TestMCPHandler_ToolsCall(t *testing.T) {
	server := NewMCPServerBuilder("test").
		WithTool("greet", "Greet someone", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}, func(input map[string]any) (any, error) {
			name := input["name"].(string)
			return "Hello, " + name + "!", nil
		}).
		Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "greet",
			"arguments": map[string]any{
				"name": "World",
			},
		},
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(MCPToolResult)
	if !ok {
		t.Fatalf("Expected MCPToolResult, got %T", resp.Result)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	// Check content contains greeting
	found := false
	for _, c := range result.Content {
		if c.Type == "text" && c.Text == "Hello, World!" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected greeting in result, got: %v", result.Content)
	}
}

func TestMCPHandler_ToolsCall_NotFound(t *testing.T) {
	server := NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "nonexistent",
		},
	}

	resp := handler.HandleRequest(req)
	if resp.Error == nil {
		t.Fatal("Expected error for nonexistent tool")
	}

	if resp.Error.Code != MCPErrorMethodNotFound {
		t.Errorf("Expected method not found error, got code %d", resp.Error.Code)
	}
}

func TestMCPHandler_UnknownMethod(t *testing.T) {
	server := NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	resp := handler.HandleRequest(req)
	if resp.Error == nil {
		t.Fatal("Expected error for unknown method")
	}

	if resp.Error.Code != MCPErrorMethodNotFound {
		t.Errorf("Expected method not found error, got code %d", resp.Error.Code)
	}
}

func TestMCPHandler_HandleBytes(t *testing.T) {
	server := NewMCPServerBuilder("test").
		WithTool("echo", "Echo input", nil, func(input map[string]any) (any, error) {
			return input["text"], nil
		}).
		Build()
	handler := NewMCPHandler(server)

	reqBytes, _ := json.Marshal(&MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "echo",
			"arguments": map[string]any{"text": "hello"},
		},
	})

	respBytes, err := handler.HandleBytes(reqBytes)
	if err != nil {
		t.Fatalf("HandleBytes failed: %v", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestMCPHandler -v
```

Expected: Compilation error

**Step 3: Implement handler**

Create `mcp_handler.go`:

```go
package sdk

import (
	"encoding/json"
	"fmt"
)

// MCPHandler processes MCP protocol messages for a server.
type MCPHandler struct {
	server      *MCPServer
	initialized bool
}

// NewMCPHandler creates a handler for the given server.
func NewMCPHandler(server *MCPServer) *MCPHandler {
	return &MCPHandler{
		server: server,
	}
}

// HandleRequest processes an MCP request and returns a response.
func (h *MCPHandler) HandleRequest(req *MCPRequest) *MCPResponse {
	resp := &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = h.handleInitialize(req.Params)
		h.initialized = true

	case "initialized":
		// Notification, no response needed
		return nil

	case "tools/list":
		resp.Result = h.handleToolsList()

	case "tools/call":
		result, err := h.handleToolsCall(req.Params)
		if err != nil {
			if mcpErr, ok := err.(*MCPError); ok {
				resp.Error = mcpErr
			} else {
				resp.Error = &MCPError{
					Code:    MCPErrorInternal,
					Message: err.Error(),
				}
			}
		} else {
			resp.Result = result
		}

	case "ping":
		resp.Result = map[string]any{}

	default:
		resp.Error = &MCPError{
			Code:    MCPErrorMethodNotFound,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		}
	}

	return resp
}

// HandleBytes processes raw JSON bytes and returns raw JSON response.
func (h *MCPHandler) HandleBytes(data []byte) ([]byte, error) {
	req, err := ParseMCPRequest(data)
	if err != nil {
		resp := &MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    MCPErrorParseError,
				Message: "failed to parse request",
				Data:    err.Error(),
			},
		}
		return MarshalMCPResponse(resp)
	}

	resp := h.HandleRequest(req)
	if resp == nil {
		// Notification, no response
		return nil, nil
	}

	return MarshalMCPResponse(resp)
}

func (h *MCPHandler) handleInitialize(params map[string]any) MCPInitializeResult {
	return MCPInitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: MCPCapabilities{
			Tools: &MCPToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: MCPServerInfo{
			Name:    h.server.Name,
			Version: h.server.Version,
		},
	}
}

func (h *MCPHandler) handleToolsList() MCPToolsListResult {
	return MCPToolsListResult{
		Tools: h.server.ListTools(),
	}
}

func (h *MCPHandler) handleToolsCall(params map[string]any) (MCPToolResult, error) {
	// Extract tool name
	name, ok := params["name"].(string)
	if !ok {
		return MCPToolResult{}, &MCPError{
			Code:    MCPErrorInvalidParams,
			Message: "missing or invalid 'name' parameter",
		}
	}

	// Extract arguments
	args, _ := params["arguments"].(map[string]any)
	if args == nil {
		args = make(map[string]any)
	}

	// Find and call tool
	tool, exists := h.server.GetTool(name)
	if !exists {
		return MCPToolResult{}, &MCPError{
			Code:    MCPErrorMethodNotFound,
			Message: fmt.Sprintf("tool not found: %s", name),
		}
	}

	// Execute handler
	result, err := tool.Handler(args)
	if err != nil {
		return MCPToolResult{
			Content: []MCPContent{NewMCPErrorContent(err)},
			IsError: true,
		}, nil
	}

	// Convert result to content
	content := resultToContent(result)
	return MCPToolResult{
		Content: content,
		IsError: false,
	}, nil
}

// resultToContent converts a handler result to MCP content.
func resultToContent(result any) []MCPContent {
	switch v := result.(type) {
	case string:
		return []MCPContent{NewMCPTextContent(v)}
	case []MCPContent:
		return v
	case MCPContent:
		return []MCPContent{v}
	default:
		// Marshal to JSON
		data, err := json.Marshal(v)
		if err != nil {
			return []MCPContent{NewMCPTextContent(fmt.Sprintf("%v", v))}
		}
		return []MCPContent{NewMCPTextContent(string(data))}
	}
}
```

**Step 4: Run test**

```bash
go test -run TestMCPHandler -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add mcp_handler.go mcp_handler_test.go
git commit -m "feat: add MCP message handler"
```

---

## Task 3: MCP Server Transport

**Files:**
- Create: `mcp_transport.go`
- Create: `mcp_transport_test.go`

**Step 1: Write test**

Create `mcp_transport_test.go`:

```go
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
)

func TestMCPServerTransport_Process(t *testing.T) {
	server := NewMCPServerBuilder("test").
		WithTool("add", "Add numbers", nil, func(input map[string]any) (any, error) {
			a := input["a"].(float64)
			b := input["b"].(float64)
			return a + b, nil
		}).
		Build()

	// Create pipes for testing
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

func TestMCPServerTransport_RunContext(t *testing.T) {
	server := NewMCPServerBuilder("test").Build()

	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	transport := NewMCPServerTransport(server, input, output)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should exit gracefully when context is cancelled
	err := transport.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Unexpected error: %v", err)
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestMCPServerTransport -v
```

Expected: Compilation error

**Step 3: Implement transport**

Create `mcp_transport.go`:

```go
package sdk

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"
)

// MCPServerTransport handles stdio communication for an MCP server.
type MCPServerTransport struct {
	handler *MCPHandler
	reader  *bufio.Reader
	writer  io.Writer
	mu      sync.Mutex
}

// NewMCPServerTransport creates a new transport for the given server.
func NewMCPServerTransport(server *MCPServer, input io.Reader, output io.Writer) *MCPServerTransport {
	return &MCPServerTransport{
		handler: NewMCPHandler(server),
		reader:  bufio.NewReader(input),
		writer:  output,
	}
}

// Run processes messages until context is cancelled or input is closed.
func (t *MCPServerTransport) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := t.ProcessOne()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				// Log error but continue processing
				continue
			}
		}
	}
}

// ProcessOne reads and handles a single message.
func (t *MCPServerTransport) ProcessOne() error {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// Skip empty lines
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	// Handle the request
	respBytes, err := t.handler.HandleBytes(line)
	if err != nil {
		return err
	}

	// Write response if present
	if respBytes != nil {
		t.mu.Lock()
		_, err := t.writer.Write(respBytes)
		if err == nil {
			_, err = t.writer.Write([]byte("\n"))
		}
		t.mu.Unlock()
		return err
	}

	return nil
}

// bytes is imported via bufio
var bytes = struct {
	TrimSpace func([]byte) []byte
}{
	TrimSpace: func(b []byte) []byte {
		start := 0
		end := len(b)
		for start < end && (b[start] == ' ' || b[start] == '\t' || b[start] == '\n' || b[start] == '\r') {
			start++
		}
		for end > start && (b[end-1] == ' ' || b[end-1] == '\t' || b[end-1] == '\n' || b[end-1] == '\r') {
			end--
		}
		return b[start:end]
	},
}
```

**Step 4: Fix imports and run test**

Update `mcp_transport.go` to use proper bytes import:

```go
package sdk

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"
)

// MCPServerTransport handles stdio communication for an MCP server.
type MCPServerTransport struct {
	handler *MCPHandler
	reader  *bufio.Reader
	writer  io.Writer
	mu      sync.Mutex
}

// NewMCPServerTransport creates a new transport for the given server.
func NewMCPServerTransport(server *MCPServer, input io.Reader, output io.Writer) *MCPServerTransport {
	return &MCPServerTransport{
		handler: NewMCPHandler(server),
		reader:  bufio.NewReader(input),
		writer:  output,
	}
}

// Run processes messages until context is cancelled or input is closed.
func (t *MCPServerTransport) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := t.ProcessOne()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				// Log error but continue processing
				continue
			}
		}
	}
}

// ProcessOne reads and handles a single message.
func (t *MCPServerTransport) ProcessOne() error {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// Skip empty lines
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	// Handle the request
	respBytes, err := t.handler.HandleBytes(line)
	if err != nil {
		return err
	}

	// Write response if present
	if respBytes != nil {
		t.mu.Lock()
		_, err := t.writer.Write(respBytes)
		if err == nil {
			_, err = t.writer.Write([]byte("\n"))
		}
		t.mu.Unlock()
		return err
	}

	return nil
}
```

```bash
go test -run TestMCPServerTransport -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add mcp_transport.go mcp_transport_test.go
git commit -m "feat: add MCP server transport"
```

---

## Task 4: MCP Server Integration with Client

**Files:**
- Modify: `client.go`
- Create: `client_mcp_test.go`

**Step 1: Write test**

Create `client_mcp_test.go`:

```go
package sdk

import (
	"testing"
)

func TestClient_WithMCPServer(t *testing.T) {
	server := NewMCPServerBuilder("test-tools").
		WithTool("greet", "Greet someone", nil, func(input map[string]any) (any, error) {
			name := input["name"].(string)
			return "Hello, " + name + "!", nil
		}).
		Build()

	client := NewClient(
		WithMCPServer(server),
	)

	if client.mcpServers == nil {
		t.Fatal("mcpServers should not be nil")
	}

	if len(client.mcpServers) != 1 {
		t.Errorf("Expected 1 MCP server, got %d", len(client.mcpServers))
	}

	if _, ok := client.mcpServers["test-tools"]; !ok {
		t.Error("MCP server 'test-tools' not found")
	}
}

func TestClient_MultipleMCPServers(t *testing.T) {
	server1 := NewMCPServerBuilder("math").
		WithTool("add", "Add numbers", nil, nil).
		Build()

	server2 := NewMCPServerBuilder("text").
		WithTool("upper", "Uppercase text", nil, nil).
		Build()

	client := NewClient(
		WithMCPServer(server1),
		WithMCPServer(server2),
	)

	if len(client.mcpServers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(client.mcpServers))
	}
}

func TestClient_MCPServerConfig(t *testing.T) {
	server := NewMCPServerBuilder("custom").
		WithDescription("Custom tools").
		WithVersion("2.0.0").
		WithTool("custom_tool", "A custom tool", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
		}, func(input map[string]any) (any, error) {
			return input["input"], nil
		}).
		Build()

	client := NewClient(
		WithMCPServer(server),
	)

	// Verify server configuration is preserved
	registeredServer := client.mcpServers["custom"]
	if registeredServer.Description != "Custom tools" {
		t.Errorf("Description mismatch: got %s", registeredServer.Description)
	}
	if registeredServer.Version != "2.0.0" {
		t.Errorf("Version mismatch: got %s", registeredServer.Version)
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestClient -v
```

Expected: Compilation error (mcpServers field doesn't exist)

**Step 3: Add MCP support to Client**

This step modifies existing client.go to add MCP server support. Add the following to the Client struct and options:

```go
// In options.go, add:

// WithMCPServer adds an SDK-hosted MCP server.
func WithMCPServer(server *MCPServer) Option {
	return func(o *Options) {
		if o.MCPServers == nil {
			o.MCPServers = make(map[string]*MCPServer)
		}
		o.MCPServers[server.Name] = server
	}
}

// In client.go, add to Client struct:
// mcpServers map[string]*MCPServer

// In NewClient, initialize:
// mcpServers: opts.MCPServers,
```

**Step 4: Run test**

```bash
go test -run TestClient -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add client.go options.go client_mcp_test.go
git commit -m "feat: integrate MCP servers with client"
```

---

## Task 5: MCP Tool Invocation Flow

**Files:**
- Modify: `query.go`
- Create: `query_mcp_test.go`

**Step 1: Write test**

Create `query_mcp_test.go`:

```go
package sdk

import (
	"testing"
)

func TestQuery_MCPToolCall(t *testing.T) {
	// Create MCP server with a tool
	server := NewMCPServerBuilder("test").
		WithTool("calculate", "Calculate expression", nil, func(input map[string]any) (any, error) {
			expr := input["expression"].(string)
			if expr == "2+2" {
				return "4", nil
			}
			return nil, &MCPError{Code: -1, Message: "unknown expression"}
		}).
		Build()

	// Create mock query with MCP server
	query := &Query{
		mcpServers: map[string]*MCPServer{"test": server},
	}

	// Simulate MCP tool call request
	input := map[string]any{
		"expression": "2+2",
	}

	result, err := query.handleMCPToolCall("test", "calculate", input)
	if err != nil {
		t.Fatalf("handleMCPToolCall failed: %v", err)
	}

	// Check result
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestQuery_MCPToolCall_ServerNotFound(t *testing.T) {
	query := &Query{
		mcpServers: map[string]*MCPServer{},
	}

	_, err := query.handleMCPToolCall("nonexistent", "tool", nil)
	if err == nil {
		t.Fatal("Expected error for nonexistent server")
	}
}

func TestQuery_MCPToolCall_ToolNotFound(t *testing.T) {
	server := NewMCPServerBuilder("test").Build()

	query := &Query{
		mcpServers: map[string]*MCPServer{"test": server},
	}

	_, err := query.handleMCPToolCall("test", "nonexistent", nil)
	if err == nil {
		t.Fatal("Expected error for nonexistent tool")
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestQuery_MCP -v
```

Expected: Compilation error

**Step 3: Implement MCP tool call handling**

Add to `query.go`:

```go
// handleMCPToolCall processes a tool call to an SDK-hosted MCP server.
func (q *Query) handleMCPToolCall(serverName, toolName string, input map[string]any) (any, error) {
	server, ok := q.mcpServers[serverName]
	if !ok {
		return nil, &MCPError{
			Code:    MCPErrorMethodNotFound,
			Message: "MCP server not found: " + serverName,
		}
	}

	return server.CallTool(toolName, input)
}
```

**Step 4: Run test**

```bash
go test -run TestQuery_MCP -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add query.go query_mcp_test.go
git commit -m "feat: add MCP tool call handling to query"
```

---

## Task 6: MCP Server Lifecycle Management

**Files:**
- Create: `mcp_lifecycle.go`
- Create: `mcp_lifecycle_test.go`

**Step 1: Write test**

Create `mcp_lifecycle_test.go`:

```go
package sdk

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMCPServerManager_Start(t *testing.T) {
	server := NewMCPServerBuilder("test").Build()

	manager := NewMCPServerManager()
	manager.Register(server)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify server is running
	if !manager.IsRunning("test") {
		t.Error("Server should be running")
	}

	manager.Stop()
}

func TestMCPServerManager_MultipleServers(t *testing.T) {
	server1 := NewMCPServerBuilder("server1").Build()
	server2 := NewMCPServerBuilder("server2").Build()

	manager := NewMCPServerManager()
	manager.Register(server1)
	manager.Register(server2)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !manager.IsRunning("server1") {
		t.Error("server1 should be running")
	}
	if !manager.IsRunning("server2") {
		t.Error("server2 should be running")
	}

	manager.Stop()

	if manager.IsRunning("server1") {
		t.Error("server1 should be stopped")
	}
}

func TestMCPServerManager_Concurrent(t *testing.T) {
	server := NewMCPServerBuilder("test").
		WithTool("echo", "Echo input", nil, func(input map[string]any) (any, error) {
			return input["text"], nil
		}).
		Build()

	manager := NewMCPServerManager()
	manager.Register(server)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	// Concurrent calls
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := manager.CallTool("test", "echo", map[string]any{
				"text": n,
			})
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent call failed: %v", err)
	}
}
```

**Step 2: Run test (should fail)**

```bash
go test -run TestMCPServerManager -v
```

Expected: Compilation error

**Step 3: Implement lifecycle management**

Create `mcp_lifecycle.go`:

```go
package sdk

import (
	"context"
	"sync"
)

// MCPServerManager manages the lifecycle of SDK-hosted MCP servers.
type MCPServerManager struct {
	servers  map[string]*MCPServer
	handlers map[string]*MCPHandler
	running  map[string]bool
	mu       sync.RWMutex
}

// NewMCPServerManager creates a new server manager.
func NewMCPServerManager() *MCPServerManager {
	return &MCPServerManager{
		servers:  make(map[string]*MCPServer),
		handlers: make(map[string]*MCPHandler),
		running:  make(map[string]bool),
	}
}

// Register adds an MCP server to the manager.
func (m *MCPServerManager) Register(server *MCPServer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.servers[server.Name] = server
	m.handlers[server.Name] = NewMCPHandler(server)
}

// Start initializes all registered servers.
func (m *MCPServerManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name := range m.servers {
		m.running[name] = true
	}

	return nil
}

// Stop shuts down all servers.
func (m *MCPServerManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name := range m.running {
		m.running[name] = false
	}
}

// IsRunning checks if a server is running.
func (m *MCPServerManager) IsRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.running[name]
}

// GetServer returns a server by name.
func (m *MCPServerManager) GetServer(name string) (*MCPServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, ok := m.servers[name]
	return server, ok
}

// CallTool invokes a tool on a server.
func (m *MCPServerManager) CallTool(serverName, toolName string, input map[string]any) (any, error) {
	m.mu.RLock()
	server, ok := m.servers[serverName]
	m.mu.RUnlock()

	if !ok {
		return nil, &MCPError{
			Code:    MCPErrorMethodNotFound,
			Message: "server not found: " + serverName,
		}
	}

	return server.CallTool(toolName, input)
}

// ListServers returns names of all registered servers.
func (m *MCPServerManager) ListServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.servers))
	for name := range m.servers {
		names = append(names, name)
	}
	return names
}

// ListTools returns all tools across all servers.
func (m *MCPServerManager) ListTools() map[string][]MCPToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]MCPToolDefinition)
	for name, server := range m.servers {
		result[name] = server.ListTools()
	}
	return result
}
```

**Step 4: Run test**

```bash
go test -run TestMCPServerManager -v
```

Expected: All tests pass

**Step 5: Commit**

```bash
git add mcp_lifecycle.go mcp_lifecycle_test.go
git commit -m "feat: add MCP server lifecycle management"
```

---

## Summary

After completing Plan 06, you have:

**MCP Protocol Types (Task 0)**
- MCPToolDefinition, MCPToolHandler, MCPTool
- MCPRequest, MCPResponse, MCPNotification
- MCPError with standard error codes
- MCPInitializeResult, MCPToolResult, MCPContent
- Helper functions for parsing/marshaling

**MCP Server Builder (Task 1)**
- MCPServer structure with tool registry
- MCPServerBuilder with fluent API
- WithTool, WithDescription, WithVersion methods
- ListTools, GetTool, CallTool methods

**MCP Message Handler (Task 2)**
- MCPHandler for request processing
- Support for initialize, tools/list, tools/call, ping
- Error handling with proper MCP error codes
- Result to content conversion

**MCP Server Transport (Task 3)**
- MCPServerTransport for stdio communication
- Newline-delimited JSON protocol
- Context-aware Run method
- Thread-safe write operations

**Client Integration (Task 4)**
- WithMCPServer option for client configuration
- Multiple MCP server support
- Server registration and lookup

**Tool Invocation Flow (Task 5)**
- handleMCPToolCall in Query
- Server/tool validation
- Result forwarding

**Lifecycle Management (Task 6)**
- MCPServerManager for multi-server management
- Start/Stop methods
- Concurrent tool call support
- Server/tool listing

**Total: ~800 additional lines of Go code across 7 tasks**

---

## Usage Example

```go
// Create custom MCP server
server := sdk.NewMCPServerBuilder("my-tools").
    WithDescription("Custom calculation tools").
    WithTool("factorial", "Calculate factorial", map[string]any{
        "type": "object",
        "properties": map[string]any{
            "n": map[string]any{"type": "integer"},
        },
        "required": []string{"n"},
    }, func(input map[string]any) (any, error) {
        n := int(input["n"].(float64))
        result := 1
        for i := 2; i <= n; i++ {
            result *= i
        }
        return result, nil
    }).
    Build()

// Use with client
client := sdk.NewClient(
    sdk.WithMCPServer(server),
)

// Claude can now use the factorial tool!
messages, _ := client.Query(ctx, "What is 10 factorial?")
```

**Plan complete and saved.** Ready for execution!

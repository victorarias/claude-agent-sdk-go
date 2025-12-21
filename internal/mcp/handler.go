package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// MCPHandler processes MCP protocol messages for a server.
type MCPHandler struct {
	server      *types.MCPServer
	initialized bool
}

// NewMCPHandler creates a handler for the given server.
func NewMCPHandler(server *types.MCPServer) *MCPHandler {
	return &MCPHandler{
		server: server,
	}
}

// HandleRequest processes an MCP request and returns a response.
func (h *MCPHandler) HandleRequest(req *MCPRequest) *MCPResponse {
	// Check for notification (no ID means notification)
	if req.ID == nil {
		h.handleNotification(req)
		return nil
	}

	resp := &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = h.handleInitialize(req.Params)
		h.initialized = true

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

func (h *MCPHandler) handleNotification(req *MCPRequest) {
	// Handle notifications (no response expected)
	switch req.Method {
	case "notifications/initialized":
		// Client acknowledges initialization, nothing to do
	case "notifications/cancelled":
		// Request cancellation, we could track this if needed
	}
}

func (h *MCPHandler) handleInitialize(params map[string]any) *MCPInitializeResult {
	return &MCPInitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: &MCPCapabilities{
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

func (h *MCPHandler) handleToolsList() *MCPToolsListResult {
	tools := make([]MCPToolDefinition, 0, len(h.server.Tools))
	for _, tool := range h.server.Tools {
		schema := tool.Schema
		if schema == nil {
			schema = map[string]any{"type": "object"}
		}
		tools = append(tools, MCPToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: schema,
		})
	}
	return &MCPToolsListResult{Tools: tools}
}

func (h *MCPHandler) handleToolsCall(params map[string]any) (*MCPToolCallResult, error) {
	// Extract tool name
	name, ok := params["name"].(string)
	if !ok {
		return nil, &MCPError{
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
		return nil, &MCPError{
			Code:    MCPErrorMethodNotFound,
			Message: fmt.Sprintf("tool not found: %s", name),
		}
	}

	// Execute handler
	result, err := tool.Handler(args)
	if err != nil {
		return &MCPToolCallResult{
			Content: []MCPContent{{Type: "text", Text: "Error: " + err.Error()}},
			IsError: true,
		}, nil
	}

	return &MCPToolCallResult{
		Content: result.Content,
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

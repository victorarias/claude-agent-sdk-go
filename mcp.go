package sdk

import (
	"context"
	"fmt"
	"sync"
)

// MCPToolHandler defines a tool that can be called via MCP.
type MCPToolHandler struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(ctx context.Context, input map[string]any) (any, error)
}

// MCPServer represents an SDK-hosted MCP server.
type MCPServer struct {
	Name  string
	Tools []*MCPToolHandler

	// Internal state
	toolsByName map[string]*MCPToolHandler
	mu          sync.RWMutex
}

// NewMCPServer creates a new MCP server.
func NewMCPServer(name string, tools []*MCPToolHandler) *MCPServer {
	server := &MCPServer{
		Name:        name,
		Tools:       tools,
		toolsByName: make(map[string]*MCPToolHandler),
	}

	for _, tool := range tools {
		server.toolsByName[tool.Name] = tool
	}

	return server
}

// GetTool returns a tool by name.
func (s *MCPServer) GetTool(name string) (*MCPToolHandler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tool, ok := s.toolsByName[name]
	return tool, ok
}

// CallTool calls a tool by name with the given input.
func (s *MCPServer) CallTool(ctx context.Context, name string, input map[string]any) (any, error) {
	tool, ok := s.GetTool(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(ctx, input)
}

// ToConfig returns the MCP server configuration for the CLI.
func (s *MCPServer) ToConfig() map[string]any {
	tools := make([]map[string]any, len(s.Tools))
	for i, tool := range s.Tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return map[string]any{
		"name":  s.Name,
		"tools": tools,
	}
}

// MCPServerBuilder provides a fluent API for building MCP servers.
type MCPServerBuilder struct {
	name  string
	tools []*MCPToolHandler
}

// NewMCPServerBuilder creates a new MCP server builder.
func NewMCPServerBuilder(name string) *MCPServerBuilder {
	return &MCPServerBuilder{
		name:  name,
		tools: make([]*MCPToolHandler, 0),
	}
}

// WithTool adds a tool to the server.
func (b *MCPServerBuilder) WithTool(
	name string,
	description string,
	inputSchema map[string]any,
	handler func(ctx context.Context, input map[string]any) (any, error),
) *MCPServerBuilder {
	b.tools = append(b.tools, &MCPToolHandler{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Handler:     handler,
	})
	return b
}

// Build creates the MCP server.
func (b *MCPServerBuilder) Build() *MCPServer {
	return NewMCPServer(b.name, b.tools)
}

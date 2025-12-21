package sdk

import (
	"fmt"
	"sync"
)

// MCPServer helper methods

// mcpToolsByName caches tool lookups by name
var (
	mcpToolsByName   = make(map[*MCPServer]map[string]*MCPTool)
	mcpToolsByNameMu sync.RWMutex
)

// getToolsMap returns or creates the tools map for a server.
func (s *MCPServer) getToolsMap() map[string]*MCPTool {
	mcpToolsByNameMu.RLock()
	if m, ok := mcpToolsByName[s]; ok {
		mcpToolsByNameMu.RUnlock()
		return m
	}
	mcpToolsByNameMu.RUnlock()

	mcpToolsByNameMu.Lock()
	defer mcpToolsByNameMu.Unlock()

	// Double-check after acquiring write lock
	if m, ok := mcpToolsByName[s]; ok {
		return m
	}

	m := make(map[string]*MCPTool)
	for _, tool := range s.Tools {
		m[tool.Name] = tool
	}
	mcpToolsByName[s] = m
	return m
}

// GetTool returns a tool by name.
func (s *MCPServer) GetTool(name string) (*MCPTool, bool) {
	m := s.getToolsMap()
	tool, ok := m[name]
	return tool, ok
}

// CallTool calls a tool by name with the given input.
func (s *MCPServer) CallTool(name string, input map[string]any) (*MCPToolResult, error) {
	tool, ok := s.GetTool(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(input)
}

// ToConfig returns the MCP server configuration for the CLI.
func (s *MCPServer) ToConfig() map[string]any {
	tools := make([]map[string]any, len(s.Tools))
	for i, tool := range s.Tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.Schema,
		}
	}

	return map[string]any{
		"name":    s.Name,
		"version": s.Version,
		"tools":   tools,
	}
}

// MCPServerBuilder provides a fluent API for building MCP servers.
type MCPServerBuilder struct {
	name    string
	version string
	tools   []*MCPTool
}

// NewMCPServerBuilder creates a new MCP server builder.
func NewMCPServerBuilder(name string) *MCPServerBuilder {
	return &MCPServerBuilder{
		name:    name,
		version: "1.0.0",
		tools:   make([]*MCPTool, 0),
	}
}

// WithVersion sets the server version.
func (b *MCPServerBuilder) WithVersion(version string) *MCPServerBuilder {
	b.version = version
	return b
}

// WithTool adds a tool to the server.
func (b *MCPServerBuilder) WithTool(
	name string,
	description string,
	schema map[string]any,
	handler MCPToolHandler,
) *MCPServerBuilder {
	b.tools = append(b.tools, &MCPTool{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     handler,
	})
	return b
}

// Build creates the MCP server.
func (b *MCPServerBuilder) Build() *MCPServer {
	return &MCPServer{
		Name:    b.name,
		Version: b.version,
		Tools:   b.tools,
	}
}

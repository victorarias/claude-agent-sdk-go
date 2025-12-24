// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestExternalMCPServerTransport_Stdio verifies that stdio MCP servers
// are included in the --mcp-config flag.
func TestExternalMCPServerTransport_Stdio(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"filesystem": {
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			Env:     map[string]string{"DEBUG": "true"},
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Find the --mcp-config flag
	var mcpConfigValue string
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			mcpConfigValue = cmd[i+1]
			break
		}
	}

	if mcpConfigValue == "" {
		t.Fatal("Expected --mcp-config flag in command for stdio server")
	}

	// Parse and validate the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	server, ok := servers["filesystem"].(map[string]any)
	if !ok {
		t.Fatal("Expected filesystem server in mcpServers")
	}

	// Verify type
	if server["type"] != "stdio" {
		t.Errorf("Expected type 'stdio', got %v", server["type"])
	}

	// Verify command
	if server["command"] != "npx" {
		t.Errorf("Expected command 'npx', got %v", server["command"])
	}

	// Verify args
	args, ok := server["args"].([]any)
	if !ok || len(args) != 3 {
		t.Errorf("Expected 3 args, got %v", server["args"])
	}
}

// TestExternalMCPServerTransport_SSE verifies that SSE MCP servers
// are included in the --mcp-config flag.
func TestExternalMCPServerTransport_SSE(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"remote-sse": {
			Type:    "sse",
			URL:     "https://api.example.com/sse",
			Headers: map[string]string{"Authorization": "Bearer token123"},
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Find the --mcp-config flag
	var mcpConfigValue string
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			mcpConfigValue = cmd[i+1]
			break
		}
	}

	if mcpConfigValue == "" {
		t.Fatal("Expected --mcp-config flag in command for SSE server")
	}

	// Parse and validate the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	server, ok := servers["remote-sse"].(map[string]any)
	if !ok {
		t.Fatal("Expected remote-sse server in mcpServers")
	}

	// Verify type
	if server["type"] != "sse" {
		t.Errorf("Expected type 'sse', got %v", server["type"])
	}

	// Verify URL
	if server["url"] != "https://api.example.com/sse" {
		t.Errorf("Expected url 'https://api.example.com/sse', got %v", server["url"])
	}

	// Verify headers
	headers, ok := server["headers"].(map[string]any)
	if !ok {
		t.Fatal("Expected headers in server config")
	}

	if headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header 'Bearer token123', got %v", headers["Authorization"])
	}
}

// TestExternalMCPServerTransport_HTTP verifies that HTTP MCP servers
// are included in the --mcp-config flag.
func TestExternalMCPServerTransport_HTTP(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"api-server": {
			Type:    "http",
			URL:     "https://api.example.com/mcp",
			Headers: map[string]string{"X-API-Key": "secret"},
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Find the --mcp-config flag
	var mcpConfigValue string
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			mcpConfigValue = cmd[i+1]
			break
		}
	}

	if mcpConfigValue == "" {
		t.Fatal("Expected --mcp-config flag in command for HTTP server")
	}

	// Parse and validate the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	server, ok := servers["api-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected api-server in mcpServers")
	}

	// Verify type
	if server["type"] != "http" {
		t.Errorf("Expected type 'http', got %v", server["type"])
	}

	// Verify URL
	if server["url"] != "https://api.example.com/mcp" {
		t.Errorf("Expected url 'https://api.example.com/mcp', got %v", server["url"])
	}
}

// TestExternalMCPServerTransport_SDKIncluded verifies that SDK MCP servers
// ARE included in the --mcp-config flag so CLI knows about available tools.
// Tool execution is handled via control protocol.
func TestExternalMCPServerTransport_SDKIncluded(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"in-process": {
			Type: "sdk",
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Should have --mcp-config flag with SDK server
	var mcpConfigValue string
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			mcpConfigValue = cmd[i+1]
			break
		}
	}

	if mcpConfigValue == "" {
		t.Fatal("Expected --mcp-config flag when SDK servers are present")
	}

	if !strings.Contains(mcpConfigValue, "in-process") {
		t.Error("Expected mcp-config to contain SDK server 'in-process'")
	}
}

// TestExternalMCPServerTransport_EmptyServers verifies that empty MCPServers
// produces no --mcp-config flag.
func TestExternalMCPServerTransport_EmptyServers(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Should NOT have --mcp-config flag
	for i, arg := range cmd {
		if arg == "--mcp-config" {
			t.Errorf("Expected NO --mcp-config flag when MCPServers is empty, but found at index %d", i)
		}
	}
}

// TestExternalMCPServerTransport_NilServers verifies that nil MCPServers
// produces no --mcp-config flag.
func TestExternalMCPServerTransport_NilServers(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = nil

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Should NOT have --mcp-config flag
	for i, arg := range cmd {
		if arg == "--mcp-config" {
			t.Errorf("Expected NO --mcp-config flag when MCPServers is nil, but found at index %d", i)
		}
	}
}

// TestExternalMCPServerTransport_MixedServers verifies that when both
// SDK and external servers are present, all servers are included
// in --mcp-config.
func TestExternalMCPServerTransport_MixedServers(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"sdk-server": {
			Type: "sdk",
		},
		"stdio-server": {
			Type:    "stdio",
			Command: "node",
			Args:    []string{"server.js"},
		},
		"sse-server": {
			Type: "sse",
			URL:  "https://example.com/sse",
		},
		"http-server": {
			Type: "http",
			URL:  "https://example.com/mcp",
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Find the --mcp-config flag
	var mcpConfigValue string
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			mcpConfigValue = cmd[i+1]
			break
		}
	}

	if mcpConfigValue == "" {
		t.Fatal("Expected --mcp-config flag in command")
	}

	// Parse and validate the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	// Should have all 4 servers (including SDK server)
	if len(servers) != 4 {
		t.Errorf("Expected 4 servers, got %d", len(servers))
	}

	// Should have sdk-server (CLI needs to know about it for tool discovery)
	if _, exists := servers["sdk-server"]; !exists {
		t.Error("SDK server should be included in --mcp-config")
	}

	// Should have stdio-server
	if _, exists := servers["stdio-server"]; !exists {
		t.Error("stdio-server should be included in --mcp-config")
	}

	// Should have sse-server
	if _, exists := servers["sse-server"]; !exists {
		t.Error("sse-server should be included in --mcp-config")
	}

	// Should have http-server
	if _, exists := servers["http-server"]; !exists {
		t.Error("http-server should be included in --mcp-config")
	}
}

// TestExternalMCPServerTransport_JSONSerialization verifies that
// MCP server configs are properly serialized to JSON.
func TestExternalMCPServerTransport_JSONSerialization(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"complex-stdio": {
			Type:    "stdio",
			Command: "/usr/local/bin/mcp-server",
			Args:    []string{"--arg1", "value1", "--arg2", "value2"},
			Env: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		"complex-sse": {
			Type: "sse",
			URL:  "https://api.example.com/sse",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"X-Custom":      "value",
			},
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Find the --mcp-config flag
	var mcpConfigValue string
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			mcpConfigValue = cmd[i+1]
			break
		}
	}

	if mcpConfigValue == "" {
		t.Fatal("Expected --mcp-config flag in command")
	}

	// Parse the config to ensure it's valid JSON
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config as JSON: %v", err)
	}

	// Re-marshal to ensure it round-trips correctly
	if _, err := json.Marshal(config); err != nil {
		t.Fatalf("Failed to re-marshal config: %v", err)
	}

	// Verify structure
	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	// Verify complex-stdio
	stdioServer, ok := servers["complex-stdio"].(map[string]any)
	if !ok {
		t.Fatal("Expected complex-stdio server in mcpServers")
	}

	if env, ok := stdioServer["env"].(map[string]any); !ok {
		t.Error("Expected env to be a map")
	} else {
		if env["KEY1"] != "value1" {
			t.Errorf("Expected KEY1=value1, got %v", env["KEY1"])
		}
	}

	// Verify complex-sse
	sseServer, ok := servers["complex-sse"].(map[string]any)
	if !ok {
		t.Fatal("Expected complex-sse server in mcpServers")
	}

	if headers, ok := sseServer["headers"].(map[string]any); !ok {
		t.Error("Expected headers to be a map")
	} else {
		if headers["X-Custom"] != "value" {
			t.Errorf("Expected X-Custom=value, got %v", headers["X-Custom"])
		}
	}
}

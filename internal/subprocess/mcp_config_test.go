// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestSSEMCPServerConfigPassedToCLI verifies that SSE MCP server configs
// are properly passed to the CLI via --mcp-config flag.
func TestSSEMCPServerConfigPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"remote-server": {
			Type:    "sse",
			URL:     "http://localhost:8080/sse",
			Headers: map[string]string{"Authorization": "Bearer token"},
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

	// Parse the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	server, ok := servers["remote-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected remote-server in mcpServers")
	}

	if server["type"] != "sse" {
		t.Errorf("Expected type 'sse', got %v", server["type"])
	}

	if server["url"] != "http://localhost:8080/sse" {
		t.Errorf("Expected url 'http://localhost:8080/sse', got %v", server["url"])
	}
}

// TestHTTPMCPServerConfigPassedToCLI verifies that HTTP MCP server configs
// are properly passed to the CLI via --mcp-config flag.
func TestHTTPMCPServerConfigPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"http-server": {
			Type:    "http",
			URL:     "http://localhost:9000/mcp",
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
		t.Fatal("Expected --mcp-config flag in command")
	}

	// Parse the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	server, ok := servers["http-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected http-server in mcpServers")
	}

	if server["type"] != "http" {
		t.Errorf("Expected type 'http', got %v", server["type"])
	}

	if server["url"] != "http://localhost:9000/mcp" {
		t.Errorf("Expected url 'http://localhost:9000/mcp', got %v", server["url"])
	}
}

// TestStdioMCPServerConfigPassedToCLI verifies that stdio MCP server configs
// are properly passed to the CLI via --mcp-config flag.
func TestStdioMCPServerConfigPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"filesystem": {
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
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

	// Verify it contains the server config
	if !strings.Contains(mcpConfigValue, "filesystem") {
		t.Error("Expected mcp-config to contain 'filesystem'")
	}

	if !strings.Contains(mcpConfigValue, "stdio") {
		t.Error("Expected mcp-config to contain 'stdio'")
	}
}

// TestSDKMCPServersPassedToCLI verifies that SDK MCP servers (in-process)
// ARE passed to the CLI so it knows about available tools.
// The actual tool execution is handled via control protocol.
func TestSDKMCPServersPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"sdk-server": {
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
		t.Fatal("Expected --mcp-config flag in command")
	}

	if !strings.Contains(mcpConfigValue, "sdk-server") {
		t.Error("SDK MCP servers should be passed to CLI")
	}
}

// TestMixedMCPServersBothPassedToCLI verifies that mixed configs
// pass both external and SDK servers to CLI.
func TestMixedMCPServersBothPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"external-sse": {
			Type: "sse",
			URL:  "http://example.com/sse",
		},
		"internal-sdk": {
			Type: "sdk",
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

	// Should contain external server
	if !strings.Contains(mcpConfigValue, "external-sse") {
		t.Error("Expected mcp-config to contain 'external-sse'")
	}

	// Should also contain SDK server (CLI needs to know about it for tool discovery)
	if !strings.Contains(mcpConfigValue, "internal-sdk") {
		t.Error("SDK servers should be included in mcp-config")
	}
}

// TestStdioMCPServerWithoutTypeFieldDefaultsToStdio verifies that stdio MCP
// server configs without an explicit type field default to "stdio" for backwards
// compatibility with the Python SDK.
func TestStdioMCPServerWithoutTypeFieldDefaultsToStdio(t *testing.T) {
	testCases := []struct {
		name         string
		serverConfig map[string]any
	}{
		{
			name: "missing type field",
			serverConfig: map[string]any{
				// No "type" field - should default to "stdio"
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			},
		},
		{
			name: "empty string type",
			serverConfig: map[string]any{
				"type":    "", // Empty string - should default to "stdio"
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := types.DefaultOptions()
			// Use map[string]any to simulate a config without the Type field
			opts.MCPServers = map[string]any{
				"filesystem": tc.serverConfig,
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

			// Parse the config
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
				t.Fatal("Expected filesystem in mcpServers")
			}

			// The server should have type "stdio" even though we didn't specify it
			serverType, ok := server["type"].(string)
			if !ok {
				t.Error("Expected type field to be present in server config")
			}
			if serverType != "stdio" {
				t.Errorf("Expected type to default to 'stdio', got %v", serverType)
			}

			// Verify command and args are preserved
			if server["command"] != "npx" {
				t.Errorf("Expected command 'npx', got %v", server["command"])
			}
		})
	}
}

// TestSDKMCPServersMapPassedToCLI verifies that SDK MCP servers from the
// SDKMCPServers map are properly passed to the CLI via --mcp-config flag.
func TestSDKMCPServersMapPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.SDKMCPServers = map[string]*types.MCPServer{
		"calculator": {
			Name:    "calculator",
			Version: "1.0.0",
			Tools: []*types.MCPTool{
				{
					Name:        "add",
					Description: "Add two numbers",
				},
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

	// Parse the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	server, ok := servers["calculator"].(map[string]any)
	if !ok {
		t.Fatal("Expected calculator in mcpServers")
	}

	// Verify it's marked as SDK type
	if server["type"] != "sdk" {
		t.Errorf("Expected type 'sdk', got %v", server["type"])
	}
}

// TestSDKMCPServersAndExternalMerged verifies that SDK MCP servers and
// external MCP servers are both included in the --mcp-config flag.
func TestSDKMCPServersAndExternalMerged(t *testing.T) {
	opts := types.DefaultOptions()
	opts.SDKMCPServers = map[string]*types.MCPServer{
		"calculator": {
			Name:    "calculator",
			Version: "1.0.0",
		},
	}
	opts.MCPServers = map[string]types.MCPServerConfig{
		"filesystem": {
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
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

	// Parse the config
	var config map[string]any
	if err := json.Unmarshal([]byte(mcpConfigValue), &config); err != nil {
		t.Fatalf("Failed to parse mcp-config: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcpServers in config")
	}

	// Should have both servers
	if _, ok := servers["calculator"]; !ok {
		t.Error("Expected calculator (SDK server) in mcpServers")
	}
	if _, ok := servers["filesystem"]; !ok {
		t.Error("Expected filesystem (external server) in mcpServers")
	}
}

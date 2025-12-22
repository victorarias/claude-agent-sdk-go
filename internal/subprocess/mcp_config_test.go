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

// TestSDKMCPServersNotPassedToCLI verifies that SDK MCP servers (in-process)
// are NOT passed to the CLI - they are handled via control protocol.
func TestSDKMCPServersNotPassedToCLI(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"sdk-server": {
			Type: "sdk",
		},
	}

	cmd := buildCommand("/path/to/claude", "", opts, true)

	// Should NOT have --mcp-config flag (SDK servers are filtered out)
	for i, arg := range cmd {
		if arg == "--mcp-config" {
			mcpConfigValue := cmd[i+1]
			if strings.Contains(mcpConfigValue, "sdk-server") {
				t.Error("SDK MCP servers should not be passed to CLI")
			}
		}
	}
}

// TestMixedMCPServersOnlyExternalPassedToCLI verifies that mixed configs
// only pass external servers to CLI, not SDK servers.
func TestMixedMCPServersOnlyExternalPassedToCLI(t *testing.T) {
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

	// Should NOT contain SDK server
	if strings.Contains(mcpConfigValue, "internal-sdk") {
		t.Error("SDK servers should be filtered from mcp-config")
	}
}

// TestStdioMCPServerWithoutTypeFieldDefaultsToStdio verifies that stdio MCP
// server configs without an explicit type field default to "stdio" for backwards
// compatibility with the Python SDK.
func TestStdioMCPServerWithoutTypeFieldDefaultsToStdio(t *testing.T) {
	opts := types.DefaultOptions()
	// Use map[string]any to simulate a config without the Type field
	opts.MCPServers = map[string]any{
		"filesystem": map[string]any{
			// No "type" field - should default to "stdio"
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
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
}

// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"testing"
)

// TestNewMCPStdioServer tests the stdio server constructor.
func TestNewMCPStdioServer(t *testing.T) {
	// Create a stdio server config
	config := NewMCPStdioServer("python", []string{"-m", "my_mcp_server"})

	// Verify Type is set correctly
	if config.Type != "stdio" {
		t.Errorf("Expected Type 'stdio', got '%s'", config.Type)
	}

	// Verify Command is set
	if config.Command != "python" {
		t.Errorf("Expected Command 'python', got '%s'", config.Command)
	}

	// Verify Args are set
	if len(config.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(config.Args))
	}
	if config.Args[0] != "-m" || config.Args[1] != "my_mcp_server" {
		t.Errorf("Args not set correctly: %v", config.Args)
	}

	// Verify Env is initialized (can be nil or empty)
	// Just check it doesn't panic when accessed
	_ = config.Env

	// Verify URL and Headers are not set for stdio
	if config.URL != "" {
		t.Error("URL should not be set for stdio server")
	}
	if config.Headers != nil && len(config.Headers) > 0 {
		t.Error("Headers should not be set for stdio server")
	}
}

// TestNewMCPStdioServerWithEnv tests stdio server with environment variables.
func TestNewMCPStdioServerWithEnv(t *testing.T) {
	config := NewMCPStdioServer("node", []string{"server.js"})

	// Add environment variables after creation
	config.Env = map[string]string{
		"API_KEY": "test-key",
		"DEBUG":   "true",
	}

	if config.Env["API_KEY"] != "test-key" {
		t.Error("Env variables not set correctly")
	}
}

// TestNewMCPSSEServer tests the SSE server constructor.
func TestNewMCPSSEServer(t *testing.T) {
	// Create an SSE server config
	config := NewMCPSSEServer("https://example.com/mcp/sse")

	// Verify Type is set correctly
	if config.Type != "sse" {
		t.Errorf("Expected Type 'sse', got '%s'", config.Type)
	}

	// Verify URL is set
	if config.URL != "https://example.com/mcp/sse" {
		t.Errorf("Expected URL 'https://example.com/mcp/sse', got '%s'", config.URL)
	}

	// Verify Headers is initialized (can be nil or empty)
	_ = config.Headers

	// Verify Command, Args are not set for SSE
	if config.Command != "" {
		t.Error("Command should not be set for SSE server")
	}
	if config.Args != nil && len(config.Args) > 0 {
		t.Error("Args should not be set for SSE server")
	}
	if config.Env != nil && len(config.Env) > 0 {
		t.Error("Env should not be set for SSE server")
	}
}

// TestNewMCPSSEServerWithHeaders tests SSE server with custom headers.
func TestNewMCPSSEServerWithHeaders(t *testing.T) {
	config := NewMCPSSEServer("https://api.example.com/sse")

	// Add headers after creation
	config.Headers = map[string]string{
		"Authorization": "Bearer token123",
		"X-Custom":      "value",
	}

	if config.Headers["Authorization"] != "Bearer token123" {
		t.Error("Headers not set correctly")
	}
}

// TestNewMCPHTTPServer tests the HTTP server constructor.
func TestNewMCPHTTPServer(t *testing.T) {
	// Create an HTTP server config
	config := NewMCPHTTPServer("https://api.example.com/mcp")

	// Verify Type is set correctly
	if config.Type != "http" {
		t.Errorf("Expected Type 'http', got '%s'", config.Type)
	}

	// Verify URL is set
	if config.URL != "https://api.example.com/mcp" {
		t.Errorf("Expected URL 'https://api.example.com/mcp', got '%s'", config.URL)
	}

	// Verify Headers is initialized (can be nil or empty)
	_ = config.Headers

	// Verify Command, Args are not set for HTTP
	if config.Command != "" {
		t.Error("Command should not be set for HTTP server")
	}
	if config.Args != nil && len(config.Args) > 0 {
		t.Error("Args should not be set for HTTP server")
	}
	if config.Env != nil && len(config.Env) > 0 {
		t.Error("Env should not be set for HTTP server")
	}
}

// TestNewMCPHTTPServerWithHeaders tests HTTP server with custom headers.
func TestNewMCPHTTPServerWithHeaders(t *testing.T) {
	config := NewMCPHTTPServer("https://mcp.example.com")

	// Add headers after creation
	config.Headers = map[string]string{
		"Authorization": "Bearer token456",
		"Content-Type":  "application/json",
	}

	if config.Headers["Content-Type"] != "application/json" {
		t.Error("Headers not set correctly")
	}
}

// TestMCPServerConfigTypes tests that all three types are distinct.
func TestMCPServerConfigTypes(t *testing.T) {
	stdio := NewMCPStdioServer("cmd", []string{})
	sse := NewMCPSSEServer("http://localhost")
	http := NewMCPHTTPServer("http://localhost")

	types := []string{stdio.Type, sse.Type, http.Type}

	// Verify all types are different
	if stdio.Type == sse.Type || stdio.Type == http.Type || sse.Type == http.Type {
		t.Errorf("Types should be distinct: %v", types)
	}

	// Verify expected values
	expectedTypes := map[string]bool{"stdio": false, "sse": false, "http": false}
	for _, typ := range types {
		if _, ok := expectedTypes[typ]; !ok {
			t.Errorf("Unexpected type: %s", typ)
		}
		expectedTypes[typ] = true
	}

	// Verify all expected types were seen
	for typ, seen := range expectedTypes {
		if !seen {
			t.Errorf("Type %s was not created", typ)
		}
	}
}

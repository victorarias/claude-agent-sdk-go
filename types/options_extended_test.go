// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"testing"
)

// TestOptions_WithSDKMCPServer tests WithSDKMCPServer option.
func TestOptions_WithSDKMCPServer(t *testing.T) {
	t.Run("WithSDKMCPServer adds server to SDKMCPServers map", func(t *testing.T) {
		server := &MCPServer{
			Name:    "test-server",
			Version: "1.0.0",
		}

		opts := DefaultOptions()
		opt := WithSDKMCPServer("test", server)
		opt(opts)

		if opts.SDKMCPServers == nil {
			t.Fatal("Expected SDKMCPServers to be initialized")
		}

		result, ok := opts.SDKMCPServers["test"]
		if !ok {
			t.Fatal("Expected server 'test' to be in SDKMCPServers")
		}

		if result.Name != "test-server" {
			t.Errorf("Expected name 'test-server', got %q", result.Name)
		}
	})
}

// TestMCPServerBuilder_WithTool tests the WithTool method.
func TestMCPServerBuilder_WithTool(t *testing.T) {
	t.Run("WithTool adds tool to builder", func(t *testing.T) {
		builder := NewMCPServerBuilder("test-server")

		handler := func(args map[string]any) (*MCPToolResult, error) {
			return &MCPToolResult{
				Content: []MCPContent{NewTextContent("result")},
			}, nil
		}

		schema := map[string]any{"type": "object"}

		builder.WithTool("calculator", "Performs calculations", schema, handler)
		server := builder.Build()

		if len(server.Tools) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(server.Tools))
		}

		if server.Tools[0].Name != "calculator" {
			t.Errorf("Expected tool name 'calculator', got %s", server.Tools[0].Name)
		}

		if server.Tools[0].Description != "Performs calculations" {
			t.Errorf("Expected description 'Performs calculations', got %s", server.Tools[0].Description)
		}
	})

	t.Run("WithToolWithAnnotations stores annotations", func(t *testing.T) {
		builder := NewMCPServerBuilder("test-server")
		readOnly := true
		handler := func(args map[string]any) (*MCPToolResult, error) { return nil, nil }

		builder.WithToolWithAnnotations(
			"reader",
			"Reads data",
			map[string]any{"type": "object"},
			&MCPToolAnnotations{ReadOnlyHint: &readOnly},
			handler,
		)
		server := builder.Build()

		if len(server.Tools) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(server.Tools))
		}
		if server.Tools[0].Annotations == nil || server.Tools[0].Annotations.ReadOnlyHint == nil {
			t.Fatal("expected annotations to be set")
		}
		if !*server.Tools[0].Annotations.ReadOnlyHint {
			t.Fatal("expected readOnlyHint=true")
		}
	})

	t.Run("WithTool returns builder for chaining", func(t *testing.T) {
		builder := NewMCPServerBuilder("test")
		handler := func(args map[string]any) (*MCPToolResult, error) { return nil, nil }
		result := builder.WithTool("test", "desc", nil, handler)
		if result != builder {
			t.Error("Expected WithTool to return the same builder instance")
		}
	})
}

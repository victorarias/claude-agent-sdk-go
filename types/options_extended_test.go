// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"testing"
)

// TestOptions_CustomTransport tests custom transport methods.
func TestOptions_CustomTransport(t *testing.T) {
	t.Run("CustomTransport returns nil by default", func(t *testing.T) {
		opts := DefaultOptions()
		if opts.CustomTransport() != nil {
			t.Error("Expected CustomTransport to return nil by default")
		}
	})

	t.Run("SetCustomTransport sets transport", func(t *testing.T) {
		opts := DefaultOptions()
		transport := &mockTransport{}
		opts.SetCustomTransport(transport)
		result := opts.CustomTransport()
		if result == nil {
			t.Error("Expected CustomTransport to return the set transport")
		}
		// Verify it's the same transport by comparing type
		if _, ok := result.(*mockTransport); !ok {
			t.Error("Expected transport to be mockTransport type")
		}
	})
}

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
		builder := NewMCPServerBuilder()
		tool := &MCPTool{
			Name:        "test",
			Description: "Test tool",
		}

		builder.WithTool(tool)
		server := builder.Build()

		if len(server.Tools) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(server.Tools))
		}

		if server.Tools[0].Name != "test" {
			t.Errorf("Expected tool name 'test', got %s", server.Tools[0].Name)
		}
	})

	t.Run("WithTool returns builder for chaining", func(t *testing.T) {
		builder := NewMCPServerBuilder()
		result := builder.WithTool(&MCPTool{Name: "test"})
		if result != builder {
			t.Error("Expected WithTool to return the same builder instance")
		}
	})
}

// mockTransport is a mock transport for testing.
type mockTransport struct{}

func (m *mockTransport) Start() error                     { return nil }
func (m *mockTransport) Stop() error                      { return nil }
func (m *mockTransport) Send(data []byte) error           { return nil }
func (m *mockTransport) Receive() ([]byte, error)         { return nil, nil }
func (m *mockTransport) IsConnected() bool                { return true }
func (m *mockTransport) SetStderrCallback(func([]byte))   {}
func (m *mockTransport) SetMaxBufferSize(size int)        {}
func (m *mockTransport) SetPermissionMode(mode string)    {}
func (m *mockTransport) HandleInterrupt() error           { return nil }
func (m *mockTransport) HandlePermission(result any) error { return nil }
func (m *mockTransport) Close() error                     { return nil }

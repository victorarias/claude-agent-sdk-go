package subprocess

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestBuildCommand_Plugins tests that plugins are passed via --plugin-dir flags.
func TestBuildCommand_Plugins(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Plugins = []types.PluginConfig{
		{
			Type: "local",
			Path: "/path/to/plugin1",
		},
		{
			Type: "local",
			Path: "/path/to/plugin2",
		},
	}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	// Should have two --plugin-dir flags
	pluginDirCount := 0
	pluginPaths := []string{}

	for i := 0; i < len(cmd)-1; i++ {
		if cmd[i] == "--plugin-dir" {
			pluginDirCount++
			pluginPaths = append(pluginPaths, cmd[i+1])
		}
	}

	if pluginDirCount != 2 {
		t.Errorf("Expected 2 --plugin-dir flags, got %d", pluginDirCount)
	}

	// Check paths are correct
	expectedPaths := map[string]bool{
		"/path/to/plugin1": false,
		"/path/to/plugin2": false,
	}

	for _, path := range pluginPaths {
		if _, exists := expectedPaths[path]; exists {
			expectedPaths[path] = true
		} else {
			t.Errorf("Unexpected plugin path: %s", path)
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected plugin path %s not found", path)
		}
	}
}

// TestBuildCommand_NoPlugins tests that no --plugin-dir flag is added when plugins is nil.
func TestBuildCommand_NoPlugins(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Plugins = nil

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	for _, arg := range cmd {
		if arg == "--plugin-dir" {
			t.Error("--plugin-dir should not be present when plugins is nil")
		}
	}
}

// TestBuildCommand_EmptyPlugins tests that no --plugin-dir flag is added when plugins is empty.
func TestBuildCommand_EmptyPlugins(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Plugins = []types.PluginConfig{}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	for _, arg := range cmd {
		if arg == "--plugin-dir" {
			t.Error("--plugin-dir should not be present when plugins is empty")
		}
	}
}

// TestBuildCommand_SinglePlugin tests a single plugin configuration.
func TestBuildCommand_SinglePlugin(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Plugins = []types.PluginConfig{
		{
			Type: "local",
			Path: "/custom/plugin/path",
		},
	}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	found := false
	for i := 0; i < len(cmd)-1; i++ {
		if cmd[i] == "--plugin-dir" {
			if cmd[i+1] == "/custom/plugin/path" {
				found = true
			}
			break
		}
	}

	if !found {
		t.Error("Expected --plugin-dir /custom/plugin/path not found")
	}
}

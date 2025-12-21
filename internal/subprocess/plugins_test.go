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

// TestWithPlugins_FunctionalOption tests that WithPlugins functional option works.
func TestWithPlugins_FunctionalOption(t *testing.T) {
	opts := types.DefaultOptions()
	types.ApplyOptions(opts, types.WithPlugins(
		types.PluginConfig{Type: "local", Path: "/plugin/one"},
		types.PluginConfig{Type: "local", Path: "/plugin/two"},
	))

	if len(opts.Plugins) != 2 {
		t.Fatalf("Expected 2 plugins, got %d", len(opts.Plugins))
	}

	if opts.Plugins[0].Path != "/plugin/one" {
		t.Errorf("Expected /plugin/one, got %s", opts.Plugins[0].Path)
	}

	if opts.Plugins[1].Path != "/plugin/two" {
		t.Errorf("Expected /plugin/two, got %s", opts.Plugins[1].Path)
	}
}

// TestBuildCommand_OnlyLocalPluginsIncluded tests that only "local" type plugins are passed to CLI.
// This matches Python SDK behavior which only supports local plugins currently.
func TestBuildCommand_OnlyLocalPluginsIncluded(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Plugins = []types.PluginConfig{
		{Type: "local", Path: "/local/plugin"},
		{Type: "remote", Path: "http://example.com/plugin"}, // Should be ignored
		{Type: "local", Path: "/another/local/plugin"},
	}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	// Count --plugin-dir flags (should only be 2, not 3)
	pluginDirCount := 0
	localPaths := []string{}

	for i := 0; i < len(cmd)-1; i++ {
		if cmd[i] == "--plugin-dir" {
			pluginDirCount++
			localPaths = append(localPaths, cmd[i+1])
		}
	}

	if pluginDirCount != 2 {
		t.Errorf("Expected 2 --plugin-dir flags (only local plugins), got %d", pluginDirCount)
	}

	// Verify remote plugin is not included
	for _, path := range localPaths {
		if path == "http://example.com/plugin" {
			t.Error("Remote plugin should not be included in command")
		}
	}
}

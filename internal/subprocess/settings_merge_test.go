// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestSettingsMerge_SandboxMergedIntoSettings tests that when both Settings and Sandbox
// are provided, the sandbox is merged INTO the settings JSON object under the "sandbox" key.
// This matches Python SDK behavior in subprocess_cli.py:118-170.
func TestSettingsMerge_SandboxMergedIntoSettings(t *testing.T) {
	opts := types.DefaultOptions()

	// Set up settings as JSON string
	opts.Settings = `{"some":"config","another":"value"}`

	// Set up sandbox config
	opts.Sandbox = &types.SandboxSettings{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Verify --settings flag is present
	var settingsValue string
	foundSettings := false
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			settingsValue = cmd[i+1]
			foundSettings = true
			break
		}
	}

	if !foundSettings {
		t.Fatal("--settings flag not found in command")
	}

	// Parse the settings JSON
	var settingsObj map[string]any
	if err := json.Unmarshal([]byte(settingsValue), &settingsObj); err != nil {
		t.Fatalf("failed to parse settings JSON: %v", err)
	}

	// Verify original settings are preserved
	if settingsObj["some"] != "config" {
		t.Errorf("original settings not preserved: got %v", settingsObj["some"])
	}
	if settingsObj["another"] != "value" {
		t.Errorf("original settings not preserved: got %v", settingsObj["another"])
	}

	// Verify sandbox is merged under "sandbox" key
	sandboxValue, ok := settingsObj["sandbox"]
	if !ok {
		t.Fatal("sandbox key not found in merged settings")
	}

	// Verify sandbox contains expected fields
	sandboxMap, ok := sandboxValue.(map[string]any)
	if !ok {
		t.Fatalf("sandbox value is not a map: %T", sandboxValue)
	}

	if sandboxMap["enabled"] != true {
		t.Errorf("sandbox.enabled: got %v, want true", sandboxMap["enabled"])
	}
	if sandboxMap["autoAllowBashIfSandboxed"] != true {
		t.Errorf("sandbox.autoAllowBashIfSandboxed: got %v, want true", sandboxMap["autoAllowBashIfSandboxed"])
	}

	// Verify --sandbox flag is NOT present when settings exist
	for i, arg := range cmd {
		if arg == "--sandbox" {
			t.Errorf("--sandbox flag should not be present when merged into settings at position %d", i)
		}
	}
}

// TestSettingsMerge_SandboxOnlyNoSettings tests that when only Sandbox is provided
// (no Settings), the sandbox is wrapped in a settings object with "sandbox" key.
func TestSettingsMerge_SandboxOnlyNoSettings(t *testing.T) {
	opts := types.DefaultOptions()

	// Only sandbox, no settings
	opts.Sandbox = &types.SandboxSettings{
		Enabled:                   true,
		EnableWeakerNestedSandbox: true,
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Should have --settings flag with sandbox wrapped
	var settingsValue string
	foundSettings := false
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			settingsValue = cmd[i+1]
			foundSettings = true
			break
		}
	}

	if !foundSettings {
		t.Fatal("--settings flag not found in command")
	}

	// Parse the settings JSON
	var settingsObj map[string]any
	if err := json.Unmarshal([]byte(settingsValue), &settingsObj); err != nil {
		t.Fatalf("failed to parse settings JSON: %v", err)
	}

	// Verify sandbox is the only key
	if len(settingsObj) != 1 {
		t.Errorf("expected only sandbox key, got %d keys: %v", len(settingsObj), settingsObj)
	}

	// Verify sandbox content
	sandboxValue, ok := settingsObj["sandbox"]
	if !ok {
		t.Fatal("sandbox key not found in settings")
	}

	sandboxMap, ok := sandboxValue.(map[string]any)
	if !ok {
		t.Fatalf("sandbox value is not a map: %T", sandboxValue)
	}

	if sandboxMap["enabled"] != true {
		t.Errorf("sandbox.enabled: got %v, want true", sandboxMap["enabled"])
	}
	if sandboxMap["enableWeakerNestedSandbox"] != true {
		t.Errorf("sandbox.enableWeakerNestedSandbox: got %v, want true", sandboxMap["enableWeakerNestedSandbox"])
	}

	// Verify --sandbox flag is NOT present
	for i, arg := range cmd {
		if arg == "--sandbox" {
			t.Errorf("--sandbox flag should not be present when wrapped in settings at position %d", i)
		}
	}
}

// TestSettingsMerge_SettingsOnlyNoSandbox tests that when only Settings is provided
// (no Sandbox), settings are passed through as-is without modification.
func TestSettingsMerge_SettingsOnlyNoSandbox(t *testing.T) {
	opts := types.DefaultOptions()

	// Only settings, no sandbox
	opts.Settings = `{"some":"config"}`

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Should have --settings flag with original value
	var settingsValue string
	foundSettings := false
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			settingsValue = cmd[i+1]
			foundSettings = true
			break
		}
	}

	if !foundSettings {
		t.Fatal("--settings flag not found in command")
	}

	// Should be the original settings value
	if settingsValue != `{"some":"config"}` {
		t.Errorf("settings value modified: got %s, want %s", settingsValue, `{"some":"config"}`)
	}

	// Verify --sandbox flag is NOT present
	for i, arg := range cmd {
		if arg == "--sandbox" {
			t.Errorf("--sandbox flag should not be present at position %d", i)
		}
	}
}

// TestSettingsMerge_NestedSandboxSettings tests merging with nested sandbox settings.
func TestSettingsMerge_NestedSandboxSettings(t *testing.T) {
	opts := types.DefaultOptions()

	// Settings with existing config
	opts.Settings = `{"model":"claude-opus-4","timeout":30}`

	// Complex sandbox config with network settings
	opts.Sandbox = &types.SandboxSettings{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
		ExcludedCommands:         []string{"rm", "mv"},
		Network: &types.SandboxNetworkConfig{
			AllowAllUnixSockets: true,
			AllowLocalBinding:   true,
		},
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Find and parse settings
	var settingsValue string
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			settingsValue = cmd[i+1]
			break
		}
	}

	var settingsObj map[string]any
	if err := json.Unmarshal([]byte(settingsValue), &settingsObj); err != nil {
		t.Fatalf("failed to parse settings JSON: %v", err)
	}

	// Verify original settings preserved
	if settingsObj["model"] != "claude-opus-4" {
		t.Errorf("model not preserved: got %v", settingsObj["model"])
	}
	if timeout, ok := settingsObj["timeout"].(float64); !ok || timeout != 30 {
		t.Errorf("timeout not preserved: got %v", settingsObj["timeout"])
	}

	// Verify sandbox merged with all nested fields
	sandboxValue, ok := settingsObj["sandbox"]
	if !ok {
		t.Fatal("sandbox key not found in merged settings")
	}
	sandboxMap, ok := sandboxValue.(map[string]any)
	if !ok {
		t.Fatalf("sandbox value is not a map: %T", sandboxValue)
	}
	if sandboxMap["enabled"] != true {
		t.Errorf("sandbox.enabled: got %v, want true", sandboxMap["enabled"])
	}
	if sandboxMap["autoAllowBashIfSandboxed"] != true {
		t.Errorf("sandbox.autoAllowBashIfSandboxed: got %v, want true", sandboxMap["autoAllowBashIfSandboxed"])
	}

	// Verify excluded commands array is preserved
	excludedCmds, ok := sandboxMap["excludedCommands"].([]any)
	if !ok {
		t.Fatalf("sandbox.excludedCommands is not an array: %T", sandboxMap["excludedCommands"])
	}
	if len(excludedCmds) != 2 {
		t.Errorf("sandbox.excludedCommands: expected 2 items, got %d", len(excludedCmds))
	}

	// Verify network config is preserved
	networkMap, ok := sandboxMap["network"].(map[string]any)
	if !ok {
		t.Fatalf("sandbox.network is not a map: %T", sandboxMap["network"])
	}
	if networkMap["allowAllUnixSockets"] != true {
		t.Errorf("sandbox.network.allowAllUnixSockets: got %v, want true", networkMap["allowAllUnixSockets"])
	}
	if networkMap["allowLocalBinding"] != true {
		t.Errorf("sandbox.network.allowLocalBinding: got %v, want true", networkMap["allowLocalBinding"])
	}
}

// TestSettingsMerge_NoSettingsNoSandbox tests that when neither Settings nor Sandbox
// are provided, no --settings or --sandbox flags are present.
func TestSettingsMerge_NoSettingsNoSandbox(t *testing.T) {
	opts := types.DefaultOptions()

	// Neither settings nor sandbox

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Verify no --settings flag
	for i, arg := range cmd {
		if arg == "--settings" {
			t.Errorf("--settings flag should not be present at position %d", i)
		}
		if arg == "--sandbox" {
			t.Errorf("--sandbox flag should not be present at position %d", i)
		}
	}
}

// TestSettingsMerge_SettingsAsFilePath tests that file path settings work correctly.
// Note: This is tricky because we can't easily distinguish file paths from JSON strings
// in the current implementation. We test that the value is passed through.
func TestSettingsMerge_SettingsAsFilePath(t *testing.T) {
	opts := types.DefaultOptions()

	// Settings as a file path (not JSON)
	opts.Settings = "/path/to/settings.json"

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Should have --settings flag with the path
	var settingsValue string
	foundSettings := false
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			settingsValue = cmd[i+1]
			foundSettings = true
			break
		}
	}

	if !foundSettings {
		t.Fatal("--settings flag not found in command")
	}

	// Should be the original path
	if settingsValue != "/path/to/settings.json" {
		t.Errorf("settings value modified: got %s, want %s", settingsValue, "/path/to/settings.json")
	}
}

// TestSettingsMerge_SettingsFilePathWithSandbox tests that when settings is a file path
// AND sandbox is provided, we read the file and merge sandbox into it.
// This matches Python SDK behavior: read file, parse JSON, merge sandbox under "sandbox" key.
func TestSettingsMerge_SettingsFilePathWithSandbox(t *testing.T) {
	// Create a temporary settings file
	tempDir := t.TempDir()
	settingsFile := tempDir + "/settings.json"
	settingsContent := `{"model":"claude-opus-4","timeout":60}`

	if err := os.WriteFile(settingsFile, []byte(settingsContent), 0644); err != nil {
		t.Fatalf("failed to create temp settings file: %v", err)
	}

	opts := types.DefaultOptions()

	// Settings as a file path
	opts.Settings = settingsFile

	// Sandbox config
	opts.Sandbox = &types.SandboxSettings{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// Find settings value
	var settingsValue string
	foundSettings := false
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			settingsValue = cmd[i+1]
			foundSettings = true
			break
		}
	}

	if !foundSettings {
		t.Fatal("--settings flag not found in command")
	}

	// Should be JSON (file was read and merged)
	if !strings.HasPrefix(settingsValue, "{") {
		t.Errorf("expected JSON output, got: %s", settingsValue)
	}

	// Parse and verify merged settings
	var settingsObj map[string]any
	if err := json.Unmarshal([]byte(settingsValue), &settingsObj); err != nil {
		t.Fatalf("failed to parse settings JSON: %v", err)
	}

	// Verify original file content is preserved
	if settingsObj["model"] != "claude-opus-4" {
		t.Errorf("model not preserved: got %v", settingsObj["model"])
	}
	if timeout, ok := settingsObj["timeout"].(float64); !ok || timeout != 60 {
		t.Errorf("timeout not preserved: got %v", settingsObj["timeout"])
	}

	// Verify sandbox was merged
	sandboxValue, ok := settingsObj["sandbox"]
	if !ok {
		t.Fatal("sandbox key not found in merged settings")
	}

	sandboxMap, ok := sandboxValue.(map[string]any)
	if !ok {
		t.Fatalf("sandbox value is not a map: %T", sandboxValue)
	}

	if sandboxMap["enabled"] != true {
		t.Errorf("sandbox.enabled: got %v, want true", sandboxMap["enabled"])
	}
	if sandboxMap["autoAllowBashIfSandboxed"] != true {
		t.Errorf("sandbox.autoAllowBashIfSandboxed: got %v, want true", sandboxMap["autoAllowBashIfSandboxed"])
	}
}

// TestSettingsMerge_SettingsFileNotFound tests error handling when settings file doesn't exist.
func TestSettingsMerge_SettingsFileNotFound(t *testing.T) {
	opts := types.DefaultOptions()

	// Settings as a non-existent file path
	opts.Settings = "/non/existent/path/settings.json"

	// Sandbox config
	opts.Sandbox = &types.SandboxSettings{
		Enabled: true,
	}

	// buildSettingsValue should return an error
	_, err := buildSettingsValue(opts)
	if err == nil {
		t.Fatal("expected error for non-existent settings file, got nil")
	}

	if !strings.Contains(err.Error(), "settings file not found") {
		t.Errorf("expected 'settings file not found' error, got: %v", err)
	}
}

// TestSettingsMerge_SettingsFileInvalidJSON tests error handling for invalid JSON in file.
func TestSettingsMerge_SettingsFileInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempDir := t.TempDir()
	settingsFile := tempDir + "/invalid.json"
	invalidContent := `{this is not valid JSON}`

	if err := os.WriteFile(settingsFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to create temp settings file: %v", err)
	}

	opts := types.DefaultOptions()
	opts.Settings = settingsFile
	opts.Sandbox = &types.SandboxSettings{
		Enabled: true,
	}

	// buildSettingsValue should return an error
	_, err := buildSettingsValue(opts)
	if err == nil {
		t.Fatal("expected error for invalid JSON in settings file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse settings file as JSON") {
		t.Errorf("expected 'failed to parse settings file as JSON' error, got: %v", err)
	}
}

// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestSandboxSettings_Fields tests that all fields are present and properly tagged.
func TestSandboxSettings_Fields(t *testing.T) {
	settings := SandboxSettings{
		Enabled:                   true,
		AutoAllowBashIfSandboxed:  true,
		ExcludedCommands:          []string{"git", "docker"},
		AllowUnsandboxedCommands:  false,
		Network:                   &SandboxNetworkConfig{AllowLocalBinding: true},
		IgnoreViolations:          &SandboxIgnoreViolations{File: []string{"/tmp"}},
		EnableWeakerNestedSandbox: true,
	}

	// All fields should be settable
	if !settings.Enabled {
		t.Error("Enabled field not set")
	}
	if !settings.AutoAllowBashIfSandboxed {
		t.Error("AutoAllowBashIfSandboxed field not set")
	}
	if len(settings.ExcludedCommands) != 2 {
		t.Error("ExcludedCommands field not set")
	}
	if settings.AllowUnsandboxedCommands {
		t.Error("AllowUnsandboxedCommands field not set")
	}
	if settings.Network == nil || !settings.Network.AllowLocalBinding {
		t.Error("Network field not set")
	}
	if settings.IgnoreViolations == nil || len(settings.IgnoreViolations.File) != 1 {
		t.Error("IgnoreViolations field not set")
	}
	if !settings.EnableWeakerNestedSandbox {
		t.Error("EnableWeakerNestedSandbox field not set")
	}
}

// TestSandboxNetworkConfig_Fields tests that all network config fields are present.
func TestSandboxNetworkConfig_Fields(t *testing.T) {
	httpPort := 8080
	socksPort := 1080

	config := SandboxNetworkConfig{
		AllowUnixSockets:    []string{"/var/run/docker.sock", "/tmp/agent.sock"},
		AllowAllUnixSockets: true,
		AllowLocalBinding:   true,
		HTTPProxyPort:       &httpPort,
		SocksProxyPort:      &socksPort,
	}

	// All fields should be settable
	if len(config.AllowUnixSockets) != 2 {
		t.Error("AllowUnixSockets field not set")
	}
	if !config.AllowAllUnixSockets {
		t.Error("AllowAllUnixSockets field not set")
	}
	if !config.AllowLocalBinding {
		t.Error("AllowLocalBinding field not set")
	}
	if config.HTTPProxyPort == nil || *config.HTTPProxyPort != 8080 {
		t.Error("HTTPProxyPort field not set")
	}
	if config.SocksProxyPort == nil || *config.SocksProxyPort != 1080 {
		t.Error("SocksProxyPort field not set")
	}
}

// TestSandboxIgnoreViolations_Fields tests that all ignore violations fields are present.
func TestSandboxIgnoreViolations_Fields(t *testing.T) {
	violations := SandboxIgnoreViolations{
		File:    []string{"/tmp/file1", "/tmp/file2"},
		Network: []string{"example.com", "api.example.com"},
	}

	// All fields should be settable
	if len(violations.File) != 2 {
		t.Error("File field not set")
	}
	if len(violations.Network) != 2 {
		t.Error("Network field not set")
	}
}

// TestSandboxSettings_JSONSerialization tests JSON marshaling and unmarshaling.
func TestSandboxSettings_JSONSerialization(t *testing.T) {
	httpPort := 8080
	original := SandboxSettings{
		Enabled:                   true,
		AutoAllowBashIfSandboxed:  true,
		ExcludedCommands:          []string{"git", "docker"},
		AllowUnsandboxedCommands:  false,
		Network: &SandboxNetworkConfig{
			AllowUnixSockets:    []string{"/var/run/docker.sock"},
			AllowAllUnixSockets: false,
			AllowLocalBinding:   true,
			HTTPProxyPort:       &httpPort,
		},
		IgnoreViolations: &SandboxIgnoreViolations{
			File:    []string{"/tmp"},
			Network: []string{"localhost"},
		},
		EnableWeakerNestedSandbox: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored SandboxSettings
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify fields
	if restored.Enabled != original.Enabled {
		t.Error("Enabled not preserved")
	}
	if restored.AutoAllowBashIfSandboxed != original.AutoAllowBashIfSandboxed {
		t.Error("AutoAllowBashIfSandboxed not preserved")
	}
	if len(restored.ExcludedCommands) != len(original.ExcludedCommands) {
		t.Error("ExcludedCommands not preserved")
	}
	if restored.AllowUnsandboxedCommands != original.AllowUnsandboxedCommands {
		t.Error("AllowUnsandboxedCommands not preserved")
	}
	if restored.EnableWeakerNestedSandbox != original.EnableWeakerNestedSandbox {
		t.Error("EnableWeakerNestedSandbox not preserved")
	}

	// Verify nested Network
	if restored.Network == nil {
		t.Fatal("Network not preserved")
	}
	if len(restored.Network.AllowUnixSockets) != len(original.Network.AllowUnixSockets) {
		t.Error("Network.AllowUnixSockets not preserved")
	}
	if restored.Network.AllowLocalBinding != original.Network.AllowLocalBinding {
		t.Error("Network.AllowLocalBinding not preserved")
	}
	if restored.Network.HTTPProxyPort == nil || *restored.Network.HTTPProxyPort != *original.Network.HTTPProxyPort {
		t.Error("Network.HTTPProxyPort not preserved")
	}

	// Verify nested IgnoreViolations
	if restored.IgnoreViolations == nil {
		t.Fatal("IgnoreViolations not preserved")
	}
	if len(restored.IgnoreViolations.File) != len(original.IgnoreViolations.File) {
		t.Error("IgnoreViolations.File not preserved")
	}
	if len(restored.IgnoreViolations.Network) != len(original.IgnoreViolations.Network) {
		t.Error("IgnoreViolations.Network not preserved")
	}
}

// TestSandboxSettings_JSONOmitEmpty tests that omitempty works correctly.
func TestSandboxSettings_JSONOmitEmpty(t *testing.T) {
	settings := SandboxSettings{
		Enabled: true,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Should only contain "enabled"
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// enabled should be present
	if _, ok := raw["enabled"]; !ok {
		t.Error("enabled field missing from JSON")
	}

	// Other fields should be omitted (except false bools which are omitted by omitempty)
	// autoAllowBashIfSandboxed should be omitted (false + omitempty)
	if _, ok := raw["autoAllowBashIfSandboxed"]; ok {
		t.Error("autoAllowBashIfSandboxed should be omitted when false")
	}
}

// TestSandboxNetworkConfig_JSONSerialization tests network config JSON handling.
func TestSandboxNetworkConfig_JSONSerialization(t *testing.T) {
	httpPort := 8080
	original := SandboxNetworkConfig{
		AllowUnixSockets:    []string{"/var/run/docker.sock"},
		AllowAllUnixSockets: true,
		AllowLocalBinding:   true,
		HTTPProxyPort:       &httpPort,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var restored SandboxNetworkConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(restored.AllowUnixSockets) != 1 {
		t.Error("AllowUnixSockets not preserved")
	}
	if !restored.AllowAllUnixSockets {
		t.Error("AllowAllUnixSockets not preserved")
	}
	if !restored.AllowLocalBinding {
		t.Error("AllowLocalBinding not preserved")
	}
	if restored.HTTPProxyPort == nil || *restored.HTTPProxyPort != 8080 {
		t.Error("HTTPProxyPort not preserved")
	}
}

// TestSandboxIgnoreViolations_JSONSerialization tests ignore violations JSON handling.
func TestSandboxIgnoreViolations_JSONSerialization(t *testing.T) {
	original := SandboxIgnoreViolations{
		File:    []string{"/tmp/file1"},
		Network: []string{"example.com"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var restored SandboxIgnoreViolations
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(restored.File) != 1 {
		t.Error("File not preserved")
	}
	if len(restored.Network) != 1 {
		t.Error("Network not preserved")
	}
}

// TestSandboxSettings_MinimalConfig tests that minimal config works.
func TestSandboxSettings_MinimalConfig(t *testing.T) {
	// Just enable sandbox, nothing else
	settings := SandboxSettings{
		Enabled: true,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal minimal config: %v", err)
	}

	var restored SandboxSettings
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal minimal config: %v", err)
	}

	if !restored.Enabled {
		t.Error("Minimal config not preserved")
	}
}

// TestSandboxSettings_FullConfig tests a fully populated config.
func TestSandboxSettings_FullConfig(t *testing.T) {
	httpPort := 8080
	socksPort := 1080

	settings := SandboxSettings{
		Enabled:                   true,
		AutoAllowBashIfSandboxed:  true,
		ExcludedCommands:          []string{"git", "docker", "kubectl"},
		AllowUnsandboxedCommands:  true,
		Network: &SandboxNetworkConfig{
			AllowUnixSockets:    []string{"/var/run/docker.sock", "/tmp/agent.sock"},
			AllowAllUnixSockets: false,
			AllowLocalBinding:   true,
			HTTPProxyPort:       &httpPort,
			SocksProxyPort:      &socksPort,
		},
		IgnoreViolations: &SandboxIgnoreViolations{
			File:    []string{"/tmp", "/var/tmp"},
			Network: []string{"localhost", "127.0.0.1"},
		},
		EnableWeakerNestedSandbox: false,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal full config: %v", err)
	}

	var restored SandboxSettings
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal full config: %v", err)
	}

	// Verify all fields
	if !restored.Enabled {
		t.Error("Enabled not preserved in full config")
	}
	if !restored.AutoAllowBashIfSandboxed {
		t.Error("AutoAllowBashIfSandboxed not preserved in full config")
	}
	if len(restored.ExcludedCommands) != 3 {
		t.Errorf("ExcludedCommands not preserved in full config, got %d items", len(restored.ExcludedCommands))
	}
	if !restored.AllowUnsandboxedCommands {
		t.Error("AllowUnsandboxedCommands not preserved in full config")
	}
	if restored.Network == nil {
		t.Fatal("Network not preserved in full config")
	}
	if len(restored.Network.AllowUnixSockets) != 2 {
		t.Error("Network.AllowUnixSockets not preserved in full config")
	}
	if restored.Network.AllowAllUnixSockets {
		t.Error("Network.AllowAllUnixSockets should be false in full config")
	}
	if !restored.Network.AllowLocalBinding {
		t.Error("Network.AllowLocalBinding not preserved in full config")
	}
	if restored.Network.HTTPProxyPort == nil || *restored.Network.HTTPProxyPort != 8080 {
		t.Error("Network.HTTPProxyPort not preserved in full config")
	}
	if restored.Network.SocksProxyPort == nil || *restored.Network.SocksProxyPort != 1080 {
		t.Error("Network.SocksProxyPort not preserved in full config")
	}
	if restored.IgnoreViolations == nil {
		t.Fatal("IgnoreViolations not preserved in full config")
	}
	if len(restored.IgnoreViolations.File) != 2 {
		t.Error("IgnoreViolations.File not preserved in full config")
	}
	if len(restored.IgnoreViolations.Network) != 2 {
		t.Error("IgnoreViolations.Network not preserved in full config")
	}
	if restored.EnableWeakerNestedSandbox {
		t.Error("EnableWeakerNestedSandbox should be false in full config")
	}
}

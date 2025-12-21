package types

import (
	"encoding/json"
	"testing"
)

// TestAgentsField tests that the Agents field exists and can be set.
func TestAgentsField(t *testing.T) {
	opts := DefaultOptions()

	// Test that Agents field can be set
	agents := map[string]AgentDefinition{
		"researcher": {
			Description: "Research assistant",
			Prompt:      "You are a research assistant.",
			Tools:       []string{"WebSearch", "Read"},
			Model:       AgentModelSonnet,
		},
		"coder": {
			Description: "Coding assistant",
			Prompt:      "You are a coding assistant.",
			Model:       AgentModelOpus,
		},
	}

	opts.Agents = agents

	if opts.Agents == nil {
		t.Error("Agents should be set")
	}

	if len(opts.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(opts.Agents))
	}

	researcher, ok := opts.Agents["researcher"]
	if !ok {
		t.Error("researcher agent should exist")
	}

	if researcher.Description != "Research assistant" {
		t.Errorf("Expected 'Research assistant', got %q", researcher.Description)
	}

	if researcher.Model != AgentModelSonnet {
		t.Errorf("Expected AgentModelSonnet, got %q", researcher.Model)
	}
}

// TestWithAgents tests the WithAgents helper function.
func TestWithAgents(t *testing.T) {
	agents := map[string]AgentDefinition{
		"test": {
			Description: "Test agent",
			Prompt:      "You are a test agent.",
		},
	}

	opts := DefaultOptions()
	ApplyOptions(opts, WithAgents(agents))

	if opts.Agents == nil {
		t.Error("Agents should be set via WithAgents")
	}

	if len(opts.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(opts.Agents))
	}

	testAgent, ok := opts.Agents["test"]
	if !ok {
		t.Error("test agent should exist")
	}

	if testAgent.Description != "Test agent" {
		t.Errorf("Expected 'Test agent', got %q", testAgent.Description)
	}
}

// TestAgentsJSONSerialization tests that agents serialize correctly to JSON.
func TestAgentsJSONSerialization(t *testing.T) {
	opts := &Options{
		Agents: map[string]AgentDefinition{
			"researcher": {
				Description: "Research assistant",
				Prompt:      "You are a research assistant.",
				Tools:       []string{"WebSearch"},
				Model:       AgentModelSonnet,
			},
		},
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	agents, ok := parsed["agents"].(map[string]any)
	if !ok {
		t.Error("agents field should be present in JSON")
	}

	researcher, ok := agents["researcher"].(map[string]any)
	if !ok {
		t.Error("researcher should be present in agents")
	}

	if researcher["description"] != "Research assistant" {
		t.Errorf("Expected 'Research assistant', got %q", researcher["description"])
	}

	if researcher["model"] != "sonnet" {
		t.Errorf("Expected 'sonnet', got %q", researcher["model"])
	}
}

// TestAgentDefinitionOmitsEmptyFields tests that empty fields are omitted in JSON.
func TestAgentDefinitionOmitsEmptyFields(t *testing.T) {
	agent := AgentDefinition{
		Description: "Test agent",
		Prompt:      "You are a test agent.",
		// Tools and Model are omitted
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, ok := parsed["tools"]; ok {
		t.Error("tools should be omitted when empty")
	}

	if _, ok := parsed["model"]; ok {
		t.Error("model should be omitted when empty")
	}

	if parsed["description"] != "Test agent" {
		t.Error("description should be present")
	}

	if parsed["prompt"] != "You are a test agent." {
		t.Error("prompt should be present")
	}
}

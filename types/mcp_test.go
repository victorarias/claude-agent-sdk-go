// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestMCPToolResultIsError tests that MCPToolResult has an IsError field
// for indicating tool execution errors (matching Python SDK behavior).
func TestMCPToolResultIsError(t *testing.T) {
	// Create a result with error
	result := MCPToolResult{
		Content: []MCPContent{
			{Type: "text", Text: "Error: division by zero"},
		},
		IsError: true,
	}

	if !result.IsError {
		t.Error("IsError should be true")
	}

	// Test JSON serialization includes isError
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, ok := parsed["isError"]; !ok {
		t.Error("JSON should include isError field")
	}

	// Test that isError=false is omitted (omitempty behavior)
	resultNoError := MCPToolResult{
		Content: []MCPContent{
			{Type: "text", Text: "Success"},
		},
		IsError: false,
	}

	data, _ = json.Marshal(resultNoError)
	parsed = make(map[string]any)
	json.Unmarshal(data, &parsed)

	if _, ok := parsed["isError"]; ok {
		t.Error("JSON should omit isError when false")
	}
}

// TestMCPContentImageSupport tests that MCPContent supports image data
// with Data and MimeType fields (matching Python SDK's ImageContent).
func TestMCPContentImageSupport(t *testing.T) {
	// Create an image content
	imageContent := MCPContent{
		Type:     "image",
		Data:     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		MimeType: "image/png",
	}

	if imageContent.Type != "image" {
		t.Error("Type should be 'image'")
	}

	if imageContent.Data == "" {
		t.Error("Data should be set for image content")
	}

	if imageContent.MimeType == "" {
		t.Error("MimeType should be set for image content")
	}

	// Test JSON serialization
	data, err := json.Marshal(imageContent)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed["type"] != "image" {
		t.Error("JSON type should be 'image'")
	}

	if parsed["data"] == nil {
		t.Error("JSON should include data field")
	}

	if parsed["mimeType"] == nil {
		t.Error("JSON should include mimeType field")
	}
}

// TestMCPContentTextOmitsImageFields tests that text content
// omits the image-specific fields (Data and MimeType).
func TestMCPContentTextOmitsImageFields(t *testing.T) {
	textContent := MCPContent{
		Type: "text",
		Text: "Hello, world!",
	}

	data, err := json.Marshal(textContent)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, ok := parsed["data"]; ok {
		t.Error("Text content should not include data field")
	}

	if _, ok := parsed["mimeType"]; ok {
		t.Error("Text content should not include mimeType field")
	}
}

// TestMCPContentMixedResult tests a result with both text and image content.
func TestMCPContentMixedResult(t *testing.T) {
	result := MCPToolResult{
		Content: []MCPContent{
			{Type: "text", Text: "Here is the generated image:"},
			{Type: "image", Data: "base64data...", MimeType: "image/png"},
		},
		IsError: false,
	}

	if len(result.Content) != 2 {
		t.Errorf("Expected 2 content items, got %d", len(result.Content))
	}

	if result.Content[0].Type != "text" {
		t.Error("First content should be text")
	}

	if result.Content[1].Type != "image" {
		t.Error("Second content should be image")
	}

	if result.Content[1].Data == "" {
		t.Error("Image content should have Data")
	}
}

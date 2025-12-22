package types

import (
	"encoding/json"
	"testing"
)

// TestNewTextContent tests creating text content with helper function.
func TestNewTextContent(t *testing.T) {
	content := NewTextContent("Hello, world!")

	if content.Type != "text" {
		t.Errorf("Expected type 'text', got %q", content.Type)
	}

	if content.Text != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got %q", content.Text)
	}

	// Verify image fields are empty
	if content.Data != "" {
		t.Error("Text content should not have Data field set")
	}

	if content.MimeType != "" {
		t.Error("Text content should not have MimeType field set")
	}
}

// TestNewTextContentJSON tests JSON serialization of text content.
func TestNewTextContentJSON(t *testing.T) {
	content := NewTextContent("Test message")

	data, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed["type"] != "text" {
		t.Error("JSON type should be 'text'")
	}

	if parsed["text"] != "Test message" {
		t.Error("JSON text should match input")
	}

	// Verify omitempty behavior
	if _, ok := parsed["data"]; ok {
		t.Error("JSON should omit empty data field")
	}

	if _, ok := parsed["mimeType"]; ok {
		t.Error("JSON should omit empty mimeType field")
	}
}

// TestNewImageContent tests creating image content with helper function.
func TestNewImageContent(t *testing.T) {
	base64Data := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	content := NewImageContent(base64Data, "image/png")

	if content.Type != "image" {
		t.Errorf("Expected type 'image', got %q", content.Type)
	}

	if content.Data != base64Data {
		t.Error("Image data should match input")
	}

	if content.MimeType != "image/png" {
		t.Errorf("Expected mimeType 'image/png', got %q", content.MimeType)
	}

	// Verify text field is empty
	if content.Text != "" {
		t.Error("Image content should not have Text field set")
	}
}

// TestNewImageContentDifferentMimeTypes tests different image MIME types.
func TestNewImageContentDifferentMimeTypes(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		mimeType string
	}{
		{
			name:     "PNG image",
			data:     "base64pngdata",
			mimeType: "image/png",
		},
		{
			name:     "JPEG image",
			data:     "base64jpegdata",
			mimeType: "image/jpeg",
		},
		{
			name:     "GIF image",
			data:     "base64gifdata",
			mimeType: "image/gif",
		},
		{
			name:     "WebP image",
			data:     "base64webpdata",
			mimeType: "image/webp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := NewImageContent(tc.data, tc.mimeType)

			if content.Type != "image" {
				t.Errorf("Expected type 'image', got %q", content.Type)
			}

			if content.Data != tc.data {
				t.Error("Data should match input")
			}

			if content.MimeType != tc.mimeType {
				t.Errorf("Expected mimeType %q, got %q", tc.mimeType, content.MimeType)
			}
		})
	}
}

// TestNewImageContentJSON tests JSON serialization of image content.
func TestNewImageContentJSON(t *testing.T) {
	base64Data := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	content := NewImageContent(base64Data, "image/png")

	data, err := json.Marshal(content)
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

	if parsed["data"] != base64Data {
		t.Error("JSON data should match input")
	}

	if parsed["mimeType"] != "image/png" {
		t.Error("JSON mimeType should match input")
	}

	// Verify text field is omitted
	if _, ok := parsed["text"]; ok {
		t.Error("JSON should omit empty text field")
	}
}

// TestContentHelpersFunctional tests using helper functions in realistic scenarios.
func TestContentHelpersFunctional(t *testing.T) {
	// Simulate a tool result with mixed content
	result := &MCPToolResult{
		Content: []MCPContent{
			NewTextContent("Processing complete. Here is the result:"),
			NewImageContent("base64imagedata", "image/png"),
			NewTextContent("Additional information..."),
		},
		IsError: false,
	}

	if len(result.Content) != 3 {
		t.Fatalf("Expected 3 content items, got %d", len(result.Content))
	}

	// Verify first item is text
	if result.Content[0].Type != "text" {
		t.Error("First item should be text")
	}
	if result.Content[0].Text != "Processing complete. Here is the result:" {
		t.Error("First item text mismatch")
	}

	// Verify second item is image
	if result.Content[1].Type != "image" {
		t.Error("Second item should be image")
	}
	if result.Content[1].Data == "" {
		t.Error("Image should have data")
	}
	if result.Content[1].MimeType != "image/png" {
		t.Error("Image should have mime type")
	}

	// Verify third item is text
	if result.Content[2].Type != "text" {
		t.Error("Third item should be text")
	}
}

// TestContentHelpersEmptyStrings tests helper functions with empty strings.
func TestContentHelpersEmptyStrings(t *testing.T) {
	// Text content with empty string
	textContent := NewTextContent("")
	if textContent.Type != "text" {
		t.Error("Type should still be 'text'")
	}
	if textContent.Text != "" {
		t.Error("Text should be empty string")
	}

	// Image content with empty data (valid use case for placeholder)
	imageContent := NewImageContent("", "image/png")
	if imageContent.Type != "image" {
		t.Error("Type should still be 'image'")
	}
	if imageContent.Data != "" {
		t.Error("Data should be empty string")
	}
	if imageContent.MimeType != "image/png" {
		t.Error("MimeType should still be set")
	}
}

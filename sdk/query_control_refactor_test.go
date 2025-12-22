package sdk

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// These tests verify that the refactored handlers use typed structs internally
// rather than map[string]any throughout the parsing and handling logic.

// TestParseControlRequest_UsingTypedStructs verifies that handleControlRequest
// uses ParseSDKControlRequest to get typed structs.
func TestParseControlRequest_UsingTypedStructs(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Register hook callback to verify typed struct is passed
	// Use channel for thread-safe signaling instead of boolean
	receivedChan := make(chan bool, 1)
	query.hookMu.Lock()
	query.hookCallbacks["cb_typed"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		// Input should be properly typed from the parsed struct
		if input == nil {
			t.Error("Expected input to not be nil")
		}
		receivedChan <- true
		return &types.HookOutput{Continue: boolPtr(true)}, nil
	}
	query.hookMu.Unlock()

	// Create a well-formed hook callback request
	hookReq := &types.SDKHookCallbackRequest{
		Subtype:    "hook_callback",
		CallbackID: "cb_typed",
		Input:      map[string]any{"test": "data"},
		ToolUseID:  stringPtr("tool_123"),
	}

	// Marshal to JSON and back to map to simulate CLI message
	reqData, err := json.Marshal(hookReq)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	var requestMap map[string]any
	if err := json.Unmarshal(reqData, &requestMap); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	controlMsg := map[string]any{
		"type":       "control_request",
		"request_id": "req_typed_test",
		"request":    requestMap,
	}

	// Handle the request
	go query.handleControlRequest(controlMsg)

	// Wait for callback to be invoked
	select {
	case <-receivedChan:
		// Success - callback was invoked
	case <-time.After(time.Second):
		t.Error("Hook callback was not invoked with typed request")
	}
}

// TestHandleControlRequest_ParseError tests that invalid requests are handled gracefully
// when using typed struct parsing.
func TestHandleControlRequest_ParseError(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send a malformed request (missing subtype)
	controlMsg := map[string]any{
		"type":       "control_request",
		"request_id": "req_malformed",
		"request": map[string]any{
			// Missing subtype field - should cause parse error with typed approach
			"callback_id": "cb_123",
		},
	}

	// Handle the request
	go query.handleControlRequest(controlMsg)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should send error response
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("Expected error response to be written")
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(written[len(written)-1]), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["type"] != "control_response" {
		t.Errorf("Expected type 'control_response', got %v", response["type"])
	}

	respData, ok := response["response"].(map[string]any)
	if !ok {
		t.Fatal("Response data is not a map")
	}

	// With typed struct parsing, this should be an error
	if respData["subtype"] != "error" {
		t.Errorf("Expected error response for malformed request, got %v", respData["subtype"])
	}
}

// TestHandleCanUseTool_TypedParsing tests that can_use_tool requests use typed parsing.
func TestHandleCanUseTool_TypedParsing(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Track if callback was invoked with correct data - use channel for thread-safe signaling
	callbackChan := make(chan bool, 1)
	query.SetCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
		// Verify typed parsing extracted the right fields
		if toolName != "TestTool" {
			t.Errorf("Expected tool_name 'TestTool', got %v", toolName)
		}
		if input["param"] != "value" {
			t.Errorf("Expected input param 'value', got %v", input["param"])
		}
		callbackChan <- true
		return &types.PermissionResultAllow{Behavior: "allow"}, nil
	})

	// Create a typed permission request
	permReq := &types.SDKControlPermissionRequest{
		Subtype:  "can_use_tool",
		ToolName: "TestTool",
		Input:    map[string]any{"param": "value"},
	}

	// Marshal and send
	reqData, err := json.Marshal(permReq)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	var requestMap map[string]any
	if err := json.Unmarshal(reqData, &requestMap); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	controlMsg := map[string]any{
		"type":       "control_request",
		"request_id": "req_perm_typed",
		"request":    requestMap,
	}

	go query.handleControlRequest(controlMsg)

	// Wait for callback to be invoked
	select {
	case <-callbackChan:
		// Success - callback was invoked
	case <-time.After(time.Second):
		t.Error("Permission callback was not invoked")
	}
}

// TestHandleControlResponse_TypedParsing tests that responses use typed parsing.
func TestHandleControlResponse_TypedParsing(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Create pending request
	respChan := make(chan map[string]any, 1)
	query.pendingMu.Lock()
	query.pendingRequests["req_typed_resp"] = respChan
	query.pendingMu.Unlock()

	// Send a properly typed response
	typedResponse := &types.ControlResponse{
		Type: "control_response",
		Response: types.ControlResponseData{
			Subtype:   "success",
			RequestID: "req_typed_resp",
			Response:  map[string]any{"result": "ok", "count": 42},
		},
	}

	// Marshal and send as raw message
	respData, err := json.Marshal(typedResponse)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	var rawResp map[string]any
	if err := json.Unmarshal(respData, &rawResp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	query.handleControlResponse(rawResp)

	// Verify response was correctly extracted
	select {
	case resp := <-respChan:
		if resp["result"] != "ok" {
			t.Errorf("Expected result 'ok', got %v", resp["result"])
		}
		// Verify numeric types are preserved through typed parsing
		count, ok := resp["count"].(float64)
		if !ok {
			t.Errorf("Expected count to be float64, got %T", resp["count"])
		}
		if count != 42 {
			t.Errorf("Expected count 42, got %v", count)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for response")
	}
}

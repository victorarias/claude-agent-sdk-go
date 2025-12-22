package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

var errTestError = errors.New("test error")

// TestHandleControlResponse_TypedStructs tests that control responses are properly
// parsed into typed structs and routed to waiting requests.
func TestHandleControlResponse_TypedStructs(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Create a pending request
	respChan := make(chan map[string]any, 1)
	query.pendingMu.Lock()
	query.pendingRequests["test_req_123"] = respChan
	query.pendingMu.Unlock()

	// Send a typed control response
	controlResponse := types.ControlResponse{
		Type: "control_response",
		Response: types.ControlResponseData{
			Subtype:   "success",
			RequestID: "test_req_123",
			Response:  map[string]any{"status": "ok", "data": "test"},
		},
	}

	// Marshal to raw message (simulating what comes from CLI)
	responseBytes, err := json.Marshal(controlResponse)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var rawResponse map[string]any
	if err := json.Unmarshal(responseBytes, &rawResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Handle the response
	query.handleControlResponse(rawResponse)

	// Verify the response was routed correctly
	select {
	case resp := <-respChan:
		if resp["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", resp["status"])
		}
		if resp["data"] != "test" {
			t.Errorf("Expected data 'test', got %v", resp["data"])
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for response")
	}
}

// TestHandleControlResponse_ErrorResponse tests error handling in control responses.
func TestHandleControlResponse_ErrorResponse(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Create a pending request
	respChan := make(chan map[string]any, 1)
	query.pendingMu.Lock()
	query.pendingRequests["test_req_456"] = respChan
	query.pendingMu.Unlock()

	// Send an error response
	controlResponse := types.ControlResponse{
		Type: "control_response",
		Response: types.ControlResponseData{
			Subtype:   "error",
			RequestID: "test_req_456",
			Error:     "something went wrong",
		},
	}

	responseBytes, err := json.Marshal(controlResponse)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var rawResponse map[string]any
	if err := json.Unmarshal(responseBytes, &rawResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	query.handleControlResponse(rawResponse)

	// Verify the error was propagated
	select {
	case resp := <-respChan:
		if resp["error"] != "something went wrong" {
			t.Errorf("Expected error 'something went wrong', got %v", resp["error"])
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for error response")
	}
}

// TestHandleControlRequest_HookCallback_TypedStruct tests that hook callback
// requests are properly parsed using typed structs.
func TestHandleControlRequest_HookCallback_TypedStruct(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Register a hook callback
	callbackInvoked := false
	query.hookMu.Lock()
	query.hookCallbacks["hook_123"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		callbackInvoked = true
		return &types.HookOutput{
			Continue: boolPtr(true),
			Decision: "approve",
		}, nil
	}
	query.hookMu.Unlock()

	// Create a hook callback request
	hookRequest := types.SDKHookCallbackRequest{
		Subtype:    "hook_callback",
		CallbackID: "hook_123",
		Input:      map[string]any{"tool_name": "Bash", "command": "ls"},
		ToolUseID:  stringPtr("tool_789"),
	}

	// Marshal to map (simulating what comes from CLI)
	reqData, err := json.Marshal(hookRequest)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	var requestMap map[string]any
	if err := json.Unmarshal(reqData, &requestMap); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Create control request wrapper
	controlRequest := map[string]any{
		"type":       "control_request",
		"request_id": "req_hook_test",
		"request":    requestMap,
	}

	// Handle the request
	go query.handleControlRequest(controlRequest)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !callbackInvoked {
		t.Error("Hook callback was not invoked")
	}

	// Verify response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("No response written")
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(written[len(written)-1]), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["type"] != "control_response" {
		t.Errorf("Expected type 'control_response', got %v", response["type"])
	}
}

// TestHandleControlRequest_CanUseTool_TypedStruct tests that permission
// requests are properly parsed using typed structs.
func TestHandleControlRequest_CanUseTool_TypedStruct(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Set up permission callback - use channel for thread-safe signaling
	callbackChan := make(chan bool, 1)
	query.SetCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
		if toolName != "Bash" {
			t.Errorf("Expected tool_name 'Bash', got %v", toolName)
		}
		callbackChan <- true
		return &types.PermissionResultAllow{
			Behavior: "allow",
		}, nil
	})

	// Create permission request
	permRequest := types.SDKControlPermissionRequest{
		Subtype:  "can_use_tool",
		ToolName: "Bash",
		Input:    map[string]any{"command": "ls"},
	}

	// Marshal to map (simulating what comes from CLI)
	reqData, err := json.Marshal(permRequest)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	var requestMap map[string]any
	if err := json.Unmarshal(reqData, &requestMap); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Create control request wrapper
	controlRequest := map[string]any{
		"type":       "control_request",
		"request_id": "req_perm_test",
		"request":    requestMap,
	}

	// Handle the request
	go query.handleControlRequest(controlRequest)

	// Wait for callback to be invoked
	select {
	case <-callbackChan:
		// Success - callback was invoked
	case <-time.After(time.Second):
		t.Error("Permission callback was not invoked")
	}

	// Verify response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("No response written")
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(written[len(written)-1]), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["type"] != "control_response" {
		t.Errorf("Expected type 'control_response', got %v", response["type"])
	}

	// Verify response structure
	responseData, ok := response["response"].(map[string]any)
	if !ok {
		t.Fatal("Response data is not a map")
	}

	if responseData["subtype"] != "success" {
		t.Errorf("Expected subtype 'success', got %v", responseData["subtype"])
	}
}

// TestSendControlResponse_TypedStruct tests that sendControlResponse creates
// properly typed responses.
func TestSendControlResponse_TypedStruct(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send a success response
	responseData := map[string]any{
		"status": "completed",
		"result": "success",
	}
	query.sendControlResponse("req_test_789", responseData, nil)

	// Verify the response structure
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("No response written")
	}

	var response types.ControlResponse
	if err := json.Unmarshal([]byte(written[0]), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Type != "control_response" {
		t.Errorf("Expected type 'control_response', got %v", response.Type)
	}

	if response.Response.Subtype != "success" {
		t.Errorf("Expected subtype 'success', got %v", response.Response.Subtype)
	}

	if response.Response.RequestID != "req_test_789" {
		t.Errorf("Expected request_id 'req_test_789', got %v", response.Response.RequestID)
	}

	if response.Response.Response["status"] != "completed" {
		t.Errorf("Expected status 'completed', got %v", response.Response.Response["status"])
	}
}

// TestSendControlResponse_TypedStruct_Error tests error responses.
func TestSendControlResponse_TypedStruct_Error(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send an error response
	query.sendControlResponse("req_error_test", nil, errTestError)

	// Verify the response structure
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("No response written")
	}

	var response types.ControlResponse
	if err := json.Unmarshal([]byte(written[0]), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Type != "control_response" {
		t.Errorf("Expected type 'control_response', got %v", response.Type)
	}

	if response.Response.Subtype != "error" {
		t.Errorf("Expected subtype 'error', got %v", response.Response.Subtype)
	}

	if response.Response.Error != "test error" {
		t.Errorf("Expected error 'test error', got %v", response.Response.Error)
	}
}

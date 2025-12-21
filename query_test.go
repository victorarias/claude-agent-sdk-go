package sdk

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNewQuery(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	if query == nil {
		t.Fatal("NewQuery returned nil")
	}

	if query.transport != transport {
		t.Error("transport not set correctly")
	}

	if !query.streaming {
		t.Error("streaming flag not set")
	}
}

func TestQuery_Start(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	err := query.Start(ctx)
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Should be able to receive messages
	transport.SendMessage(map[string]any{"type": "system", "subtype": "init"})

	select {
	case msg := <-query.Messages():
		if msg.MessageType() != "system" {
			t.Errorf("got type %v, want system", msg.MessageType())
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}

	query.Close()
}

func TestQuery_RawMessages(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendMessage(map[string]any{"type": "test", "custom": true})

	select {
	case msg := <-query.RawMessages():
		if msg["type"] != "test" {
			t.Errorf("got type %v, want test", msg["type"])
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for raw message")
	}
}

func TestQuery_ResultReceived(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send a result message
	transport.SendMessage(map[string]any{
		"type":       "result",
		"subtype":    "success",
		"session_id": "test_123",
	})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !query.ResultReceived() {
		t.Error("expected result to be received")
	}
}

func TestQuery_Close(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if err := query.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should be able to close again without error
	if err := query.Close(); err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}

func TestQuery_SendControlRequest(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate response in background
	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"status": "ok"},
				},
			})
		}
	}()

	response, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 5*time.Second)

	if err != nil {
		t.Errorf("sendControlRequest failed: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("unexpected response: %v", response)
	}
}

func TestQuery_SendControlRequest_Timeout(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Don't send a response - should timeout
	_, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 100*time.Millisecond)

	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestQuery_SendControlRequest_Error(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "error",
					"request_id": reqID,
					"error":      "something went wrong",
				},
			})
		}
	}()

	_, err := query.sendControlRequest(map[string]any{
		"subtype": "test",
	}, 5*time.Second)

	if err == nil {
		t.Error("expected error response")
	}
}

func TestQuery_SendControlRequest_NonStreaming(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, false) // non-streaming

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	_, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 100*time.Millisecond)

	if err == nil {
		t.Error("expected error for non-streaming mode")
	}
}

func TestQuery_Initialize(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate init response
	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"commands": []any{"/help", "/clear"},
					},
				},
			})
		}
	}()

	result, err := query.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if result == nil {
		t.Error("expected initialization result")
	}
}

func TestQuery_Initialize_WithHooks(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			// Verify hooks are in request
			request := req["request"].(map[string]any)
			if request["hooks"] == nil {
				t.Error("expected hooks in request")
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	hooks := map[HookEvent][]HookMatcher{
		HookPreToolUse: {
			{
				Matcher: strPtr("Bash"),
				Hooks: []HookCallback{
					func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
						cont := true
						return &HookOutput{Continue: &cont}, nil
					},
				},
			},
		},
	}

	_, err := query.Initialize(hooks)
	if err != nil {
		t.Errorf("Initialize with hooks failed: %v", err)
	}
}

func TestQuery_Initialize_NonStreaming(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, false) // non-streaming

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Non-streaming should return nil without error
	result, err := query.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-streaming")
	}
}

func strPtr(s string) *string { return &s }

func TestQuery_Interrupt(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.Interrupt()
	if err != nil {
		t.Errorf("Interrupt failed: %v", err)
	}
}

func TestQuery_SetPermissionMode(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.SetPermissionMode(PermissionBypass)
	if err != nil {
		t.Errorf("SetPermissionMode failed: %v", err)
	}
}

func TestQuery_SetModel(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			// Verify model is in request
			request := req["request"].(map[string]any)
			if request["model"] != "claude-opus-4" {
				t.Errorf("expected model claude-opus-4, got %v", request["model"])
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.SetModel("claude-opus-4")
	if err != nil {
		t.Errorf("SetModel failed: %v", err)
	}
}

func TestQuery_RewindFiles(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			// Verify user_message_id is in request
			request := req["request"].(map[string]any)
			if request["user_message_id"] != "msg_123" {
				t.Errorf("expected user_message_id msg_123, got %v", request["user_message_id"])
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.RewindFiles("msg_123")
	if err != nil {
		t.Errorf("RewindFiles failed: %v", err)
	}
}

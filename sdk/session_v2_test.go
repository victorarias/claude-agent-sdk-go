// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func startUnstableV2MockResponder(t *testing.T, transport *MockTransport, sessionID string) func() {
	t.Helper()
	stopCh := make(chan struct{})
	go func() {
		processed := 0
		for {
			select {
			case <-stopCh:
				return
			default:
			}

			if !transport.WaitForWrite(25 * time.Millisecond) {
				continue
			}

			written := transport.Written()
			for processed < len(written) {
				raw := written[processed]
				processed++

				var msg map[string]any
				if err := json.Unmarshal([]byte(raw), &msg); err != nil {
					continue
				}

				switch msg["type"] {
				case "control_request":
					reqID, _ := msg["request_id"].(string)
					req, _ := msg["request"].(map[string]any)
					subtype, _ := req["subtype"].(string)

					response := map[string]any{}
					if subtype == "initialize" {
						response["session_id"] = sessionID
					}

					transport.SendMessage(map[string]any{
						"type": "control_response",
						"response": map[string]any{
							"subtype":    "success",
							"request_id": reqID,
							"response":   response,
						},
					})

				case "user":
					transport.SendMessage(map[string]any{
						"type":       "assistant",
						"uuid":       "assistant_1",
						"session_id": sessionID,
						"message": map[string]any{
							"model": "claude-sonnet-4-5",
							"content": []any{
								map[string]any{"type": "text", "text": "Hello from mock."},
							},
						},
					})
					transport.SendMessage(map[string]any{
						"type":            "result",
						"subtype":         "success",
						"uuid":            "result_1",
						"duration_ms":     float64(1),
						"duration_api_ms": float64(1),
						"is_error":        false,
						"num_turns":       float64(1),
						"session_id":      sessionID,
						"result":          "ok",
					})
				}
			}
		}
	}()

	return func() {
		close(stopCh)
	}
}

func TestUnstableV2CreateSession(t *testing.T) {
	transport := NewMockTransport()
	stop := startUnstableV2MockResponder(t, transport, "sess_create")
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := UnstableV2CreateSession(ctx, types.WithTransport(transport))
	if err != nil {
		t.Fatalf("UnstableV2CreateSession failed: %v", err)
	}
	defer session.Close()

	sid, err := session.SessionID()
	if err != nil {
		t.Fatalf("SessionID should be available after initialize: %v", err)
	}
	if sid != "sess_create" {
		t.Fatalf("expected session ID sess_create, got %s", sid)
	}
}

func TestUnstableV2ResumeSession(t *testing.T) {
	transport := NewMockTransport()
	stop := startUnstableV2MockResponder(t, transport, "sess_resume")
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := UnstableV2ResumeSession(ctx, "sess_resume", types.WithTransport(transport))
	if err != nil {
		t.Fatalf("UnstableV2ResumeSession failed: %v", err)
	}
	defer session.Close()

	sid, err := session.SessionID()
	if err != nil {
		t.Fatalf("SessionID should be available after initialize: %v", err)
	}
	if sid != "sess_resume" {
		t.Fatalf("expected session ID sess_resume, got %s", sid)
	}
}

func TestUnstableV2SessionSendAndStream(t *testing.T) {
	transport := NewMockTransport()
	stop := startUnstableV2MockResponder(t, transport, "sess_stream")
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := UnstableV2CreateSession(ctx, types.WithTransport(transport))
	if err != nil {
		t.Fatalf("UnstableV2CreateSession failed: %v", err)
	}
	defer session.Close()

	if err := session.Send("hello"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	gotAssistant := false
	gotResult := false
	timeout := time.After(2 * time.Second)
	for !gotAssistant || !gotResult {
		select {
		case msg, ok := <-session.Stream():
			if !ok {
				t.Fatal("stream closed unexpectedly")
			}
			switch msg.(type) {
			case *types.AssistantMessage:
				gotAssistant = true
			case *types.ResultMessage:
				gotResult = true
			}
		case <-timeout:
			t.Fatalf("timed out waiting for stream messages (assistant=%t result=%t)", gotAssistant, gotResult)
		}
	}
}

func TestUnstableV2Prompt(t *testing.T) {
	transport := NewMockTransport()
	stop := startUnstableV2MockResponder(t, transport, "sess_prompt")
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := UnstableV2Prompt(ctx, "hello", types.WithTransport(transport))
	if err != nil {
		t.Fatalf("UnstableV2Prompt failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result message")
	}
	if result.Subtype != "success" {
		t.Fatalf("expected success subtype, got %s", result.Subtype)
	}
	if result.SessionID != "sess_prompt" {
		t.Fatalf("expected session_id=sess_prompt, got %s", result.SessionID)
	}
}
